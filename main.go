package main

import (
	"cpanel/home"
	"cpanel/loop"
	"cpanel/manager"
	"time"
)

func main() {
	//检查是否经过配置
	go home.Web()
	go manager.Web()
	go loop.WorkQueue()
	go loop.Watch()
	time.Sleep(1 * time.Hour)
}
