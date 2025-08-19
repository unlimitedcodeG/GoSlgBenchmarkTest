package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// EventType 事件类型
type EventType string

const (
	EventConnect        EventType = "CONNECT"
	EventDisconnect     EventType = "DISCONNECT"
	EventLogin          EventType = "LOGIN"
	EventHeartbeat      EventType = "HEARTBEAT"
	EventMessageSend    EventType = "MESSAGE_SEND"
	EventMessageReceive EventType = "MESSAGE_RECEIVE"
	EventError          EventType = "ERROR"
	EventReconnect      EventType = "RECONNECT"
	EventClose          EventType = "CLOSE"
)

// CloseCode WebSocket关闭代码
type CloseCode int

const (
	CloseNormal          CloseCode = 1000
	CloseGoingAway       CloseCode = 1001
	CloseProtocolError   CloseCode = 1002
	CloseUnsupported     CloseCode = 1003
	CloseNoStatus        CloseCode = 1005
	CloseAbnormal        CloseCode = 1006
	CloseInvalidData     CloseCode = 1007
	ClosePolicyViolation CloseCode = 1008
	CloseTooBig          CloseCode = 1009
	CloseInternalError   CloseCode = 1011
	CloseServiceRestart  CloseCode = 1012
	CloseTryAgainLater   CloseCode = 1013
	CloseTLSHandshake    CloseCode = 1015
)

// SessionEvent 会话事件
type SessionEvent struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	ClientTime  time.Time              `json:"client_time"`
	ServerTime  time.Time              `json:"server_time"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Opcode      uint16                 `json:"opcode,omitempty"`
	MessageSize int                    `json:"message_size,omitempty"`
	MessageHash string                 `json:"message_hash,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CloseCode   CloseCode              `json:"close_code,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageFrame 消息帧记录
type MessageFrame struct {
	RawData     []byte    `json:"raw_data"`
	Opcode      uint16    `json:"opcode"`
	Body        []byte    `json:"body"`
	Timestamp   time.Time `json:"timestamp"`
	Direction   string    `json:"direction"` // "send" or "receive"
	SequenceNum uint64    `json:"sequence_num,omitempty"`
}

// SessionStats 会话统计
type SessionStats struct {
	StartTime          time.Time             `json:"start_time"`
	EndTime            time.Time             `json:"end_time"`
	Duration           time.Duration         `json:"duration"`
	TotalEvents        int64                 `json:"total_events"`
	MessagesSent       int64                 `json:"messages_sent"`
	MessagesReceived   int64                 `json:"messages_received"`
	BytesSent          int64                 `json:"bytes_sent"`
	BytesReceived      int64                 `json:"bytes_received"`
	ReconnectCount     int64                 `json:"reconnect_count"`
	ErrorCount         int64                 `json:"error_count"`
	AverageLatency     time.Duration         `json:"average_latency"`
	MinLatency         time.Duration         `json:"min_latency"`
	MaxLatency         time.Duration         `json:"max_latency"`
	LatencyPercentiles map[int]time.Duration `json:"latency_percentiles"`
}

// SessionRecorder 会话录制器
type SessionRecorder struct {
	sessionID string
	startTime time.Time
	events    []*SessionEvent
	frames    []*MessageFrame
	stats     *SessionStats

	// 统计计数器
	eventCounter   atomic.Int64
	messageCounter atomic.Int64
	byteCounter    atomic.Int64
	reconnectCount atomic.Int64
	errorCount     atomic.Int64

	// 延迟统计
	latencySum   atomic.Int64
	latencyCount atomic.Int64
	minLatency   atomic.Int64
	maxLatency   atomic.Int64

	// 同步控制
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	isActive atomic.Bool
}

