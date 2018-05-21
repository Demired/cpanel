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
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)

}

func init() {
	orm.RegisterDataBase("default", "sqlite3", "./db/cpanel_manager.db", 30)
}

func login(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	token := string(rpwd.Init(16, true, true, true, false))
	sess.Set("loginToken", token)
	t, _ := template.ParseFiles("html/manager/login.html")
	t.Execute(w, map[string]string{"token": token})
}

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
	fmt.Println(sha1passwd)
	fmt.Println("------------")
	fmt.Println(manager.ID)
	fmt.Println(manager.Email)
	fmt.Println(manager.Passwd)
	fmt.Println("------------")
	fmt.Println(manager.Passwd == sha1passwd)
	if manager.Passwd != sha1passwd {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v"})
	w.Write(msg)
}
