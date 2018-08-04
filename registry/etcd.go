package registry

import (
	"github.com/coreos/etcd/clientv3"
	"sync"
	"strconv"
	"time"
	"github.com/Mr-GaoSai/goagent/log"
	"context"
	"bytes"
	"strings"
	"os"
	"encoding/json"
)

// 有关etcd struct的配置

const default_keep_alive_time = 3

type EtcdRegistry struct {
	client           *clientv3.Client    // etcd客户端
	serviceHostCache map[string][]Host   // 本地缓存发现的地址，通过异步线程WatchEtcd更改
	serviceEtcdCache map[string]*Service // 本地hash映射缓存地址
	serviceCache     map[string]*Service // 服务端的缓存
	lock             sync.RWMutex        // 缓存更改需要读写锁
	name             string              // 应用名
}

func (this *EtcdRegistry) Register(service Service, host *Host) error {
	resp, err := this.client.Grant(context.TODO(), default_keep_alive_time) // default_keep_alive_time s
	if err != nil {
		log.Error(err)
		return err
	}
	key := this.mkKey(service, host)
	value := this.mkValue(service, host)
	//put
	_, err = this.client.Put(context.TODO(), key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("etcd注册节点:", key, value)
	this.addService(service.Hash, &service)
	//异步线程 心跳
	go this.keepAlive(resp.ID)
	return nil
}
func (this *EtcdRegistry) Find(service, version, method, argsTypes string) (*Host, string) {
	etcdKey := this.mkKey1(service, version, method, argsTypes)
	//读写锁管理缓存
	this.lock.RLock()
	// 存在[]host缓存吗？
	value, ook := this.serviceHostCache[etcdKey]
	hash := this.serviceEtcdCache[etcdKey]
	this.lock.RUnlock()
	if ook {
		return GetHostWithWieghtRandom(value), hash.Hash
	}
	// 写锁锁定
	this.lock.Lock()
	defer this.lock.Unlock()
	// 双重锁验证
	value, ook = this.serviceHostCache[etcdKey]
	hash = this.serviceEtcdCache[etcdKey]
	if ook {
		return GetHostWithWieghtRandom(value), hash.Hash
	}
	// 双重验证都进入
	// 从etcd数据库中拿数据
	res, _ := this.client.Get(context.TODO(), etcdKey, clientv3.WithPrefix())
	slice, ok := this.serviceHostCache[etcdKey]
	if !ok {
		slice = make([]Host, 0)
	}
	for _, kv := range res.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)
		host := this.praseKey(key, value)
		if host == nil {
			continue
		}
		for i := 0; i < host.Ratio; i++ {
			slice = append(slice, *host)
		}
		if hash == nil {
			hash = this.praseService(service, version, method, argsTypes, key, value)
		}
	}
	this.serviceHostCache[etcdKey] = slice
	this.serviceEtcdCache[etcdKey] = hash
	log.Info("获取hosts:", etcdKey, slice)
	log.Info("得到service的hash", etcdKey, *hash)
	// 异步watch机制
	go this.watch(etcdKey)
	return GetHostWithWieghtRandom(slice), hash.Hash
}

// consumer使用的，从etcd中获取service的hash
func (this *EtcdRegistry) FindHash(service, version, method, argsTypes string) string {
	etcdKey := this.mkKey1(service, version, method, argsTypes)
	this.lock.RLock()
	defer this.lock.RUnlock()
	value, ok := this.serviceEtcdCache[etcdKey]
	if !ok {
		log.Debug("需要先Find后FindHash才可以得到数据")
		return "hash"
	}
	return value.Hash
}

func (this *EtcdRegistry) addService(hash string, service *Service) {
	this.serviceCache[hash] = service
}

func (this *EtcdRegistry) GetService(hash string) (*Service, bool) {
	s, ok := this.serviceCache[hash]
	return s, ok
}

// 这个方法需要改进
func (this *EtcdRegistry) keepAlive(leaseid clientv3.LeaseID) {
	// <-chan *clientv3.LeaseKeepAliveResponse
	ch, err := this.client.KeepAlive(context.TODO(), leaseid)
	if err != nil {
		log.Error(err)
		return
	}
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				log.Error("keep alive channel closed")
				return
			} else {
				//log.Debug(ka)
			}
		}
	}
}

func (this *EtcdRegistry) watch(etcdKey string) {
	rch := this.client.Watch(context.Background(), etcdKey, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)
				this.addCache(etcdKey, key, value)
			case clientv3.EventTypeDelete:
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)
				this.delCache(etcdKey, key, value)
			}
		}
	}
}

