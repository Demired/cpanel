package main

import (
	"cpanel/config"
	"cpanel/control"
	"cpanel/loop"
	"cpanel/table"
	"cpanel/tools"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Demired/rpwd"
)

var cLog = config.CLog
var cSession = config.CSession

func main() {
	go loop.Watch()
	go loop.WorkQueue()
	http.HandleFunc("/", index)
	http.HandleFunc("/verify", verify)
	http.HandleFunc("/repwd", repwd)
	http.HandleFunc("/edit", editAPI)
	http.HandleFunc("/list", list)
	http.HandleFunc("/login.html", login)
	http.HandleFunc("/login", loginAPI)
	http.HandleFunc("/forget.html", forget)
	http.HandleFunc("/logout", logoutAPI)
	http.HandleFunc("/userInfo.html", userInfo)
	http.HandleFunc("/userInfo", userInfoAPI)
	http.HandleFunc("/register.html", register)
	http.HandleFunc("/register", registerAPI)
	http.HandleFunc("/info.html", info)
	http.HandleFunc("/load.json", loadJSON)
	http.HandleFunc("/start", start)
	http.HandleFunc("/shutdown", shutdown)
	http.HandleFunc("/reboot", reboot)
	http.HandleFunc("/create", createAPI)
	http.HandleFunc("/404.html", notFound)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/repasswd.html", repasswd)
	http.HandleFunc("/alarm.html", alarm)
	http.HandleFunc("/alarm", alarmAPI)
	http.HandleFunc("/repasswd", repasswdAPI)
	http.HandleFunc("/undefine", undefine)
	http.HandleFunc("/edit.html", edit)
	http.HandleFunc("/create.html", create)
	http.ListenAndServe(":8100", nil)
}

func verify(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	email := req.URL.Query().Get("email")
	var tmpVerify = table.Verify
	orm, _ := control.Bdb()
	fb := orm.SetTable("Verify").SetPK("ID").Where("Code = ? and Email = ?", code, email).Find(&tmpVerify)
	if fb != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "验证失败"), http.StatusFound)
		return
	}
	var t = time.Now()
	t.ParseDuration("-24h")
	if t.After(tmpVerify.Ctime) {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "已过期，请通过找回密码，重新发起验证","/forget.html"), http.StatusFound)
		return
	}
	if tmpVerify.Status == 1 {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "不要重复验证"), http.StatusFound)
		return
	}
	if tmpVerify.Type == "verify"{
		var vData = make(map[string]interface{})
		vData["Status"] = 1
		vData["Vtime"] = time.Now()
		orm.SetTable("Verify").SetPK("ID").Where("Code = ? and Email = ?", code, email).Update(vData)
		var uData = make(map[string]interface{})
		uData["Status"] = 1
		orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Update(vData)
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "邮箱验证完毕，请登录", "/login.html"), http.StatusFound)
		return
	}else if tmpVerify.Type == "forget"{
		t, _ := template.ParseFiles("html/verify.html")
		t.Execute(w, map[string]string{"email": email, "code": code})
	}
}

func repwd(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	code := req.PostFormValue("code")
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	fb := orm.SetTable("Verify").SetPK("ID").Where("Code = ? and Email = ?", code, email).Find(&tmpVerify)
	if fb != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "验证失败"), http.StatusFound)
		return
	}
	var t time.Now()
	t.ParseDuration("-24h")
	if t.After(tmpVerify.Ctime) {
		msg,_:=json.Marshal(er{Ret:"e",Msg:"已过期，请通过找回密码，重新发起验证"})
		w.Write(msg)
		return
	}
	if tmpVerify.Status == 1 {
		msg,_:=json.Marshal(er{Ret:"e",Msg:"不要重复验证"})
		w.Write(msg)
		return
	}
	passReg := regexp.MustCompile(`^[\w!@#$%^&*()-_=+]*$`)
	if !passReg.Match([]byte(passwd)) {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码只支持数字，大小写字母，和\"!@#$%^&*()_-=+\""})
		w.Write(msg)
		return
	}
	var vData = make(map[string]interface{})
	vData["Status"] = 1
	vData["Vtime"] = time.Now()
	orm.SetTable("Verify").SetPK("ID").Where("Code = ? and Email = ?", code, email).Update(vData)
	
	h := sha1.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	var uData = make(map[string]interface{})
	uData["Passwd"] = string(bs)
	orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Update(vData)
	t, _ := template.ParseFiles("html/forget.html")
	t.Execute(w, nil)
}

func forget(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/forget.html")
	t.Execute(w, nil)
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

	t, _ := template.ParseFiles("html/notFound.html")
	t.Execute(w, map[string]string{"msg": msg, "url": url})
}

