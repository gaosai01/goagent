package tcp

import (
	"net"
	"time"
	"github.com/Mr-GaoSai/goagent/pool"
	"github.com/Mr-GaoSai/goagent/log"
	"bufio"
)

// tcp链接的对象，可以通过TcpPool对象生成
type TcpClient struct {
	conn           net.Conn       // 连接
	ut             time.Time      // 最近一次使用时间,用于定时关闭
	packageHandler PackageHandler // 编码器解码器
	scanner        *bufio.Scanner
	url            string
}

// 和tcp连接有关的写操作
func (this *TcpClient) Write(v interface{}) error {
	// 写入对象,调用 编码的方式
	// 最好的方式为边解码边写入连接池，经过测试并不是这样，边解压边写入的方式会增大网络曾压力，多个小包传输数据，时间更长
	err := this.packageHandler.Pack(v, this.conn)
	if err != nil {
		return err
	}
	return nil
}

// 和tcp链接相关的读操作
func (this *TcpClient) Read() (interface{}, error) {
	defer func() {
		if p := recover(); p != nil {
			log.Error("tcp Read panic", p)
		}
	}()
	if !this.scanner.Scan() {
		return nil, this.scanner.Err()
	}
	ans, err := this.packageHandler.UnPack(this.scanner.Bytes())
	if err != nil {
		return ans, err
	}
	return ans, nil
}

// 和tcp链接有关的read出来的对象的释放操作
func (this *TcpClient) Release(v interface{}) {
	this.packageHandler.Release(v)
}

// tcp链接关闭
func (this *TcpClient) Close() {
	this.conn.Close()
}

// 获取tcp链接的远程地址信息
func (this *TcpClient) Url() string {
	return this.url
}

// tcp链接池
type TcpPool struct {
	pool *pool.Pool
	url  string
}

// tcp 客户端采用连接池的方式, url访问地址, 连接池内连接数量, 元素生存时间, 包处理器
func NewTcpPool(url string, num int, duration time.Duration, handler PackageHandler) (*TcpPool, error) {
	tp := TcpPool{}
	tp.url = url
	pl, err := pool.New(0, num, func() interface{} {
		tc := TcpClient{}
		conn, err := net.Dial("tcp", url)
		if err != nil {
			log.Error("create client connection error: %v\n", err)
			return nil
		}
		tc.conn = conn
		tc.url = url
		tc.ut = time.Now()
		tc.packageHandler = handler
		//解决粘包问题的对象
		tc.scanner = bufio.NewScanner(tc.conn)
		tc.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = handler.Split(data, atEOF)
			return
		})
		return &tc
	})
	if err != nil {
		log.Error("新建tcp连接池错误url:", url, err)
		return nil, err
	}

	// 连接池的有效期
	pl.Ping = func(v interface{}) bool {
		tc := v.(*TcpClient)
		if tc == nil {
			return false
		}
		return !time.Now().After(tc.ut.Add(duration))
	}
	// 关闭操作
	pl.Close = func(i interface{}) {
		i.(*TcpClient).conn.Close()
	}
	tp.pool = pl
	return &tp, nil
}

// 获取TcpClient对象
func (this *TcpPool) Get() (*TcpClient, error) {
	client, err := this.pool.Get()
	if err != nil {
		return nil, err
	}
	tc := client.(*TcpClient)
	tc.ut = time.Now()
	return tc, nil
}

// 获取tcp池的远程地址
func (this *TcpPool) Url() string {
	return this.url
}

// 将tcpClient对象放回池子
func (this *TcpPool) Put(tc *TcpClient) {
	if tc == nil {
		return
	}
	this.pool.Put(tc)
}

// tcp链接池关闭
func (this *TcpPool) Close() {
	this.pool.Destroy()
}
