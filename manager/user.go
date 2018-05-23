package manager

import (
	"cpanel/table"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/orm"
)

//用户列表模板
func userList(w http.ResponseWriter, req *http.Request) {
	var managers []table.Manager
	o := orm.NewOrm()
	_, err := o.Raw("Select * form manager where status = ?", "1").QueryRows(&managers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(managers)
}
