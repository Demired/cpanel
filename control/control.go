package control

import (
	"cpanel/control"
	"encoding/json"
	"fmt"

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

func SetPsswd(vname string, username string, passwd string) error {
	encryptPasswd, err := crypt.Crypt(passwd, "$6$Pk3YRrQamkzbN6wY")
	if err != nil {
		return err
	}
	dom, err := control.Connect().LookupDomainByName(vname)
	s, _, err := dom.GetState()
	if int(s) == 1 {
		t := fmt.Sprintf("vm:%s,passwd:%s", vname, encryptPasswd)
		fmt.Println(t)
		err = dom.SetUserPassword(username, encryptPasswd, libvirt.DOMAIN_PASSWORD_ENCRYPTED)
		if err != nil {
			return err
		}
		msg, err := json.Marshal(er{Ret: "v", Msg: "密码修改成功"})
		if err != nil {
			return err
		}
		return nil
	}
	return errors.new("vps not run")
}
