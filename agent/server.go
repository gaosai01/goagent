package agent

import (
	"goagent/tcp"
	"goagent/agent/protocol"
	etcd "goagent/registry"
	"goagent/config"
)

func NewServer(register etcd.IRegister, appConfig *config.AppConfig) {
	//dubboUrl string, maxConn int, etcd registry.IRegister
	msgHandler := protocol.NewAgentMessageHandler(appConfig.Server.DubboUrl, appConfig.Server.MaxDubboConnCount, register)
	pkgHandler := protocol.NewAgentProviderPackageHandler()
	//port int, msgHandler MessageHandler, pkgHandler PackageHandler, maxConn int
	server := tcp.NewTcpServer(appConfig.Server.Port, msgHandler, pkgHandler, appConfig.Server.MaxServerConnCount)
	server.Start()
}
