package slg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/protocol"
	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/wsclient"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// TestSLGProtocolIntegration 测试SLG协议与测试框架的集成
func TestSLGProtocolIntegration(t *testing.T) {
	// 启动测试服务器
	server := testserver.New(testserver.DefaultServerConfig(":18100"))
	require.NoError(t, server.Start())
	defer server.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18100/ws", "slg-test-token")
	client := wsclient.New(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.Connect(ctx))
	defer client.Close()

	// 测试发送SLG相关的玩家操作
	action := &gamev1.PlayerAction{
		ActionSeq:       1,
		PlayerId:        "slg-player-001",
		ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
		ClientTimestamp: time.Now().UnixMilli(),
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Move{
				Move: &gamev1.MoveAction{
					TargetPosition: &gamev1.Position{X: 100.0, Y: 200.0, Z: 0.0},
					MoveSpeed:      10.0,
				},
			},
		},
	}

	err := client.SendAction(action)
	require.NoError(t, err)

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	t.Log("SLG协议集成测试通过")
}

// TestSLGMessageSerialization 测试SLG消息序列化
func TestSLGMessageSerialization(t *testing.T) {
	// TODO: 这里应该测试实际的SLG协议消息
	// 当生成了SLG协议代码后，可以导入并测试

	// 示例：测试通用的位置消息
	position := &gamev1.Position{
		X: 123.45,
		Y: 678.90,
		Z: 0.0,
	}

	data, err := proto.Marshal(position)
	require.NoError(t, err)
	require.Greater(t, len(data), 0)

	// 反序列化
	decoded := &gamev1.Position{}
	err = proto.Unmarshal(data, decoded)
	require.NoError(t, err)

	assert.InDelta(t, position.X, decoded.X, 0.001)
	assert.InDelta(t, position.Y, decoded.Y, 0.001)
	assert.InDelta(t, position.Z, decoded.Z, 0.001)
}

// TestSLGFrameEncoding 测试SLG消息的帧编码
func TestSLGFrameEncoding(t *testing.T) {
	// 创建一个模拟的SLG消息
	message := &gamev1.BattlePush{
		Seq:       12345,
		BattleId:  "slg_battle_001",
		StateHash: []byte{0x01, 0x02, 0x03, 0x04},
		Units: []*gamev1.BattleUnit{
			{
				UnitId:   "slg_unit_001",
				Hp:       1000,
				Mp:       500,
				Position: &gamev1.Position{X: 10.5, Y: 20.3, Z: 0.0},
				Status:   gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// 序列化
	data, err := proto.Marshal(message)
	require.NoError(t, err)

	// 帧编码
	frame := protocol.EncodeFrame(protocol.OpBattlePush, data)
	require.Greater(t, len(frame), protocol.FrameHeaderSize)

	// 帧解码
	opcode, body, err := protocol.DecodeFrame(frame)
	require.NoError(t, err)
	assert.Equal(t, protocol.OpBattlePush, opcode)

	// 反序列化
	decoded := &gamev1.BattlePush{}
	err = proto.Unmarshal(body, decoded)
	require.NoError(t, err)

	// 验证数据完整性
	assert.Equal(t, message.Seq, decoded.Seq)
	assert.Equal(t, message.BattleId, decoded.BattleId)
	assert.Equal(t, message.StateHash, decoded.StateHash)
	assert.Equal(t, len(message.Units), len(decoded.Units))

	if len(decoded.Units) > 0 {
		originalUnit := message.Units[0]
		decodedUnit := decoded.Units[0]

		assert.Equal(t, originalUnit.UnitId, decodedUnit.UnitId)
		assert.Equal(t, originalUnit.Hp, decodedUnit.Hp)
		assert.Equal(t, originalUnit.Mp, decodedUnit.Mp)
		assert.Equal(t, originalUnit.Status, decodedUnit.Status)
	}
}

// TestSLGLoadTest 测试SLG场景下的负载
func TestSLGLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过负载测试")
	}

	// 启动测试服务器
	serverConfig := testserver.DefaultServerConfig(":18101")
	serverConfig.MaxConnections = 100
	serverConfig.PushInterval = 50 * time.Millisecond // 模拟高频推送

	server := testserver.New(serverConfig)
	require.NoError(t, server.Start())

	// 确保服务器完全关闭，即使测试失败
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		t.Log("🛑 正在优雅关闭测试服务器...")
		if err := server.Shutdown(ctx); err != nil {
			t.Logf("服务器关闭错误: %v", err)
		}
		t.Log("✅ 测试服务器已关闭")
	}()

	time.Sleep(100 * time.Millisecond)

	// 模拟多个SLG玩家同时连接
	const numPlayers = 20
	clients := make([]*wsclient.Client, numPlayers)

	// 连接所有客户端
	for i := 0; i < numPlayers; i++ {
		config := wsclient.DefaultClientConfig("ws://127.0.0.1:18101/ws",
			"slg-player-"+string(rune('A'+i)))
		clients[i] = wsclient.New(config)

		err := clients[i].Connect(context.Background())
		require.NoError(t, err)
	}

	// 确保所有客户端都被完全关闭
	defer func() {
		t.Log("🧹 正在清理所有客户端连接...")

		// 首先通知所有客户端停止操作
		for i, client := range clients {
			if client != nil {
				t.Logf("   通知客户端 %d 停止...", i)
				client.Close() // 这会触发stopChan
			}
		}

		// 等待一段时间让goroutine优雅退出
		time.Sleep(200 * time.Millisecond)

		// 再次确认所有客户端都已关闭
		for i, client := range clients {
			if client != nil {
				stats := client.GetStats()
				t.Logf("   客户端 %d 最终状态: %v", i, stats["state"])
			}
		}
		t.Log("✅ 所有客户端已清理完成")
	}()

	// 模拟SLG操作：移动、攻击、建造等
	for i, client := range clients {
		action := &gamev1.PlayerAction{
			ActionSeq:       uint64(i + 1),
			PlayerId:        "slg-player-" + string(rune('A'+i)),
			ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
			ClientTimestamp: time.Now().UnixMilli(),
			ActionData: &gamev1.ActionData{
				Data: &gamev1.ActionData_Move{
					Move: &gamev1.MoveAction{
						TargetPosition: &gamev1.Position{
							X: float32(i * 10),
							Y: float32(i * 15),
							Z: 0.0,
						},
						MoveSpeed: 5.0,
					},
				},
			},
		}

		err := client.SendAction(action)
		require.NoError(t, err)
	}

	// 等待所有操作处理完成
	time.Sleep(2 * time.Second)

	// 验证服务器统计
	stats := server.GetStats()
	t.Logf("SLG负载测试统计: %+v", stats)

	// 基本断言
	assert.Equal(t, int32(numPlayers), stats["current_connections"])
	assert.GreaterOrEqual(t, stats["total_connections"], uint64(numPlayers))

	// 验证消息传输成功
	totalMessages := stats["total_messages"].(uint64)
	assert.Greater(t, totalMessages, uint64(0), "应该收到至少一条消息")

	// 验证连接稳定性
	assert.Less(t, stats["total_connections"], uint64(numPlayers*2), "连接数不应过多")

	t.Log("✅ SLG负载测试断言全部通过")
}
