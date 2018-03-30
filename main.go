package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/template"

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
		_, sss, err := dom.GetState()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		vvm.Status = sss
		vvvm = append(vvvm, vvm)
	}

	t, _ := template.ParseFiles("html/list_bak.html")
	t.Execute(w, vvvm)
}

func listVM(w http.ResponseWriter, req *http.Request) {
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

	//开机
	//入库

	return io.Copy(desFile, srcFile)
}

func createAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}

	defer req.Body.Close()

	vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))

	if err != nil {
		fmt.Println("vmemory value err")
		w.Write([]byte("vm memory value err"))
		return
	}

	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		fmt.Println("vm cpu number value err")
		w.Write([]byte("vm cpu number value err"))
		return
	}
	vpasswd := req.PostFormValue("vpasswd")
	if vpasswd == "" {
		vpasswd = string(rpwd.Init(16, true, true, true, true))
		w.Write([]byte(fmt.Sprintf("you passwd is:%s\n", vpasswd)))
	}

	var tvm vm

	tvm.Vcpu = vcpu
	tvm.Vmemory = vmemory
	tvm.Passwd = vpasswd
	tvm.Mac = rmac()
	tvm.Br = "br1"
	tvm.Vname = string(rpwd.Init(8, true, true, true, false))

	xml := createKvmXML(tvm)
	dom, err := connect().DomainDefineXML(xml)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	rname, _ := dom.GetName()

	createSysDisk(rname)
	fmt.Println(rname)
	w.Write([]byte(fmt.Sprintf("you vm name is:%s\n", rname)))
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
	return "1e:16:3e:77:e2:ed"
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
