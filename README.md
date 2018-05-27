# cpanel

control panel

## 端口

iptables -I INPUT -p tcp --dport 8100 -j ACCEPT

### 生产环境

系统 centos7.4

```sh
yum install virtual-devel virtual
```

## 编译环境

```sh

```

## 编译方法

```sh
go build main.go
```