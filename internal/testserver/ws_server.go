package testserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/protocol"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// ServerConfig 测试服务器配置
type ServerConfig struct {
	Addr                   string
	PushInterval           time.Duration // 推送间隔
	EnableBattlePush       bool          // 是否启用战斗推送
	EnableRandomDisconnect bool          // 是否随机断连
	DisconnectProbability  float64       // 断连概率 (0.0-1.0)
	MaxConnections         int           // 最大连接数
	ReadBufferSize         int
	WriteBufferSize        int
	EnableCompression      bool
}

// DefaultServerConfig 返回默认配置
func DefaultServerConfig(addr string) *ServerConfig {
	return &ServerConfig{
		Addr:                   addr,
		PushInterval:           100 * time.Millisecond,
		EnableBattlePush:       true,
		EnableRandomDisconnect: false,
		DisconnectProbability:  0.01, // 1%概率断连
		MaxConnections:         1000,
		ReadBufferSize:         1024,
		WriteBufferSize:        1024,
		EnableCompression:      true,
	}
}

// ConnectionStats 连接统计信息
type ConnectionStats struct {
	ConnectedAt      time.Time
	MessagesReceived atomic.Uint64
	MessagesSent     atomic.Uint64
	LastActivity     atomic.Int64 // unix nano
	BytesReceived    atomic.Uint64
	BytesSent        atomic.Uint64
}

// Connection 表示一个WebSocket连接
type Connection struct {
	ID       string
	Conn     *websocket.Conn
	PlayerID string
	Stats    *ConnectionStats

	// 控制标志
	stopChan  chan struct{}
	closeOnce sync.Once
	mu        sync.RWMutex
}

// safeClose 安全关闭连接的stopChan
func (c *Connection) safeClose() {
	c.closeOnce.Do(func() {
		close(c.stopChan)
	})
}

// Server 测试用WebSocket服务器
type Server struct {
	config   *ServerConfig
	server   *http.Server
	upgrader websocket.Upgrader

	// 连接管理
	connections sync.Map // map[string]*Connection
	connCount   atomic.Int32
	connWg      sync.WaitGroup // 等待所有连接goroutine退出

	// 后台任务管理
	bgWg   sync.WaitGroup // 等待后台goroutine退出
	stopCh chan struct{}  // 停止信号

	// 序列号生成器
	seqGenerator atomic.Uint64

	// 控制标志
	forceDisconnect atomic.Bool
	isRunning       atomic.Bool

	// 统计信息
	totalConnections atomic.Uint64
	totalMessages    atomic.Uint64
	startTime        time.Time

	mu sync.RWMutex
}

// New 创建新的测试服务器
func New(config *ServerConfig) *Server {
	if config == nil {
		config = DefaultServerConfig(":8080")
	}

	server := &Server{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:    config.ReadBufferSize,
			WriteBufferSize:   config.WriteBufferSize,
			EnableCompression: config.EnableCompression,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有源
			},
		},
		stopCh:    make(chan struct{}),
		startTime: time.Now(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handleWebSocket)
	mux.HandleFunc("/stats", server.handleStats)
	mux.HandleFunc("/control", server.handleControl)

	server.server = &http.Server{
		Addr:    config.Addr,
		Handler: mux,
	}

	return server
}

// Start 启动服务器
func (s *Server) Start() error {
	if !s.isRunning.CompareAndSwap(false, true) {
		return fmt.Errorf("server is already running")
	}

	log.Printf("Starting test server on %s", s.config.Addr)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// 给服务器足够的时间启动
	// 不要通过连接测试，因为这可能干扰后续的客户端连接
	time.Sleep(200 * time.Millisecond)

	// 启动推送任务
	if s.config.EnableBattlePush {
		s.bgWg.Add(1)
		go s.battlePushLoop()
	}

	return nil
}

// StartTLS 启动TLS服务器
func (s *Server) StartTLS(cert tls.Certificate) error {
	if !s.isRunning.CompareAndSwap(false, true) {
		return fmt.Errorf("server is already running")
	}

	s.server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	log.Printf("Starting TLS test server on %s", s.config.Addr)

	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	if s.config.EnableBattlePush {
		s.bgWg.Add(1)
		go s.battlePushLoop()
	}

	return nil
}

