package registry_test

import (
	etcd "github.com/Mr-GaoSai/goagent/registry"
	"github.com/Mr-GaoSai/goagent/log"
	"testing"
)

func TestGetEtcdRegistry(t *testing.T) {
	// 获取Etcd注册中心
	registry, err := etcd.GetEtcdRegistry("go-agent", []string{"127.0.0.1:2379"})
	if err != nil {
		log.Error(err)
		return
	}
	// 生成需要注册的Service对象
	service := etcd.Service{}
	service.Name = "IHelloService"
	service.Method = "hash"
	service.ArgTypes = "String"
	service.Version = "default"
	service.Hash = "h1"
	host := etcd.Host{}
	ip := etcd.GetInternalIp()
	port := 8080
	host.Url = etcd.CreateUrl(ip, port)
	host.Ratio = 1
	// 注册etcd上，传输service对象和本机信息
	registry.Register(service, &host)
	// 从etcd上查找是否存在刚才存储的对象
	h, _ := registry.Find(service.Name, service.Version, service.Method, service.ArgTypes)
	// 打印获取到的对象
	log.Debug(*h)
	// 获取service的hash值
	str := registry.FindHash(service.Name, service.Version, service.Method, service.ArgTypes)
	log.Debug(str)
}
