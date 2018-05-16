package main

import (
	"cpanel/home"
	"cpanel/loop"
	"cpanel/manager"
)

func main() {
	go loop.Watch()
	go loop.WorkQueue()
	go home.Web()
	go manager.Web()
}
