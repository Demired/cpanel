package main

import (
	"cpanel/control"
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
	libvirt "github.com/libvirt/libvirt-go"
	_ "github.com/mattn/go-sqlite3"

	"github.com/Demired/rpwd"
)

var q = make(chan string)

var mac = make(chan string)

func main() {
	go vmWatch()
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

var t = make(map[string]uint64)

func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, nil)
}

type wa struct {
	Vname  string
	CPU    int
	Memory int
	Ctime  int
}
type watch struct {
	Vname  string
	CPU    int
	Memory uint64
	Ctime  int64
}

func vmWatch() {
	w := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-w.C:
			doms, err := control.Connect().ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
			if err != nil {
				fmt.Println(err.Error())
			}
			db, err := sql.Open("sqlite3", "./db/cpanel.db")
			if err != nil {
				fmt.Println("打开数据库失败", err.Error())
				continue
			}
			orm := beedb.New(db)

			for _, dom := range doms {
				name, _ := dom.GetName()
				info, err := dom.GetInfo()
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				var wd watch
				wd.Ctime = time.Now().Unix()
				var cpurate float32
				if lastCPUTime, ok := t[name]; ok {
					cpurate = float32((info.CpuTime-lastCPUTime)*100) / float32(20*info.NrVirtCpu*10000000)
				}
				if cpurate < 1 {
					wd.CPU = 1
				} else {
					wd.CPU = int(cpurate)
				}
				wd.Memory = info.Memory
				err := orm.Save(&w)
				if err != nil {
					fmt.Println("写入数据失败", err.Error())
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
		fmt.Println(err.Error())
	}
	s, _, err := dom.GetState()
	if err != nil {
		fmt.Println(err.Error())
	}
	if int(s) == 1 {
		info, err := dom.GetInfo()
		if err != nil {
			fmt.Println(info)
		}
	}
	db, _ := sql.Open("sqlite3", "./db/cpanel.db")
	defer db.Close()
	sql := fmt.Sprintf("SELECT Vname,IPv4,IPv6,LocalIP,Mac,Vcpu,Bandwidth,Vmemory,Status FROM vm WHERE vname = '%s';", vname)
	rows, _ := db.Query(sql)
	defer rows.Close()
	var vvm vm
	if rows.Next() {
		rows.Scan(&vvm.Vname, &vvm.IPv4, &vvm.IPv6, &vvm.LocalIP, &vvm.Mac, &vvm.Vcpu, &vvm.Bandwidth, &vvm.Vmemory, &vvm.Status)
	}
	var vmInfo = make(map[string]string)
	vmInfo["Vname"] = vvm.Vname
	vmInfo["IPv4"] = vvm.IPv4
	vmInfo["IPv6"] = vvm.IPv6
	vmInfo["Mac"] = vvm.Mac
	vmInfo["LocalIP"] = vvm.LocalIP
	vmInfo["Bandwidth"] = fmt.Sprintf("%d", vvm.Bandwidth)
	vmInfo["Vmemory"] = fmt.Sprintf("%d", vvm.Vmemory)
	vmInfo["Vcpu"] = fmt.Sprintf("%d", vvm.Vcpu)
	vmInfo["Status"] = fmt.Sprintf("%d", s)
	// vmInfo[""]
	t, _ := template.ParseFiles("html/info.html")
	// vmInfoJ, _ := json.Marshal(vmInfo)
	t.Execute(w, vmInfo)
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
		var ww wa
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
	var vvvm []vm
	for rows.Next() {
		var vvm vm
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
	defer req.Body.Close()
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
	var tvm vm

	tvm.Vcpu = vcpu
	tvm.Vmemory = vmemory
	tvm.Passwd = vpasswd
	tvm.Mac = tools.Rmac()
	tvm.Br = "br1"
	tvm.Bandwidth = bandwidth
	tvm.Vname = string(rpwd.Init(8, true, true, true, false))

	xml := createKvmXML(tvm)
	_, err = control.Connect().DomainDefineXML(xml)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	_, err = createSysDisk(tvm.Vname)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机硬盘失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	stmt, err := db.Prepare("INSERT INTO vm(UID,Vname, Vcpu, Vmemory, Mac, Bandwidth, Status,IPv4,IPv6,LocalIP) values(?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	_, err = stmt.Exec(1, tvm.Vname, tvm.Vcpu, tvm.Vmemory, tvm.Mac, tvm.Bandwidth, 1, "", "", "")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入数据失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	str := fmt.Sprintf("%s/%s", tvm.Vname, tvm.Passwd)
	fmt.Println(str)
	q <- str
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", tvm.Passwd)})
	w.Write(msg)
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

func createKvmXML(tvm vm) string {
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

type vm struct {
	ID        int    `json:"id"`
	IPv4      string `json:"ipv4"`
	IPv6      string `json:"ipv6"`
	LocalIP   string `json:"local"`
	Ctime     string `json:"ctime"`
	Utime     string `json:"utime"`
	Vcpu      int    `json:"vcpu"`
	Status    int    `json:"status"`
	Etime     string `json:"etime"`   //Expire time
	Vmemory   int    `json:"vmemory"` //GiB
	Passwd    string `json:"vpasswd"`
	Vname     string `json:"vname"`
	Br        string `json:"br"`
	Mac       string `json:"mac"`
	Bandwidth int    `json:"bandwidth"` //Mbps
}
