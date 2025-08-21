package logger

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LogMessage 日志消息结构
type LogMessage struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Module    string    `json:"module"`
	TestID    *int64    `json:"test_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// WebSocketLogger WebSocket日志广播器
type WebSocketLogger struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan LogMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

// NewWebSocketLogger 创建新的WebSocket日志器
func NewWebSocketLogger() *WebSocketLogger {
	return &WebSocketLogger{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan LogMessage, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 启动WebSocket日志器
func (wsl *WebSocketLogger) Run() {
	for {
		select {
		case client := <-wsl.register:
			wsl.mu.Lock()
			wsl.clients[client] = true
			wsl.mu.Unlock()
			log.Printf("WebSocket客户端已连接，当前连接数: %d", len(wsl.clients))

		case client := <-wsl.unregister:
			wsl.mu.Lock()
			if _, ok := wsl.clients[client]; ok {
				delete(wsl.clients, client)
				client.Close()
				wsl.mu.Unlock()
				log.Printf("WebSocket客户端已断开，当前连接数: %d", len(wsl.clients))
			} else {
				wsl.mu.Unlock()
			}

		case message := <-wsl.broadcast:
			wsl.mu.RLock()
			for client := range wsl.clients {
				select {
				case <-time.After(time.Second):
					wsl.mu.RUnlock()
					wsl.mu.Lock()
					delete(wsl.clients, client)
					client.Close()
					wsl.mu.Unlock()
					wsl.mu.RLock()
				default:
					if err := client.WriteJSON(message); err != nil {
						log.Printf("发送日志消息失败: %v", err)
						wsl.mu.RUnlock()
						wsl.mu.Lock()
						delete(wsl.clients, client)
						client.Close()
						wsl.mu.Unlock()
						wsl.mu.RLock()
					}
				}
			}
			wsl.mu.RUnlock()
		}
	}
}

// LogInfo 记录信息日志
func (wsl *WebSocketLogger) LogInfo(module, message string, testID *int64) {
	logMsg := LogMessage{
		Level:     "INFO",
		Message:   message,
		Module:    module,
		TestID:    testID,
		Timestamp: time.Now(),
	}

	// 同时输出到控制台
	if testID != nil {
		log.Printf("[%s] [Test-%d] %s: %s", logMsg.Level, *testID, module, message)
	} else {
		log.Printf("[%s] %s: %s", logMsg.Level, module, message)
	}

	select {
	case wsl.broadcast <- logMsg:
	default:
		// 如果通道满了，丢弃消息避免阻塞
	}
}

// LogError 记录错误日志
func (wsl *WebSocketLogger) LogError(module, message string, testID *int64) {
	logMsg := LogMessage{
		Level:     "ERROR",
		Message:   message,
		Module:    module,
		TestID:    testID,
		Timestamp: time.Now(),
	}

	// 同时输出到控制台
	if testID != nil {
		log.Printf("[%s] [Test-%d] %s: %s", logMsg.Level, *testID, module, message)
	} else {
		log.Printf("[%s] %s: %s", logMsg.Level, module, message)
	}

	select {
	case wsl.broadcast <- logMsg:
	default:
	}
}

// LogSuccess 记录成功日志
func (wsl *WebSocketLogger) LogSuccess(module, message string, testID *int64) {
	logMsg := LogMessage{
		Level:     "SUCCESS",
		Message:   message,
		Module:    module,
		TestID:    testID,
		Timestamp: time.Now(),
	}

	// 同时输出到控制台
	if testID != nil {
		log.Printf("[%s] [Test-%d] %s: %s", logMsg.Level, *testID, module, message)
	} else {
		log.Printf("[%s] %s: %s", logMsg.Level, module, message)
	}

	select {
	case wsl.broadcast <- logMsg:
	default:
	}
}

// LogWarning 记录警告日志
func (wsl *WebSocketLogger) LogWarning(module, message string, testID *int64) {
	logMsg := LogMessage{
		Level:     "WARNING",
		Message:   message,
		Module:    module,
		TestID:    testID,
		Timestamp: time.Now(),
	}

	// 同时输出到控制台
	if testID != nil {
		log.Printf("[%s] [Test-%d] %s: %s", logMsg.Level, *testID, module, message)
	} else {
		log.Printf("[%s] %s: %s", logMsg.Level, module, message)
	}

	select {
	case wsl.broadcast <- logMsg:
	default:
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// HandleWebSocket 处理WebSocket连接
func (wsl *WebSocketLogger) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 注册客户端
	wsl.register <- conn

	// 发送欢迎消息
	welcomeMsg := LogMessage{
		Level:     "INFO",
		Message:   "已连接到Unity SLG测试平台日志流",
		Module:    "WebSocket",
		Timestamp: time.Now(),
	}
	conn.WriteJSON(welcomeMsg)

	// 处理客户端断开
	defer func() {
		wsl.unregister <- conn
		conn.Close()
	}()

	// 保持连接活跃
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket连接错误: %v", err)
			}
			break
		}
	}
}

// 全局日志器实例
var GlobalLogger *WebSocketLogger

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger() {
	GlobalLogger = NewWebSocketLogger()
	go GlobalLogger.Run()
}

// 便捷函数
func LogInfo(module, message string, testID *int64) {
	if GlobalLogger != nil {
		GlobalLogger.LogInfo(module, message, testID)
	}
}

func LogError(module, message string, testID *int64) {
	if GlobalLogger != nil {
		GlobalLogger.LogError(module, message, testID)
	}
}

func LogSuccess(module, message string, testID *int64) {
	if GlobalLogger != nil {
		GlobalLogger.LogSuccess(module, message, testID)
	}
}

func LogWarning(module, message string, testID *int64) {
	if GlobalLogger != nil {
		GlobalLogger.LogWarning(module, message, testID)
	}
}
