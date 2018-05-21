package main

import (
	"cpanel/table"

	"github.com/astaxie/beego/orm"
)

func main() {
	orm.RegisterModel(new(table.Manager))
	orm.RegisterModel(new(table.Compose))
	orm.RegisterDataBase("default", "sqlite3", "./db/cpanel_manager.db", 30)
	orm.RunSyncdb("default", false, true)
}
