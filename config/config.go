package config

import (
	"github.com/Mr-GaoSai/goagent/registry"
	"io/ioutil"
	"github.com/Mr-GaoSai/goagent/log"
	"gopkg.in/yaml.v2"
)

const (
	CONSUMER = iota
	PROVIDER
)

type AppServer struct {
	Port               int                // 服务监听的端口
	DubboUrl           string             // dubbo服务的地址
	MaxServerConnCount int                // tcp服务器最大的连接数
	MaxDubboConnCount  int                // 与dubbo服务连接的连接池的大小
	DubboServices      []registry.Service // server 注册的服务
	Ratio              int                // 负载指数
}

type AppClient struct {
	Port         int
	MaxConnCount int
}

type AppConfig struct {
	AppName string
	Role    int
	EtcdUrl string
	Server  AppServer
	Client  AppClient
}

func (this *AppConfig) GetConfig(path string) {
	log.Debug("配置文件路径", path)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("yamlFile.Get err ", err)
	}
	log.Debug("配置信息", string(yamlFile))
	err = yaml.Unmarshal(yamlFile, this)
	if err != nil {
		log.Error("Unmarshal Error", err)
	}
}
