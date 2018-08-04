package registry_test

import (
	etcd "github.com/Mr-GaoSai/goagent/registry"
	"github.com/Mr-GaoSai/goagent/log"
	"testing"
)

func TestGetEtcdRegistry(t *testing.T) {
	registry, err := etcd.GetEtcdRegistry("go-agent", []string{"127.0.0.1:2379"})
	if err != nil {
		log.Error(err)
		return
	}
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
	registry.Register(service, &host)
	h := registry.Find(service.Name, service.Version, service.Method, service.ArgTypes)
	log.Debug(*h)
	str := registry.FindHash(service.Name, service.Version, service.Method, service.ArgTypes)
	log.Debug(str)
}
