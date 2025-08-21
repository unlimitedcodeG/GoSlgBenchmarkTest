package wsclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/protocol"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// ClientState 客户端连接状态
type ClientState int32

const (
	StateDisconnected ClientState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateClosed
)

func (s ClientState) String() string {
	switch s {
	case StateDisconnected:
		return "DISCONNECTED"
	case StateConnecting:
		return "CONNECTING"
	case StateConnected:
		return "CONNECTED"
	case StateReconnecting:
		return "RECONNECTING"
	case StateClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

// PushHandler 推送消息处理器
type PushHandler func(opcode uint16, message proto.Message)

// StateChangeHandler 状态变化处理器
type StateChangeHandler func(oldState, newState ClientState)

// RTTHandler RTT变化处理器
type RTTHandler func(rtt time.Duration)

// ClientConfig 客户端配置
type ClientConfig struct {
	URL               string
	Token             string
	ClientVersion     string
	DeviceID          string
	HandshakeTimeout  time.Duration
	HeartbeatInterval time.Duration
	PingTimeout       time.Duration
	ReconnectInterval time.Duration
	MaxReconnectTries int
	EnableCompression bool
	UserAgent         string
}

// DefaultClientConfig 返回默认配置
func DefaultClientConfig(url, token string) *ClientConfig {
	return &ClientConfig{
		URL:               url,
		Token:             token,
		ClientVersion:     "1.0.0",
		DeviceID:          "test-device",
		HandshakeTimeout:  10 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		PingTimeout:       5 * time.Second,
		ReconnectInterval: 2 * time.Second,
		MaxReconnectTries: 10,
		EnableCompression: true,
		UserAgent:         "GoSlgBenchmarkTest/1.0",
	}
}

// Client WebSocket客户端，支持自动重连、心跳、消息去重
type Client struct {
	config *ClientConfig
	dialer *websocket.Dialer
	conn   *websocket.Conn
	state  atomic.Int32

	// 消息处理
	onPush        PushHandler
	onStateChange StateChangeHandler
	onRTT         RTTHandler

	// 同步控制
	mu            sync.RWMutex
	writeMu       sync.Mutex // 专用于WebSocket写入同步
	stopChan      chan struct{}
	reconnectChan chan struct{}

	// 序列号管理（用于消息去重）
	lastSeq atomic.Uint64

	// 心跳和RTT统计
	lastPingSeq  atomic.Int32
	lastPingTime atomic.Int64 // unix nano
	avgRTT       atomic.Int64 // nano seconds

	// 重连控制
	reconnectCount atomic.Int32
	reconnects     atomic.Int32 // 重连次数统计

	// 帧解码器
	frameDecoder *protocol.FrameDecoder
}

// New 创建新的WebSocket客户端
func New(config *ClientConfig) *Client {
	if config == nil {
		panic("config cannot be nil")
	}

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = config.HandshakeTimeout
	dialer.EnableCompression = config.EnableCompression

	client := &Client{
		config:        config,
		dialer:        &dialer,
		stopChan:      make(chan struct{}),
		reconnectChan: make(chan struct{}, 1),
		frameDecoder:  protocol.NewFrameDecoder(),
	}

	client.setState(StateDisconnected)
	return client
}

// SetPushHandler 设置推送消息处理器
func (c *Client) SetPushHandler(handler PushHandler) {
	c.onPush = handler
}

// SetStateChangeHandler 设置状态变化处理器
func (c *Client) SetStateChangeHandler(handler StateChangeHandler) {
	c.onStateChange = handler
}

// SetRTTHandler 设置RTT变化处理器
func (c *Client) SetRTTHandler(handler RTTHandler) {
	c.onRTT = handler
}

// Connect 连接到服务器
func (c *Client) Connect(ctx context.Context) error {
	if !c.compareAndSwapState(StateDisconnected, StateConnecting) {
		return errors.New("client is not in disconnected state")
	}

	if err := c.doConnect(ctx); err != nil {
		c.setState(StateDisconnected)
		return err
	}

	c.setState(StateConnected)

	// 启动后台任务
	go c.heartbeatLoop()
	go c.readLoop()
	go c.reconnectLoop()

	return nil
}

// doConnect 执行实际的连接逻辑
func (c *Client) doConnect(ctx context.Context) error {
	headers := http.Header{
		"User-Agent": []string{c.config.UserAgent},
	}

	conn, resp, err := c.dialer.DialContext(ctx, c.config.URL, headers)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// 执行登录握手
	return c.doLogin(ctx)
}

// doLogin 执行登录流程
func (c *Client) doLogin(ctx context.Context) error {
	loginReq := &gamev1.LoginReq{
		Token:         c.config.Token,
		ClientVersion: c.config.ClientVersion,
		DeviceId:      c.config.DeviceID,
	}

	// 发送登录请求
	if err := c.sendMessage(protocol.OpLoginReq, loginReq); err != nil {
		return fmt.Errorf("send login request failed: %w", err)
	}

	// 等待登录响应
	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.HandshakeTimeout)
	defer cancel()

	opcode, message, err := c.readMessage(timeoutCtx)
	if err != nil {
		return fmt.Errorf("read login response failed: %w", err)
	}

	if opcode != protocol.OpLoginResp {
		return fmt.Errorf("unexpected opcode for login response: %d", opcode)
	}

	loginResp, ok := message.(*gamev1.LoginResp)
	if !ok {
		return errors.New("invalid login response message type")
	}

	if !loginResp.Ok {
		return fmt.Errorf("login failed: player_id=%s", loginResp.PlayerId)
	}

	log.Printf("Login successful: player_id=%s, session_id=%s",
		loginResp.PlayerId, loginResp.SessionId)

	return nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	if !c.compareAndSwapState(StateConnected, StateClosed) &&
		!c.compareAndSwapState(StateReconnecting, StateClosed) &&
		!c.compareAndSwapState(StateDisconnected, StateClosed) {
		return nil // 已经关闭
	}

	close(c.stopChan)

	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		return conn.Close()
	}

	return nil
}

