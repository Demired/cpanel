package manager

import (
	"cpanel/table"
	"fmt"
	"html/template"
	"net/http"

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

//添加套餐
func addCompose(w http.ResponseWriter, req *http.Request) {
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

	vcpu := req.PostFormValue("vcpu")
	ipv4 := req.PostFormValue("ipv4")
	ipv6 := req.PostFormValue("ipv6")
	bandwidth := req.PostFormValue("bandwidth")

	fmt.Println(vcpu)
	fmt.Println(ipv4)
	fmt.Println(ipv6)
	fmt.Println(bandwidth)
}
