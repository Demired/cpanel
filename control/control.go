package control

import (
	"fmt"

	libvirt "github.com/libvirt/libvirt-go"
)

func Start(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	return dom.Create()
}

func Connect() *libvirt.Connect {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return conn
}

func Shutdown(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	return dom.Shutdown()
}

func Reboot(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	return dom.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT)
}
