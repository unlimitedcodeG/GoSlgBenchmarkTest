package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	"GoSlgBenchmarkTest/internal/protocol"
	"GoSlgBenchmarkTest/internal/session"
)

// 命令行参数
var (
	listenAddr = flag.String("listen", ":8080", "代理监听地址")
	targetURL  = flag.String("target", "", "真实游戏服务器WebSocket地址")
	sessionID  = flag.String("session", "", "录制会话ID")
	verbose    = flag.Bool("verbose", false, "启用详细日志")
)

// ProxyConnection 代理连接
type ProxyConnection struct {
	clientConn *websocket.Conn // Unity客户端连接
	serverConn *websocket.Conn // 游戏服务器连接
	recorder   *session.SessionRecorder
	sessionID  string
	verbose    bool
}

// RecordingProxy 录制代理服务器
type RecordingProxy struct {
	listenAddr string
	targetURL  string
	upgrader   websocket.Upgrader
	recorder   *session.SessionRecorder
	verbose    bool
}

func main() {
	flag.Parse()

	fmt.Println("🎬 Unity录制代理服务器")
	fmt.Println("========================")
	fmt.Println()

	// 验证参数
	if *targetURL == "" {
		log.Fatal("❌ 必须指定目标服务器地址 (--target)")
	}

	if *sessionID == "" {
		*sessionID = fmt.Sprintf("proxy_session_%d", time.Now().Unix())
	}

	fmt.Printf("👂 监听地址: %s\n", *listenAddr)
	fmt.Printf("🎯 目标服务器: %s\n", *targetURL)
	fmt.Printf("📹 会话ID: %s\n", *sessionID)

	// 创建录制代理
	proxy := &RecordingProxy{
		listenAddr: *listenAddr,
		targetURL:  *targetURL,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有源
			},
		},
		recorder: session.NewSessionRecorder(*sessionID),
		verbose:  *verbose,
	}

	// 设置HTTP路由
	http.HandleFunc("/ws", proxy.handleWebSocket)
	http.HandleFunc("/status", proxy.handleStatus)
	http.HandleFunc("/start-recording", proxy.handleStartRecording)
	http.HandleFunc("/stop-recording", proxy.handleStopRecording)

	// 启动HTTP服务器
	server := &http.Server{
		Addr:    *listenAddr,
		Handler: nil,
	}

	go func() {
		fmt.Printf("🚀 代理服务器启动: http://%s\n", *listenAddr)
		fmt.Printf("📊 状态监控: http://%s/status\n", *listenAddr)
		fmt.Println()
		fmt.Println("💡 Unity客户端连接地址:")
		fmt.Printf("   ws://%s/ws\n", *listenAddr)
		fmt.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🔄 正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ 服务器关闭错误: %v", err)
	}

	// 导出录制数据
	fmt.Println("💾 导出录制数据...")
	proxy.exportRecording()

	fmt.Println("✅ 服务器已关闭")
}

// handleWebSocket 处理WebSocket连接
func (p *RecordingProxy) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级客户端连接
	clientConn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ 升级客户端连接失败: %v", err)
		return
	}
	defer clientConn.Close()

	if p.verbose {
		fmt.Printf("🔗 Unity客户端已连接: %s\n", r.RemoteAddr)
	}

	// 连接游戏服务器
	serverURL, err := url.Parse(p.targetURL)
	if err != nil {
		log.Printf("❌ 无效的服务器URL: %v", err)
		return
	}

	serverConn, _, err := websocket.DefaultDialer.Dial(p.targetURL, nil)
	if err != nil {
		log.Printf("❌ 连接游戏服务器失败: %v", err)
		return
	}
	defer serverConn.Close()

	if p.verbose {
		fmt.Printf("🔗 已连接游戏服务器: %s\n", serverURL.Host)
	}

	// 创建代理连接
	proxyConn := &ProxyConnection{
		clientConn: clientConn,
		serverConn: serverConn,
		recorder:   p.recorder,
		sessionID:  p.recorder.GetSession().ID,
		verbose:    p.verbose,
	}

	// 记录连接事件
	p.recorder.RecordEvent(session.EventConnect, map[string]interface{}{
		"client_addr": r.RemoteAddr,
		"server_url":  p.targetURL,
		"proxy_mode":  "recording",
	})

	// 启动双向代理
	proxyConn.startProxy()
}

// startProxy 启动双向代理
func (pc *ProxyConnection) startProxy() {
	// 创建取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动客户端到服务器的代理
	go func() {
		defer cancel()
		pc.proxyClientToServer(ctx)
	}()

	// 启动服务器到客户端的代理
	go func() {
		defer cancel()
		pc.proxyServerToClient(ctx)
	}()

	// 等待任一方向断开
	<-ctx.Done()

	if pc.verbose {
		fmt.Println("🔌 代理连接已断开")
	}

	// 记录断开事件
	pc.recorder.RecordEvent(session.EventDisconnect, map[string]interface{}{
		"reason": "proxy_disconnected",
	})
}

