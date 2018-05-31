package log

import (
	"log"
	"os"
	"fmt"
)

// 此处需要优化的
func Init() {
	logFile, logErr := os.OpenFile("/root/logs/agent.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("Fail to find server start Failed")
		return
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
func Debug(iface ...interface{}) {
	//log.Println(iface)
}

func Info(iface ...interface{}) {
	log.Println(iface)
}

func Error(iface ...interface{}) {
	log.Println(iface)
}
