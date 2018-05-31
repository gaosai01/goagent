package dubbo

import (
	"bytes"
	"encoding/json"
)

func short2bytes(bytes []byte, v uint16, offset int) {
	bytes[offset+1] = byte(v & 0xFF)
	bytes[offset] = byte(v >> 8 & 0xFF)
}

func int2bytes(bytes []byte, v int32, offset int) {
	bytes[offset+3] = byte(v)
	bytes[offset+2] = byte(v >> 8)
	bytes[offset+1] = byte(v >> 16)
	bytes[offset] = byte(v >> 24)
}

func long2bytes(bytes []byte, v int64, offset int) {
	bytes[offset+7] = byte(v)
	bytes[offset+6] = byte(v >> 8)
	bytes[offset+5] = byte(v >> 16)
	bytes[offset+4] = byte(v >> 24)
	bytes[offset+3] = byte(v >> 32)
	bytes[offset+2] = byte(v >> 40)
	bytes[offset+1] = byte(v >> 48)
	bytes[offset] = byte(v >> 56)
}

func bytes2short(bytes []byte) int {
	var ans int32 = 0
	list := make([]int32, 4, 4)
	for i, b := range bytes[:4] {
		list[i] = int32(b)
	}
	ans += list[0] << 24
	ans += list[1] << 16
	ans += list[2] << 8
	ans += list[3]
	return int(ans)
}

func toJson(str string) string {
	var buf bytes.Buffer
	buf.WriteString("\"")
	buf.WriteString(str)
	buf.WriteString("\"")
	return buf.String()
}
func toNull() string {
	return "null"
}

func toBytes(m map[string]string) []byte {
	bs, _ := json.Marshal(m)
	return bs
}