func userInfo(w http.ResponseWriter, req *http.Request) {
	sess, err := cSession.SessionStart(w, req)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你没有登录", fmt.Sprintf("login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}

	var userInfo table.User
	orm, _ := control.Bdb()
	fb := orm.SetTable("User").SetPK("ID").Where("ID = ?", uid).Find(&userInfo)
	if fb != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "用户不存在"), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/userInfo.html")
	t.Execute(w, userInfo)
}

func index(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, _ := sess.Get("uid").(int)
	t, _ := template.ParseFiles("html/index.html")
	t.Execute(w, map[string]int{"uid": uid})
}

func register(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/register.html")
	t.Execute(w, nil)
}

func userInfoAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	sess, err := cSession.SessionStart(w, req)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "系统错误，请联系管理员"})
		w.Write(msg)
		return
	}
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "未登录"})
		w.Write(msg)
		return
	}
	var date = make(map[string]interface{})
	date["Username"] = req.PostFormValue("username")
	date["Tel"] = req.PostFormValue("tel")
	date["Realname"] = req.PostFormValue("realname")
	date["Idtype"] = req.PostFormValue("idtype")
	date["Idnumber"] = req.PostFormValue("idnumber")
	date["City"] = req.PostFormValue("city")
	date["Company"] = req.PostFormValue("company")
	date["Address"] = req.PostFormValue("address")
	if req.PostFormValue("idtype") == "1" {
		date["idtype"] = 1
	} else if req.PostFormValue("idtype") == "2" {
		date["idtype"] = 2
	}
	if req.PostFormValue("sex") == "0" {
		date["Sex"] = 0
	} else {
		date["Sex"] = 1
	}
	if req.PostFormValue("username") != "" {
		usernameReg := regexp.MustCompile(`^[a-zA-Z0-9]{4,16}$`)
		if !usernameReg.Match([]byte(req.PostFormValue("username"))) {
			msg, _ := json.Marshal(er{Ret: "e", Param: "username-box", Msg: "用户名只允许大小写字母和数字"})
			w.Write(msg)
			return
		}
	}
	if req.PostFormValue("tel") != "" {
		telReg := regexp.MustCompile(`^1[0-9]{10}$`)
		if !telReg.Match([]byte(req.PostFormValue("tel"))) {
			msg, _ := json.Marshal(er{Ret: "e", Param: "tel-box", Msg: "手机号有误"})
			w.Write(msg)
			return
		}
	}
	orm, _ := control.Bdb()
	_, err = orm.SetTable("User").SetPK("ID").Where("ID = ?", uid).Update(date)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "修改失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "修改成功"})
	w.Write(msg)
}

func forgetAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	email := req.PostFormValue("email")
	if email == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "邮箱不能为空"})
		w.Write(msg)
		return
	}
	var tmpUser table.User
	orm, _ := control.Bdb()
	err := orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Find(&tmpUser)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "账号不存在"})
		w.Write(msg)
		return
	}
	var tmpVerify table.Verify
	err = orm.SetTable("Verify").SetPK("ID").Where("Email = ?", email).Find(&tmpVerify)
	tmpVerify
	if tmpUser.Status == 0 {
		//注册
	} else {
		//找回密码
	}

}

func registerAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	if email == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "邮箱不能为空"})
		w.Write(msg)
		return
	}
	if passwd == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码不能为空"})
		w.Write(msg)
		return
	}
	emailReg := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if !emailReg.Match([]byte(email)) {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "请检查邮箱拼写是否有误"})
		w.Write(msg)
		return
	}
	passReg := regexp.MustCompile(`^[\w!@#$%^&*()-_=+]*$`)
	if !passReg.Match([]byte(passwd)) {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码只支持数字，大小写字母，和\"!@#$%^&*()_-=+\""})
		w.Write(msg)
		return
	}
	h := sha1.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	var tmpUser table.User
	orm, _ := control.Bdb()
	fb := orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Find(&tmpUser)
	if fb == nil {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "邮箱已注册，你可以尝试登录或者找回密码"})
		w.Write(msg)
		return
	}
	var userInfo table.User
	userInfo.Email = email
	userInfo.Passwd = string(bs)
	userInfo.Utime = time.Now()
	userInfo.Ctime = time.Now()
	userInfo.Status = 0
	err := orm.SetTable("User").SetPK("ID").Save(&userInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	cLog.Info("%注册成功", email)
	//注册验证
	var v table.Verify
	v.Ctime = time.Now()
	v.Vtime = time.Now()
	v.Code = string(rpwd.Init(16, true, true, true, false))
	v.Email = email
	v.Status = 0
	orm.SetTable("Verify").SetPK("ID").Save(&v)
	htmlBody := fmt.Sprintf("<h1>注册验证</h1><p>点击<a href='http://172.16.1.181:8100/verify?code=%s&email=%s'>链接</a>验证注册，非本人操作请忽略</p>", v.Code, v.Email)
	tools.SendMail(email, "注册验证", htmlBody)
	msg, _ := json.Marshal(er{Ret: "v", Msg: "注册完毕,请前往邮箱查收验证邮件"})
	w.Write(msg)
}