// proxyClientToServer 代理客户端到服务器的消息
func (pc *ProxyConnection) proxyClientToServer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 从客户端读取消息
		messageType, data, err := pc.clientConn.ReadMessage()
		if err != nil {
			if pc.verbose {
				fmt.Printf("❌ 读取客户端消息失败: %v\n", err)
			}
			return
		}

		// 记录客户端发送的消息
		pc.recordMessage("client_to_server", data)

		// 转发到服务器
		if err := pc.serverConn.WriteMessage(messageType, data); err != nil {
			if pc.verbose {
				fmt.Printf("❌ 转发到服务器失败: %v\n", err)
			}
			return
		}
	}
}

// proxyServerToClient 代理服务器到客户端的消息
func (pc *ProxyConnection) proxyServerToClient(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 从服务器读取消息
		messageType, data, err := pc.serverConn.ReadMessage()
		if err != nil {
			if pc.verbose {
				fmt.Printf("❌ 读取服务器消息失败: %v\n", err)
			}
			return
		}

		// 记录服务器发送的消息
		pc.recordMessage("server_to_client", data)

		// 转发到客户端
		if err := pc.clientConn.WriteMessage(messageType, data); err != nil {
			if pc.verbose {
				fmt.Printf("❌ 转发到客户端失败: %v\n", err)
			}
			return
		}
	}
}

// recordMessage 记录消息
func (pc *ProxyConnection) recordMessage(direction string, data []byte) {
	if len(data) < 2 {
		return // 消息太短，跳过
	}

	// 尝试解析协议帧
	opcode, body, err := protocol.DecodeFrame(data)
	if err != nil {
		// 如果不是协议帧，记录原始数据
		pc.recorder.RecordMessage(direction, data, 0, data, 0)
		return
	}

	// 记录协议消息
	sequenceNum := uint64(0) // 这里可以从消息中提取序列号
	pc.recorder.RecordMessage(direction, data, opcode, body, sequenceNum)

	// 记录事件
	eventType := session.EventMessageSend
	if direction == "server_to_client" {
		eventType = session.EventMessageReceive
	}

	pc.recorder.RecordEvent(eventType, map[string]interface{}{
		"direction":    direction,
		"opcode":       opcode,
		"message_size": len(data),
		"body_size":    len(body),
	})

	if pc.verbose {
		fmt.Printf("📝 记录消息: %s, opcode=%d, size=%d\n", direction, opcode, len(data))
	}
}

// handleStatus 处理状态查询
func (p *RecordingProxy) handleStatus(w http.ResponseWriter, r *http.Request) {
	session := p.recorder.GetSession()
	stats := p.recorder.GetStats()

	status := map[string]interface{}{
		"session_id":   session.ID,
		"start_time":   session.StartTime,
		"events_count": len(session.Events),
		"frames_count": len(session.Frames),
		"target_url":   p.targetURL,
		"listen_addr":  p.listenAddr,
		"recording":    true,
		"stats":        stats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleStartRecording 处理开始录制请求
func (p *RecordingProxy) handleStartRecording(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🎬 收到开始录制请求")

	response := map[string]interface{}{
		"status":     "recording",
		"session_id": p.recorder.GetSession().ID,
		"message":    "Recording already active",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStopRecording 处理停止录制请求
func (p *RecordingProxy) handleStopRecording(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🛑 收到停止录制请求")

	// 导出录制数据
	p.exportRecording()

	response := map[string]interface{}{
		"status":  "stopped",
		"message": "Recording data exported",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// exportRecording 导出录制数据
func (p *RecordingProxy) exportRecording() {
	recordedSession := p.recorder.GetSession()

	// 导出JSON
	jsonData, err := p.recorder.ExportJSON()
	if err != nil {
		log.Printf("❌ 导出JSON失败: %v", err)
		return
	}

	filename := fmt.Sprintf("recordings/proxy_session_%s.json", recordedSession.ID)
	if err := os.MkdirAll("recordings", 0755); err != nil {
		log.Printf("❌ 创建目录失败: %v", err)
		return
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		log.Printf("❌ 保存文件失败: %v", err)
		return
	}

	fmt.Printf("💾 录制数据已保存: %s (%d 字节)\n", filename, len(jsonData))
	fmt.Printf("📊 录制统计: %d 个事件, %d 个消息帧\n",
		len(recordedSession.Events), len(recordedSession.Frames))
}
