package manager

import (
	"cpanel/config"
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

var cLog = config.CLog

var cSession = config.CSession

// Web func is manager entry
func Web() {
	homeMux := http.NewServeMux()
	homeMux.HandleFunc("/login.html", login)
	homeMux.HandleFunc("/login", loginAPI)
	homeMux.HandleFunc("/compose", compose)
	homeMux.HandleFunc("/composes", composes)
	homeMux.HandleFunc("/index.html", index)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)
}

func init() {
	orm.RegisterModel(new(table.Compose))
	orm.RegisterDataBase("default", "sqlite3", "./db/cpanel_manager.db", 30)
}

func compose(w http.ResponseWriter, req *http.Request) {
	var composes []table.Compose
	o := orm.NewOrm()
	res, err := o.Raw("Select * form compose where status = ?", "1").QueryRows(&composes)
	// err := o.Read(&compose)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(res)
	fmt.Println(composes)
	t, _ := template.ParseFiles("html/manager/compose.html")
	t.Execute(w, nil)
}

//套餐列表
func composes(w http.ResponseWriter, req *http.Request) {
	o := orm.NewOrm()
	o.Raw("select * from Composes where status = 1")
	fmt.Println("123")
}

// index web template
func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/manager/index.html")
	t.Execute(w, nil)
}

//login web template
func login(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	token := string(rpwd.Init(16, true, true, true, false))
	sess.Set("loginToken", token)
	t, _ := template.ParseFiles("html/manager/login.html")
	t.Execute(w, map[string]string{"token": token})
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
	fmt.Println(manager.Passwd == sha1passwd)
	if manager.Passwd != sha1passwd {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	sess.Set("uid", manager.ID)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "登录成功"})
	w.Write(msg)
}

//vm func
//输出所有虚拟机
func vm() {
	//
}

//翻页
func vmList() {

}

//创建虚拟机套餐
func createVMType() {

}

//管理用户列表

//禁用客户

//查看ret状况

//查看财务状况
