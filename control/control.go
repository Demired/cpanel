package control

import (
	"fmt"

	libvirt "github.com/libvirt/libvirt-go"
)

func Start(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println("123")
	return nil
}

func Connect() *libvirt.Connect {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return conn
}
