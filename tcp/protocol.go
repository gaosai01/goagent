package tcp

import "io"

// 粘包，拆包的解析接口
type PackageHandler interface {
	// 将对象写入连接里，传输到套接字的另一端
	Pack(interface{}, io.Writer) error
	// Split函数在包UnPack前的调用,这个才是真正的防止粘包和拆包的函数
	Split([]byte, bool) (int, []byte, error)
	// 将byte[]解析成对象
	UnPack([]byte) (interface{}, error)
	// 提供释放byte[]解析出来的对象
	Release(interface{})
}

// PackageHandler的UnPack解析出来的对象后会调用此方法
type MessageHandler interface {
	Handle(*Session)
}
