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
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
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
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	if req.Method != "POST" {
		//TODO提交方式有误
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "提交失败"), http.StatusFound)
		return
	}
	var compose table.Compose
	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "cpu数量必须填写"})
		w.Write(msg)
		return
	}
	name := req.PostFormValue("name")
	if name == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "套餐名"})
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
	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "带宽必须填写"})
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
	price, err := strconv.Atoi(req.PostFormValue("price"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "单价必须填写"})
		w.Write(msg)
		return
	}
	compose.Bandwidth = bandwidth
	compose.Vcpu = vcpu
	compose.IPv4 = ipv4
	compose.IPv6 = ipv6
	compose.Price = price
	compose.Vmemory = vmemory
	compose.Status = 1
	compose.TotalFlow = totalflow
	o := orm.NewOrm()
	_, err = o.Insert(&compose)
	if err != nil {
		cLog.Error(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "添加失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "添加套餐成功"})
	w.Write(msg)
}
