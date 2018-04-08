package control

import (
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

func SetPasswd(vname string,userName string passwd string) error {
	encryptPasswd, err := crypt.Crypt(passwd, "$6$Pk3YRrQamkzbN6wY")
	if err != nil {
		return err
	}
	dom, err := control.Connect().LookupDomainByName(vname)
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

// func SetPsswd(vname string, username string, passwd string) error {
// 	encryptPasswd, err := crypt.Crypt(passwd, "$6$Pk3YRrQamkzbN6wY")
// 	if err != nil {
// 		return err
// 	}
// 	dom, err := control.Connect().LookupDomainByName(vname)
// 	if err != nil {
// 		return err
// 	}
// 	s, _, err := dom.GetState()
// 	if int(s) == 1 {
// 		err = dom.SetUserPassword(username, encryptPasswd, libvirt.DOMAIN_PASSWORD_ENCRYPTED)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}
// 	return errors.new("vps not run")
// }
