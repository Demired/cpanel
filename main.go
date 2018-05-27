package main

import (
	"cpanel/config"
	"cpanel/home"
	"cpanel/loop"
	"cpanel/manager"
	"cpanel/table"
	"cpanel/tools"
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--init" {
		err := InstallDB()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("init complete")
		}
		return
	}
	//检查是否经过配置
	go home.Web()
	go manager.Web()
	go loop.WorkQueue()
	go loop.Watch()

	time.Sleep(1 * time.Hour)
}

// InstallDB 安装数据库
func InstallDB() error {
	err := orm.RunSyncdb("default", false, false)
	if err != nil {
		return err
	}
	o := orm.NewOrm()
	var manager table.Manager
	manager.Email = config.Yaml.ManagerEmail
	manager.Passwd = tools.SumSha1(config.Yaml.ManagerPasswd)
	_, err = o.Insert(&manager)
	if err != nil {
		fmt.Println("insert ok")
		return err
	}
	return nil
}

func init() {
	orm.RegisterModel(new(table.Virtual))
	orm.RegisterModel(new(table.Billing))
	orm.RegisterModel(new(table.Prompt))
	orm.RegisterModel(new(table.User))
	orm.RegisterModel(new(table.Verify))
	orm.RegisterModel(new(table.Watch))
	orm.RegisterModel(new(table.Compose))
	orm.RegisterModel(new(table.Manager))
	orm.RegisterDataBase("default", "sqlite3", config.Yaml.DBPath, 30)
}
