package registry

import (
	"net"
	"math/rand"
	"bytes"
	"strconv"
)

// 获取本机的Ip地址
func GetInternalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// 随机负载均衡算法，这个不是根据权重，权重在etcd哪里设置了
func GetHostWithWieghtRandom(hosts []Host) *Host {
	if len(hosts) == 0 {
		return nil
	}
	return &hosts[rand.Intn(len(hosts))]
}

// 工具函数，通过ip和端口生成远程地址
func CreateUrl(ip string, port int) string {
	var buf bytes.Buffer
	buf.WriteString(ip)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(port))
	return buf.String()
}
