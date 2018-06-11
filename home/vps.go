package home

import (
	"cpanel/control"
	"cpanel/loop"
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/Demired/rpwd"
	"github.com/astaxie/beego/orm"
)

//创建虚拟机
func createAPI(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/login.html?url=%s", req.URL.String()), http.StatusFound)
		return
	}
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "内存大小必须为整数"})
		w.Write(msg)
		return
	}
	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "cpu个数必须为整数"})
		w.Write(msg)
		return
	}
	vpasswd := req.PostFormValue("vpasswd")
	if vpasswd == "" {
		vpasswd = string(rpwd.Init(16, true, true, true, false))
	}
	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "带宽必须为整数"})
		w.Write(msg)
		return
	}
	sys := req.PostFormValue("sys")
	if sys == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "镜像版本必填"})
		w.Write(msg)
		return
	}
	var vInfo table.Virtual
	autopay := req.PostFormValue("autopay")
	if autopay == "0" {
		vInfo.AutoPay = 0
	} else {
		vInfo.AutoPay = 1
	}

	cycle := req.PostFormValue("cycle")
	if cycle == "2" {
		vInfo.Cycle = 2
	} else if cycle == "1" {
		vInfo.Cycle = 1
	} else {
		vInfo.Cycle = 0
	}

	//计算费用

	//查询余额

	//扣费

	//事务支持

	vInfo.UID = uid
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "创建虚拟机硬盘失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	xml := createKvmXML(vInfo)
	connect := control.Connect()
	defer connect.Close()
	_, err = connect.DomainDefineXML(xml)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "创建虚拟机失败", Data: err.Error()})
		w.Write(msg)
		return
	}

	o := orm.NewOrm()
	_, err = o.Insert(&vInfo)
	if err != nil {
		cLog.Error(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		//手动回滚
		return
	}
	loop.VmInit <- vInfo.Vname
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	w.Write(msg)
}

func list(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var virtuals []table.Virtual
	var userInfo table.User
	o := orm.NewOrm()
	o.Raw("select * from virtual where status = ? and uid = ?", "1", uid).QueryRows(&virtuals)
	o.Raw("select * from user where id = ?", uid).QueryRow(&userInfo)
	for k, v := range virtuals {
		connect := control.Connect()
		defer connect.Close()
		dom, err := connect.LookupDomainByName(v.Vname)
		if err != nil {
			cLog.Warn(err.Error())
			continue
		}
		s, _, err := dom.GetState()
		if err != nil {
			cLog.Warn(err.Error())
			continue
		}
		virtuals[k].Status = int(s)
	}
	t, _ := template.ParseFiles("html/home/list.html", "html/home/public/header.html", "html/home/public/footer.html")
	t.Execute(w, map[string]interface{}{"virtuals": virtuals, "userName": userInfo.Username})
}

func info(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	// orm, err := control.Bdb()
	Vname := req.URL.Query().Get("Vname")
	var virtual table.Virtual
	var userInfo table.User
	// err = orm.SetTable("Virtual").Where("Vname = ? and Uid = ?", Vname, uid).Find(&vvm)
	o := orm.NewOrm()
	o.Raw("select * from user where id = ?", uid).QueryRow(&userInfo)
	err := o.Raw("select * from virtual where Vname = ? and Uid = ?", Vname, uid).QueryRow(&virtual)
	if err != nil {
		cLog.Info(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "服务器不存在", "/list"), http.StatusFound)
		return
	}
	err = control.CheckEtime(Vname)
	if err != nil {
		cLog.Info(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "服务器已到期", "/list"), http.StatusFound)
		return
	}
	connect := control.Connect()
	defer connect.Close()
	dom, err := connect.LookupDomainByName(Vname)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	s, _, err := dom.GetState()
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	virtual.Status = int(s)
	t, _ := template.ParseFiles("html/home/info.html", "html/home/public/header.html", "html/home/public/footer.html")
	t.Execute(w, map[string]interface{}{"virtual": virtual, "userName": userInfo.Username})
}

func loadJSON(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	o := orm.NewOrm()
	startTime, err := strconv.Atoi(req.URL.Query().Get("start"))
	if err != nil {
		startTime = int(time.Now().Unix()) - 3600
	}
	endTime, err := strconv.Atoi(req.URL.Query().Get("end"))
	if err != nil {
		endTime = int(time.Now().Unix())
	}
	var watchs []table.Watch
	_, err = o.Raw("select * from watch where vname = ? and ctime > ? and ctime < ?", Vname, startTime, endTime).QueryRows(&watchs)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var virtual table.Virtual
	err = o.Raw("select * from virtual where vname = ?", Vname).QueryRow(&virtual)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var cpus [][]int
	var memorys [][]int
	var up [][]int
	var down [][]int

	for _, v := range watchs {
		up = append(up, []int{v.Ctime, v.Up})
		down = append(down, []int{v.Ctime, v.Down})
		memorys = append(memorys, []int{v.Ctime, v.Memory})
		cpus = append(cpus, []int{v.Ctime, v.CPU})
	}
	var date = make(map[string]interface{})
	date["maxMemory"] = virtual.Vmemory * 1024
	date["cpus"] = cpus
	date["memorys"] = memorys
	date["up"] = up
	date["down"] = down
	dj, _ := json.Marshal(date)
	w.Write(dj)
}
