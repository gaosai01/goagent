package protocol

import (
	"io"
	"sync"
	"goagent/tcp"
	"strings"
	"bytes"
)

func NewAgentProviderPackageHandler() tcp.PackageHandler {
	headerPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, HEADER_LENGTH, HEADER_LENGTH)
		},
	}
	requestPool := &sync.Pool{
		New: func() interface{} {
			return &AgentRequest{}
		},
	}
	obj := AgentProviderPackageHandler{}
	obj.headerPool = headerPool
	obj.requestPool = requestPool
	return &obj
}

type AgentProviderPackageHandler struct {
	headerPool  *sync.Pool
	requestPool *sync.Pool
}

func (this *AgentProviderPackageHandler) Pack(v interface{}, writer io.Writer) error {
	response, ok := v.(*AgentResponse)
	if !ok {
		return ARGS_ERROR
	}
	header := this.headerPool.Get().([]byte)
	short2bytes(header, response.Header.Magic, 0)
	header[2] = byte(response.Header.Command)
	int2bytes(header, response.Header.Len, 3)
	bytesBuf := bytes.NewBuffer(make([]byte, 0))
	_, err := bytesBuf.Write(header)
	this.headerPool.Put(header)
	if err != nil {
		return err
	}
	_, err = io.WriteString(bytesBuf, response.Body)
	if err != nil {
		return err
	}
	_, err = bytesBuf.WriteTo(writer)
	if err != nil {
		return err
	}
	return nil
}

func (this *AgentProviderPackageHandler) Split(data []byte, eof bool) (int, []byte, error) {
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

func (this *AgentProviderPackageHandler) UnPack(data []byte) (interface{}, error) {
	request := this.requestPool.Get().(*AgentRequest)
	request.Header.Magic = bytes2short(data, 0)
	request.Header.Command = uint8(data[2])
	request.Header.Len = bytes2int(data, 3)
	length := int(HEADER_LENGTH + request.Header.Len)
	if length > len(data) {
		return nil, SPLIT_ERROR
	}
	str := string(data[HEADER_LENGTH:length])
	k := strings.Index(str, "\n")
	if k == -1 {
		return nil, BODY_ERROR
	}
	request.Body.Hash = str[:k]
	request.Body.Parameter = str[k+1:]
	return request, nil
}

func (this *AgentProviderPackageHandler) Release(data interface{}) {
	request, ok := data.(*AgentRequest)
	if ok {
		this.requestPool.Put(request)
	}
}
