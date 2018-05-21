package manager

import (
	"cpanel/config"
	"cpanel/table"
	"cpanel/tools"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/Demired/rpwd"
	"github.com/astaxie/beego/orm"
)

var cLog = config.CLog

var cSession = config.CSession

// Web func is manager entry
func Web() {
	homeMux := http.NewServeMux()
	homeMux.HandleFunc("/login.html", login)
	homeMux.HandleFunc("/login", loginAPI)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Yaml.ManagerPort), homeMux)

}

func init() {
	orm.RegisterDataBase("default", "sqlite3", "./db/cpanel_manager.db", 30)
}

func login(w http.ResponseWriter, req *http.Request) {
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	token := string(rpwd.Init(16, true, true, true, false))
	sess.Set("loginToken", token)
	t, _ := template.ParseFiles("html/manager/login.html")
	t.Execute(w, map[string]string{"token": token})
}

func loginAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/login.html", http.StatusFound)
	}
	email := req.PostFormValue("email")
	passwd := req.PostFormValue("passwd")
	token := req.PostFormValue("token")
	sess, _ := cSession.SessionStart(w, req)
	defer sess.SessionRelease(w)
	loginToken := sess.Get("loginToken")
	if token != loginToken {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	o := orm.NewOrm()
	var manager table.Manager
	err := o.Raw("select * from Manager where Email = ?", email).QueryRow(&manager)
	if err != nil {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "用户不存在"})
		w.Write(msg)
		return
	}
	fmt.Println(manager)

	//产生一个散列值得方式是 sha1.New()，sha1.Write(bytes)，然后 sha1.Sum([]byte{})。这里我们从一个新的散列开始。
	h := sha1.New()
	//写入要处理的字节。如果是一个字符串，需要使用[]byte(s) 来强制转换成字节数组。
	h.Write([]byte(passwd))
	//这个用来得到最终的散列值的字符切片。Sum 的参数可以用来都现有的字符切片追加额外的字节切片：一般不需要要。
	bs := h.Sum(nil)
	//SHA1 值经常以 16 进制输出，例如在 git commit 中。使用%x 来将散列结果格式化为 16 进制字符串。
	sha1passwd := fmt.Sprintf("%x\n", bs)
	fmt.Println(sha1passwd)
	if manager.Passwd != sha1passwd {
		msg, _ := json.Marshal(tools.Er{Ret: "e", Msg: "密码错误"})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(tools.Er{Ret: "v"})
	w.Write(msg)
}
