package testutil

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/config"
	"GoSlgBenchmarkTest/internal/wsclient"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// TestClient 测试客户端包装器
type TestClient struct {
	*wsclient.Client
	config *config.TestConfig
	t      *testing.T

	// 消息收集
	mu               sync.RWMutex
	receivedMessages []ReceivedMessage
	rttReadings      []time.Duration
	stateChanges     []StateChange
}

// ReceivedMessage 接收到的消息
type ReceivedMessage struct {
	Opcode    uint16
	Message   proto.Message
	Timestamp time.Time
}

// StateChange 状态变化
type StateChange struct {
	OldState  wsclient.ClientState
	NewState  wsclient.ClientState
	Timestamp time.Time
}

// NewTestClient 创建测试客户端
func NewTestClient(t *testing.T, serverURL, token string) *TestClient {
	cfg := config.GetTestConfig()

	clientConfig := wsclient.DefaultClientConfig(serverURL, token)
	clientConfig.HeartbeatInterval = cfg.Client.Heartbeat.Interval

	client := wsclient.New(clientConfig)

	tc := &TestClient{
		Client:           client,
		config:           cfg,
		t:                t,
		receivedMessages: make([]ReceivedMessage, 0),
		rttReadings:      make([]time.Duration, 0),
		stateChanges:     make([]StateChange, 0),
	}

	// 设置消息处理器
	tc.setupHandlers()

	return tc
}

// setupHandlers 设置各种处理器
func (tc *TestClient) setupHandlers() {
	// 消息处理器
	tc.Client.SetPushHandler(func(opcode uint16, message proto.Message) {
		tc.mu.Lock()
		tc.receivedMessages = append(tc.receivedMessages, ReceivedMessage{
			Opcode:    opcode,
			Message:   message,
			Timestamp: time.Now(),
		})
		tc.mu.Unlock()
		tc.t.Logf("📥 Received message: opcode=%d, type=%T", opcode, message)
	})

	// RTT处理器
	tc.Client.SetRTTHandler(func(rtt time.Duration) {
		tc.mu.Lock()
		tc.rttReadings = append(tc.rttReadings, rtt)
		tc.mu.Unlock()
		tc.t.Logf("📊 RTT: %v", rtt)
	})

	// 状态变化处理器
	tc.Client.SetStateChangeHandler(func(oldState, newState wsclient.ClientState) {
		tc.mu.Lock()
		tc.stateChanges = append(tc.stateChanges, StateChange{
			OldState:  oldState,
			NewState:  newState,
			Timestamp: time.Now(),
		})
		tc.mu.Unlock()
		tc.t.Logf("🔄 State change: %s -> %s", oldState.String(), newState.String())
	})
}

// ConnectWithTimeout 带超时连接
func (tc *TestClient) ConnectWithTimeout(ctx context.Context) error {
	err := tc.Client.Connect(ctx)
	if err != nil {
		tc.t.Logf("❌ Client connection failed: %v", err)
		return err
	}

	tc.t.Logf("✅ Client connected successfully")
	return nil
}

// ConnectAndWait 连接并等待就绪
func (tc *TestClient) ConnectAndWait() error {
	ctx, cancel := context.WithTimeout(context.Background(), tc.config.TestScenarios.BasicConnection.Timeout)
	defer cancel()

	if err := tc.ConnectWithTimeout(ctx); err != nil {
		return err
	}

	// 等待连接稳定
	time.Sleep(tc.config.TestScenarios.BasicConnection.ValidationDelay)
	return nil
}

// SendTestAction 发送测试操作
func (tc *TestClient) SendTestAction(seq uint64, playerID string) error {
	action := &gamev1.PlayerAction{
		ActionSeq:       seq,
		PlayerId:        playerID,
		ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
		ClientTimestamp: time.Now().UnixMilli(),
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Move{
				Move: &gamev1.MoveAction{
					TargetPosition: &gamev1.Position{
						X: float32(seq % 100),
						Y: float32(seq * 2 % 100),
						Z: 0,
					},
					MoveSpeed: 5.0,
				},
			},
		},
	}

	err := tc.Client.SendAction(action)
	if err != nil {
		tc.t.Logf("❌ Send action failed: %v", err)
	} else {
		tc.t.Logf("📤 Sent action: seq=%d, player=%s", seq, playerID)
	}
	return err
}

// SendMultipleActions 发送多个操作
func (tc *TestClient) SendMultipleActions(count int, playerID string) error {
	for i := 0; i < count; i++ {
		if err := tc.SendTestAction(uint64(i+1), playerID); err != nil {
			return fmt.Errorf("failed to send action %d: %v", i+1, err)
		}

		// 添加短暂延迟避免过快发送
		if tc.config.TestScenarios.Messaging.WarmupDelay > 0 {
			time.Sleep(tc.config.TestScenarios.Messaging.WarmupDelay)
		}
	}
	return nil
}