// SendAction 发送玩家操作
func (c *Client) SendAction(action *gamev1.PlayerAction) error {
	if c.getState() != StateConnected {
		return errors.New("client is not connected")
	}

	return c.sendMessage(protocol.OpPlayerAction, action)
}

// sendMessage 发送protobuf消息
func (c *Client) sendMessage(opcode uint16, message proto.Message) error {
	body, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	frame := protocol.EncodeFrame(opcode, body)

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("connection is nil")
	}

	// 使用专用的写入锁防止并发写入
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteMessage(websocket.BinaryMessage, frame)
}

// readMessage 读取单个消息
func (c *Client) readMessage(ctx context.Context) (uint16, proto.Message, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return 0, nil, errors.New("connection is nil")
	}

	// 设置读取超时
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetReadDeadline(deadline)
	}

	messageType, rawData, err := conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}

	if messageType != websocket.BinaryMessage {
		return 0, nil, errors.New("received non-binary message")
	}

	opcode, body, err := protocol.DecodeFrame(rawData)
	if err != nil {
		return 0, nil, fmt.Errorf("decode frame failed: %w", err)
	}

	message, err := c.unmarshalMessage(opcode, body)
	if err != nil {
		return 0, nil, fmt.Errorf("unmarshal message failed: %w", err)
	}

	return opcode, message, nil
}

// unmarshalMessage 根据操作码反序列化消息
func (c *Client) unmarshalMessage(opcode uint16, body []byte) (proto.Message, error) {
	var message proto.Message

	switch opcode {
	case protocol.OpLoginResp:
		message = &gamev1.LoginResp{}
	case protocol.OpHeartbeatResp:
		message = &gamev1.HeartbeatResp{}
	case protocol.OpBattlePush:
		message = &gamev1.BattlePush{}
	case protocol.OpActionResp:
		message = &gamev1.PlayerAction{} // 这里应该定义专门的响应消息
	case protocol.OpError:
		message = &gamev1.ErrorResp{}
	default:
		return nil, fmt.Errorf("unknown opcode: %d", opcode)
	}

	if len(body) > 0 {
		if err := proto.Unmarshal(body, message); err != nil {
			return nil, err
		}
	}

	return message, nil
}

// heartbeatLoop 心跳循环
func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if c.getState() == StateConnected {
				c.sendHeartbeat()
				c.checkPing()
			}
		}
	}
}

// sendHeartbeat 发送心跳
func (c *Client) sendHeartbeat() {
	seq := c.lastPingSeq.Add(1)
	now := time.Now()
	c.lastPingTime.Store(now.UnixNano())

	heartbeat := &gamev1.Heartbeat{
		ClientUnixMs: now.UnixMilli(),
		PingSeq:      seq,
	}

	if err := c.sendMessage(protocol.OpHeartbeat, heartbeat); err != nil {
		log.Printf("Send heartbeat failed: %v", err)
		c.triggerReconnect()
	}
}

// checkPing 检查ping超时
func (c *Client) checkPing() {
	lastPingTime := time.Unix(0, c.lastPingTime.Load())
	if time.Since(lastPingTime) > c.config.PingTimeout {
		log.Printf("Ping timeout, triggering reconnect")
		c.triggerReconnect()
	}
}

