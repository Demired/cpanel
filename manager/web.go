package manager

import (
	"cpanel/config"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"rpwd"
)

var cLog = config.CLog

var cSession = config.CSession

func Web() {
	http.HandleFunc("/login.html", login)
	http.HandleFunc("/login", loginAPI)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), nil)
}

func login(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	token := string(rpwd.Init(16, true, true, true, false))
	sess.Set("loginToken", token)
	t, _ := template.ParseFiles("html/managre/login.html")
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
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码不能为空"})
		w.Write(msg)
		return
	}
}
