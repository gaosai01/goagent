package protocol

import (
	"github.com/Mr-GaoSai/goagent/tcp"
	"github.com/Mr-GaoSai/goagent/registry"
	"github.com/Mr-GaoSai/goagent/log"
	"github.com/Mr-GaoSai/goagent/agent/dubbo"
	"time"
)

func NewAgentMessageHandler(dubboUrl string, maxConn int, etcd registry.IRegister) tcp.MessageHandler {
	handler := AgentMessageHandler{}
	handler.etcd = etcd
	var err error
	handler.tcpPool, err = tcp.NewTcpPool(dubboUrl, maxConn, 30*time.Second, dubbo.NewDubboPackageHandler())
	if err != nil {
		log.Error("与dubbo建立连接出现err", err)
	}
	gentResponse := NewAgentResponse("Error")
	handler.defaultAgentResponse = gentResponse
	return &handler
}

type AgentMessageHandler struct {
	etcd                 registry.IRegister
	tcpPool              *tcp.TcpPool
	defaultAgentResponse *AgentResponse
}

func (this *AgentMessageHandler) Handle(session *tcp.Session) {
	log.Debug("tcp server 接收到新请求")
	request := session.Read().(*AgentRequest)
	service, ok := this.etcd.GetService(request.Body.Hash)
	if !ok {
		log.Error("provider的配置里面不存在hash的服务,", request.Body.Hash)
		return
	}
	dubboReq := dubbo.NewDubboRequest()
	dubboReq.ServiceName = service.Name
	dubboReq.Method = service.Method
	dubboReq.MethodParameterTypes = service.ArgTypes
	dubboReq.MethodArgs = request.Body.Parameter
	dubboReq.Attachments["path"] = service.Name
	client, err := this.tcpPool.Get()
	if err != nil {
		log.Error("与dubbo服务相连接的tcp连接池获取失败", err)
		session.Write(this.defaultAgentResponse)
		return
	}
	err = client.Write(dubboReq)
	dubbo.ReleaseDubboRequest(dubboReq)
	if err != nil {
		log.Error("写入dubborequest请求时出现error", err)
		session.Write(this.defaultAgentResponse)
		return
	}
	dubboRes, err := client.Read()
	if err != nil {
		log.Error("读取dubbo请求的返回值时出现error", err)
		session.Write(this.defaultAgentResponse)
		return
	}
	this.tcpPool.Put(client)
	ans := dubboRes.(string)
	log.Debug("写入答案:", ans)
	agentRes := NewAgentResponse(ans)
	err = session.Write(agentRes)
	ReleaseAgentResponse(agentRes)
	if err != nil {
		log.Error("向consumeragent写入数据出现error", err)
		session.Close()
		return
	}

}
