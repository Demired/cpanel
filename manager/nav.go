package manager

import (
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
)

//导航列表
func nav(w http.ResponseWriter, req *http.Request) {
	o := orm.NewOrm()
	//判断登录
	sess, err := cSession.SessionStart(w, req)
	if err != nil {
		cLog.Warn(err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "系统错误，请联系管理员"), http.StatusFound)
		return
	}
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你没有登录", fmt.Sprintf("login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	var navs []table.Nav
	//读表
	_, err = o.Raw("select * from Nav ").QueryRows(&navs)
	if err != nil {
		cLog.Info("nav读表发生错误，错误信息：%s", err.Error())
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "发生错误"), http.StatusFound)
		return
	}
	t, _ := template.ParseFiles("html/manager/nav.html", "html/manager/public/header.html", "html/manager/public/footer.html")
	t.Execute(w, map[string]interface{}{"navs": navs, "mid": mid})
}

//添加导航标签
func saveNav(w http.ResponseWriter, req *http.Request) {
	sess, err := cSession.SessionStart(w, req)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "添加失败"})
		w.Write(msg)
		return
	}
	defer sess.SessionRelease(w)
	_, e := sess.Get("mid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你没有登录", Param: "login"})
		w.Write(msg)
		return
	}
	id := req.PostFormValue("id")
	fmt.Println(id)
	order, err := strconv.Atoi(req.PostFormValue("order"))
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数错误"})
		w.Write(msg)
		return
	}
	var name = req.PostFormValue("name")
	var url = req.PostFormValue("url")
	//TODO 检查是否符合规则

	var nav table.Nav
	nav.Name = name
	nav.Order = order
	nav.URL = url
	o := orm.NewOrm()
	_, err = o.Insert(&nav)
	if err != nil {
		cLog.Info("添加标签发生错误")
	}
}

//删除导航标签
func delNav(w http.ResponseWriter, req *http.Request) {

}
