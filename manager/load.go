package manager

import (
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
	}
}
