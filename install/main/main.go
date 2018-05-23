package main

import (
	"cpanel/config"
	"cpanel/table"

	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	orm.RegisterModel(new(table.Compose))
	orm.RegisterDataBase("default", "sqlite3", config.Yaml.DBPath, 30)
	orm.RunSyncdb("default", false, true)
}
