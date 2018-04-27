package main

import (
	"cpanel/control"
	"cpanel/table"
	"cpanel/tools"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/astaxie/beedb"
	"github.com/astaxie/beego/logs"
	libvirt "github.com/libvirt/libvirt-go"
	_ "github.com/mattn/go-sqlite3"

	"github.com/Demired/rpwd"
)

var q = make(chan string) //设置初始密码的chan

var mac = make(chan string) //获取初始内网ip寄存的mac地址

var cLog = logs.NewLogger(1)

var logFile = "/var/log/cpanel.log"

func main() {
	cLog.SetLogger("file", `{"filename":"`+logFile+`"}`)
	cLog.SetLevel(logs.LevelInformational)

	go watch()
	go workQueue()
	http.HandleFunc("/", index)
	http.HandleFunc("/edit", editAPI)
	http.HandleFunc("/list", list)
	http.HandleFunc("/info.html", info)
	http.HandleFunc("/load.json", loadJSON)
	http.HandleFunc("/start", start)
	http.HandleFunc("/shutdown", shutdown)
	http.HandleFunc("/reboot", reboot)
	http.HandleFunc("/create", createAPI)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/repasswd.html", repasswd)
	http.HandleFunc("/alarm.html", alarm)
	http.HandleFunc("/alarm", alarmAPI)
	http.HandleFunc("/repasswd", repasswdAPI)
	http.HandleFunc("/undefine", undefine)
	http.HandleFunc("/edit.html", edit)
	http.HandleFunc("/create.html", create)
	http.ListenAndServe(":8100", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, nil)
}

func favicon(w http.ResponseWriter, req *http.Request) {
	path := "./html/images/favicon.ico"
	http.ServeFile(w, req, path)
}

func watch() {
	var t = make(map[string]uint64)
	w := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-w.C:
			doms, err := control.Connect().ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
			if err != nil {
				cLog.Warn(err.Error())
				continue
			}
			db, err := sql.Open("sqlite3", "./db/cpanel.db")
			if err != nil {
				cLog.Warn("打开数据库失败", err.Error())
				continue
			}
			orm := beedb.New(db)
			for _, dom := range doms {
				name, err := dom.GetName()
				if err != nil {
					cLog.Warn(err.Error())
					continue
				}
				info, err := dom.GetInfo()
				if err != nil {
					cLog.Warn(err.Error())
					continue
				}
				var virtual table.Virtual
				if err := orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", name).Find(&virtual); err != nil {
					cLog.Warn("读取虚拟机信息失败", err.Error())
					continue
				}
				var cpurate int
				if lastCPUTime, ok := t[name]; ok {
					cpurate = int(float32((info.CpuTime-lastCPUTime)*100) / float32(20*info.NrVirtCpu*10000000))
				}
				if cpurate < 1 {
					cpurate = 1
				}
				var watch table.Watch
				watch.CPU = cpurate
				watch.Vname = name
				watch.Ctime = int(time.Now().Unix())
				watch.Memory = int(info.Memory)
				if err = orm.SetTable("Watch").SetPK("ID").Save(&watch); err != nil {
					cLog.Warn("写入数据失败", err.Error())
					continue
				}
				t[name] = info.CpuTime
				dom.Free()
			}
			db.Close()
		}
	}
}

func create(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/create.html")
	t.Execute(w, nil)
}

func edit(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	var vvm table.Virtual
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	// err = orm.SetTable("Virtual").Where("Vname = ?", Vname).Find(&vvm)
	// if err != nil {
	// 	cLog.Warn(err.Error())
	// 	return
	// }
	fmt.Println(Vname)
	t, _ := template.ParseFiles("html/create.html")
	t.Execute(w, nil)
}

func info(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	dom, err := control.Connect().LookupDomainByName(Vname)
	if err != nil {
		cLog.Warn(err.Error())
	}
	s, _, err := dom.GetState()
	if err != nil {
		cLog.Warn(err.Error())
	}
	if int(s) == 1 {
		_, err := dom.GetInfo()
		if err != nil {
			cLog.Warn(err.Error())
		}
	}
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		cLog.Warn("打开数据库失败", err.Error())
		return
	}
	orm := beedb.New(db)
	var vvm table.Virtual
	err = orm.SetTable("Virtual").Where("Vname = ?", Vname).Find(&vvm)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	vvm.Status = int(s)
	t, _ := template.ParseFiles("html/info.html")
	t.Execute(w, vvm)
}

func loadJSON(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	Vname := req.URL.Query().Get("Vname")
	db, _ := sql.Open("sqlite3", "./db/cpanel.db")
	defer db.Close()
	startTime, err := strconv.Atoi(req.URL.Query().Get("start"))
	if err != nil {
		startTime = int(time.Now().Unix()) - 3600
	}
	endTime, err := strconv.Atoi(req.URL.Query().Get("end"))
	if err != nil {
		endTime = int(time.Now().Unix())
	}
	var watchs []table.Watch
	orm := beedb.New(db)
	err = orm.SetTable("Watch").Where("Vname = ? and Ctime > ? and Ctime < ?", Vname, startTime, endTime).FindAll(&watchs)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var virtual table.Virtual
	err = orm.SetTable("Virtual").Where("Vname = ?", Vname).Find(&virtual)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var cpus [][]int
	var memorys [][]int
	for _, v := range watchs {
		memorys = append(memorys, []int{v.Ctime, v.Memory})
		cpus = append(cpus, []int{v.Ctime, v.CPU})
	}
	var date = make(map[string]interface{})
	date["maxMemory"] = virtual.Vmemory * 1024
	date["cpus"] = cpus
	date["memorys"] = memorys
	dj, _ := json.Marshal(date)
	w.Write(dj)
}

