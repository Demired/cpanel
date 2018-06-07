package home

import (
	"cpanel/table"
	"cpanel/tools"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/Demired/rpwd"
	"github.com/astaxie/beego/orm"
)

func login(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Query().Get("url")
	//检查url域名
	t, _ := template.ParseFiles("html/home/login.html")
	t.Execute(w, url)
}

func loginAPI(w http.ResponseWriter, req *http.Request) {
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	if email == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "用户名不能为空"})
		w.Write(msg)
		return
	}
	if passwd == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "passwd", Msg: "密码不能为空"})
		w.Write(msg)
		return
	}
	emailReg := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if !emailReg.Match([]byte(email)) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "请检查邮箱拼写是否有误"})
		w.Write(msg)
		return
	}
	passReg := regexp.MustCompile(`^[\w!@#$%^&*()-_=+]*$`)
	if !passReg.Match([]byte(passwd)) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "passwd", Msg: "密码只支持数字，大小写字母，和\"!@#$%^&*()_-=+\""})
		w.Write(msg)
		return
	}
	var user table.User
	o := orm.NewOrm()
	err := o.Raw("select * from user where email = ?", email).QueryRow(&user)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "用户不存在"})
		w.Write(msg)
		return
	}
	if user.Status != 1 {
		var msg []byte
		switch user.Status {
		case 0:
			msg, _ = json.Marshal(tools.Er{Ret: "e", Msg: "邮箱验证未通过"})
			break
		case -1:
			msg, _ = json.Marshal(tools.Er{Ret: "e", Msg: "帐号已禁用"})
			break
		}
		w.Write(msg)
		return
	}
	h := sha1.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	if user.Passwd != string(bs) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "passwd", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	sess.Set("uid", user.ID)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "登录成功"})
	w.Write(msg)
}

func registerAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	if email == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "邮箱不能为空"})
		w.Write(msg)
		return
	}
	if passwd == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "passwd", Msg: "密码不能为空"})
		w.Write(msg)
		return
	}
	emailReg := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if !emailReg.Match([]byte(email)) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "请检查邮箱拼写是否有误"})
		w.Write(msg)
		return
	}
	passReg := regexp.MustCompile(`^[\w!@#$%^&*()-_=+]*$`)
	if !passReg.Match([]byte(passwd)) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "passwd", Msg: "密码只支持数字，大小写字母，和\"!@#$%^&*()_-=+\""})
		w.Write(msg)
		return
	}
	h := sha1.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	var tmpUser table.User
	o := orm.NewOrm()
	err := o.Raw("select * from user where Email = ?", email).QueryRow(&tmpUser)
	if err == nil {
		cLog.Info("邮箱：%s，重复注册", email)
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "邮箱已注册，你可以尝试登录或者找回密码"})
		w.Write(msg)
		return
	}
	var userInfo table.User
	userInfo.Email = email
	userInfo.Passwd = string(bs)
	userInfo.Status = 0
	_, err = o.Insert(&userInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	cLog.Info("%s注册成功", email)
	//注册验证
	var verify table.Verify
	verify.Ctime = time.Now()
	verify.Vtime = time.Now()
	verify.Type = "verify"
	verify.Code = fmt.Sprintf("%x", string(rpwd.Init(16, true, true, true, false)))
	verify.Email = email
	verify.Status = 0
	_, err = o.Insert(&verify)
	htmlBody := fmt.Sprintf("<h1>注册验证</h1><p>点击<a href='http://172.16.1.181:8100/verify?code=%s&email=%s'>链接</a>验证注册，非本人操作请忽略</p>", verify.Code, verify.Email)
	tools.SendMail(email, "注册验证", htmlBody)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "注册完毕,请前往邮箱查收验证邮件"})
	w.Write(msg)
}

func verify(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	email := req.URL.Query().Get("email")
	o := orm.NewOrm()
	var thisUser table.User
	err := o.Raw("select * from user where Email = ?", email).QueryRow(&thisUser)
	if err != nil {
		cLog.Info("邮箱不存在,邮箱:%s", email)
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "邮箱不存在"), http.StatusFound)
		return
	}
	if thisUser.Status == 1 {
		cLog.Info("用户重复验证,邮箱:%s", email)
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "重复验证"), http.StatusFound)
		return
	}

	var thisVerify table.Verify
	err = o.Raw("select * from verify where Code = ? and Email = ?", code, email).QueryRow(&thisVerify)
	if err != nil {
		cLog.Info("邮箱验证失败，邮箱：%s", email)
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "验证失败"), http.StatusFound)
		return
	}
	var nowTime = time.Now()
	subTime, _ := time.ParseDuration("-24h")
	lastTime := nowTime.Add(subTime)
	if lastTime.After(thisVerify.Ctime) {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "链接已过期，请通过找回密码，重新发起验证", "/forget.html"), http.StatusFound)
		return
	}
	if thisVerify.Status == 1 {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "链接已作废，请通过找回密码，重新发起验证"), http.StatusFound)
		return
	}
	if thisVerify.Type == "verify" {
		thisVerify.Status = 1
		o.Update(&thisVerify)
		thisUser.Status = 1
		o.Update(&thisUser)
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "邮箱验证完毕，请登录", "/login.html"), http.StatusFound)
		return
	} else if thisVerify.Type == "forget" {
		t, _ := template.ParseFiles("html/home/verify.html")
		t.Execute(w, map[string]string{"email": email, "code": code})
	}
}
