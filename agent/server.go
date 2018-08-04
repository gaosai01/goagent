package agent

import (
	"github.com/Mr-GaoSai/goagent/tcp"
	"github.com/Mr-GaoSai/goagent/agent/protocol"
	etcd "github.com/Mr-GaoSai/goagent/registry"
	"github.com/Mr-GaoSai/goagent/config"
)

func StartProviderAgent(register etcd.IRegister, appConfig *config.AppConfig) {
	//dubboUrl string, maxConn int, etcd registry.IRegister
	msgHandler := protocol.NewAgentMessageHandler(appConfig.Server.DubboUrl, appConfig.Server.MaxDubboConnCount, register)
	pkgHandler := protocol.NewAgentProviderPackageHandler()
	//port int, msgHandler MessageHandler, pkgHandler PackageHandler, maxConn int
	server := tcp.NewTcpServer(appConfig.Server.Port, msgHandler, pkgHandler, appConfig.Server.MaxServerConnCount)
	server.StartAsync()
}
