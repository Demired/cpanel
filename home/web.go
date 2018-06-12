package home

import (
	"cpanel/config"
	"cpanel/control"
	"cpanel/loop"
	"cpanel/table"
	"cpanel/tools"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/Demired/rpwd"
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
)

var cLog = config.CLog

var cSession = config.CSession

// Web func
func Web() {

	homeMux := http.NewServeMux()

	homeMux.HandleFunc("/", index)
	homeMux.HandleFunc("/verify", verify)
	homeMux.HandleFunc("/edit", editAPI)
	homeMux.HandleFunc("/list", list)
	homeMux.HandleFunc("/login.html", login)
	homeMux.HandleFunc("/composes", composes)
	homeMux.HandleFunc("/login", loginAPI)
	homeMux.HandleFunc("/forget.html", forget)
	homeMux.HandleFunc("/forget", forgetAPI)
	homeMux.HandleFunc("/logout", logoutAPI)
	homeMux.HandleFunc("/userInfo.html", userInfo)
	homeMux.HandleFunc("/userInfo", userInfoAPI)
	homeMux.HandleFunc("/register.html", register)
	homeMux.HandleFunc("/register", registerAPI)
	homeMux.HandleFunc("/info.html", info)
	homeMux.HandleFunc("/setpwd", setpwd)
	homeMux.HandleFunc("/load.json", loadJSON)
	homeMux.HandleFunc("/start", start)
	homeMux.HandleFunc("/shutdown", shutdown)
	homeMux.HandleFunc("/reboot", reboot)
	homeMux.HandleFunc("/create", createAPI)
	homeMux.HandleFunc("/404.html", notFound)
	homeMux.HandleFunc("/favicon.ico", favicon)
	homeMux.HandleFunc("/repasswd.html", repasswd)
	homeMux.HandleFunc("/alarm.html", alarm)
	homeMux.HandleFunc("/alarm", alarmAPI)
	homeMux.HandleFunc("/repasswd", repasswdAPI)
	homeMux.HandleFunc("/undefine", undefine)
	homeMux.HandleFunc("/edit.html", edit)
	homeMux.HandleFunc("/create.html", create)
	homeMux.HandleFunc("/cart", cartAPI)
	homeMux.HandleFunc("/cart.html", cart)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.HomePort), homeMux)

}

// func init() {
// 	// orm.RegisterModel(new(table.Virtual))
// 	// orm.RegisterModel(new(table.Billing))
// 	// orm.RegisterModel(new(table.Prompt))
// 	// orm.RegisterModel(new(table.User))
// 	// orm.RegisterModel(new(table.Verify))
// 	// orm.RegisterModel(new(table.Watch))
// 	// orm.RegisterModel(new(table.Compose))
// 	// orm.RegisterModel(new(table.Manager))
// 	// orm.RegisterDataBase("default", "sqlite3", config.Yaml.DBPath, 30)
// }

func ref(w http.ResponseWriter, req *http.Request) {
	//设置cookie登录，注册埋点
	//跳转/home
}

func logoutAPI(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	sess.Delete("uid")
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "注销完毕"})
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
	t, _ := template.ParseFiles("html/home/create.html")
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
	t, _ := template.ParseFiles("html/home/create.html")
	t.Execute(w, vvm)
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
	t, _ := template.ParseFiles("html/home/repasswd.html")
	t.Execute(w, Vname)
	return
}

func setpwd(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	orm, _ := control.Bdb()
	code := req.PostFormValue("code")
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	var tmpVerify table.Verify
	fb := orm.SetTable("Verify").SetPK("ID").Where("Code = ? and Email = ?", code, email).Find(&tmpVerify)
	if fb != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "验证失败"), http.StatusFound)
		return
	}
	var nowTime = time.Now()
	subTime, _ := time.ParseDuration("-24h")
	lastTime := nowTime.Add(subTime)
	if lastTime.After(tmpVerify.Ctime) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "链接已过期，请通过找回密码，重新发起验证"})
		w.Write(msg)
		return
	}
	if tmpVerify.Status == 1 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "链接已作废，请通过找回密码，重新发起验证"})
		w.Write(msg)
		return
	}
	passReg := regexp.MustCompile(`^[\w!@#$%^&*()-_=+]*$`)
	if !passReg.Match([]byte(passwd)) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "passwd", Msg: "密码只支持数字，大小写字母，和\"!@#$%^&*()_-=+\""})
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
	orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Update(uData)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "重置完毕"})
	w.Write(msg)
	return
}

