package manager

import (
	"cpanel/config"
	"cpanel/table"
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/orm"
)

var cLog = config.CLog

var cSession = config.CSession

// Web func is manager entry
func Web() {
	homeMux := http.NewServeMux()
	homeMux.HandleFunc("/", index)
	homeMux.HandleFunc("/vps", vpsList)
	homeMux.HandleFunc("/init", initDB)
	homeMux.HandleFunc("/login", loginAPI)
	homeMux.HandleFunc("/logout", logout)
	homeMux.HandleFunc("/compose", compose)
	homeMux.HandleFunc("/userList", userList)
	homeMux.HandleFunc("/login.html", login)
	homeMux.HandleFunc("/index.html", index)
	homeMux.HandleFunc("/404.html", notFound)
	homeMux.HandleFunc("/addCompose", addCompose)
	homeMux.HandleFunc("/editCompose", editCompose)
	homeMux.HandleFunc("/addComposeInfo", addComposeInfo)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)
}

func init() {
	orm.RegisterModel(new(table.Virtual))
	// orm.RegisterModel(new(table.Billing))
	// orm.RegisterModel(new(table.Prompt))
	// orm.RegisterModel(new(table.User))
	// orm.RegisterModel(new(table.Verify))
	// orm.RegisterModel(new(table.Watch))
	// orm.RegisterModel(new(table.Compose))
	// orm.RegisterModel(new(table.Manager))
	orm.RegisterDataBase("default", "sqlite3", config.Yaml.DBPath, 30)
}

// index web template
func index(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("html/manager/index.html")
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, _ := sess.Get("uid").(int)
	t.Execute(w, map[string]int{"uid": uid})
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