// Shutdown 关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	if !s.isRunning.CompareAndSwap(true, false) {
		return nil
	}

	log.Printf("Shutting down test server...")

	// 发送停止信号给后台任务
	close(s.stopCh)

	// 关闭所有连接
	s.connections.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)
		s.closeConnection(conn, "Server shutdown")
		return true
	})

	// 等待所有连接处理goroutine退出
	s.connWg.Wait()

	// 等待所有后台goroutine退出
	s.bgWg.Wait()

	return s.server.Shutdown(ctx)
}

// ForceDisconnectAll 强制断开所有连接
func (s *Server) ForceDisconnectAll() {
	s.forceDisconnect.Store(true)
	log.Printf("Force disconnecting all connections")

	s.connections.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)
		s.closeConnection(conn, "Force disconnect")
		return true
	})

	s.forceDisconnect.Store(false)
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if s.connCount.Load() >= int32(s.config.MaxConnections) {
		http.Error(w, "Too many connections", http.StatusServiceUnavailable)
		return
	}

	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	connID := fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), s.totalConnections.Add(1))
	conn := &Connection{
		ID:       connID,
		Conn:     wsConn,
		Stats:    &ConnectionStats{ConnectedAt: time.Now()},
		stopChan: make(chan struct{}),
	}
	conn.Stats.LastActivity.Store(time.Now().UnixNano())

	s.connections.Store(connID, conn)
	s.connCount.Add(1)

	log.Printf("New connection: %s from %s", connID, r.RemoteAddr)

	// 处理连接
	s.handleConnection(conn)
}

// handleConnection 处理单个连接的生命周期
func (s *Server) handleConnection(conn *Connection) {
	s.connWg.Add(1)
	defer func() {
		s.closeConnection(conn, "Connection ended")
		s.connWg.Done()
	}()

	// 等待登录
	if !s.handleLogin(conn) {
		return
	}

	// 启动消息处理循环
	s.connWg.Add(1)
	go s.messageReadLoop(conn)

	// 主循环：处理推送和断连检查
	ticker := time.NewTicker(s.config.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-conn.stopChan:
			return
		case <-ticker.C:
			// 检查是否需要随机断连
			if s.config.EnableRandomDisconnect && s.shouldDisconnect() {
				log.Printf("Random disconnect: %s", conn.ID)
				return
			}

			// 检查强制断连
			if s.forceDisconnect.Load() {
				return
			}
		}
	}
}

// handleLogin 处理登录流程
func (s *Server) handleLogin(conn *Connection) bool {
	conn.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	messageType, rawData, err := conn.Conn.ReadMessage()
	if err != nil {
		log.Printf("Read login message failed: %v", err)
		return false
	}

	if messageType != websocket.BinaryMessage {
		log.Printf("Expected binary message for login")
		return false
	}

	opcode, body, err := protocol.DecodeFrame(rawData)
	if err != nil {
		log.Printf("Decode login frame failed: %v", err)
		return false
	}

	if opcode != protocol.OpLoginReq {
		log.Printf("Expected login request, got opcode: %d", opcode)
		return false
	}

	loginReq := &gamev1.LoginReq{}
	if err := proto.Unmarshal(body, loginReq); err != nil {
		log.Printf("Unmarshal login request failed: %v", err)
		return false
	}

	// 简单验证（在真实环境中应该验证token）
	playerID := fmt.Sprintf("player_%s_%d", loginReq.DeviceId, time.Now().Unix())

	// 使用锁保护PlayerID的写入，避免与广播的读取产生数据竞争
	conn.mu.Lock()
	conn.PlayerID = playerID
	conn.mu.Unlock()

	// 发送登录响应
	loginResp := &gamev1.LoginResp{
		Ok:         true,
		PlayerId:   playerID,
		SessionId:  fmt.Sprintf("session_%s", conn.ID),
		ServerTime: time.Now().UnixMilli(),
	}

	if err := s.sendMessage(conn, protocol.OpLoginResp, loginResp); err != nil {
		log.Printf("Send login response failed: %v", err)
		return false
	}

	log.Printf("Login successful: %s -> %s", conn.ID, playerID)
	return true
}

