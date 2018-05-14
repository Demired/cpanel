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
	AutoPay    int       `json:"autopay"`
	ACpu       int       `json:"acpu"`
	ABandwidth int       `json:"abandwidth"`
	AMemory    int       `json:"amemory"`
	ADisk      int       `json:"adisk"`
	AStatus    int       `json:"astatus"`
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

type User struct {
	ID       int
	Username string
	Passwd   string
	Tel      string
	Email    string
	Realname string
	Idnumber string
	Idtype   int
	Sex      int
	Address  string
	Company  string
	City     string
	Status   int
	Utime    time.Time
	Ctime    time.Time
}

type Verify struct {
	ID     int
	Email  string
	Code   string
	Type   string
	Status int
	Ctime  time.Time
	Vtime  time.Time
}

type Billing struct {
	ID    int
	UID   int
	Value int
	Model string
	Desc  string
	Ctime time.Time
}

// ID INTEGER PRIMARY KEY AUTOINCREMENT,
// Username CHAR(20) NOT NULL,
// Passwd CHAR(50) NOT NULL,
// Tel CHAR(20) NOT NULL,
// Email CHAR(20) NOT NULL,
// Realname CHAR(20) NOT NULL,
// Idnumber CHAR(20) NOT NULL,
// Sex INT NOT NULL,
// Company CHAR(20) NOT NULL,
// City CHAR(20) NOT NULL,
// Status INT NOT NULL,
// Utime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
// Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
