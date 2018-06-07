package table

import (
	"time"
)

// Virtual struct
type Virtual struct {
	ID         int       `orm:"column(id);auto" json:"id"`
	UID        int       `orm:"column(uid)" json:"uid"`
	IPv4       string    `orm:"size(30);column(ipv4)" json:"ipv4"`
	IPv6       string    `orm:"size(50);column(ipv6)" json:"ipv6"`
	LocalIP    string    `orm:"size(30);column(localIP)" json:"local"`
	Vcpu       int       `orm:"column(vcpu)" json:"vcpu"`
	Status     int       `json:"status"`
	Vmemory    int       `json:"vmemory"` //GiB
	Passwd     string    `json:"vpasswd"`
	Vname      string    `json:"vname"`
	Tag        string    `json:"tag"`
	Br         string    `json:"br"`
	Mac        string    `json:"mac"`
	Sys        string    `json:"sys"`
	Bandwidth  int       `json:"bandwidth"`                //Mbps
	Etime      time.Time `orm:"auto_now_add" json:"etime"` //Expire time
	Ctime      time.Time `orm:"auto_now_add" json:"ctime"`
	Utime      time.Time `orm:"auto_now_add" json:"utime"`
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
	ID     int    `orm:"column(id);auto"`
	Vname  string `orm:"size(50)"`
	Up     int    //上行流量
	Down   int    //下行流量
	Read   int    //硬盘读取
	Write  int    //硬盘写入
	CPU    int    //cpu时间片
	Memory int    //内存占用
	Ctime  int    //时间戳
}

// User struct
type User struct {
	ID       int    `orm:"column(id);auto"`
	Username string `orm:"size(50)"`
	Passwd   string `orm:"size(50)"`
	Tel      string `orm:"size(20)"`
	Email    string `orm:"size(320)"`
	Realname string `orm:"size(50)"`
	Idnumber string `orm:"size(50)"`
	Idtype   int    //证件类型 0 身份证 1 军官证
	Sex      int
	Address  string `orm:"size(50)"` //地址
	Company  string `orm:"size(30)"` //公司
	City     string `orm:"size(20)"` //城市
	Status   int
	Utime    time.Time `orm:"auto_now_add"`
	Ctime    time.Time `orm:"auto_now_add;type(datetime)"`
}

// Verify struct
type Verify struct {
	ID     int    `orm:"column(id);auto"`
	Email  string `orm:"size(320)"`
	Code   string `orm:"size(50)"`
	Type   string `orm:"size(10)"`
	Status int
	Ctime  time.Time `orm:"auto_now_add;type(datetime)"`
	Vtime  time.Time `orm:"auto_now_add"`
}

// Billing strcut
type Billing struct {
	ID    int       `orm:"column(id);auto"`
	UID   int       `orm:"column(uid)"`
	Value int       `orm:"column(value)"`
	Model string    `orm:"size(50)"`
	Desc  string    `orm:"size(50)"`
	Ctime time.Time `orm:"auto_now_add;type(datetime)"`
}

// Manager strcut
type Manager struct {
	ID     int       `orm:"column(id);auto"`
	Email  string    `orm:"size(320)"`
	Passwd string    `orm:"size(50)"`
	Status int       `orm:"column(status)"`
	Ctime  time.Time `orm:"auto_now_add;type(datetime)"`
	Utime  time.Time `orm:"auto_now"`
}

// Compose strcut
// 配置组合表
type Compose struct {
	ID        int `orm:"column(id);auto"`
	UID       int `orm:"column(uid)"`
	Name      string
	IPv4      int
	IPv6      int
	Vcpu      int
	Bandwidth int
	Vmemory   int
	TotalFlow int
	Price     int
	Status    int
	Ctime     time.Time `orm:"auto_now;type(datetime)"`
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

//Prompt strcut
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

//购物车

//Cart struct
type Cart struct {
	ID     int `orm:"column(id);auto"`
	UID    int
	CID    int
	Num    int
	Status int
	Ctime  time.Time `orm:"auto_now_add;type(datetime)"`
}
