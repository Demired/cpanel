package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/session"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	LogFile         string `yaml:"LogFile"`
	CookieName      string `yaml:"CookieName"`
	Gclifetime      int64  `yaml:"Gclifetime"`
	CookieLifeTime  int    `yaml:"CookieLifeTime"`
	EnableSetCookie bool   `yaml:"EnableSetCookie"`
	RegionID        string `yaml:"RegionId"`
	Alias           string `yaml:"Alias"`
	Domain          string `yaml:"Domain"`
	AccessKeyID     string `yaml:"AccessKeyId"`
	AccessKeySecret string `yaml:"AccessKeySecret"`
	AccountName     string `yaml:"AccountName"`
	ReplyAddress    string `yaml:"ReplyAddress"`
	ManagerPort     int    `yaml:"ManagerPort"`
	HomePort        int    `yaml:"HomePort"`
}

var CLog = logs.NewLogger(1)
var CSession *session.Manager

var Yaml Config

func init() {
	//获取配置文件
	conf, err := ioutil.ReadFile("./config.yaml")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	yaml.Unmarshal(conf, &Yaml)

	CLog.SetLogger("file", `{"filename":"`+Yaml.LogFile+`"}`)
	CLog.SetLevel(logs.LevelInformational)
	CSession, _ = session.NewManager("memory",
		&session.ManagerConfig{
			CookieName:      Yaml.CookieName,
			Gclifetime:      Yaml.Gclifetime,
			CookieLifeTime:  Yaml.CookieLifeTime,
			EnableSetCookie: Yaml.EnableSetCookie,
		})

	go CSession.GC()
}
