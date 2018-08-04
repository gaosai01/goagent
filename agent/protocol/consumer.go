package protocol

import (
	"io"
	"sync"
	"github.com/Mr-GaoSai/goagent/tcp"
	"bytes"
)

func NewAgentConsumerPackageHandler() tcp.PackageHandler {
	headerPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, HEADER_LENGTH, HEADER_LENGTH)
		},
	}
	responsePool := &sync.Pool{
		New: func() interface{} {
			return &AgentResponse{}
		},
	}
	obj := AgentConsumerPackageHandler{}
	obj.headerPool = headerPool
	obj.responsePool = responsePool
	return &obj
}

type AgentConsumerPackageHandler struct {
	headerPool   *sync.Pool
	responsePool *sync.Pool
}

func (this *AgentConsumerPackageHandler) Pack(v interface{}, writer io.Writer) error {
	request, ok := v.(*AgentRequest)
	if !ok {
		return ARGS_ERROR
	}
	header := this.headerPool.Get().([]byte)
	short2bytes(header, request.Header.Magic, 0)
	header[2] = byte(request.Header.Command)
	int2bytes(header, request.Header.Len, 3)
	bytesBuf := bytes.NewBuffer(make([]byte, 0))
	_, err := bytesBuf.Write(header)
	this.headerPool.Put(header)
	if err != nil {
		return err
	}
	io.WriteString(bytesBuf, request.Body.Hash)
	io.WriteString(bytesBuf, "\n")
	_, err = io.WriteString(bytesBuf, request.Body.Parameter)
	if err != nil {
		return err
	}
	_, err = bytesBuf.WriteTo(writer)
	if err != nil {
		return err
	}
	return nil
}

func (this *AgentConsumerPackageHandler) Split(data []byte, eof bool) (int, []byte, error) {
	if len(data) <= HEADER_LENGTH {
		return 0, nil, nil
	}
	length := int(bytes2int(data, 3))
	rLen := length + HEADER_LENGTH
	if len(data) < rLen {
		if !eof {
			return 0, nil, nil
		}
		return 0, nil, MSG_ERROR
	}
	return rLen, data[:rLen], nil
}

func (this *AgentConsumerPackageHandler) UnPack(data []byte) (interface{}, error) {
	response := this.responsePool.Get().(*AgentResponse)
	response.Header.Magic = bytes2short(data, 0)
	response.Header.Command = uint8(data[2])
	response.Header.Len = bytes2int(data, 3)
	length := int(HEADER_LENGTH + response.Header.Len)
	if length > len(data) {
		return nil, SPLIT_ERROR
	}
	response.Body = string(data[HEADER_LENGTH:length])
	return response, nil
}

func (this *AgentConsumerPackageHandler) Release(data interface{}) {
	response, ok := data.(*AgentResponse)
	if ok {
		this.responsePool.Put(response)
	}
}
