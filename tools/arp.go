package tools

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// type ARP struct {
// 	IP  string
// 	MAC string
// }

func Arp(mac string) string {
	buf, err := ioutil.ReadFile("/proc/net/arp")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	k := strings.Split(string(buf), "\n")

	for index, value := range k {
		if index == 0 || len(value) < 20 {
			continue
		}
		reg := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+).*(\w{2}:\w{2}:\w{2}:\w{2}:\w{2}:\w{2})`).FindAllStringSubmatch(value, -1)
		if len(reg) < 1 {
			continue
		}
		if reg[0][2] == mac {
			return reg[0][1]
		}
	}
	return ""
}
