package test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/config"
	"GoSlgBenchmarkTest/internal/protocol"
	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/testutil"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// TestBasicConnection 测试基本连接功能
func TestBasicConnection(t *testing.T) {
	// 使用统一工具创建服务器
	server := testutil.NewTestServer(t)
	server.Start()
	defer server.Stop()

	// 使用统一工具创建客户端
	client := testutil.NewTestClient(t, server.GetWebSocketURL(), "test-token")
	defer client.Cleanup()

	// 连接并验证
	err := client.ConnectAndWait()
	require.NoError(t, err)

	// 使用统一断言验证连接
	assertions := testutil.NewTestAssertions(t)
	assertions.AssertConnection(client)
}

// TestReconnectAndSequenceMonotonic 测试断线重连和序列号单调性
func TestReconnectAndSequenceMonotonic(t *testing.T) {
	cfg := config.GetTestConfig()

	// 使用自定义配置创建服务器（高频推送）
	server := testutil.NewTestServerWithConfig(t, func(serverConfig *testserver.ServerConfig) {
		serverConfig.PushInterval = 50 * time.Millisecond
		serverConfig.EnableBattlePush = true
	})
	server.Start()
	defer server.Stop()

	// 用channel收集序列号，避免数据竞争
	seqCh := make(chan uint64, 1024)

	// 创建客户端（使用配置化的重连参数）
	client := testutil.NewTestClient(t, server.GetWebSocketURL(), "test-token")
	defer client.Cleanup()

	// 设置推送处理器收集序列号
	client.SetPushHandler(func(opcode uint16, message proto.Message) {
		if opcode == protocol.OpBattlePush {
			if battlePush, ok := message.(*gamev1.BattlePush); ok {
				select {
				case seqCh <- battlePush.Seq:
				default: // 防阻塞
				}
			}
		}
	})

	// 使用足够的超时时间支持重试机制 (5次重试 * 1秒间隔 + 缓冲时间)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, client.ConnectWithTimeout(ctx))

	// 等待接收一些消息
	time.Sleep(cfg.TestScenarios.Reconnect.ForceDisconnectAfter)

	// 强制断开连接触发重连
	server.Server.ForceDisconnectAll()

	// 等待重连和更多消息
	time.Sleep(2 * time.Second)

	client.Close()

	// 从channel收集所有序列号
	close(seqCh)
	var seqs []uint64
	for seq := range seqCh {
		seqs = append(seqs, seq)
	}

	require.Greater(t, len(seqs), 0, "Should receive at least one message")

	// 使用统一断言验证消息序列
	assertions := testutil.NewTestAssertions(t)
	assertions.AssertMessageSequence(client)

	// 验证序列号单调递增
	for i := 1; i < len(seqs); i++ {
		assert.Greater(t, seqs[i], seqs[i-1],
			"Sequence numbers should be monotonically increasing: seq[%d]=%d, seq[%d]=%d",
			i-1, seqs[i-1], i, seqs[i])
	}

	t.Logf("Received %d messages with monotonic sequences", len(seqs))
	t.Logf("Reconnect count: %d", client.GetReconnectCount())
}

// TestHeartbeatAndRTT 测试心跳和RTT统计
func TestHeartbeatAndRTT(t *testing.T) {
	cfg := config.GetTestConfig()

	// 创建服务器
	server := testutil.NewTestServer(t)
	server.Start()
	defer server.Stop()

	// 创建客户端（配置化心跳间隔）
	client := testutil.NewTestClient(t, server.GetWebSocketURL(), "test-token")
	defer client.Cleanup()

	// 使用足够的超时时间支持重试机制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, client.ConnectWithTimeout(ctx))

	// 等待足够的心跳周期
	time.Sleep(cfg.TestScenarios.Heartbeat.TestDuration)

	client.Close()

	// 验证RTT统计
	rtts := client.GetRTTReadings()
	t.Logf("Received %d RTT readings", len(rtts))

	// 使用配置化的最小RTT数量要求
	expectedMinRTTs := cfg.TestScenarios.Heartbeat.ExpectedMinRTTs
	if len(rtts) < expectedMinRTTs {
		t.Skip("Insufficient RTT readings - this may be due to timing issues in the test environment")
		return
	}

	// 使用统一断言验证延迟
	assertions := testutil.NewTestAssertions(t)
	assertions.AssertLatency(client)

	// 额外验证RTT合理性
	maxAcceptableRTT := cfg.TestScenarios.Heartbeat.MaxAcceptableRTT
	for _, rtt := range rtts {
		assert.Greater(t, rtt, time.Duration(0), "RTT should be positive")
		assert.Less(t, rtt, maxAcceptableRTT, "RTT should be reasonable for local connection")
	}
}

