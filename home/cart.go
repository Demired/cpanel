package home

import (
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
)

// CartAPI func
func cartAPI(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PostFormValue("id"))
	action := req.PostFormValue("action")
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "参数有误"})
		w.Write(msg)
		return
	}
	//检查是否存在套餐
	o := orm.NewOrm()
	var compose table.Compose
	compose.ID = id
	err = o.Read(&compose)
	if err != nil {
		cLog.Info("套餐不存在，%d", id)
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "套餐不存在"})
		w.Write(msg)
		return
	}
	//检查登录
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	uid, e := sess.Get("uid").(int)
	if !e {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "你还没有登录", Param: "login"})
		w.Write(msg)
		return
	}
	//检查是否已经有这个商品
	//已经有了，就增加数量
	//没有就新增
	var cart table.Cart
	err = o.Raw("select * from cart where status = 1 and cid = ? and uid = ?", id, uid).QueryRow(&cart)
	if err != nil {
		if action == "minus" {
			msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "没找到"})
			w.Write(msg)
			return
		}
		cLog.Info("没找了")
		cart.CID = id
		cart.UID = uid
		cart.Status = 1
		cart.Num = 1
		_, err := o.Insert(&cart)
		if err != nil {
			cLog.Info("加入失败，%d", id)
			msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "加入失败"})
			w.Write(msg)
			return
		}
	} else {
		if action == "minus" {
			if cart.Num == 1 {
				//删除这个条记录
				cart.Status = 0
			}
			cart.Num--
		} else {
			cart.Num++
		}
		_, err := o.Update(&cart)
		if err != nil {
			msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "操作失败"})
			w.Write(msg)
			return
		}
	}
	var cartComposes []table.CartCompose
	//查总价
	o.Raw("select compose.id,cart.num,compose.price,compose.name from cart,compose where cart.status = 1 and cart.uid = ? and cart.cid = compose.id", uid).QueryRows(&cartComposes)
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "操作成功", Data: cartComposes})
	w.Write(msg)
}

//Cart func
func cart() {
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
	t, _ := template.ParseFiles("html/home/cart.html", "html/home/public/header.html", "html/home/public/footer.html")
	var total = 0
	for _, v := range cartComposes {
		total += v.Price * v.Num
	}
	t.Execute(w, map[string]interface{}{"carts": cartComposes, "userName": userInfo.Username, "total": total})
}
