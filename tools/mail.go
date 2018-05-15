package tools

import (
	"cpanel/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dm"
)

var cLog = config.CLog

func SendMail(address, subject, htmlBody string) {
	dmClient, err := dm.NewClientWithAccessKey(config.Yaml.RegionID, config.Yaml.AccessKeyID, config.Yaml.AccessKeySecret)
	if err != nil {
		panic(err)
	}
	request := dm.CreateSingleSendMailRequest()
	request.ToAddress = address
	request.AccountName = config.Yaml.AccountName
	request.ReplyAddress = config.Yaml.ReplyAddress
	request.Subject = subject
	request.ReplyToAddress = "false"
	request.AddressType = "0"
	request.FromAlias = config.Yaml.Alias
	request.HtmlBody = htmlBody
	request.Domain = config.Yaml.Domain
	_, err = dmClient.SingleSendMail(request)
	if err != nil {
		cLog.Warn(err.Error())
	}
}
