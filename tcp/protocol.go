package tcp

import "io"

type PackageHandler interface {
	Pack(interface{}, io.Writer) error
	Split([]byte, bool) (int, []byte, error)
	UnPack([]byte) (interface{}, error)
	Release(interface{})
}

type MessageHandler interface {
	Handle(*Session)
}
