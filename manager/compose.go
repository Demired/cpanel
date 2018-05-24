package manager

import (
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
)

//套餐列表 套餐模板
func compose(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("uid").(int)
	if !e {
		//TODO跳转登录页面
		return
	}
	var composes []table.Compose
	o := orm.NewOrm()
	_, err := o.Raw("Select * from compose where status = ?", "1").QueryRows(&composes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t, _ := template.ParseFiles("html/manager/compose.html")
	t.Execute(w, composes)
}

//套餐列表
func composes(w http.ResponseWriter, req *http.Request) {
	o := orm.NewOrm()
	o.Raw("select * from Composes where status = 1")
	fmt.Println("123")
}

func addCompose(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/manager/addCompose.html")
	t.Execute(w, nil)
}

//添加套餐
func addComposeInfo(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("uid").(int)
	if !e {
		//TODO跳转登录页面
		return
	}

	if req.Method != "POST" {
		//TODO提交方式有误
		return
	}

	var compose table.Compose
	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "带宽必须填写"})
		w.Write(msg)
		return
	}
	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "cpu数量必须填写"})
		w.Write(msg)
		return
	}
	ipv4, err := strconv.Atoi(req.PostFormValue("ipv4"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "ipv4个数必须填写"})
		w.Write(msg)
		return
	}
	ipv6, err := strconv.Atoi(req.PostFormValue("ipv6"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "ipv6个数必须填写"})
		w.Write(msg)
		return
	}
	price, err := strconv.Atoi(req.PostFormValue("price"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "单价必须填写"})
		w.Write(msg)
		return
	}
	vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "内存必须填写"})
		w.Write(msg)
		return
	}
	totalflow, err := strconv.Atoi(req.PostFormValue("totalflow"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "流量上限必须填写"})
		w.Write(msg)
		return
	}

	compose.BandWidth = bandwidth
	compose.Vcpu = vcpu
	compose.IPv4 = ipv4
	compose.IPv6 = ipv6
	compose.Price = price
	compose.Vmemory = vmemory
	compose.TotalFlow = totalflow

	o := orm.NewOrm()

	res, err := o.Insert(&compose)

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res)

	// compose.Vcpu = req.PostFormValue("vcpu")
	// compose.IPv4 = req.PostFormValue("ipv4")
	// compose.IPv6 = req.PostFormValue("ipv6")
	// compose.Price = req.PostFormValue("price")
	// compose.Vmemory = req.PostFormValue("vmemory")
	// compose.TotalFlow = req.PostFormValue("totalflow")
	fmt.Println(compose)
}
