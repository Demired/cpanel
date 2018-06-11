package home

import (
	"cpanel/table"
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/orm"
)

func composes(w http.ResponseWriter, req *http.Request) {
	o := orm.NewOrm()
	var composes []table.Compose
	_, err := o.Raw("select * from compose where status = 1").QueryRows(&composes)
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "查询失败"), http.StatusFound)
		return
	}
	//用户信息
	var userInfo table.User
	var carts []table.Cart
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if e {
		//查找用户信息
		o.Raw("select * from user where id = ?", uid).QueryRow(&userInfo)
		o.Raw("select * from cart where status = 1 and uid = ?", uid).QueryRows(&carts)
	}
	t, _ := template.ParseFiles("html/home/compose.html", "html/home/public/header.html", "html/home/public/footer.html")
	t.Execute(w, map[string]interface{}{"composes": composes, "carts": carts, "userName": userInfo.Username})
}