// readLoop 消息读取循环
func (c *Client) readLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			if c.getState() != StateConnected {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			opcode, message, err := c.readMessage(context.Background())
			if err != nil {
				log.Printf("Read message failed: %v", err)
				c.triggerReconnect()
				continue
			}

			c.handleMessage(opcode, message)
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(opcode uint16, message proto.Message) {
	switch opcode {
	case protocol.OpHeartbeatResp:
		c.handleHeartbeatResp(message.(*gamev1.HeartbeatResp))
	case protocol.OpBattlePush:
		c.handleBattlePush(message.(*gamev1.BattlePush))
	default:
		if c.onPush != nil {
			c.onPush(opcode, message)
		}
	}
}

// handleHeartbeatResp 处理心跳响应
func (c *Client) handleHeartbeatResp(resp *gamev1.HeartbeatResp) {
	pingTime := time.Unix(0, c.lastPingTime.Load())
	if pingTime.IsZero() {
		return // 没有发送过心跳
	}

	rtt := time.Since(pingTime)
	if rtt <= 0 {
		return // 无效的RTT
	}

	// 更新平均RTT（简单移动平均）
	oldAvg := time.Duration(c.avgRTT.Load())
	newAvg := (oldAvg + rtt) / 2
	c.avgRTT.Store(int64(newAvg))

	if c.onRTT != nil {
		c.onRTT(rtt)
	}
}

// handleBattlePush 处理战斗推送（带去重）
func (c *Client) handleBattlePush(push *gamev1.BattlePush) {
	// 序列号去重：要求单调递增
	lastSeq := c.lastSeq.Load()
	if push.Seq <= lastSeq {
		log.Printf("Duplicate or out-of-order message, seq=%d, lastSeq=%d",
			push.Seq, lastSeq)
		return
	}

	c.lastSeq.Store(push.Seq)

	if c.onPush != nil {
		c.onPush(protocol.OpBattlePush, push)
	}
}

// reconnectLoop 重连循环
func (c *Client) reconnectLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		case <-c.reconnectChan:
			c.doReconnect()
		}
	}
}

// triggerReconnect 触发重连
func (c *Client) triggerReconnect() {
	if c.getState() == StateConnected {
		c.setState(StateReconnecting)
		select {
		case c.reconnectChan <- struct{}{}:
		default:
		}
	}
}

// doReconnect 执行重连
func (c *Client) doReconnect() {
	count := c.reconnectCount.Add(1)
	if count > int32(c.config.MaxReconnectTries) {
		log.Printf("Max reconnect tries exceeded, giving up")
		c.setState(StateDisconnected)
		return
	}

	log.Printf("Reconnecting... (attempt %d/%d)", count, c.config.MaxReconnectTries)

	// 关闭旧连接
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	// 指数退避
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = c.config.ReconnectInterval
	backOff.MaxElapsedTime = time.Duration(c.config.MaxReconnectTries) * c.config.ReconnectInterval

	ctx := context.Background()
	err := backoff.Retry(func() error {
		return c.doConnect(ctx)
	}, backOff)

	if err != nil {
		log.Printf("Reconnect failed: %v", err)
		c.setState(StateDisconnected)
	} else {
		log.Printf("Reconnected successfully")
		c.setState(StateConnected)
		c.reconnectCount.Store(0) // 重置重连计数
		c.incrReconnect()         // 增加重连成功计数
	}
}

// getState 获取当前状态
func (c *Client) getState() ClientState {
	return ClientState(c.state.Load())
}

// setState 设置状态
func (c *Client) setState(newState ClientState) {
	oldState := ClientState(c.state.Swap(int32(newState)))
	if oldState != newState && c.onStateChange != nil {
		c.onStateChange(oldState, newState)
	}
}

// compareAndSwapState 原子性状态切换
func (c *Client) compareAndSwapState(oldState, newState ClientState) bool {
	swapped := c.state.CompareAndSwap(int32(oldState), int32(newState))
	if swapped && c.onStateChange != nil {
		c.onStateChange(oldState, newState)
	}
	return swapped
}

// Reconnects 获取重连次数（线程安全）
func (c *Client) Reconnects() int {
	return int(c.reconnects.Load())
}

// incrReconnect 增加重连次数（线程安全）
func (c *Client) incrReconnect() {
	c.reconnects.Add(1)
}

// setLastSeq 设置最后序列号（线程安全）
func (c *Client) setLastSeq(v uint64) {
	c.lastSeq.Store(v)
}

// getLastSeq 获取最后序列号（线程安全）
func (c *Client) getLastSeq() uint64 {
	return c.lastSeq.Load()
}

// GetStats 获取客户端统计信息
func (c *Client) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"state":           c.getState().String(),
		"last_seq":        c.lastSeq.Load(),
		"reconnect_count": c.reconnectCount.Load(),
		"reconnects":      c.reconnects.Load(),
		"avg_rtt_ms":      time.Duration(c.avgRTT.Load()).Milliseconds(),
	}
}
