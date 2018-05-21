package manager

import (
	"cpanel/config"
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/Demired/rpwd"
	"github.com/astaxie/beego/orm"
)

var cLog = config.CLog

var cSession = config.CSession

func Web() {

	homeMux := http.NewServeMux()

	homeMux.HandleFunc("/login.html", login)
	homeMux.HandleFunc("/login", loginAPI)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)

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

	// orm.RegisterModel(new(table.Manager))

	orm.RegisterDataBase("default", "sqlite3", "./db/cpanel_manager.db", 30)

	// orm.RunSyncdb("default", false, true)

	o := orm.NewOrm()

	var manager table.Manager

	err := o.Raw("select * form Manager where Email = ?", email).QueryRow(&manager)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(manager)

	if manager.Passwd == passwd {
		fmt.Println("login ok")
	}
	// fmt.Println(email)
	// fmt.Println(passwd)
}