// NewSessionRecorder 创建新的会话录制器
func NewSessionRecorder(sessionID string) *SessionRecorder {
	ctx, cancel := context.WithCancel(context.Background())

	recorder := &SessionRecorder{
		sessionID: sessionID,
		startTime: time.Now(),
		events:    make([]*SessionEvent, 0, 1000),
		frames:    make([]*MessageFrame, 0, 1000),
		stats: &SessionStats{
			StartTime:          time.Now(),
			LatencyPercentiles: make(map[int]time.Duration),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	recorder.isActive.Store(true)

	// 记录会话开始事件
	recorder.RecordEvent(EventConnect, map[string]interface{}{
		"session_id": sessionID,
		"start_time": recorder.startTime,
	})

	return recorder
}

// RecordEvent 记录事件
func (r *SessionRecorder) RecordEvent(eventType EventType, metadata map[string]interface{}) {
	if !r.isActive.Load() {
		return
	}

	now := time.Now()
	event := &SessionEvent{
		ID:         fmt.Sprintf("event_%d", r.eventCounter.Add(1)),
		Type:       eventType,
		Timestamp:  now,
		ClientTime: now,
		ServerTime: now,
		Metadata:   metadata,
	}

	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()

	// 更新统计
	r.updateStats(event)
}

// RecordMessage 记录消息
func (r *SessionRecorder) RecordMessage(direction string, rawData []byte, opcode uint16, body []byte, sequenceNum uint64) {
	if !r.isActive.Load() {
		return
	}

	now := time.Now()
	frame := &MessageFrame{
		RawData:     rawData,
		Opcode:      opcode,
		Body:        body,
		Timestamp:   now,
		Direction:   direction,
		SequenceNum: sequenceNum,
	}

	r.mu.Lock()
	r.frames = append(r.frames, frame)
	r.mu.Unlock()

	// 记录消息事件
	eventType := EventMessageSend
	if direction == "receive" {
		eventType = EventMessageReceive
	}

	r.RecordEvent(eventType, map[string]interface{}{
		"opcode":       opcode,
		"message_size": len(rawData),
		"body_size":    len(body),
		"sequence_num": sequenceNum,
		"direction":    direction,
	})

	// 更新消息统计
	if direction == "send" {
		r.messageCounter.Add(1)
		r.byteCounter.Add(int64(len(rawData)))
	} else {
		r.messageCounter.Add(1)
		r.byteCounter.Add(int64(len(rawData)))
	}
}

// RecordLatency 记录延迟
func (r *SessionRecorder) RecordLatency(latency time.Duration) {
	if !r.isActive.Load() || latency <= 0 {
		return
	}

	latencyNano := latency.Nanoseconds()

	// 更新延迟统计
	r.latencySum.Add(latencyNano)
	r.latencyCount.Add(1)

	// 更新最小延迟
	for {
		current := r.minLatency.Load()
		if current == 0 || latencyNano < current {
			if r.minLatency.CompareAndSwap(current, latencyNano) {
				break
			}
		} else {
			break
		}
	}

	// 更新最大延迟
	for {
		current := r.maxLatency.Load()
		if latencyNano > current {
			if r.maxLatency.CompareAndSwap(current, latencyNano) {
				break
			}
		} else {
			break
		}
	}
}

// RecordReconnect 记录重连事件
func (r *SessionRecorder) RecordReconnect(attempt int, duration time.Duration, success bool) {
	r.reconnectCount.Add(1)

	metadata := map[string]interface{}{
		"attempt":  attempt,
		"duration": duration,
		"success":  success,
	}

	r.RecordEvent(EventReconnect, metadata)
}

// RecordError 记录错误事件
func (r *SessionRecorder) RecordError(err error, metadata map[string]interface{}) {
	r.errorCount.Add(1)

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["error"] = err.Error()

	r.RecordEvent(EventError, metadata)
}

// RecordClose 记录关闭事件
func (r *SessionRecorder) RecordClose(closeCode CloseCode, reason string) {
	metadata := map[string]interface{}{
		"close_code": closeCode,
		"reason":     reason,
	}

	r.RecordEvent(EventClose, metadata)

	// 停止录制
	r.Stop()
}

// Stop 停止录制
func (r *SessionRecorder) Stop() {
	if !r.isActive.CompareAndSwap(true, false) {
		return
	}

	r.cancel()

	// 计算最终统计
	r.calculateFinalStats()

	// 记录会话结束事件
	r.RecordEvent(EventDisconnect, map[string]interface{}{
		"end_time": time.Now(),
		"duration": time.Since(r.startTime),
	})
}

// GetSession 获取完整会话记录
func (r *SessionRecorder) GetSession() *Session {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &Session{
		ID:        r.sessionID,
		StartTime: r.startTime,
		EndTime:   time.Now(),
		Events:    append([]*SessionEvent{}, r.events...),
		Frames:    append([]*MessageFrame{}, r.frames...),
		Stats:     r.stats,
	}
}

// GetEvents 获取事件列表
func (r *SessionRecorder) GetEvents() []*SessionEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]*SessionEvent{}, r.events...)
}