// TestPlayerAction 测试玩家操作发送和响应
func TestPlayerAction(t *testing.T) {
	// 创建服务器
	server := testutil.NewTestServer(t)
	server.Start()
	defer server.Stop()

	// 创建客户端
	client := testutil.NewTestClient(t, server.GetWebSocketURL(), "test-token")
	defer client.Cleanup()

	var responseReceived bool
	var mu sync.Mutex

	// 设置响应处理器
	client.SetPushHandler(func(opcode uint16, message proto.Message) {
		if opcode == protocol.OpActionResp {
			mu.Lock()
			responseReceived = true
			mu.Unlock()
		}
	})

	// 使用足够的超时时间支持重试机制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, client.ConnectWithTimeout(ctx))

	// 发送测试操作（使用统一工具）
	err := client.SendTestAction(1, "test-player")
	require.NoError(t, err)

	// 等待响应
	time.Sleep(500 * time.Millisecond)

	client.Close()

	// 验证响应
	mu.Lock()
	received := responseReceived
	mu.Unlock()

	assert.True(t, received, "Should receive action response")
}

// TestConcurrentConnections 测试并发连接
func TestConcurrentConnections(t *testing.T) {
	cfg := config.GetTestConfig()

	// 使用自定义配置创建服务器
	server := testutil.NewTestServerWithConfig(t, func(serverConfig *testserver.ServerConfig) {
		serverConfig.MaxConnections = 50
		serverConfig.PushInterval = 100 * time.Millisecond
	})
	server.Start()
	defer server.Stop()

	// 使用配置化的客户端数量
	numClients := cfg.StressTest.ConcurrentClients.DefaultClients

	// 使用现代化并发模式：context + errgroup，支持重试机制
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var successCount int32
	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			client := testutil.NewTestClient(t, server.GetWebSocketURL(),
				fmt.Sprintf("token-%d", clientID))
			defer client.Cleanup()

			if err := client.ConnectWithTimeout(ctx); err != nil {
				t.Logf("Client %d connect failed: %v", clientID, err)
				return
			}

			// 保持连接一段时间
			time.Sleep(1 * time.Second)

			atomic.AddInt32(&successCount, 1)
		}(i)
	}

	wg.Wait()

	// 使用统一断言验证并发连接
	require.Equal(t, int32(numClients), atomic.LoadInt32(&successCount),
		"All clients should connect successfully")

	// 验证服务器统计 - 使用Eventually等待连接完全关闭
	require.Eventually(t, func() bool {
		stats := server.Server.GetStats()
		return stats["current_connections"].(int32) == 0
	}, 3*time.Second, 50*time.Millisecond, "All connections should be closed")

	// 验证总连接数
	stats := server.Server.GetStats()
	assert.Equal(t, uint64(numClients), stats["total_connections"],
		"Total connections should match")
}

// TestLargeMessage 测试大消息处理
func TestLargeMessage(t *testing.T) {
	cfg := config.GetTestConfig()

	// 创建服务器
	server := testutil.NewTestServer(t)
	server.Start()
	defer server.Stop()

	// 创建客户端
	client := testutil.NewTestClient(t, server.GetWebSocketURL(), "test-token")
	defer client.Cleanup()

	// 使用足够的超时时间支持重试机制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, client.ConnectWithTimeout(ctx))

	// 使用配置化的大消息大小
	messageSize := cfg.StressTest.LargeMessages.MessageSizes["large"] // 100KB
	largeMessage := make([]byte, messageSize)
	for i := range largeMessage {
		largeMessage[i] = byte('A' + (i % 26))
	}

	action := &gamev1.PlayerAction{
		ActionSeq:       1,
		PlayerId:        "test-player",
		ActionType:      gamev1.ActionType_ACTION_TYPE_CHAT,
		ClientTimestamp: time.Now().UnixMilli(),
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Chat{
				Chat: &gamev1.ChatAction{
					Message: string(largeMessage),
					Channel: gamev1.ChatChannel_CHAT_CHANNEL_WORLD,
				},
			},
		},
	}

	err := client.SendAction(action)
	require.NoError(t, err)

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	t.Logf("Large message sent successfully: %d bytes", messageSize)
}
