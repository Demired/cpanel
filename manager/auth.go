package manager

import (
	"cpanel/table"
	"cpanel/tools"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/Demired/rpwd"

	"github.com/astaxie/beego/orm"
)

//login web template
func login(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	url := req.URL.Query().Get("url")
	defer sess.SessionRelease(w)
	token := string(rpwd.Init(16, true, true, true, false))
	sess.Set("loginToken", token)
	t, _ := template.ParseFiles("html/manager/login.html")
	t.Execute(w, map[string]string{"token": token, "url": url})
}

//login api
func loginAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/login.html", http.StatusFound)
	}
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	token := req.PostFormValue("token")
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	loginToken := sess.Get("loginToken")
	if token != loginToken {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	o := orm.NewOrm()
	var manager table.Manager
	err := o.Raw("select * from Manager where Email = ?", email).QueryRow(&manager)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "用户不存在"})
		w.Write(msg)
		return
	}
	h := sha1.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	sha1passwd := fmt.Sprintf("%x", bs)
	if manager.Passwd != sha1passwd {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	sess.Set("uid", manager.ID)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "登录成功"})
	w.Write(msg)
}

func notFound(w http.ResponseWriter, req *http.Request) {
	msg := req.URL.Query().Get("msg")
	url := req.URL.Query().Get("url")
	if msg == "" {
		msg = "页面找不到"
	}
	if url == "" {
		url = "/"
	}
	t, _ := template.ParseFiles("html/manager/notFound.html")
	t.Execute(w, map[string]string{"msg": msg, "url": url})
}
