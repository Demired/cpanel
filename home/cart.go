package home

import (
	"cpanel/table"
	"cpanel/tools"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
)

//购物车
func cart(w http.ResponseWriter, req *http.Request) {

}

//必须登录
func addCart(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PostFormValue("id"))
	del := req.PostFormValue("del")
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
		if del == "del" {
			msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "没找到"})
			w.Write(msg)
		}
		cLog.Info("没找了")
		cart.CID = id
		cart.UID = uid
		_, err := o.Insert(&cart)
		if err != nil {
			cLog.Info("加入失败，%d", id)
			msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "加入失败"})
			w.Write(msg)
			return
		}
	} else {
		if del == "del" {
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
	msg, _ := json.Marshal(tools.Er{Ret: "v", Msg: "操作成功"})
	w.Write(msg)
}
