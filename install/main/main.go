package main

import (
	"cpanel/table"

	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	orm.RegisterModel(new(table.Compose))
	orm.RegisterDataBase("default", "sqlite3", "/root/go/src/cpanel/db/cpanel_manager.db", 30)
	orm.RunSyncdb("default", false, true)
}
