package manager

import (
	"cpanel/table"
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/orm"
)

func vpsList(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		//TODO跳转登录页面
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var composes []table.Virtual
	o := orm.NewOrm()
	_, err := o.Raw("Select * from Virtual where status = ?", "1").QueryRows(&composes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t, _ := template.ParseFiles("html/manager/compose.html")
	t.Execute(w, map[string]interface{}{"composes": composes, "uid": uid})
}