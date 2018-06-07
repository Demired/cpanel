package manager

import (
	"cpanel/config"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var cLog = config.CLog

var cSession = config.CSession

// Web func is manager entry
func Web() {
	homeMux := http.NewServeMux()
	homeMux.HandleFunc("/", index)
	homeMux.HandleFunc("/vps", vpsList)
	homeMux.HandleFunc("/login", loginAPI)
	homeMux.HandleFunc("/logout", logout)
	homeMux.HandleFunc("/compose", compose)
	homeMux.HandleFunc("/upUser", upUser)
	homeMux.HandleFunc("/downUser", downUser)
	homeMux.HandleFunc("/userList", userList)
	homeMux.HandleFunc("/login.html", login)
	homeMux.HandleFunc("/index.html", index)
	homeMux.HandleFunc("/404.html", notFound)
	homeMux.HandleFunc("/upVps", upVps)
	homeMux.HandleFunc("/downVps", downVps)
	homeMux.HandleFunc("/addCompose", addCompose)
	homeMux.HandleFunc("/editCompose", editCompose)
	homeMux.HandleFunc("/addComposeInfo", addComposeInfo)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)
}

// index web template
func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/manager/index.html")
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, _ := sess.Get("mid").(int)
	t.Execute(w, map[string]int{"mid": mid})
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