// GetFrames 获取消息帧列表
func (r *SessionRecorder) GetFrames() []*MessageFrame {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]*MessageFrame{}, r.frames...)
}

// GetStats 获取统计信息
func (r *SessionRecorder) GetStats() *SessionStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.stats
}

// ExportJSON 导出为JSON格式
func (r *SessionRecorder) ExportJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 过滤系统事件，只保留业务相关事件
	filteredEvents := make([]*SessionEvent, 0, len(r.events))
	for _, event := range r.events {
		// 跳过系统事件，只保留业务事件
		if event.Type == EventHeartbeat || event.Type == EventLogin || event.Type == EventDisconnect {
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}

	session := &Session{
		ID:        r.sessionID,
		StartTime: r.startTime,
		EndTime:   time.Now(),
		Events:    filteredEvents,
		Frames:    append([]*MessageFrame{}, r.frames...),
		Stats:     r.stats,
	}
	
	return json.MarshalIndent(session, "", "  ")
}

// updateStats 更新统计信息
func (r *SessionRecorder) updateStats(event *SessionEvent) {
	switch event.Type {
	case EventMessageSend:
		r.stats.MessagesSent++
	case EventMessageReceive:
		r.stats.MessagesReceived++
	case EventReconnect:
		r.stats.ReconnectCount++
	case EventError:
		r.stats.ErrorCount++
	}
}

// calculateFinalStats 计算最终统计
func (r *SessionRecorder) calculateFinalStats() {
	r.stats.EndTime = time.Now()
	r.stats.Duration = r.stats.EndTime.Sub(r.stats.StartTime)
	r.stats.TotalEvents = int64(len(r.events))
	r.stats.MessagesSent = r.messageCounter.Load()
	r.stats.MessagesReceived = r.messageCounter.Load()
	r.stats.BytesSent = r.byteCounter.Load()
	r.stats.BytesReceived = r.byteCounter.Load()
	r.stats.ReconnectCount = r.reconnectCount.Load()
	r.stats.ErrorCount = r.errorCount.Load()

	// 计算延迟统计
	latencyCount := r.latencyCount.Load()
	if latencyCount > 0 {
		latencySum := r.latencySum.Load()
		r.stats.AverageLatency = time.Duration(latencySum / latencyCount)
		r.stats.MinLatency = time.Duration(r.minLatency.Load())
		r.stats.MaxLatency = time.Duration(r.maxLatency.Load())

		// 计算延迟百分位数（简化版本）
		r.calculateLatencyPercentiles()
	}
}

// calculateLatencyPercentiles 计算延迟百分位数
func (r *SessionRecorder) calculateLatencyPercentiles() {
	// 这里可以实现更复杂的百分位数计算
	// 暂时使用简单的统计
	r.stats.LatencyPercentiles[50] = r.stats.AverageLatency
	r.stats.LatencyPercentiles[90] = r.stats.MaxLatency
	r.stats.LatencyPercentiles[95] = r.stats.MaxLatency
	r.stats.LatencyPercentiles[99] = r.stats.MaxLatency
}
