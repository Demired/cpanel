package manager

import (
	"cpanel/table"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/orm"
)

//用户列表模板
func userList(w http.ResponseWriter, req *http.Request) {
	var users []table.User
	o := orm.NewOrm()
	_, err := o.Raw("Select * from user").QueryRows(&users)
	if err != nil {
		cLog.Warn(err.Error())
		return
	}
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid := sess.Get("min").(int)
	t, _ := template.ParseFiles("html/manager/userList.html")
	t.Execute(w, map[string]interface{}{"users": users, "mid": mid})
}