func repasswd(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	db, _ := sql.Open("sqlite3", "./db/cpanel.db")
	orm := beedb.New(db)
	var watch table.Watch
	err := orm.SetTable("Watch").Find(&watch)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, "/list", http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/repasswd.html")
	t.Execute(w, Vname)
	return
}

func list(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	defer db.Close()
	orm := beedb.New(db)
	var vvvm []table.Virtual
	err = orm.SetTable("Virtual").Where("Status = ?", "1").FindAll(&vvvm)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	for k, v := range vvvm {
		dom, err := control.Connect().LookupDomainByName(v.Vname)
		if err != nil {
			cLog.Warn(err.Error())
			continue
		}
		s, _, err := dom.GetState()
		if err != nil {
			cLog.Warn(err.Error())
			continue
		}
		vvvm[k].Status = int(s)
	}
	t, _ := template.ParseFiles("html/list.html")
	t.Execute(w, vvvm)
}

func createSysDisk(Vname, mirror string) (w int64, err error) {
	mirrorPath := fmt.Sprintf("/virt/mirror/%s.qcow2", mirror)
	srcFile, err := os.Open(mirrorPath)
	if err != nil {
		cLog.Info(err.Error())
		return 0, err
	}
	defer srcFile.Close()
	diskPath := fmt.Sprintf("/virt/disk/%s.qcow2", Vname)
	desFile, err := os.Create(diskPath)
	if err != nil {
		fmt.Println(err)
	}
	defer desFile.Close()
	return io.Copy(desFile, srcFile)
}

func start(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	err := control.CheckEtime(Vname)
	if err != nil {
		cLog.Warn("检查到期:%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "服务器已经到期", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Start(Vname)
	if err != nil {
		cLog.Warn("虚拟机开启:%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "开机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "正在开机"})
	w.Write(msg)
}

func repasswdAPI(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	passwd := req.PostFormValue("passwd")
	err := control.SetPasswd(Vname, "root", passwd)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "密码已重置"})
	w.Write(msg)
}

