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
	http.HandleFunc("/list", list)
	http.HandleFunc("/info.html", info)
	http.HandleFunc("/load.json", loadJSON)
	http.HandleFunc("/start", start)
	http.HandleFunc("/shutdown", shutdown)
	http.HandleFunc("/reboot", reboot)
	http.HandleFunc("/create", createAPI)
	http.HandleFunc("/repasswd.html", repasswd)
	http.HandleFunc("/repasswd", repasswdAPI)
	http.HandleFunc("/undefine", undefine)
	// http.HandleFunc("/edit.html", edit)
	http.HandleFunc("/create.html", create)
	http.ListenAndServe(":8100", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, nil)
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
				var cpurate float32
				if lastCPUTime, ok := t[name]; ok {
					cpurate = float32((info.CpuTime-lastCPUTime)*100) / float32(20*info.NrVirtCpu*10000000)
					if cpurate < 1 {
						cpurate = 1
					}
				}
				var watch table.Watch
				watch.CPU = int(cpurate)
				watch.Vname = name
				watch.Ctime = int(time.Now().Unix())
				watch.Memory = int(info.Memory)
				if err = orm.SetTable("watch").Save(&watch); err != nil {
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

// func edit(w http.ResponseWriter, req *http.Request) {
// 	vname := req.URL.Query().Get("vname")
// 	db, _ := sql.Open("sqlite3", "./db/cpanel.db")
// 	sql := fmt.Sprintf("SELECT Vname,IPv4,IPv6,LocalIP,Mac,Vcpu,Vmemory,Status FROM vm WHERE Vname = '%s';", vname)
// 	rows, _ := db.Query(sql)
// 	if rows.Next() == true {
// 		var vvm vm
// 		// err := rows.Scan(&vvm.Vname, &vvm.IPv4, &vvm.IPv6, &vvm.LocalIP, &vvm.Mac, &vvm.Vcpu, &vvm.Vmemory, &vvm.Status)
// 		// if

// 	}
// 	t, _ := template.ParseFiles("html/create.html")
// 	t.Execute(w, nil)
// }

func info(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vname := req.URL.Query().Get("vname")
	dom, err := control.Connect().LookupDomainByName(vname)
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
	err = orm.SetTable("Virtual").Where("vname=?", vname).Find(&vvm)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	fmt.Println(vvm)
	// sql := fmt.Sprintf("SELECT Vname,IPv4,IPv6,LocalIP,Mac,Vcpu,Bandwidth,Vmemory,Status FROM vm WHERE vname = '%s';", vname)
	// rows, _ := db.Query(sql)
	// var vvm vm
	// if rows.Next() {
	// 	rows.Scan(&vvm.Vname, &vvm.IPv4, &vvm.IPv6, &vvm.LocalIP, &vvm.Mac, &vvm.Vcpu, &vvm.Bandwidth, &vvm.Vmemory, &vvm.Status)
	// }
	var vmInfo = make(map[string]string)
	// vmInfo["Vname"] = vvm.Vname
	// vmInfo["IPv4"] = vvm.IPv4
	// vmInfo["IPv6"] = vvm.IPv6
	// vmInfo["Mac"] = vvm.Mac
	// vmInfo["LocalIP"] = vvm.LocalIP
	// vmInfo["Bandwidth"] = fmt.Sprintf("%d", vvm.Bandwidth)
	// vmInfo["Vmemory"] = fmt.Sprintf("%d", vvm.Vmemory)
	// vmInfo["Vcpu"] = fmt.Sprintf("%d", vvm.Vcpu)
	// vmInfo["Status"] = fmt.Sprintf("%d", s)
	// vmInfo[""]
	vmInfoJ, _ := json.Marshal(vmInfo)
	t, _ := template.ParseFiles("html/info.html")
	t.Execute(w, vmInfoJ)
}

func loadJSON(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vname := req.URL.Query().Get("vname")
	db, _ := sql.Open("sqlite3", "./db/cpanel.db")
	defer db.Close()
	startTime, err := strconv.Atoi(req.URL.Query().Get("start"))
	if err != nil {
		startTime = int(time.Now().Unix()) - 3600
	}
	sql := fmt.Sprintf("SELECT Vname,CPU,Ctime FROM watch WHERE Vname = '%s' AND Ctime > '%d';", vname, startTime)
	rows, _ := db.Query(sql)
	var cpus [][]int
	for rows.Next() {
		var ww table.Watch
		err := rows.Scan(&ww.Vname, &ww.CPU, &ww.Ctime)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		cpus = append(cpus, []int{ww.Ctime, ww.CPU})
	}
	var date = make(map[string]interface{})
	date["cpus"] = cpus
	dj, _ := json.Marshal(date)
	w.Write(dj)
}

func repasswd(w http.ResponseWriter, req *http.Request) {
	vname := req.URL.Query().Get("vname")
	db, _ := sql.Open("sqlite3", "./db/cpanel.db")
	sql := fmt.Sprintf("SELECT id FROM vm WHERE Vname = '%s';", vname)
	rows, _ := db.Query(sql)
	if rows.Next() == true {
		t, _ := template.ParseFiles("html/repasswd.html")
		t.Execute(w, vname)
		return
	}
	http.Redirect(w, req, "/list", http.StatusFound)
	return
}

func list(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	rows, err := db.Query("SELECT Vname,IPv4,IPv6,LocalIP,Mac,Vcpu,Bandwidth,Vmemory,Status FROM vm WHERE Status = 1 LIMIT 100;")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()
	var vvvm []table.Virtual
	for rows.Next() {
		var vvm table.Virtual
		err := rows.Scan(&vvm.Vname, &vvm.IPv4, &vvm.IPv6, &vvm.LocalIP, &vvm.Mac, &vvm.Vcpu, &vvm.Bandwidth, &vvm.Vmemory, &vvm.Status)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		vvm.LocalIP = tools.Arp(vvm.Mac)
		dom, err := control.Connect().LookupDomainByName(vvm.Vname)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		s, _, err := dom.GetState()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		vvm.Status = int(s)
		vvvm = append(vvvm, vvm)
	}
	db.Close()
	t, _ := template.ParseFiles("html/list.html")
	t.Execute(w, vvvm)
}

func createSysDisk(vname string) (w int64, err error) {
	srcFile, err := os.Open("/virt/disk/centos.qcow2")
	if err != nil {
		fmt.Println(err)
	}
	defer srcFile.Close()

	desFile, err := os.Create("/virt/disk/" + vname + ".qcow2")
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
	vname := req.PostFormValue("vname")
	err := control.Start(vname)
	if err != nil {
		fmt.Println(err.Error())
		msg, err := json.Marshal(er{Ret: "e", Msg: "开机失败", Data: err.Error()})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		w.Write(msg)
		return
	}
	msg, err := json.Marshal(er{Ret: "v", Msg: "正在开机"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Write(msg)
}

func repasswdAPI(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	vname := req.PostFormValue("vname")
	passwd := req.PostFormValue("passwd")
	err := control.SetPasswd(vname, "root", passwd)
	if err != nil {
		msg, err := json.Marshal(er{Ret: "e", Msg: err.Error()})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		w.Write(msg)
		return
	}
	msg, err := json.Marshal(er{Ret: "v", Msg: "密码已重置"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Write(msg)
}

type er struct {
	Ret  string `json:"ret"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func shutdown(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	vname := req.PostFormValue("vname")
	err := control.Shutdown(vname)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	msg, err := json.Marshal(er{Ret: "v", Msg: "正在关机"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Write(msg)
}

func reboot(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	vname := req.PostFormValue("vname")
	err := control.Reboot(vname)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	msg, err := json.Marshal(er{Ret: "v", Msg: "正在重启"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Write(msg)
}

//创建虚拟机
func createAPI(w http.ResponseWriter, req *http.Request) {
	// if req.Method != "POST" {
	// 	http.Redirect(w, req, "/create.html", http.StatusFound)
	// 	return
	// }
	// vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))

	// if err != nil {
	// 	msg, _ := json.Marshal(er{Ret: "e", Msg: "内存大小必须为整数"})
	// 	w.Write(msg)
	// 	return
	// }

	// vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	// if err != nil {
	// 	msg, _ := json.Marshal(er{Ret: "e", Msg: "cpu个数必须为整数"})
	// 	w.Write(msg)
	// 	return
	// }

	// vpasswd := req.PostFormValue("vpasswd")
	// if vpasswd == "" {
	// 	vpasswd = string(rpwd.Init(16, true, true, true, false))
	// }

	// bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	// if err != nil {
	// 	msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽必须位整数"})
	// 	w.Write(msg)
	// 	return
	// }
	var vInfo table.Vv
	vInfo.Vname = string(rpwd.Init(8, true, true, true, false))

	// vInfo.Vcpu = vcpu
	// vInfo.Vmemory = vmemory
	// vInfo.Passwd = vpasswd
	// vInfo.Mac = tools.Rmac()
	// vInfo.Br = "br1"
	// vInfo.Bandwidth = bandwidth
	// vInfo.Ctime = time.Now()
	// vInfo.Etime = time.Now()
	// vInfo.Utime = time.Now()

	// xml := createKvmXML(vInfo)
	// _, err = control.Connect().DomainDefineXML(xml)
	// if err != nil {
	// 	msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机失败", Data: err.Error()})
	// 	w.Write(msg)
	// 	return
	// }
	// _, err = createSysDisk(vInfo.Vname)
	// if err != nil {
	// 	msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机硬盘失败", Data: err.Error()})
	// 	w.Write(msg)
	// 	return
	// }
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	orm := beedb.New(db)

	err = orm.SetTable("vv").SetPK("ID").Save(vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	// q <- fmt.Sprintf("%s/%s", vInfo.Vname, vInfo.Passwd)
	// msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	// w.Write(msg)
}

func undefine(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	vname := req.PostFormValue("vname")
	disk := fmt.Sprintf("/virt/disk/%s.qcow2", vname)
	os.Remove(disk)
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	stmt, err := db.Prepare("UPDATE vm SET Status = 0 WHERE Vname = ?")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	_, err = stmt.Exec(vname)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	control.Undefine(vname)
	// if err != nil {
	// 	msg, _ := json.Marshal(er{Ret: "e", Msg: "删除失败", Data: err.Error()})
	// 	w.Write(msg)
	// 	return
	// }
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
