package table

import (
	"time"
)

type Virtual struct {
	ID         int       `json:"id"`
	UID        int       `json:"uid"`
	IPv4       string    `json:"ipv4"`
	IPv6       string    `json:"ipv6"`
	LocalIP    string    `json:"local"`
	Vcpu       int       `json:"vcpu"`
	Status     int       `json:"status"`
	Vmemory    int       `json:"vmemory"` //GiB
	Passwd     string    `json:"vpasswd"`
	Vname      string    `json:"vname"`
	Tag        string    `json:"tag"`
	Br         string    `json:"br"`
	Mac        string    `json:"mac"`
	Sys        string    `json:"sys"`
	Bandwidth  int       `json:"bandwidth"` //Mbps
	Etime      time.Time `json:"etime"`     //Expire time
	Ctime      time.Time `json:"ctime"`
	Utime      time.Time `json:"utime"`
	ACpu       string    `json:"acpu"`
	ABandwidth string    `json:"abandwidth"`
	AMemory    string    `json:"amemory"`
	ADisk      string    `json:"adisk"`
	AStatus    string    `json:"astatus"`
}

type Watch struct {
	ID     int
	Vname  string
	Up     int
	Down   int
	Read   int
	Write  int
	CPU    int
	Memory int
	Ctime  int
}
