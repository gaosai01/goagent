package agent

import (
	"github.com/Mr-GaoSai/goagent/agent/protocol"
	"github.com/Mr-GaoSai/goagent/tcp"
	"github.com/Mr-GaoSai/goagent/log"
	"bytes"
	"strconv"
	etcd "github.com/Mr-GaoSai/goagent/registry"
	"github.com/Mr-GaoSai/goagent/config"
	"sync"
	"fmt"
	"time"
	"github.com/valyala/fasthttp"
)

var register etcd.IRegister
var appConfig *config.AppConfig
var tcpPoolMap map[string]*tcp.TcpPool
var mu sync.RWMutex
var pkgHandler tcp.PackageHandler

func StartConsumerAgent(r etcd.IRegister, ac *config.AppConfig) {
	register = r
	appConfig = ac
	tcpPoolMap = make(map[string]*tcp.TcpPool)
	//url string, num int, duration time.Duration, handler PackageHandler
	// 包解析工具
	pkgHandler = protocol.NewAgentConsumerPackageHandler()
	// 开启http服务器
	var buf bytes.Buffer
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(appConfig.Client.Port))
	log.Info("consumerAgent开启，监听port:", appConfig.Client.Port)
	// 开启http服务器
	go func() {
		err := fasthttp.ListenAndServe(buf.String(), callAgent)
		if err != nil {
			log.Error("启动http服务器error", err)
		}
	}()
}

// http request上传3个字段，interface，method，parameter字段
func callAgent(ctx *fasthttp.RequestCtx) {
	req := ctx.Request.PostArgs()
	// 获取http request的interface字段
	iface := string(req.Peek("interface"))
	// 获取http request的method字段
	method := string(req.Peek("method"))
	parameterTypesString := string(req.Peek("parameterTypesString"))
	// 获取http request的paramter字段
	parameter := string(req.Peek("parameter"))
	host, hash := register.Find(iface, "default", method, parameterTypesString)
	if host == nil {
		log.Error("不存在provider-agent")
		fmt.Fprintln(ctx, "Error")
		return
	}
	// 创建agent传输对象
	agentReq := protocol.NewAgentRequest(hash, parameter)
	// 创建与agent的连接
	tcpClient, err := getTcpClient(host)
	if err != nil {
		protocol.ReleaseAgentRequest(agentReq)
		log.Error("TcpPool.Get", err)
		fmt.Fprintln(ctx, "Error")
		return
	}
	// defer 如果出现运行中恐慌，需要需要关闭tcp连接
	defer func() {
		if p := recover(); p != nil {
			log.Error("callProviderAgent2 error:", p)
			tcpClient.Close()
		}
	}()
	// 写数据
	tcpClient.Write(agentReq)
	// sync.pool释放
	protocol.ReleaseAgentRequest(agentReq)
	if err != nil {
		tcpClient.Close()
		fmt.Fprintln(ctx, "Error")
		log.Error("write to agent error", err)
		return
	}
	// 读取响应
	agentRes, err := tcpClient.Read()
	if err != nil {
		tcpClient.Close()
		fmt.Fprintln(ctx, "Error")
		log.Error("read from agent error", err)
		return
	}
	// 在读取完成后释放连接
	putTcpClient(tcpClient)
	// 反馈给http请求
	res, ook := agentRes.(*protocol.AgentResponse)
	if !ook {
		tcpClient.Close()
		fmt.Fprintln(ctx, "Error")
		log.Error("read from agent error")
		return
	}
	fmt.Fprint(ctx, res.Body)
	tcpClient.Release(res)
	return
}

/*
url -> tcpPool对象，懒加载，读写锁
 */
func getTcpClient(host *etcd.Host) (*tcp.TcpClient, error) {
	mu.RLock()
	pool, ok := tcpPoolMap[host.Url]
	mu.RUnlock()
	if !ok {
		mu.Lock()
		pool, ok = tcpPoolMap[host.Url]
		if !ok {
			var err error
			//url string, num int, duration time.Duration, handler PackageHandler
			// 穿件tcp连接池，和每一个远程服务器都有一个连接
			pool, err = tcp.NewTcpPool(host.Url, appConfig.Client.MaxConnCount*host.Ratio, 30*time.Second, pkgHandler)
			if err != nil {
				log.Error("tcp pool create", err)
				mu.Unlock()
				return nil, err
			}
			// 连接池存储到字典里面
			tcpPoolMap[host.Url] = pool
		}
		mu.Unlock()
	}
	return pool.Get()
}

/*
由于先调用的get，然后put，所以不需要锁
 */
func putTcpClient(client *tcp.TcpClient) {
	pool, ok := tcpPoolMap[client.Url()]
	if !ok {
		client.Close()
		return
	}
	pool.Put(client)
}
