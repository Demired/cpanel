package arp

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type I struct {
	IP  string
	MAC string
}

func Get() []I {
	buf, err := ioutil.ReadFile("/proc/net/arp")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	k := strings.Split(string(buf), "\n")

	var mps []I
	for index, value := range k {
		if index == 0 || len(value) < 20 {
			continue
		}
		reg := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+).*(\w{2}:\w{2}:\w{2}:\w{2}:\w{2}:\w{2})`).FindAllStringSubmatch(value, -1)

		if len(reg) < 1 {
			continue
		}

		mps = append(mps, I{IP: reg[0][1], MAC: reg[0][2]})
	}
	return mps
	// fmt.Println(mps)
}
