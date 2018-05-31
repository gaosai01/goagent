package protocol

import (
	"errors"
	"sync"
)

const (
	REQUEST   = iota
	RESPONSE
	HEARTBEAT
)

const (
	HEADER_LENGTH = 7
)
const MAGIC uint16 = 123

var (
	BODY_ERROR  = errors.New("agent body error")
	MAGIC_ERROR = errors.New("agent magic header error")
	ARGS_ERROR  = errors.New("参数错误")
)

var (
	MSG_ERROR   = errors.New("eof数据不完整")
	SPLIT_ERROR = errors.New("Split粘包问题解决失败")
)

type AgentHeader struct {
	Magic   uint16
	Command uint8
	Len     uint32
}

type AgentBody struct {
	Hash      string
	Parameter string
}

type AgentRequest struct {
	Header AgentHeader
	Body   AgentBody
}

type AgentResponse struct {
	Header AgentHeader
	Body   string
}

var requestPool *sync.Pool
var responsePool *sync.Pool

func init() {
	requestPool = &sync.Pool{
		New: func() interface{} {
			return &AgentRequest{}
		},
	}

	responsePool = &sync.Pool{
		New: func() interface{} {
			return &AgentResponse{}
		},
	}
}

func NewAgentRequest(hash, parameter string) *AgentRequest {
	req := requestPool.Get().(*AgentRequest)
	req.Header.Magic = MAGIC
	req.Header.Command = REQUEST
	req.Header.Len = uint32(len(hash) + 1 + len(parameter))
	req.Body.Hash = hash
	req.Body.Parameter = parameter
	return req
}
func ReleaseAgentRequest(req *AgentRequest) {
	requestPool.Put(req)
}
func NewAgentResponse(body string) *AgentResponse {
	res := responsePool.Get().(*AgentResponse)
	res.Header.Magic = MAGIC
	res.Header.Command = RESPONSE
	res.Header.Len = uint32(len(body))
	res.Body = body
	return res
}
func ReleaseAgentResponse(res *AgentResponse) {
	responsePool.Put(res)
}

func bytes2short(bytes []byte, offset int) uint16 {
	var ans uint16 = 0
	list := make([]uint16, 2, 2)
	for i, b := range bytes[offset : offset+2] {
		list[i] = uint16(b)
	}
	ans += list[0] << 8
	ans += list[1] << 0
	return ans
}

func bytes2int(bytes []byte, offset int) uint32 {
	var ans uint32 = 0
	list := make([]uint32, 4, 4)
	for i, b := range bytes[offset : offset+4] {
		list[i] = uint32(b)
	}
	ans += list[0] << 24
	ans += list[1] << 16
	ans += list[2] << 8
	ans += list[3] << 0
	return ans
}

func short2bytes(bytes []byte, v uint16, offset int) {
	bytes[offset+1] = byte(v & 0xFF)
	bytes[offset] = byte(v >> 8 & 0xFF)
}

func int2bytes(bytes []byte, v uint32, offset int) {
	bytes[offset+3] = byte(v)
	bytes[offset+2] = byte(v >> 8)
	bytes[offset+1] = byte(v >> 16)
	bytes[offset] = byte(v >> 24)
}
