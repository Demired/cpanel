package main

import (
	"cpanel/home"
	"cpanel/loop"
	"cpanel/manager"
)

func main() {

	//检查是否经过配置

	go loop.Watch()
	go loop.WorkQueue()
	go home.Web()
	manager.Web()
}
