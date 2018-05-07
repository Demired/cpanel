package init

import (
	"github.com/astaxie/beego/logs"
)

var CLog = logs.NewLogger(1)

// var CSession session.Manager

var logFile = "/var/log/cpanel.log"

func init() {
	CLog.SetLogger("file", `{"filename":"`+logFile+`"}`)
	CLog.SetLevel(logs.LevelInformational)
	// CSession, _ := session.NewManager("memory", &session.ManagerConfig{CookieName: "gosession", Gclifetime: 3600})
	// go CSession.GC()
}
