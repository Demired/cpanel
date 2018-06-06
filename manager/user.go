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

//用户列表模板
func userList(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	if !e {
		//TODO跳转登录页面
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var users []table.User
	o := orm.NewOrm()
	_, err := o.Raw("Select * from user").QueryRows(&users)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "查询失败", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/manager/userList.html")
	t.Execute(w, map[string]interface{}{"users": users, "mid": mid})
}

//禁用用户
func downUser(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
	if !e {
		//TODO跳转登录页面
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录", Param: "login"})
		w.Write(msg)
		return
	}
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数不全"})
		w.Write(msg)
		return
	}
	var user table.User
	o := orm.NewOrm()
	user.ID = id
	err = o.Read(&user)
	// err = o.Raw("select * from user").QueryRow(&user)
	// 查找用户
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "用户不存在"})
		w.Write(msg)
		return
	}
	if user.Status == 0 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "重复禁用"})
		w.Write(msg)
		return
	}
	user.Status = 0
	_, err = o.Update(&user)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "禁用失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "禁用完毕"})
	w.Write(msg)
	return
}

//启用用户
func upUser(w http.ResponseWriter, req *http.Request) {

}