func forget(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/home/forget.html")
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
	t, _ := template.ParseFiles("html/home/notFound.html")
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
	o := orm.NewOrm()
	userInfo.ID = uid
	err = o.Read(&userInfo)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "用户不存在"), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/home/userInfo.html", "html/home/public/header.html", "html/home/public/footer.html")
	t.Execute(w, map[string]interface{}{"userInfo": userInfo, "userName": userInfo.Username})
}

func index(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	var userInfo table.User
	if e {
		o := orm.NewOrm()
		o.Raw("select * from user where id = ?", uid).QueryRow(&userInfo)
	}
	t, _ := template.ParseFiles("html/home/index.html", "html/home/public/header.html", "html/home/public/footer.html")
	t.Execute(w, map[string]interface{}{"userName": userInfo.Username})
}

func register(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/home/register.html")
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "系统错误，请联系管理员"})
		w.Write(msg)
		return
	}
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "未登录"})
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
			msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "username-box", Msg: "用户名只允许大小写字母和数字"})
			w.Write(msg)
			return
		}
	}
	if req.PostFormValue("tel") != "" {
		telReg := regexp.MustCompile(`^1[0-9]{10}$`)
		if !telReg.Match([]byte(req.PostFormValue("tel"))) {
			msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "tel-box", Msg: "手机号有误"})
			w.Write(msg)
			return
		}
	}
	orm, _ := control.Bdb()
	_, err = orm.SetTable("User").SetPK("ID").Where("ID = ?", uid).Update(date)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "修改失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "修改成功"})
	w.Write(msg)
}

func forgetAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	email := req.PostFormValue("email")
	if email == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "邮箱不能为空"})
		w.Write(msg)
		return
	}
	var tmpUser table.User
	orm, _ := control.Bdb()
	err := orm.SetTable("User").SetPK("ID").Where("Email = ?", email).Find(&tmpUser)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "账号不存在"})
		w.Write(msg)
		return
	}

	var nowTime = time.Now()
	subTime, _ := time.ParseDuration("-24h")
	lastTime := nowTime.Add(subTime)

	var tmpVerify []table.Verify
	err = orm.SetTable("Verify").SetPK("ID").Where("Email = ? and Ctime > ?", email, lastTime).FindAll(&tmpVerify)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Param: "email", Msg: "账号不存在"})
		w.Write(msg)
		return
	}
	if len(tmpVerify) > 5 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "找回密码太频繁"})
		w.Write(msg)
		return
	}
	var v table.Verify
	v.Email = email
	v.Code = string(rpwd.Init(16, true, true, true, false))
	v.Ctime = time.Now()
	v.Status = 0
	if tmpUser.Status == 0 {
		v.Type = "verify"
		htmlBody := fmt.Sprintf("<h1>注册验证</h1><p>点击<a href='http://172.16.1.181:8100/verify?code=%s&email=%s'>链接</a>验证注册，非本人操作请忽略</p>", v.Code, v.Email)
		tools.SendMail(email, "注册验证", htmlBody)
		//发送邮件
		//注册
	} else {
		v.Type = "forget"
		htmlBody := fmt.Sprintf("<h1>找回密码</h1><p>点击<a href='http://172.16.1.181:8100/verify?code=%s&email=%s'>链接</a>找回密码，非本人操作请忽略</p>", v.Code, v.Email)
		tools.SendMail(email, "找回密码", htmlBody)
		//找回密码
	}
	orm.SetTable("Verify").SetPK("ID").Save(&v)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "邮件已发送"})
	w.Write(msg)
	return
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	Vname := req.PostFormValue("Vname")
	var tmpVirtual table.Virtual
	o := orm.NewOrm()
	err := o.Raw("select * from virtual where uid = ? and vname = ?", uid, Vname).QueryRow(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}
	err = control.CheckEtime(Vname)
	if err != nil {
		cLog.Warn("检查到期：%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "服务器已经到期", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Start(Vname)
	if err != nil {
		cLog.Warn("虚拟机开启:%s,%s", Vname, err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "开机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "正在开机"})
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "密码已重置"})
	w.Write(msg)
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	Vname := req.PostFormValue("Vname")
	var tmpVirtual table.Virtual
	o := orm.NewOrm()
	err := o.Raw("select * from virtual where uid = ? and vname = ?", uid, Vname).QueryRow(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}
	err = control.Shutdown(Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "关机失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "正在关机"})
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	o := orm.NewOrm()
	var tmpVirtual table.Virtual
	Vname := req.PostFormValue("Vname")
	err := o.Raw("select * from virtual where uid = ? and vname = ?", uid, Vname).QueryRow(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "权限不足"})
		w.Write(msg)
		return
	}
	err = control.Reboot(Vname)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "重启失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "正在重启"})
	w.Write(msg)
}

func editAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/create.html", http.StatusFound)
		return
	}
	var sourceVirtual table.Virtual
	vname := req.PostFormValue("vname")
	o := orm.NewOrm()
	_, err := o.Raw("Select * from Virtual where vname = ?", vname).QueryRows(&sourceVirtual)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "机器不存在"})
		w.Write(msg)
		return
	}
	if time.Now().After(sourceVirtual.Etime) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "机器已经到期"})
		w.Write(msg)
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

	bandwidth, err := strconv.Atoi(req.PostFormValue("bandwidth"))
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "带宽必须位整数"})
		w.Write(msg)
		return
	}

	if sourceVirtual.Vmemory != vmemory {

	}

	sys := req.PostFormValue("sys")
	if sys == "" {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "镜像必填"})
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
	_, err = o.Insert(&vInfo)
	if err != nil {
		cLog.Info(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "写入失败", Data: err.Error()})
		//手动回滚
		w.Write(msg)
		return
	}
	loop.VmInit <- vInfo.Vname
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: fmt.Sprintf("你的虚拟机密码是：%s", vInfo.Passwd)})
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
	// orm, err := control.Bdb()
	// if err != nil {
	// 	cLog.Warn(err.Error())
	// 	http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
	// 	return
	// }
	o := orm.NewOrm()
	var dInfo table.Virtual
	// err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ? and Uid = ?", Vname, uid).Find(&dInfo)
	err := o.Raw("select * from virtual where Vname = ? and Uid = ?", Vname, uid).QueryRow(&dInfo)
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
	t, _ := template.ParseFiles("html/home/alarm.html")
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}

	var dInfo table.Virtual
	Vname := req.PostFormValue("Vname")
	// err = orm.SetTable("Virtual").SetPK("ID").Where("Vname = ? and Uid = ?", Vname, uid).Find(&dInfo)
	o := orm.NewOrm()
	err := o.Raw("select * from virtual where Vname = ? and Uid = ?", Vname, uid).QueryRow(&dInfo)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "没有权限", Data: err.Error()})
		w.Write(msg)
		return
	}
	if time.Now().After(dInfo.Etime) {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "虚拟机已到期"})
		w.Write(msg)
		return
	}
	//检查虚拟机所有者
	AStatus, _ := strconv.Atoi(req.PostFormValue("AStatus"))
	if AStatus == 0 {
		if dInfo.AStatus == 0 {
			msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "报警未开启"})
			w.Write(msg)
			return
		}
		// t := make(map[string]interface{})
		// t["AStatus"] = 0
		dInfo.AStatus = 0
		// orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", Vname).Update(t)
		_, err := o.Update(&dInfo)
		if err != nil {
			msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "关闭失败"})
			w.Write(msg)
			return
		}
		msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "报警已关闭"})
		w.Write(msg)
		return
	}
	// ACpu INT NOT NULL,
	// ABandwidth INT NOT NULL,
	// AMemory INT NOT NULL,
	// ADisk INT NOT NULL,
	ACpu, err := strconv.Atoi(req.PostFormValue("ACpu"))
	if err != nil || ACpu > 100 || ACpu < 0 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "cpu报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	ABandwidth, err := strconv.Atoi(req.PostFormValue("ABandwidth"))
	if err != nil || ABandwidth > 100 || ABandwidth < 0 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "带宽报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	AMemory, err := strconv.Atoi(req.PostFormValue("AMemory"))
	if err != nil || AMemory > 100 || AMemory < 0 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "内存报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}
	ADisk, err := strconv.Atoi(req.PostFormValue("ADisk"))
	if err != nil || ADisk > 100 || ADisk < 0 {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "内存报警百分比阀值必须为整数"})
		w.Write(msg)
		return
	}

	dInfo.ACpu = ACpu
	dInfo.ABandwidth = ABandwidth
	dInfo.AMemory = AMemory
	dInfo.ADisk = ADisk
	dInfo.AStatus = 1
	_, err = o.Update(&dInfo)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "设置失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "设置成功"})
	w.Write(msg)
}

func undefine(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	Vname := req.PostFormValue("Vname")
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "请登录", Param: "login"})
		w.Write(msg)
		return
	}
	o := orm.NewOrm()
	var tmpVirtual table.Virtual
	err := o.Raw("select * from virtual where uid = ? and vname = ?", uid, Vname).QueryRow(&tmpVirtual)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "权限不足", Data: err.Error()})
		w.Write(msg)
		return
	}
	disk := fmt.Sprintf("/virt/disk/%s.qcow2", Vname)
	os.Remove(disk)
	tmpVirtual.Status = 0
	_, err = o.Update(&tmpVirtual)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "删除失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	err = control.Undefine(Vname)
	if err != nil {
		cLog.Error(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "销毁失败", Data: err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "已删除"})
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
