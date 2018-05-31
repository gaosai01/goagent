package dubbo

import (
	"io"
	"bytes"
	"errors"
	"sync"
	"goagent/tcp"
)

const (
	HEADER_LENGTH = 12 // 12 * 8 位
	// magic header
	MAGIC      = uint16(0xdabb)
	MAGIC_HIGH = byte(0xda)
	MAGIC_LOW  = byte(0xbb)
	// message flag.
	FLAG_REQUEST = byte(0x80)
	FLAG_TWOWAY  = byte(0x40)
	FLAG_EVENT   = byte(0x20) // for heartbeat
)

var (
	MSG_ERROR   = errors.New("eof数据不完整")
	SPLIT_ERROR = errors.New("Split粘包问题解决失败")
	mu          sync.Mutex
	dubboId     int64 = 0
)

type DubboRequest struct {
	Id                   int64
	DubboVersion         string
	ServiceName          string
	ServiceVersion       string
	Method               string
	MethodParameterTypes string
	MethodArgs           string
	Attachments          map[string]string
}

var dubboRequestPool *sync.Pool

func init() {
	dubboRequestPool = &sync.Pool{
		New: func() interface{} {
			return &DubboRequest{}
		},
	}
}
func NewDubboRequest() *DubboRequest {
	request := dubboRequestPool.Get().(*DubboRequest)
	mu.Lock()
	dubboId = dubboId + 1
	request.Id = dubboId
	mu.Unlock()
	request.DubboVersion = "2.0.1"
	request.ServiceVersion = ""
	request.Attachments = make(map[string]string)
	return request
}
func ReleaseDubboRequest(req *DubboRequest) {
	dubboRequestPool.Put(req)
}

func NewDubboPackageHandler() tcp.PackageHandler {
	headerPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, HEADER_LENGTH, HEADER_LENGTH)
		},
	}

	obj := DubboPackageHandler{}
	obj.headerPool = headerPool
	return &obj
}

type DubboPackageHandler struct {
	headerPool *sync.Pool
}

func (this *DubboPackageHandler) Pack(v interface{}, writer io.Writer) error {
	request := v.(*DubboRequest)
	//log.Debug("dubbo request", *request)
	header := this.headerPool.Get().([]byte)
	// 16位Magic
	header[0] = MAGIC_HIGH
	header[1] = MAGIC_LOW
	// Req/Res (1 bit)， 2 Way (1 bit)， Event (1 bit)， Serialization ID (5 bit)
	header[2] = 6
	header[2] |= FLAG_REQUEST
	header[2] |= FLAG_TWOWAY
	// Request ID (64 bits)
	long2bytes(header, request.Id, 4)
	bytesBuf := bytes.NewBuffer(make([]byte, 0))
	bytesBuf.Write(header)
	this.headerPool.Put(header)
	// 开始解释Variable Part部分, 长度+data
	var buf bytes.Buffer
	buf.WriteString(toJson(request.DubboVersion))
	buf.WriteString("\n")
	buf.WriteString(toJson(request.ServiceName))
	buf.WriteString("\n")
	//buf.WriteString(toJson(request.ServiceVersion))
	buf.WriteString(toNull())
	buf.WriteString("\n")
	buf.WriteString(toJson(request.Method))
	buf.WriteString("\n")
	buf.WriteString(toJson(request.MethodParameterTypes))
	buf.WriteString("\n")
	buf.WriteString(toJson(request.MethodArgs))
	buf.WriteString("\n")
	buf.Write(toBytes(request.Attachments))
	buf.WriteString("\n")
	buf.WriteString("\n")
	data := buf.Bytes()
	// 长度
	size := make([]byte, 4, 4)
	int2bytes(size, int32(len(data)), 0)
	bytesBuf.Write(size)
	// data
	bytesBuf.Write(data)
	_, err := bytesBuf.WriteTo(writer)
	if err == nil {
		return nil
	}
	return err
}
func (this *DubboPackageHandler) Split(data []byte, eof bool) (int, []byte, error) {
	if len(data) <= HEADER_LENGTH+4 {
		return 0, nil, nil
	}
	blen := bytes2short(data[HEADER_LENGTH : HEADER_LENGTH+4])
	alllen := HEADER_LENGTH + 4 + blen
	if len(data) < alllen {
		if eof {
			return 0, nil, MSG_ERROR
		}
		return 0, nil, nil
	}
	return alllen, data[:alllen], nil

}
func (this *DubboPackageHandler) UnPack(data []byte) (interface{}, error) {
	blen := bytes2short(data[HEADER_LENGTH : HEADER_LENGTH+4])
	alllen := HEADER_LENGTH + 4 + blen
	ans := string(data[HEADER_LENGTH+6 : alllen-1])
	return ans, nil
}

func (this *DubboPackageHandler) Release(v interface{}) {

}
