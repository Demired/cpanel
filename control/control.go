package control

import (
	"fmt"

	"github.com/Demired/rpwd"
	"github.com/amoghe/go-crypt"
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
		return nil
	}
	return conn
}

func GetState(vname string) (int, error) {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return 0, err
	}
	s, _, _ := dom.GetState()
	return int(s), nil
}

func Shutdown(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	return dom.Shutdown()
}

func Undefine(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	return dom.Undefine()
}

func Reboot(vname string) error {
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	return dom.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT)
}

func SetPasswd(vname string, userName string, passwd string) error {
	salt := fmt.Sprintf("$6$%s", rpwd.Init(12, true, true, true, false))
	encryptPasswd, err := crypt.Crypt(passwd, salt)
	if err != nil {
		return err
	}
	dom, err := Connect().LookupDomainByName(vname)
	if err != nil {
		return err
	}
	s, _, err := dom.GetState()
	if int(s) == 1 {
		err = dom.SetUserPassword(userName, encryptPasswd, libvirt.DOMAIN_PASSWORD_ENCRYPTED)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
