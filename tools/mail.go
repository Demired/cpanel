package tools

import (
	"cpanel/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dm"
)

var cLog = config.CLog

func SendMail(address, subject, htmlBody string) {
	dmClient, err := dm.NewClientWithAccessKey("cn-hangzhou", "LTAIG47RA2EJ06qP", "jmEqG2mrGxXNpGSkTB6lWYY9xdUnfN")
	if err != nil {
		panic(err)
	}
	request := dm.CreateSingleSendMailRequest()
	request.ToAddress = address
	request.AccountName = "send@mail.0x8c.com"
	request.ReplyAddress = "zhangyuan8087@gmail.com"
	request.Subject = subject
	request.ReplyToAddress = "false"
	request.AddressType = "0"
	request.FromAlias = "cpanl"
	request.HtmlBody = htmlBody
	//"<h1>注册验证</h1><p>点击<a href='https://t.0x8c.com/qwertyuiop'>链接</a>验证注册，非本人操作请忽略</p>"
	request.Domain = "dm.aliyuncs.com"
	_, err = dmClient.SingleSendMail(request)
	if err != nil {
		cLog.Warn(err.Error())
	}
}
