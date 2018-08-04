package pool

import (
	"sync"
	"errors"
)

// 一个链接吃
type Pool struct {
	// 链接池中新建连接时会调用此方法
	New func() interface{}
	// 判断连接池中的连接是否还有心跳的方法
	Ping func(interface{}) bool
	// 关闭某个连接
	Close func(interface{})
	// 存储的连接的chan对象
	store chan interface{}
	// 锁
	mu sync.Mutex
	// 连接池的最大容量
	maxCap int
}

// 新建连接池的方法
func New(initCap, maxCap int, newFunc func() interface{}) (*Pool, error) {
	if maxCap == 0 || initCap > maxCap {
		return nil, errors.New("不合法的initCap参数")
	}
	p := new(Pool)
	p.store = make(chan interface{}, maxCap)
	p.maxCap = maxCap
	if newFunc != nil {
		p.New = newFunc
	}
	for i := 0; i < initCap; i++ {
		v, err := p.create()
		if err != nil {
			return p, err
		}
		p.store <- v
	}
	return p, nil
}

// 获取连接池中连接数
func (p *Pool) Len() int {
	return len(p.store)
}

// 从连接池中获取一个连接
func (p *Pool) Get() (interface{}, error) {
	if p.store == nil {
		// 存储不存在，创建一个新连接
		return p.create()
	}
	for {
		select {
		// 循环等待创建
		case v := <-p.store:
			if p.Ping != nil && p.Ping(v) == false {
				p.Close(v)
				continue
			}
			return v, nil
		default:
			return p.create()
		}
	}
}

// 将连接放回连接池
func (p *Pool) Put(v interface{}) {
	if p.store == nil {
		return
	}
	select {
	case p.store <- v:
		return
	default:
		// pool is full, close passed connection
		if p.Close != nil {
			p.Close(v)
		}
		return
	}
}

// 销毁连接池
func (p *Pool) Destroy() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.store == nil {
		// 连接池已经销毁
		return
	}
	close(p.store)
	for v := range p.store {
		if p.Close != nil {
			p.Close(v)
		}
	}
	p.store = nil
}

// 新建连接
func (p *Pool) create() (interface{}, error) {
	if p.store == nil {
		p.mu.Lock()
		if p.store == nil {
			p.store = make(chan interface{}, p.maxCap)
		}
		p.mu.Unlock()
	}
	if p.New == nil {
		return nil, errors.New("pool.New方法不能为空")
	}
	ans := p.New()
	if ans == nil {
		return nil, errors.New("新建连接失败")
	}
	return ans, nil
}
