package tcp

import (
	"strconv"
	"net"
	"bytes"
	"github.com/Mr-GaoSai/goagent/log"
	"bufio"
	"time"
)

// session, close, write, read
type Session struct {
	conn    net.Conn
	data    interface{}
	scanner *bufio.Scanner // 处理tcp粘包
	server  *TcpServer
}

func (this *Session) Close() {
	this.conn.Close()
}
func (this *Session) Read() interface{} {
	return this.data
}
func (this *Session) Write(iface interface{}) error {
	err := this.server.packageHandler.Pack(iface, this.conn)
	if err != nil {
		return err
	}
	return nil
}

// tcp server
type TcpServer struct {
	port           int
	messageHandler MessageHandler
	packageHandler PackageHandler
	message        chan *Session
	maxConn        int // 最大连接数
	curConn        int // 当前连接数
	//closeMessageThread chan int
	//messageChan        chan *Session
}

func NewTcpServer(port int, msgHandler MessageHandler, pkgHandler PackageHandler,
	maxConn int) *TcpServer {
	server := TcpServer{}
	server.port = port
	server.maxConn = maxConn
	server.curConn = 0
	server.messageHandler = msgHandler
	server.packageHandler = pkgHandler
	server.message = make(chan *Session, maxConn)
	//server.closeMessageThread = make(chan int, maxConn)
	//server.messageChan = make(chan *Session, maxConn)
	return &server
}

// 启动tcp server
func (this *TcpServer) Start() {
	var buf bytes.Buffer
	buf.WriteString("0.0.0.0:")
	buf.WriteString(strconv.Itoa(this.port))
	netListen, err := net.Listen("tcp", buf.String())
	if err != nil {
		log.Error(err)
		return
	}
	defer netListen.Close()
	for {
		conn, err := netListen.Accept()
		if err != nil {
			continue
		}
		if this.curConn >= this.maxConn {
			log.Error("连接失败,已达到最大连接数")
			conn.Close()
			continue
		}
		this.curConn++
		session := Session{}
		session.conn = conn
		session.scanner = bufio.NewScanner(conn)
		session.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = this.packageHandler.Split(data, atEOF)
			return
		})
		session.server = this
		go this.packageThread(&session) // 解包协程
		//go this.messageThread()         // 处理消息协程
		// 下一步开发计划
		// 如果技术能力足够，需要查看fasthttp是怎么进行多路复用的，估计能提高性能
	}
}

func (this *TcpServer) StartAsync() {
	go this.Start()
}

func (this *TcpServer) packageThread(session *Session) {
	// 打印接收到的数据包
	defer session.conn.Close()
	for session.scanner.Scan() {
		ans, err := this.packageHandler.UnPack(session.scanner.Bytes())
		if err != nil {
			log.Error("tcp server UnPack err", err)
			break
		}
		session.data = ans
		//this.messageChan <- session
		this.messageHandler.Handle(session)
		this.packageHandler.Release(session.data)
	}
	time.Sleep(50 * time.Microsecond)
	//this.closeMessageThread <- 1
	this.curConn--
}

//func (this *TcpServer) messageThread() {
//	for {
//		select {
//		case session := <-this.messageChan:
//			this.messageHandler.Handle(session)
//			this.packageHandler.Release(session.data)
//		case <-this.closeMessageThread:
//			return
//		}
//	}
//}
