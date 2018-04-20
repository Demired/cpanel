package table

import (
	"time"
)

type Virtual struct {
	ID        int       `json:"id"`
	IPv4      string    `json:"ipv4"`
	IPv6      string    `json:"ipv6"`
	LocalIP   string    `json:"local"`
	Vcpu      int       `json:"vcpu"`
	Status    int       `json:"status"`
	Vmemory   int       `json:"vmemory"` //GiB
	Passwd    string    `json:"vpasswd"`
	Vname     string    `json:"vname"`
	Br        string    `json:"br"`
	Mac       string    `json:"mac"`
	Bandwidth int       `json:"bandwidth"` //Mbps
	Etime     time.Time `json:"etime"`     //Expire time
	Ctime     time.Time `json:"ctime"`
	Utime     time.Time `json:"utime"`
}

type Watch struct {
	ID     int
	Vname  string
	CPU    int
	Memory int
	Ctime  int
}
