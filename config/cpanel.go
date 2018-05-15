package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/session"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	LogFile         string
	CookieName      string
	Gclifetime      int64
	CookieLifeTime  int
	EnableSetCookie bool
}

var CLog = logs.NewLogger(1)
var CSession *session.Manager

var Yaml config

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