// messageReadLoop 消息读取循环
func (s *Server) messageReadLoop(conn *Connection) {
	defer func() {
		conn.safeClose()
		s.connWg.Done()
	}()

	conn.Conn.SetReadLimit(512 * 1024) // 512KB限制

	for {
		select {
		case <-conn.stopChan:
			return
		default:
			conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			messageType, rawData, err := conn.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Connection read error: %v", err)
				}
				return
			}

			conn.Stats.MessagesReceived.Add(1)
			conn.Stats.BytesReceived.Add(uint64(len(rawData)))
			conn.Stats.LastActivity.Store(time.Now().UnixNano())
			s.totalMessages.Add(1)

			if messageType == websocket.PingMessage {
				conn.Conn.WriteMessage(websocket.PongMessage, rawData)
				continue
			}

			if messageType != websocket.BinaryMessage {
				continue
			}

			s.handleMessage(conn, rawData)
		}
	}
}

// handleMessage 处理接收到的消息
func (s *Server) handleMessage(conn *Connection, rawData []byte) {
	opcode, body, err := protocol.DecodeFrame(rawData)
	if err != nil {
		log.Printf("Decode frame failed: %v", err)
		return
	}

	switch opcode {
	case protocol.OpHeartbeat:
		s.handleHeartbeat(conn, body)
	case protocol.OpPlayerAction:
		s.handlePlayerAction(conn, body)
	default:
		log.Printf("Unknown opcode: %d", opcode)
	}
}

// handleHeartbeat 处理心跳消息
func (s *Server) handleHeartbeat(conn *Connection, body []byte) {
	heartbeat := &gamev1.Heartbeat{}
	if err := proto.Unmarshal(body, heartbeat); err != nil {
		log.Printf("Unmarshal heartbeat failed: %v", err)
		return
	}

	now := time.Now()
	clientTime := time.UnixMilli(heartbeat.ClientUnixMs)
	rtt := now.Sub(clientTime)

	resp := &gamev1.HeartbeatResp{
		ServerUnixMs: now.UnixMilli(),
		PingSeq:      heartbeat.PingSeq,
		RttMs:        int32(rtt.Milliseconds()),
	}

	s.sendMessage(conn, protocol.OpHeartbeatResp, resp)
}

// handlePlayerAction 处理玩家操作
func (s *Server) handlePlayerAction(conn *Connection, body []byte) {
	action := &gamev1.PlayerAction{}
	if err := proto.Unmarshal(body, action); err != nil {
		log.Printf("Unmarshal player action failed: %v", err)
		return
	}

	log.Printf("Received action from %s: type=%v, seq=%d",
		conn.PlayerID, action.ActionType, action.ActionSeq)

	// 这里可以添加游戏逻辑处理
	// 暂时直接回复成功
	resp := &gamev1.PlayerAction{
		ActionSeq:       action.ActionSeq,
		PlayerId:        conn.PlayerID,
		ActionType:      action.ActionType,
		ClientTimestamp: action.ClientTimestamp,
	}

	s.sendMessage(conn, protocol.OpActionResp, resp)
}

// battlePushLoop 战斗推送循环
func (s *Server) battlePushLoop() {
	defer s.bgWg.Done()

	ticker := time.NewTicker(s.config.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if s.connCount.Load() == 0 {
				continue
			}

			seq := s.seqGenerator.Add(1)
			battlePush := &gamev1.BattlePush{
				Seq:       seq,
				BattleId:  fmt.Sprintf("battle_%d", seq/100), // 每100个消息一个战斗
				StateHash: []byte{byte(seq), byte(seq >> 8), byte(seq >> 16)},
				Units: []*gamev1.BattleUnit{
					{
						UnitId: fmt.Sprintf("unit_%d", seq%10),
						Hp:     int32(100 - (seq % 100)),
						Mp:     int32(50 + (seq % 50)),
						Position: &gamev1.Position{
							X: float32(seq % 100),
							Y: float32((seq * 2) % 100),
							Z: 0,
						},
						Status: gamev1.UnitStatus(seq%4 + 1),
					},
				},
				Timestamp: time.Now().UnixMilli(),
			}

			// 广播给所有连接
			s.broadcastMessage(protocol.OpBattlePush, battlePush)
		}
	}
}

