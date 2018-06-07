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
	t, _ := template.ParseFiles("html/home/compose.html")
	t.Execute(w, map[string]interface{}{"composes": composes})
}
