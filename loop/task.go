package loop

import (
	"cpanel/config"
	"cpanel/control"
	"cpanel/table"
	"fmt"
	"strings"
	"time"

	libvirt "github.com/libvirt/libvirt-go"
)

// var InitPass = make(chan string) //设置初始密码的chan
var VmInit = make(chan string)

var Bill = make(chan string)

var Alarm = make(chan string)

var cLog = config.CLog

func Watch() {
	var t = make(map[string]uint64)
	w := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-w.C:
			connect := control.Connect()
			doms, err := connect.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
			if err != nil {
				cLog.Warn(err.Error())
				continue
			}
			orm, err := control.Bdb()
			if err != nil {
				cLog.Warn(err.Error())
				continue
			}
			for _, dom := range doms {
				name, err := dom.GetName()
				if err != nil {
					cLog.Warn(err.Error())
					continue
				}
				info, err := dom.GetInfo()
				if err != nil {
					cLog.Warn(err.Error())
					continue
				}

				var virtual table.Virtual
				if err := orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", name).Find(&virtual); err != nil {
					cLog.Warn("读取虚拟机信息失败", err.Error())
					continue
				}
				var cpurate int
				if lastCPUTime, ok := t[name]; ok {
					cpurate = int(float32((info.CpuTime-lastCPUTime)*100) / float32(20*info.NrVirtCpu*10000000))
				}
				if cpurate < 1 {
					cpurate = 1
				}
				var watch table.Watch
				watch.CPU = cpurate
				watch.Vname = name
				watch.Ctime = int(time.Now().Unix())
				watch.Memory = int(info.Memory)
				intface, err := dom.InterfaceStats(fmt.Sprintf("lan-%s", name))
				if err != nil {
					cLog.Warn(err.Error())
				} else {
					watch.Up = int(intface.TxBytes)
					watch.Down = int(intface.RxBytes)
				}
				if err = orm.SetTable("Watch").SetPK("ID").Save(&watch); err != nil {
					cLog.Warn("写入数据失败", err.Error())
					continue
				}
				//检查是否到期
				if time.Now().After(virtual.Etime) {
					err := control.Shutdown(name)
					if err != nil {
						cLog.Warn("关机失败", err.Error())
						continue
					}
					Bill <- fmt.Sprintf("%s", name)
					continue
				}
				//TODO 将要到期

				if virtual.AStatus == 1 {
					//TODO 连续3次
					//检查是否超过阀值
					if cpurate/100 > virtual.ACpu {
						Alarm <- fmt.Sprintf("cpu/%s", name)
						cLog.Warn("in alarm")
					}
				}
				t[name] = info.CpuTime
				dom.Free()
			}
			connect.Close()
		}
	}
}

func WorkQueue() {
	for {
		select {
		case vname := <-VmInit:
			fmt.Printf("正在初始化的虚拟机，%s\n", vname)
			control.Start(vname)
			orm, err := control.Bdb()
			if err != nil {
				cLog.Warn(err.Error())
			}
			var vm table.Virtual
			orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", vname).Find(&vm)
			for {
				connect := control.Connect()
				defer connect.Close()
				net, _ := connect.LookupNetworkByName("lan")
				dhcps, err := net.GetDHCPLeases()
				if err != nil {
					cLog.Warn(err.Error())
					continue
				}
				for _, dhcp := range dhcps {
					if dhcp.Mac == vm.Mac {
						//ip地址入库
						var date = make(map[string]interface{})
						date["LocalIP"] = dhcp.IPaddr
						orm.SetTable("Virtual").SetPK("ID").Where("Vname = ?", vname).Update(date)
						//设置外网ip
						//设置密码
						control.SetPasswd(vm.Vname, "root", vm.Passwd)
						return
					}
				}
				time.Sleep(3 * time.Second)
			}
		case str := <-Alarm:
			cLog.Warn("out alarm")
			data := strings.Split(str, "/")
			//发短信 邮件 通知
			cLog.Warn("%s,%s使用率过高超越阀值", data[1], data[0])
		case Vname := <-Bill:
			cLog.Warn("%s已到期", Vname)
			//发短信 邮件 通知
		}
	}
}