// sendMessage 发送消息给指定连接
func (s *Server) sendMessage(conn *Connection, opcode uint16, message proto.Message) error {
	body, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	frame := protocol.EncodeFrame(opcode, body)

	conn.mu.Lock()
	defer conn.mu.Unlock()

	conn.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err = conn.Conn.WriteMessage(websocket.BinaryMessage, frame)
	if err == nil {
		conn.Stats.MessagesSent.Add(1)
		conn.Stats.BytesSent.Add(uint64(len(frame)))
	}

	return err
}

// broadcastMessage 广播消息给所有连接
func (s *Server) broadcastMessage(opcode uint16, message proto.Message) {
	body, err := proto.Marshal(message)
	if err != nil {
		log.Printf("Marshal broadcast message failed: %v", err)
		return
	}

	frame := protocol.EncodeFrame(opcode, body)

	// 收集需要关闭的连接，避免在Range过程中修改map
	var failedConns []*Connection

	s.connections.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)

		conn.mu.Lock()
		// 只向已认证的连接推送消息
		if conn.PlayerID == "" {
			conn.mu.Unlock()
			return true // 跳过未认证的连接
		}

		conn.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		err := conn.Conn.WriteMessage(websocket.BinaryMessage, frame)
		if err == nil {
			conn.Stats.MessagesSent.Add(1)
			conn.Stats.BytesSent.Add(uint64(len(frame)))
		}
		conn.mu.Unlock()

		if err != nil {
			log.Printf("Broadcast to %s failed: %v", conn.ID, err)
			failedConns = append(failedConns, conn)
		}

		return true
	})

	// 在Range完成后关闭失败的连接
	for _, conn := range failedConns {
		s.closeConnection(conn, "Broadcast failed")
	}
}

// closeConnection 关闭连接
func (s *Server) closeConnection(conn *Connection, reason string) {
	s.connections.Delete(conn.ID)
	s.connCount.Add(-1)

	conn.mu.Lock()
	if conn.Conn != nil {
		conn.Conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason),
			time.Now().Add(time.Second))
		conn.Conn.Close()
	}
	conn.mu.Unlock()

	select {
	case <-conn.stopChan:
	default:
		conn.safeClose()
	}

	log.Printf("Connection closed: %s, reason: %s", conn.ID, reason)
}

// shouldDisconnect 判断是否应该随机断连
func (s *Server) shouldDisconnect() bool {
	// 简单的随机断连逻辑
	return time.Now().UnixNano()%1000 < int64(s.config.DisconnectProbability*1000)
}

// handleStats 处理统计信息请求
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.GetStats()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
        "running": %t,
        "uptime_seconds": %.1f,
        "current_connections": %d,
        "total_connections": %d,
        "total_messages": %d,
        "sequence_number": %d
    }`,
		stats["running"],
		stats["uptime_seconds"],
		stats["current_connections"],
		stats["total_connections"],
		stats["total_messages"],
		stats["sequence_number"])
}

// handleControl 处理控制命令
func (s *Server) handleControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action := r.URL.Query().Get("action")
	switch action {
	case "disconnect_all":
		s.ForceDisconnectAll()
		fmt.Fprintf(w, "Disconnected all connections")
	case "toggle_push":
		// 这里可以添加开关推送的逻辑
		fmt.Fprintf(w, "Push toggled")
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

// GetStats 获取服务器统计信息
func (s *Server) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":             s.isRunning.Load(),
		"uptime_seconds":      time.Since(s.startTime).Seconds(),
		"current_connections": s.connCount.Load(),
		"total_connections":   s.totalConnections.Load(),
		"total_messages":      s.totalMessages.Load(),
		"sequence_number":     s.seqGenerator.Load(),
	}
}

// GetConnectionStats 获取连接统计信息
func (s *Server) GetConnectionStats() map[string]*ConnectionStats {
	stats := make(map[string]*ConnectionStats)

	s.connections.Range(func(key, value interface{}) bool {
		connID := key.(string)
		conn := value.(*Connection)
		stats[connID] = conn.Stats
		return true
	})

	return stats
}