type er struct {
	Ret  string `json:"ret"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func shutdown(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	err := control.CheckEtime(Vname)
	if err != nil {
		cLog.Warn("检查到期：%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "服务器已到期", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Shutdown(Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "关机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "正在关机"})
	w.Write(msg)
}

func reboot(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	err := control.Reboot(Vname)
	if err != nil {
		cLog.Info(err.Error())
		return
	}
	msg, err := json.Marshal(er{Ret: "v", Msg: "正在重启"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Write(msg)
}

func editAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存大小必须为整数"})
		w.Write(msg)
		return
	}
	fmt.Println(vmemory)
}

func alarm(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	var dInfo table.Virtual
	orm := beedb.New(db)
	err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Find(&dInfo)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "发生错误", Data: err.Error()})
		w.Write(msg)
		return
	}
	if time.Now().After(dInfo.Etime) {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "虚拟机已到期"})
		w.Write(msg)
		return
	}
	if dInfo.AStatus == 0 {
		dInfo.ABandwidth = 0
		dInfo.ACpu = 0
		dInfo.ADisk = 0
		dInfo.AMemory = 0
	}
	//检查虚拟机所有者
	t, _ := template.ParseFiles("html/alarm.html")
	t.Execute(w, dInfo)
}

func alarmAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	var dInfo table.Virtual
	orm := beedb.New(db)
	Vname := req.PostFormValue("Vname")
	err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Find(&dInfo)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "发生错误", Data: err.Error()})
		w.Write(msg)
		return
	}
	if time.Now().After(dInfo.Etime) {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "虚拟机已到期"})
		w.Write(msg)
		return
	}
	//检查虚拟机所有者
	AStatus, _ := strconv.Atoi(req.PostFormValue("AStatus"))
	if AStatus == 0 {
		if dInfo.AStatus == 0 {
			msg, _ := json.Marshal(er{Ret: "v", Msg: "报警未开启"})
			w.Write(msg)
			return
		}
		t := make(map[string]interface{})
		t["AStatus"] = 0
		orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
		msg, _ := json.Marshal(er{Ret: "v", Msg: "报警已关闭"})
		w.Write(msg)
		return
	}
	// ACpu INT NOT NULL,
	// ABandwidth INT NOT NULL,
	// AMemory INT NOT NULL,
	// ADisk INT NOT NULL,
	ACpu, err := strconv.Atoi(req.PostFormValue("ACpu"))
	if err != nil || ACpu > 100 || ACpu < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "cpu报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	ABandwidth, err := strconv.Atoi(req.PostFormValue("ABandwidth"))
	if err != nil || ABandwidth > 100 || ABandwidth < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	AMemory, err := strconv.Atoi(req.PostFormValue("AMemory"))
	if err != nil || AMemory > 100 || AMemory < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	ADisk, err := strconv.Atoi(req.PostFormValue("ADisk"))
	if err != nil || ADisk > 100 || ADisk < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	t := make(map[string]interface{})
	t["ACpu"] = ACpu
	t["ABandwidth"] = ABandwidth
	t["AMemory"] = AMemory
	t["ADisk"] = ADisk
	t["AStatus"] = 1
	_, err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "设置失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "设置成功"})
	w.Write(msg)
}

//创建虚拟机
func createAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))

	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存大小必须为整数"})
		w.Write(msg)
		return
	}

	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "cpu个数必须为整数"})
		w.Write(msg)
		return
	}

	vpasswd := req.PostFormValue("vpasswd")
	if vpasswd == "" {
		vpasswd = string(rpwd.Init(16, true, true, true, false))
	}

	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽必须位整数"})
		w.Write(msg)
		return
	}

	sys := req.PostFormValue("sys")
	if sys == "" {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "镜像必填"})
		w.Write(msg)
		return
	}

	var vInfo table.Virtual
	vInfo.Vname = string(rpwd.Init(8, true, true, true, false))
	vInfo.Vcpu = vcpu
	vInfo.Vmemory = vmemory
	vInfo.Passwd = vpasswd
	vInfo.Mac = tools.Rmac()
	vInfo.Br = "br1"
	vInfo.Status = 1
	vInfo.Bandwidth = bandwidth
	vInfo.Ctime = time.Now()
	vInfo.Etime = time.Now().Add(24 * 30 * time.Hour)
	vInfo.Utime = time.Now()
	vInfo.Sys = sys

	_, err = createSysDisk(vInfo.Vname, vInfo.Sys)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机硬盘失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	xml := createKvmXML(vInfo)
	_, err = control.Connect().DomainDefineXML(xml)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	orm := beedb.New(db)
	err = orm.SetTable("Virtual").SetPK("ID").Save(&vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	q <- fmt.Sprintf("%s/%s", vInfo.Vname, vInfo.Passwd)
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	w.Write(msg)
}

func undefine(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	Vname := req.PostFormValue("Vname")
	disk := fmt.Sprintf("/virt/disk/%s.qcow2", Vname)
	os.Remove(disk)
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	orm := beedb.New(db)
	t := make(map[string]interface{})
	t["Status"] = 0
	_, err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "删除失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Undefine(Vname)
	if err != nil {
		cLog.Error(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "销毁失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "已删除"})
	w.Write(msg)
}

func workQueue() {
	for {
		select {
		case str := <-q:
			by := strings.Split(str, "/")
			err := control.Start(by[0])
			if err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(1 * time.Minute)
			control.SetPasswd(by[0], "root", by[1])
		}
	}
}

func createKvmXML(tvm table.Virtual) string {
	// name := "test"
	var templateXML = `
	<domain type='kvm'>
		<name>` + tvm.Vname + `</name>
		<memory unit="GiB">` + fmt.Sprintf("%d", tvm.Vmemory) + `</memory>
		<os>
			<type>hvm</type>
		</os>
		<features>
			<acpi/>
			<apic/>
			<pae/>
		</features>
		<clock offset='utc'/>
		<on_poweroff>destroy</on_poweroff>
		<on_reboot>restart</on_reboot>
		<on_crash>destroy</on_crash>
		<devices>
			<emulator>/usr/libexec/qemu-kvm</emulator>
			<disk type="file" device="disk">
				<driver name='qemu' type='qcow2'/>
				<source file="/virt/disk/` + tvm.Vname + `.qcow2"/>
				<target dev="hdb" bus="ide"/>
			</disk>
			<interface type='bridge'>
				<mac address='` + tvm.Mac + `'/>
				<source bridge='` + tvm.Br + `'/>
				<bandwidth>
					<inbound average='` + fmt.Sprintf("%d", tvm.Bandwidth*1000) + `' peak='` + fmt.Sprintf("%d", tvm.Bandwidth*3000) + `' burst='` + fmt.Sprintf("%d", tvm.Bandwidth*1024) + `'/>
					<outbound average='` + fmt.Sprintf("%d", tvm.Bandwidth*1000) + `' peak='` + fmt.Sprintf("%d", tvm.Bandwidth*3000) + `' burst='` + fmt.Sprintf("%d", tvm.Bandwidth*1024) + `'/>
				</bandwidth>
			</interface>
			<serial type='pty'>
				<target port='1'/>
			</serial>
			<console type='pty'>
				<target type='serial' port='1'/>
			</console>
			<console type='pty'>
				<target type='virtio' port='1'/>
			</console>
			<channel type='unix'>
				<target type='virtio' name='org.qemu.guest_agent.0' state='connected'/>
				<address type='virtio-serial' controller='0' bus='0' port='1'/>
			</channel>
		</devices>
	</domain>`
	return templateXML
}
