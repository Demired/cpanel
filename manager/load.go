package manager

import (
	"cpanel/config"
	"cpanel/table"
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/orm"
)

//初始化系统
//TODO 加入初始化判断
//     加入超级管理员录入
func initDB(w http.ResponseWriter, req *http.Request) {
	err := orm.RunSyncdb("default", false, true)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var manager table.Manager
	manager.Email = config.Yaml.ManagerEmail
	h := sha1.New()
	h.Write([]byte(config.Yaml.ManagerPasswd))
	bs := h.Sum(nil)
	sha1passwd := fmt.Sprintf("%x", bs)
	manager.Passwd = sha1passwd
	o := orm.NewOrm()
	ret, err := o.Insert(&manager)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(ret)
	fmt.Println(manager)
}
