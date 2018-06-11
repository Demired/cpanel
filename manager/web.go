package manager

import (
	"cpanel/config"
	"cpanel/table"
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/orm"

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
	homeMux.HandleFunc("/downCompose", downCompose)
	homeMux.HandleFunc("/upCompose", upCompose)
	homeMux.HandleFunc("/deleteCompose", deleteCompose)
	homeMux.HandleFunc("/editCompose", editCompose)
	homeMux.HandleFunc("/addComposeInfo", addComposeInfo)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)
}

// index web template
func index(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	t, _ := template.ParseFiles("html/manager/index.html", "html/manager/public/header.html", "html/manager/public/footer.html")
	var manager table.Manager
	if e {
		o := orm.NewOrm()
		err := o.Raw("select * from manager where id = ?", mid).QueryRow(&manager)
		if err != nil {
			cLog.Warn("管理员信息查询失败%s", err.Error())
			http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "管理员信息查询失败"), http.StatusFound)
			return
		}
	}
	t.Execute(w, map[string]string{"email": manager.Email})
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
