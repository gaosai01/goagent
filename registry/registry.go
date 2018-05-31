package registry

import (
	"strings"
)

// 主机信息
type Host struct {
	Url   string
	Ratio int //负载比例
}

// 服务的配置
type Service struct {
	Version, Name, Method, Hash, ArgTypes string
}

// 由于method args type 带有 / 符号，当注册到etcd中时会导致一些问题
func encodeArgTypes(types string) string {
	return strings.Replace(types, "/", "20", -1)
}

func decodeArgTypes(types string) string {
	return strings.Replace(types, "20", "/", -1)
}

// 定义接口用于抽象出etcd或者zookeeper
type IRegister interface {
	Register(service Service, host *Host) error
	Find(service, version, method, argsTypes string) (*Host, string) // 由于 consumer上传参数不携带hash，所以需要返还
	GetService(hash string) (*Service, bool)
}
