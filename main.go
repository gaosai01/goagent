package main

import (
	"github.com/Mr-GaoSai/goagent/agent"
	etcd "github.com/Mr-GaoSai/goagent/registry"
	"flag"
	"github.com/Mr-GaoSai/goagent/log"
	"github.com/Mr-GaoSai/goagent/config"
)

func main() {
	/*
	1。 log
	2。 使用异步client和server，chan阻塞
	 */
	log.Init()

	// 配置文件
	path := flag.String("config", "config/provider-l.yaml", "this is config file path")
	flag.Parse()
	appConfig := config.AppConfig{}
	appConfig.GetConfig(*path)
	log.Info(appConfig)

	// 注册中心
	register, _ := etcd.GetEtcdRegistry(appConfig.AppName, []string{appConfig.EtcdUrl})

	// 启动服务
	if appConfig.Role == config.CONSUMER {
		agent.NewClient(register, &appConfig)
	} else {
		// 向 etcd 中注册服务
		ip := etcd.GetInternalIp()
		log.Info("本机ip为：", ip)
		host := etcd.Host{}
		host.Url = etcd.CreateUrl(ip, appConfig.Server.Port)
		host.Ratio = appConfig.Server.Ratio
		for _, v := range appConfig.Server.DubboServices {
			register.Register(v, &host)
		}
		// 开启tcp服务器
		agent.NewServer(register, &appConfig)
	}

}
