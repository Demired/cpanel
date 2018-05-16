package main

import (
	"cpanel/config"
	"cpanel/home"
	"cpanel/loop"

	_ "github.com/mattn/go-sqlite3"
)


func main() {
	go loop.Watch()
	go loop.WorkQueue()
	go home.Web()
	go manger.Web()
}