func ref(w http.ResponseWriter, req *http.Request) {

}

func login(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Query().Get("url")
	//检查url域名
	t, _ := template.ParseFiles("html/login.html")
	t.Execute(w, url)
}

func logoutAPI(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	sess.Delete("uid")
	msg, _ := json.Marshal(er{Ret: "v", Msg: "注销完毕"})
	w.Write(msg)
}

func loginAPI(w http.ResponseWriter, req *http.Request) {
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	if email == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "用户名不能为空"})
		w.Write(msg)
		return
	}
	if passwd == "" {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码不能为空"})
		w.Write(msg)
		return
	}
	emailReg := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if !emailReg.Match([]byte(email)) {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "请检查邮箱拼写是否有误"})
		w.Write(msg)
		return
	}
	passReg := regexp.MustCompile(`^[\w!@#$%^&*()-_=+]*$`)
	if !passReg.Match([]byte(passwd)) {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码只支持数字，大小写字母，和\"!@#$%^&*()_-=+\""})
		w.Write(msg)
		return
	}
	var user table.User
	orm, _ := control.Bdb()
	err := orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Find(&user)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Param: "email", Msg: "用户不存在"})
		w.Write(msg)
		return
	}
	if user.Status == 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "邮箱验证未通过"})
		w.Write(msg)
		return
	}
	h := sha1.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	if user.Passwd != string(bs) {
		msg, _ := json.Marshal(er{Ret: "e", Param: "passwd", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	sess.Set("uid", user.ID)
	msg, _ := json.Marshal(er{Ret: "v", Msg: "登录成功"})
	w.Write(msg)
}

func favicon(w http.ResponseWriter, req *http.Request) {
	path := "./html/images/favicon.ico"
	http.ServeFile(w, req, path)
}

func create(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	_, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/create.html")
	t.Execute(w, nil)
}

func edit(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	var vvm table.Virtual
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	orm.SetTable("Virtual").Where("Vname = ?", Vname).Find(&vvm)
	if time.Now().After(vvm.Etime) {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "服务器已经到期", "/list"), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/create.html")
	t.Execute(w, vvm)
}

func info(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	Vname := req.URL.Query().Get("Vname")
	var vvm table.Virtual
	err = orm.SetTable("Virtual").Where("Vname = ? and Uid = ?", Vname, uid).Find(&vvm)
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
	vvm.Status = int(s)
	t, _ := template.ParseFiles("html/info.html")
	t.Execute(w, vvm)
}

func loadJSON(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	Vname := req.URL.Query().Get("Vname")
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	startTime, err := strconv.Atoi(req.URL.Query().Get("start"))
	if err != nil {
		startTime = int(time.Now().Unix()) - 3600
	}
	endTime, err := strconv.Atoi(req.URL.Query().Get("end"))
	if err != nil {
		endTime = int(time.Now().Unix())
	}
	var watchs []table.Watch
	err = orm.SetTable("Watch").Where("Vname = ? and Ctime > ? and Ctime < ?", Vname, startTime, endTime).FindAll(&watchs)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var virtual table.Virtual
	err = orm.SetTable("Virtual").Where("Vname = ?", Vname).Find(&virtual)
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

func repasswd(w http.ResponseWriter, req *http.Request) {
	Vname := req.URL.Query().Get("Vname")
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var watch table.Watch
	err = orm.SetTable("Watch").Find(&watch)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, "/list", http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/repasswd.html")
	t.Execute(w, Vname)
	return
}

func list(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var vvvm []table.Virtual
	err = orm.SetTable("Virtual").Where("Status = ? and Uid = ?", "1", uid).FindAll(&vvvm)
	for k, v := range vvvm {
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
		vvvm[k].Status = int(s)
	}
	t, _ := template.ParseFiles("html/list.html")
	t.Execute(w, vvvm)
}

func createSysDisk(Vname, mirror string) (w int64, err error) {
	mirrorPath := fmt.Sprintf("/virt/mirror/%s.qcow2", mirror)
	srcFile, err := os.Open(mirrorPath)
	if err != nil {
		cLog.Info(err.Error())
		return 0, err
	}
	defer srcFile.Close()
	diskPath := fmt.Sprintf("/virt/disk/%s.qcow2", Vname)
	desFile, err := os.Create(diskPath)
	if err != nil {
		fmt.Println(err)
	}
	defer desFile.Close()
	return io.Copy(desFile, srcFile)
}

func start(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	Vname := req.PostFormValue("Vname")
	var tmpVirtual table.Virtual
	err = orm.SetTable("Virtual").SetPK("ID").Where("Uid = ? and Vname = ?", uid, Vname).Find(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}
	err = control.CheckEtime(Vname)
	if err != nil {
		cLog.Warn("检查到期：%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "服务器已经到期", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Start(Vname)
	if err != nil {
		cLog.Warn("虚拟机开启:%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "开机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "正在开机"})
	w.Write(msg)
}

func repasswdAPI(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	passwd := req.PostFormValue("passwd")
	err := control.SetPasswd(Vname, "root", passwd)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "密码已重置"})
	w.Write(msg)
}

type er struct {
	Ret   string `json:"ret"`
	Msg   string `json:"msg"`
	Data  string `json:"data"`
	Param string `json:"param"`
}

func shutdown(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	Vname := req.PostFormValue("Vname")
	var tmpVirtual table.Virtual
	err = orm.SetTable("Virtual").SetPK("ID").Where("Uid = ? and Vname = ?", uid, Vname).Find(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}
	err = control.Shutdown(Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "关机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "正在关机"})
	w.Write(msg)
}

func reboot(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	Vname := req.PostFormValue("Vname")
	var tmpVirtual table.Virtual
	err = orm.SetTable("Virtual").SetPK("ID").Where("Uid = ? and Vname = ?", uid, Vname).Find(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}
	err = control.Reboot(Vname)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "重启失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "正在重启"})
	w.Write(msg)
}

func editAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var sourceVirtual table.Virtual
	vname := req.PostFormValue("vname")
	err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", vname).Find(&sourceVirtual)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "机器不存在"})
		w.Write(msg)
		return
	}
	if time.Now().After(sourceVirtual.Etime) {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "机器已经到期"})
		w.Write(msg)
		return
	}

	vmemory, err := strconv.Atoi(req.PostFormValue("vmemory"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存大小必须为整数"})
		w.Write(msg)
		return
	}

	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "cpu个数必须为整数"})
		w.Write(msg)
		return
	}

	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽必须位整数"})
		w.Write(msg)
		return
	}

	if sourceVirtual.Vmemory != vmemory {

	}

	sys := req.PostFormValue("sys")
	if sys == "" {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "镜像必填"})
		w.Write(msg)
		return
	}

	var vInfo table.Virtual
	vInfo.Vname = string(rpwd.Init(8, true, true, true, false))
	vInfo.Vcpu = vcpu
	vInfo.Vmemory = vmemory
	vInfo.Mac = tools.Rmac()
	vInfo.Br = "br1"
	vInfo.Status = 1
	vInfo.Bandwidth = bandwidth
	vInfo.Utime = time.Now()
	vInfo.Sys = sys

	_, err = createSysDisk(vInfo.Vname, vInfo.Sys)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机硬盘失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	xml := createKvmXML(vInfo)
	connect := control.Connect()
	defer connect.Close()
	_, err = connect.DomainDefineXML(xml)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机失败", Data: err.Error()})
		w.Write(msg)
		return
	}

	err = orm.SetTable("Virtual").SetPK("ID").Save(&vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	loop.VmInit <- vInfo.Vname
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	w.Write(msg)
}

