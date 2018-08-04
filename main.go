package main

import (
	"github.com/Mr-GaoSai/goagent/agent"
	etcd "github.com/Mr-GaoSai/goagent/registry"
	"flag"
	"github.com/Mr-GaoSai/goagent/log"
	"github.com/Mr-GaoSai/goagent/config"
	"os"
	"os/signal"
	"syscall"
)

const GoagentVersion = "1.0.test"

func main() {
	log.Init()
	log.Info("启动goagent,版本:" + GoagentVersion)
	// 读取配置文件
	path := flag.String("c", "/etc/linux/conf.yaml", "this is config file path")
	flag.Parse()
	appConfig := config.AppConfig{}
	err := appConfig.GetConfig(*path)
	if err != nil {
		log.Error("配置文件不存在", err)
		return
	}
	// 打印配置信息
	log.Info(appConfig)

	// 注册中心
	register, _ := etcd.GetEtcdRegistry(appConfig.AppName, []string{appConfig.EtcdUrl})

	// 启动服务
	if appConfig.Role == config.CONSUMER {
		// 这个方法是非阻塞的
		agent.StartConsumerAgent(register, &appConfig)
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
		// 开启tcp服务器作为provider agent，非阻塞的
		agent.StartProviderAgent(register, &appConfig)
	}

	initSignal()

}

func initSignal() {
	signals := make(chan os.Signal, 1)
	// It is not possible to block SIGKILL or syscall.SIGSTOP
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-signals
		log.Info("get signal %s", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			log.Info("agent关闭")
			return
		}
	}
}
