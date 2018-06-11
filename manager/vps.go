package manager

import (
	"cpanel/control"
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
)

func vpsList(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	if !e {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s&url=%s", "你还没有登录", fmt.Sprintf("/login.html?url=%s", req.URL.String())), http.StatusFound)
		return
	}
	o := orm.NewOrm()

	var manager table.Manager
	manager.ID = mid
	err := o.Read(&manager)
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "管理员信息查询失败"), http.StatusFound)
		return
	}
	var virtuals []table.Virtual
	_, err = o.Raw("Select * from Virtual").QueryRows(&virtuals)
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("/404.html?msg=%s", "虚拟机查询失败"), http.StatusFound)
		return
	}
	for k, v := range virtuals {
		connect := control.Connect()
		defer connect.Close()
		dom, err := connect.LookupDomainByName(v.Vname)
		if err != nil {
			cLog.Warn(err.Error())
			continue
		}
		s, _, err := dom.GetState()
		if err != nil {
			cLog.Warn(err.Error())
			continue
		}
		virtuals[k].Status = int(s)
	}
	t, _ := template.ParseFiles("html/manager/vps.html", "html/manager/public/header.html", "html/manager/public/footer.html")
	t.Execute(w, map[string]interface{}{"virtuals": virtuals, "email": manager.Email})
}

//关机
func downVps(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录", Param: "login"})
		w.Write(msg)
		return
	}
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数不全"})
		w.Write(msg)
		return
	}
	o := orm.NewOrm()
	virtual := table.Virtual{ID: id}
	err = o.Read(&virtual)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "机器不存在"})
		w.Write(msg)
		return
	}
	//检查机器运行状态
	cLog.Info("管理员：%d,关闭%s机器", mid, virtual.Vname)
	err = control.Shutdown(virtual.Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "关机失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "正在关机"})
	w.Write(msg)
}

//开机
func upVps(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	mid, e := sess.Get("mid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录", Param: "login"})
		w.Write(msg)
		return
	}
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数不全"})
		w.Write(msg)
		return
	}
	o := orm.NewOrm()
	virtual := table.Virtual{ID: id}
	err = o.Read(&virtual)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "机器不存在"})
		w.Write(msg)
		return
	}
	err = control.CheckEtime(virtual.Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "机器已到期"})
		w.Write(msg)
		return
	}
	err = control.CheckEtime(virtual.Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "服务器已到期"})
		w.Write(msg)
		return
	}
	cLog.Info("管理员：%d，开启%s机器", mid, virtual.Vname)
	err = control.Start(virtual.Vname)
	if err != nil {
		cLog.Warn(err.Error())
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "开机失败"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "正在开机"})
	w.Write(msg)
}