func alarm(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	Vname := req.URL.Query().Get("Vname")
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	var dInfo table.Virtual
	err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ? and Uid = ?", Vname, uid).Find(&dInfo)
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "没有权限"), http.StatusFound)
		return
	}
	if time.Now().After(dInfo.Etime) {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "虚拟机已到期", "/list"), http.StatusFound)
		return
	}
	if dInfo.AStatus == 0 {
		dInfo.ABandwidth = 0
		dInfo.ACpu = 0
		dInfo.ADisk = 0
		dInfo.AMemory = 0
	}
	//检查虚拟机所有者
	t, _ := template.ParseFiles("html/alarm.html")
	t.Execute(w, dInfo)
}

func alarmAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	var dInfo table.Virtual
	Vname := req.PostFormValue("Vname")
	err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ? and Uid = ?", Vname, uid).Find(&dInfo)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "没有权限", Data: err.Error()})
		w.Write(msg)
		return
	}
	if time.Now().After(dInfo.Etime) {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "虚拟机已到期"})
		w.Write(msg)
		return
	}
	//检查虚拟机所有者
	AStatus, _ := strconv.Atoi(req.PostFormValue("AStatus"))
	if AStatus == 0 {
		if dInfo.AStatus == 0 {
			msg, _ := json.Marshal(er{Ret: "v", Msg: "报警未开启"})
			w.Write(msg)
			return
		}
		t := make(map[string]interface{})
		t["AStatus"] = 0
		orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
		msg, _ := json.Marshal(er{Ret: "v", Msg: "报警已关闭"})
		w.Write(msg)
		return
	}
	// ACpu INT NOT NULL,
	// ABandwidth INT NOT NULL,
	// AMemory INT NOT NULL,
	// ADisk INT NOT NULL,
	ACpu, err := strconv.Atoi(req.PostFormValue("ACpu"))
	if err != nil || ACpu > 100 || ACpu < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "cpu报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	ABandwidth, err := strconv.Atoi(req.PostFormValue("ABandwidth"))
	if err != nil || ABandwidth > 100 || ABandwidth < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	AMemory, err := strconv.Atoi(req.PostFormValue("AMemory"))
	if err != nil || AMemory > 100 || AMemory < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	ADisk, err := strconv.Atoi(req.PostFormValue("ADisk"))
	if err != nil || ADisk > 100 || ADisk < 0 {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	t := make(map[string]interface{})
	t["ACpu"] = ACpu
	t["ABandwidth"] = ABandwidth
	t["AMemory"] = AMemory
	t["ADisk"] = ADisk
	t["AStatus"] = 1
	_, err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "设置失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "设置成功"})
	w.Write(msg)
}

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
		msg, _ := json.Marshal(er{Ret: "e", Msg: "内存大小必须为整数"})
		w.Write(msg)
		return
	}
	vcpu, err := strconv.Atoi(req.PostFormValue("vcpu"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "cpu个数必须为整数"})
		w.Write(msg)
		return
	}
	vpasswd := req.PostFormValue("vpasswd")
	if vpasswd == "" {
		vpasswd = string(rpwd.Init(16, true, true, true, false))
	}
	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "带宽必须位整数"})
		w.Write(msg)
		return
	}
	sys := req.PostFormValue("sys")
	if sys == "" {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "镜像必填"})
		w.Write(msg)
		return
	}
	var vInfo table.Virtual
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
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机硬盘失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	xml := createKvmXML(vInfo)
	connect := control.Connect()
	defer connect.Close()
	_, err = connect.DomainDefineXML(xml)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "创建虚拟机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	err = orm.SetTable("Virtual").SetPK("ID").Save(&vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	loop.VmInit <- vInfo.Vname
	msg, _ := json.Marshal(er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
	w.Write(msg)
}

