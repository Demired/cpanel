package main

import (
	"cpanel/control"
	"cpanel/log"
	"cpanel/loop"
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Demired/rpwd"
)

var cLog = log.CLog

func main() {
	go loop.Watch()
	go loop.WorkQueue()
	http.HandleFunc("/", index)
	http.HandleFunc("/edit", editAPI)
	http.HandleFunc("/list", list)
	http.HandleFunc("/login.html", login)
	http.HandleFunc("/login", loginAPI)
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

func login(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/login.html")
	t.Execute(w, nil)
}

func loginAPI(w http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	passwd := req.PostFormValue("passwd")
	if username == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "username", Msg: "用户名不能为空"})
		w.Write(msg)
	}
	if passwd == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码不能为空"})
		w.Write(msg)
	}
	// orm, _ := control.Bdb()
	// orm.SetTable("")
	msg, _ := json.Marshal(er{Ret: "v", Msg: "登录成功"})
	w.Write(msg)
}

func favicon(w http.ResponseWriter, req *http.Request) {
	path := "./html/images/favicon.ico"
	http.ServeFile(w, req, path)
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
	orm.SetTable("Virtual").Where("Vname = ?", Vname).Find(&vvm)
	if time.Now().After(vvm.Etime) {
		w.Write([]byte("服务器已到期"))
		return
	}
	t, _ := template.ParseFiles("html/create.html")
	t.Execute(w, vvm)
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
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
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
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	startTime, err := strconv.Atoi(req.URL.Query().Get("start"))
	if err != nil {
		startTime = int(time.Now().Unix()) - 3600
	}
	endTime, err := strconv.Atoi(req.URL.Query().Get("end"))
	if err != nil {
		endTime = int(time.Now().Unix())
	}
	var watchs []table.Watch
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
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var watch table.Watch
	err = orm.SetTable("Watch").Find(&watch)
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
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
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
		cLog.Warn("检查到期：%s,%s", Vname, err.Error())
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
	Ret   string `json:"ret"`
	Msg   string `json:"msg"`
	Data  string `json:"data"`
	Param string `json:"param"`
}

func shutdown(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	err := control.Shutdown(Vname)
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
		msg, _ := json.Marshal(er{Ret: "e", Msg: "重启失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "正在重启"})
	w.Write(msg)
}

func editAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var sourceVirtual table.Virtual
	vname := req.PostFormValue("vname")
	err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", vname).Find(&sourceVirtual)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "机器不存在"})
		w.Write(msg)
		return
	}
	if time.Now().After(sourceVirtual.Etime) {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "机器已经到期"})
		w.Write(msg)
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

	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽必须位整数"})
		w.Write(msg)
		return
	}

	if sourceVirtual.Vmemory != vmemory {

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
	vInfo.Mac = tools.Rmac()
	vInfo.Br = "br1"
	vInfo.Status = 1
	vInfo.Bandwidth = bandwidth
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

	err = orm.SetTable("Virtual").SetPK("ID").Save(&vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	loop.VmInit <- vInfo.Vname
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	w.Write(msg)
}

func alarm(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var dInfo table.Virtual
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
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var dInfo table.Virtual
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
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	err = orm.SetTable("Virtual").SetPK("ID").Save(&vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	loop.VmInit <- vInfo.Vname
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	w.Write(msg)
}

func undefine(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	Vname := req.PostFormValue("Vname")
	disk := fmt.Sprintf("/virt/disk/%s.qcow2", Vname)
	os.Remove(disk)
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
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

func createKvmXML(tvm table.Virtual) string {
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
			<interface type='network'>
				<mac address='` + tvm.Mac + `'/>
				<source network='lan'/>
				<bandwidth>
					<inbound average='` + fmt.Sprintf("%d", tvm.Bandwidth*125) + `' peak='` + fmt.Sprintf("%d", tvm.Bandwidth*375) + `' burst='` + fmt.Sprintf("%d", tvm.Bandwidth*128) + `'/>
					<outbound average='` + fmt.Sprintf("%d", tvm.Bandwidth*125) + `' peak='` + fmt.Sprintf("%d", tvm.Bandwidth*375) + `' burst='` + fmt.Sprintf("%d", tvm.Bandwidth*128) + `'/>
				</bandwidth>
				<target dev='lan-` + tvm.Vname + `'/>
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
