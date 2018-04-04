package control

import (
	"fmt"
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
