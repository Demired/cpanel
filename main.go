package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Demired/rpwd"
	libvirt "github.com/libvirt/libvirt-go"
)

func connect() *libvirt.Connect {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return conn
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/list", list)
	http.HandleFunc("/start", start)
	http.HandleFunc("/shutdown", shutdown)
	http.HandleFunc("/reboot", reboot)
	http.HandleFunc("/list_b", listVM)
	http.HandleFunc("/create", createAPI)
	http.HandleFunc("/create.html", create)
	http.ListenAndServe(":8100", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, nil)
}

func create(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/create.html")
	t.Execute(w, nil)
}

func list(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	db, err := sql.Open("sqlite3", "./db/cpanel.db")
	rows, err := db.Query("SELECT Vname,IPv4,IPv6,LocalIP,Vcpu,Vmemory,Status FROM vm LIMIT 10;")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()
	var vvvm []vm
	for rows.Next() {
		var vvm vm
		err := rows.Scan(&vvm.Vname, &vvm.IPv4, &vvm.IPv6, &vvm.LocalIP, &vvm.Vcpu, &vvm.Vmemory, &vvm.Status)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		dom, err := connect().LookupDomainByName(vvm.Vname)
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

	t, _ := template.ParseFiles("html/list_bak.html")
	t.Execute(w, vvvm)
}

func listVM(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	doms, err := connect().ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_NO_AUTOSTART)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, dom := range doms {
		name, err := dom.GetName()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		_, sss, err := dom.GetState()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("vm name is %s,vm state is %d \n", name, sss)
		dom.Free()
	}
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
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	defer req.Body.Close()
	vname := req.PostFormValue("vname")
	err := contrl(vname, 1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	msg, err := json.Marshal(er{Ret: "v", Msg: "正在开机"})
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
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	defer req.Body.Close()
	vname := req.PostFormValue("vname")
	err := contrl(vname, 2)
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
	fmt.Println("reboot")
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	defer req.Body.Close()
	vname := req.PostFormValue("vname")
	err := contrl(vname, 3)
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

func contrl(vname string, c int) error {
	dom, err := connect().LookupDomainByName(vname)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if c == 1 {
		err = dom.Create()
	} else if c == 2 {
		err = dom.Shutdown()
	} else if c == 3 {
		err = dom.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT)
	} else if c == 4 {
		err = dom.Destroy()
	}
	return err
}

func resetPWD(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	defer req.Body.Close()
}

//创建虚拟机
func createAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}

	defer req.Body.Close()

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

	var tvm vm

	tvm.Vcpu = vcpu
	tvm.Vmemory = vmemory
	tvm.Passwd = vpasswd
	tvm.Mac = rmac()
	tvm.Br = "br1"
	tvm.Vname = string(rpwd.Init(8, true, true, true, false))

	xml := createKvmXML(tvm)
	_, err = connect().DomainDefineXML(xml)
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
		msg, _ := json.Marshal(er{Ret: "e", Msg: "数据看打开失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	stmt, err := db.Prepare("INSERT INTO vm(Vname, Vcpu, Vmemory, Status, IPv4, IPv6, LocalIP) values(?,?,?,?,?,?,?)")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "数据库写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	_, err = stmt.Exec(tvm.Vname, tvm.Vcpu, tvm.Vmemory, 1, "", "", "")
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "数据库写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	go setPasswdQueue(tvm.Vname, tvm.Passwd)
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", tvm.Passwd)})
	w.Write(msg)
}

//创建完毕修改密码
func setPasswdQueue(vname string, passwd string) {
	dom, err := connect().LookupDomainByName(vname)
	dom.Create()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ticker := time.NewTicker(time.Second * 20)
	i := 0
	for _ = range ticker.C {
		i++
		s, _, _ := dom.GetState()
		if int(s) == 1 || i > 5 {
			ticker.Stop()
		}
	}
	err = dom.SetUserPassword("root", passwd, libvirt.DOMAIN_PASSWORD_ENCRYPTED)
	if err != nil {
		//记录初始化密码失败
	}
	return
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
				<target type='virtio' name='org.qemu.guest_agent.0'/>
				<address type='virtio-serial' controller='0' bus='0' port='1'/>
			</channel>
		</devices>
	</domain>`
	return templateXML
}

func rmac() string {
	str := "0123456789abcdef"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 12; i++ {
		if i%2 == 0 && i != 0 {
			result = append(result, ':')
		}
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

type vm struct {
	ID      int    `json:"id"`
	IPv4    string `json:"ipv4"`
	IPv6    string `json:"ipv6"`
	LocalIP string `json:"local"`
	Ctime   string `json:"ctime"`
	Utime   string `json:"utime"`
	Vcpu    int    `json:"vcpu"`
	Status  int    `json:"status"`
	Etime   string `json:"etime"`   //Expire time
	Vmemory int    `json:"vmemory"` //GiB
	Passwd  string `json:"vpasswd"`
	Vname   string `json:"vname"`
	Br      string `json:"br"`
	Mac     string `json:"mac"`
}