func undefine(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	orm, err := control.Bdb()
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	Vname := req.PostFormValue("Vname")
	var tmpVirtual table.Virtual
	err = orm.SetTable("Virtual").SetPK("ID").Where("Uid = ? and Vname = ?", uid, Vname).Find(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}

	defer req.Body.Close()
	disk := fmt.Sprintf("/virt/disk/%s.qcow2", Vname)
	os.Remove(disk)
	t := make(map[string]interface{})
	t["Status"] = 0
	_, err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "删除失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Undefine(Vname)
	if err != nil {
		cLog.Error(err.Error())
		msg, _ := json.Marshal(er{Ret: "e", Msg: "销毁失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(er{Ret: "v", Msg: "已删除"})
	w.Write(msg)
}

func createKvmXML(tvm table.Virtual) string {
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
			<interface type='network'>
				<mac address='` + tvm.Mac + `'/>
				<source network='lan'/>
				<bandwidth>
					<inbound average='` + fmt.Sprintf("%d", tvm.Bandwidth*125) + `' peak='` + fmt.Sprintf("%d", tvm.Bandwidth*375) + `' burst='` + fmt.Sprintf("%d", tvm.Bandwidth*128) + `'/>
					<outbound average='` + fmt.Sprintf("%d", tvm.Bandwidth*125) + `' peak='` + fmt.Sprintf("%d", tvm.Bandwidth*375) + `' burst='` + fmt.Sprintf("%d", tvm.Bandwidth*128) + `'/>
				</bandwidth>
				<target dev='lan-` + tvm.Vname + `'/>
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
				<target type='virtio' name='org.qemu.guest_agent.0' state='connected'/>
				<address type='virtio-serial' controller='0' bus='0' port='1'/>
			</channel>
		</devices>
	</domain>`
	return templateXML
}
