package control

import (
	"fmt"

	libvirt "github.com/libvirt/libvirt-go"
)

func Connect() *libvirt.Connect {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return conn
}
