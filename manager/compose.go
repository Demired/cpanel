package manager

import (
	"cpanel/table"
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/orm"
)

//套餐列表 套餐模板
func compose(w http.ResponseWriter, req *http.Request) {
	var composes []table.Compose
	o := orm.NewOrm()
	res, err := o.Raw("Select * from compose where status = ?", "1").QueryRows(&composes)
	// err := o.Read(&compose)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(res)
	fmt.Println(composes)
	t, _ := template.ParseFiles("html/manager/compose.html")
	t.Execute(w, nil)
}

//套餐列表
func composes(w http.ResponseWriter, req *http.Request) {
	o := orm.NewOrm()
	o.Raw("select * from Composes where status = 1")
	fmt.Println("123")
}
