package tools

import (
	"crypto/sha1"
	"fmt"
)

// Er struct
type Er struct {
	Ret   string `json:"ret"`
	Msg   string `json:"msg"`
	Data  string `json:"data"`
	Param string `json:"param"`
}

// SumSha1 str
func SumSha1(passwd string) string {
	h := sha1.New()
	h.Write([]byte(passwd))
	sha1PasswdByte := h.Sum(nil)
	return fmt.Sprintf("%x", sha1PasswdByte)
}
