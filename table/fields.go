package table

import (
	"time"
)

// Virtual struct
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
	Cycle      int       `json:"cycle"`
	ACpu       int       `json:"acpu"`
	ABandwidth int       `json:"abandwidth"`
	AMemory    int       `json:"amemory"`
	ADisk      int       `json:"adisk"`
	AStatus    int       `json:"astatus"`
}

// Watch struct
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

// User struct
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

// Verify struct
type Verify struct {
	ID     int
	Email  string
	Code   string
	Type   string
	Status int
	Ctime  time.Time
	Vtime  time.Time
}

// Billing strcut
type Billing struct {
	ID    int
	UID   int
	Value int //金额 单位：分
	// Model string
	Desc  string //描述
	Ctime time.Time
}

// Manager strcut
type Manager struct {
	ID      int       `orm:"auto"`
	Email   string    `orm:"size(320)"`
	Passwd  string    `orm:"size(50)"`
	Ctime   time.Time `orm:"auto_now_add;type(datetime)"`
	Updated time.Time `orm:"auto_now"`
}

// Compose strcut
type Compose struct {
	ID    int       `orm:"auto"`
	IPv4  int       `orm:"int8"`
	IPv6  int       `orm:"int8"`
	Vcpu  int       `orm:"int8"`
	Ctime time.Time `orm:"auto_now"`
}

// CREATE TABLE IF NOT EXISTS Manager (
// 	ID INTEGER PRIMARY KEY AUTOINCREMENT,
// 	Email CHAR(320) NOT NULL,
// 	Passwd CHAR(50) NOT NULL,
// 	Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
// );

// CREATE TABLE IF NOT EXISTS Prompt (
// 	ID INTEGER PRIMARY KEY AUTOINCREMENT,
// 	Vname CHAR(20) NOT NULL,
// 	Email CHAR(20) NOT NULL,
// 	Type CHAR(20) NOT NULL,
// 	Subject CHAR(20) NOT NULL,
// 	Desc CHAR(50) NOT NULL,
// 	Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
// );

//提示盒子
type Prompt struct {
	ID      int
	UID     int
	Vname   string
	Email   string
	Type    string //提示类型 余额不足 到期提醒 将要扣款通知
	Subject string //主题
	Desc    string //描述
	Ctime   time.Time
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