// WaitForMessages 等待接收指定数量的消息
func (tc *TestClient) WaitForMessages(expectedCount int, timeout time.Duration) ([]ReceivedMessage, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		tc.mu.RLock()
		currentCount := len(tc.receivedMessages)
		tc.mu.RUnlock()

		if currentCount >= expectedCount {
			tc.mu.RLock()
			messages := make([]ReceivedMessage, len(tc.receivedMessages))
			copy(messages, tc.receivedMessages)
			tc.mu.RUnlock()
			return messages, nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	tc.mu.RLock()
	actualCount := len(tc.receivedMessages)
	tc.mu.RUnlock()

	return nil, fmt.Errorf("timeout waiting for messages: expected %d, got %d", expectedCount, actualCount)
}

// WaitForRTTReadings 等待RTT读数
func (tc *TestClient) WaitForRTTReadings(expectedCount int, timeout time.Duration) ([]time.Duration, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		tc.mu.RLock()
		currentCount := len(tc.rttReadings)
		tc.mu.RUnlock()

		if currentCount >= expectedCount {
			tc.mu.RLock()
			readings := make([]time.Duration, len(tc.rttReadings))
			copy(readings, tc.rttReadings)
			tc.mu.RUnlock()
			return readings, nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	tc.mu.RLock()
	actualCount := len(tc.rttReadings)
	tc.mu.RUnlock()

	return nil, fmt.Errorf("timeout waiting for RTT readings: expected %d, got %d", expectedCount, actualCount)
}

// GetReceivedMessages 获取接收到的消息
func (tc *TestClient) GetReceivedMessages() []ReceivedMessage {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	messages := make([]ReceivedMessage, len(tc.receivedMessages))
	copy(messages, tc.receivedMessages)
	return messages
}

// GetRTTReadings 获取RTT读数
func (tc *TestClient) GetRTTReadings() []time.Duration {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	readings := make([]time.Duration, len(tc.rttReadings))
	copy(readings, tc.rttReadings)
	return readings
}

// GetStateChanges 获取状态变化
func (tc *TestClient) GetStateChanges() []StateChange {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	changes := make([]StateChange, len(tc.stateChanges))
	copy(changes, tc.stateChanges)
	return changes
}

// GetAverageRTT 获取平均RTT
func (tc *TestClient) GetAverageRTT() time.Duration {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if len(tc.rttReadings) == 0 {
		return 0
	}

	var total time.Duration
	for _, rtt := range tc.rttReadings {
		total += rtt
	}
	return total / time.Duration(len(tc.rttReadings))
}

// ClearStats 清除统计数据
func (tc *TestClient) ClearStats() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.receivedMessages = tc.receivedMessages[:0]
	tc.rttReadings = tc.rttReadings[:0]
	tc.stateChanges = tc.stateChanges[:0]
}

// ValidateConnection 验证连接状态
func (tc *TestClient) ValidateConnection() error {
	stats := tc.Client.GetStats()
	expectedState := tc.config.TestScenarios.BasicConnection.ExpectedState

	if state, ok := stats["state"].(string); ok && state == expectedState {
		return nil
	}

	return fmt.Errorf("unexpected connection state: expected %s, got %v", expectedState, stats["state"])
}

// ForceDisconnect 强制断开连接（用于重连测试）
func (tc *TestClient) ForceDisconnect() {
	tc.Client.Close()
	tc.t.Logf("🔌 Forced client disconnect")
}

// WaitForReconnect 等待重连
func (tc *TestClient) WaitForReconnect(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if err := tc.ValidateConnection(); err == nil {
			tc.t.Logf("🔄 Client reconnected successfully")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for reconnection")
}

// Cleanup 清理资源
func (tc *TestClient) Cleanup() {
	if tc.Client != nil {
		tc.Client.Close()
		tc.t.Logf("🧹 Client cleanup completed")
	}
}

// GetConnectionCount 获取连接数（用于状态变化统计）
func (tc *TestClient) GetConnectionCount() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	count := 0
	for _, change := range tc.stateChanges {
		if change.NewState == wsclient.StateConnected {
			count++
		}
	}
	return count
}

// GetReconnectCount 获取重连次数
func (tc *TestClient) GetReconnectCount() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	reconnects := 0
	for _, change := range tc.stateChanges {
		if change.OldState == wsclient.StateDisconnected && change.NewState == wsclient.StateConnected {
			reconnects++
		}
	}

	// 减去初始连接
	if reconnects > 0 {
		reconnects--
	}

	return reconnects
}
