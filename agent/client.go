package agent

import (
	"goagent/agent/protocol"
	"goagent/tcp"
	"goagent/log"
	"bytes"
	"strconv"
	etcd "goagent/registry"
	"goagent/config"
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

func NewClient(r etcd.IRegister, ac *config.AppConfig) {
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
	log.Info("start consumer-agent，port:", appConfig.Client.Port)

	err := fasthttp.ListenAndServe(buf.String(), callAgent)
	if err != nil {
		log.Error("启动http服务器error", err)
	}
}

//var pingjun float64 = 0
//var num float64 = 0.0
//var nummu sync.Mutex
func callAgent(ctx *fasthttp.RequestCtx) {
	//start_time := time.Now().UnixNano() / 100000
	//defer func() {
	//	end_time := time.Now().UnixNano() / 100000
	//	cha_time := end_time - start_time
	//	log.Info("此次访问时间:", cha_time)
	//	nummu.Lock()
	//	pingjun = pingjun*num + float64(cha_time)
	//	num += 1
	//	pingjun = pingjun / num
	//	nummu.Unlock()
	//	log.Info("平均访问时间:", pingjun)
	//}()
	req := ctx.Request.PostArgs()
	iface := string(req.Peek("interface"))
	method := string(req.Peek("method"))
	parameterTypesString := string(req.Peek("parameterTypesString"))
	parameter := string(req.Peek("parameter"))
	host, hash := register.Find(iface, "default", method, parameterTypesString)
	if host == nil {
		log.Error("不存在provider-agent")
		fmt.Fprintln(ctx, "Error")
		return
	}
	//log.Debug("parameter", parameter)
	//log.Debug(*host)
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
			pool, err = tcp.NewTcpPool(host.Url, appConfig.Client.MaxConnCount*host.Ratio, 30*time.Second, pkgHandler)
			if err != nil {
				log.Error("tcp pool create", err)
				mu.Unlock()
				return nil, err
			}
			tcpPoolMap[host.Url] = pool
		}
		mu.Unlock()
	}
	return pool.Get()
}

/*
由于先调用的get，然后put，所以不许要锁
 */
func putTcpClient(client *tcp.TcpClient) {
	pool, ok := tcpPoolMap[client.Url()]
	if !ok {
		client.Close()
		return
	}
	pool.Put(client)
}
