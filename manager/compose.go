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
	mid, e := sess.Get("mid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var composes []table.Compose
	o := orm.NewOrm()
	_, err := o.Raw("select * from compose where status > 0").QueryRows(&composes)
	if err != nil {
		cLog.Warn("查询套餐列表失败%s", err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "查询失败"), http.StatusFound)
		return
	}
	var manager table.Manager
	err = o.Raw("select * from manager where id = ?", mid).QueryRow(&manager)
	if err != nil {
		cLog.Warn("管理员信息查询失败%s", err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "管理员信息查询失败"), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/manager/compose.html", "html/manager/public/header.html", "html/manager/public/footer.html")
	t.Execute(w, map[string]interface{}{"composes": composes, "email": manager.Email})
}

//添加套餐
func addCompose(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var manager table.Manager
	o := orm.NewOrm()
	err := o.Raw("select * from manager where id = ?", mid).QueryRow(&manager)
	if err != nil {
		cLog.Warn("管理员信息查询失败%s", err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "管理员信息查询失败"), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/manager/addCompose.html", "html/manager/public/header.html", "html/manager/public/footer.html")
	t.Execute(w, map[string]interface{}{"email": manager.Email})
}

//添加套餐
func addComposeInfo(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
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
	compose.Name = name
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

//编辑套餐
func editCompose(w http.ResponseWriter, req *http.Request) {
	//判断是否登录
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	//判断提交方式
	if req.Method != "GET" {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "发生错误", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	//展示页面

}

//下架套餐
func downCompose(w http.ResponseWriter, req *http.Request) {
	//判断是否登录
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录"})
		w.Write(msg)
		return
	}
	//判断提交方式
	if req.Method != "POST" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "提交失败"})
		w.Write(msg)
		return
	}
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数有误"})
		w.Write(msg)
		return
	}
	var compose table.Compose
	compose.ID = id
	o := orm.NewOrm()
	err = o.Read(&compose)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "查询失败"})
		w.Write(msg)
		return
	}
	if compose.Status == 2 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "重复下架"})
		w.Write(msg)
		return
	}
	compose.Status = 2
	_, err = o.Update(&compose)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "更新失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "下架完毕"})
	w.Write(msg)
}

//上架套餐
func upCompose(w http.ResponseWriter, req *http.Request) {
	//判断是否登录
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录"})
		w.Write(msg)
		return
	}
	//判断提交方式
	if req.Method != "POST" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "提交失败"})
		w.Write(msg)
		return
	}
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数有误"})
		w.Write(msg)
		return
	}
	var compose table.Compose
	compose.ID = id
	o := orm.NewOrm()
	err = o.Read(&compose)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "查询失败"})
		w.Write(msg)
		return
	}
	if compose.Status == 1 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "重复上架"})
		w.Write(msg)
		return
	}
	compose.Status = 1
	_, err = o.Update(&compose)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "更新失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "上架完毕"})
	w.Write(msg)
}

func deleteCompose(w http.ResponseWriter, req *http.Request) {
	//判断是否登录
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录"})
		w.Write(msg)
		return
	}
	//判断提交方式
	if req.Method != "POST" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "提交失败"})
		w.Write(msg)
		return
	}
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数有误"})
		w.Write(msg)
		return
	}
	var compose table.Compose
	compose.ID = id
	o := orm.NewOrm()
	err = o.Read(&compose)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "查询失败"})
		w.Write(msg)
		return
	}
	if compose.Status == -1 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "重复删除"})
		w.Write(msg)
		return
	}
	compose.Status = -1
	_, err = o.Update(&compose)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "更新失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "删除完毕"})
	w.Write(msg)
}
