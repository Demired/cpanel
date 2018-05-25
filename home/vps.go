package home

import (
	"cpanel/control"
	"cpanel/loop"
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
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
	_, err := o.Insert(&vInfo)
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