func (this *EtcdRegistry) addCache(etcdKey, str, value string) {
	host := this.praseKey(str, value)
	if host == nil {
		return
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	slice, _ := this.serviceHostCache[etcdKey]
	for _, old := range slice {
		if old.Url == host.Url {
			return
		}
	}
	log.Info("添加新节点", etcdKey, host)
	for i := 0; i < host.Ratio; i++ {
		slice = append(slice, *host)
	}
	this.serviceHostCache[etcdKey] = slice
}

func (this *EtcdRegistry) delCache(etcdKey, str, value string) {
	host := this.praseKey(str, value)
	if host == nil {
		return
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	slice, _ := this.serviceHostCache[etcdKey]
	// remove 移除元素,不知道是否会出现遍历中删除的error
	for i := range slice {
		if slice[i].Url == host.Url {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	log.Info("移除节点", etcdKey, host)
	this.serviceHostCache[etcdKey] = slice
}

func (this *EtcdRegistry) mkKey(service Service, host *Host) string {
	var buf strings.Builder
	argTypes := encodeArgTypes(service.ArgTypes)

	size := 7 + len(this.name) + len(service.Name) + len(service.Version)
	size += len(service.Method) + len(argTypes) + len(service.Hash) + len(host.Url)

	buf.Grow(size)

	buf.WriteString("/")
	buf.WriteString(this.name)
	buf.WriteString("/")
	buf.WriteString(service.Name)
	buf.WriteString("/")
	buf.WriteString(service.Version)
	buf.WriteString("/")
	buf.WriteString(service.Method)
	buf.WriteString("/")
	buf.WriteString(argTypes)
	buf.WriteString("/")
	buf.WriteString(service.Hash)
	buf.WriteString("/")
	buf.WriteString(host.Url)
	return buf.String()
}

func (this *EtcdRegistry) mkKey1(service, version, method, argsTypes string) string {
	var buf strings.Builder
	argsTypes = encodeArgTypes(argsTypes)
	size := 5 + len(this.name) + len(service) + len(version) + len(method) + len(argsTypes)
	buf.Grow(size)
	buf.WriteString("/")
	buf.WriteString(this.name)
	buf.WriteString("/")
	buf.WriteString(service)
	buf.WriteString("/")
	buf.WriteString(version)
	buf.WriteString("/")
	buf.WriteString(method)
	buf.WriteString("/")
	buf.WriteString(argsTypes)
	return buf.String()
}

func (*EtcdRegistry) praseKey(key string, value string) *Host {
	index := strings.LastIndex(key, "/")
	if index == -1 {
		return nil
	}
	sub := key[index+1:]
	ratio, err := strconv.Atoi(value)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &Host{Url: sub, Ratio: ratio}
}

func (*EtcdRegistry) praseService(service, version, method, argsTypes, key, value string) *Service {
	index := strings.LastIndex(key, "/")
	ans := Service{}
	ans.Name = service
	ans.Version = version
	ans.Method = method
	ans.ArgTypes = argsTypes
	if index == -1 {
		ans.Hash = "error"
		return &ans
	}
	// 解析hash
	sub := key[:index]
	index = strings.LastIndex(sub, "/")
	if index == -1 {
		ans.Hash = "error"
		return &ans
	}
	hash := sub[index+1:]
	ans.Hash = hash
	return &ans
}

func (*EtcdRegistry) mkValue(service Service, host *Host) string {
	return strconv.Itoa(host.Ratio)
}

// 外部获取实例方法
func GetEtcdRegistry(appName string, etcd_urls []string) (*EtcdRegistry, error) {
	register := EtcdRegistry{}
	register.name = appName
	register.serviceHostCache = make(map[string][]Host)
	register.serviceEtcdCache = make(map[string]*Service)
	register.serviceCache = make(map[string]*Service)
	var err error
	register.client, err = clientv3.New(clientv3.Config{
		Endpoints:   etcd_urls,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	log.Info("etcd连接成功,url：", etcd_urls)
	return &register, nil
}

// 从json文件中读取service数组配置，然后注册
func RegisterFromJsonFile(path string, register IRegister, ratio, port int) {
	file, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var slice []Service
	decoder.Decode(&slice)
	ip := GetInternalIp()
	log.Info("本机ip为：", ip)
	host := Host{}
	host.Url = CreateUrl(ip, port)
	host.Ratio = ratio
	for _, service := range slice {
		register.Register(service, &host)
	}
}

func CreateUrl(ip string, port int) string {
	var buf bytes.Buffer
	buf.WriteString(ip)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(port))
	return buf.String()
}
