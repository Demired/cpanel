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
	var cartComposes []table.CartCompose
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if e {
		//查找用户信息
		o.Raw("select * from user where id = ?", uid).QueryRow(&userInfo)
		o.Raw("select compose.id,cart.num,compose.price,compose.name from cart,compose where cart.status = 1 and cart.uid = ? and cart.cid = compose.id", uid).QueryRows(&cartComposes)
	}
	t, _ := template.ParseFiles("html/home/compose.html", "html/home/public/header.html", "html/home/public/footer.html")
	var total = 0
	for _, v := range cartComposes {
		total += v.Price * v.Num
	}
	t.Execute(w, map[string]interface{}{"composes": composes, "carts": cartComposes, "userName": userInfo.Username, "total": total})
}
