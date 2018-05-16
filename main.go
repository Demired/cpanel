package main

import (
	"cpanel/loop"
	"cpanel/manager"
)

func main() {
	go loop.Watch()
	go loop.WorkQueue()
	// go home.Web()
	manager.Web()
}
