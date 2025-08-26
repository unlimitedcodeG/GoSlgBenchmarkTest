package logger

import "log"

// InitLogger 初始化日志器
func InitLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Logger initialized")
}
