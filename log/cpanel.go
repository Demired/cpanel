package log

import (
	"github.com/astaxie/beego/logs"
)

var CLog = logs.NewLogger(1)

var logFile = "./cpanel.log"

func init() {
	CLog.SetLogger("file", `{"filename":"`+logFile+`"}`)
	CLog.SetLevel(logs.LevelInformational)
}
