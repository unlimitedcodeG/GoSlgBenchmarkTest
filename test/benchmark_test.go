package test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/protocol"
	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/wsclient"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// BenchmarkSingleClientRoundtrip 基准测试单客户端往返延迟
func BenchmarkSingleClientRoundtrip(b *testing.B) {
	server := testserver.New(testserver.DefaultServerConfig(":18090"))
	server.Start()
	defer server.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)

	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18090/ws", "bench-token")
	client := wsclient.New(config)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// 预热
	for i := 0; i < 10; i++ {
		action := &gamev1.PlayerAction{
			ActionSeq:       uint64(i),
			PlayerId:        "bench-player",
			ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
			ClientTimestamp: time.Now().UnixMilli(),
		}
		client.SendAction(action)
		time.Sleep(time.Millisecond)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		action := &gamev1.PlayerAction{
			ActionSeq:       uint64(i),
			PlayerId:        "bench-player",
			ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
			ClientTimestamp: time.Now().UnixMilli(),
		}

		start := time.Now()
		client.SendAction(action)
		// 简单等待，实际应用中会有响应回调
		time.Sleep(100 * time.Microsecond)
		b.StopTimer()
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/op")
		b.StartTimer()
	}
}

// BenchmarkConcurrentClients 基准测试并发客户端
func BenchmarkConcurrentClients(b *testing.B) {
	serverConfig := testserver.DefaultServerConfig(":18091")
	serverConfig.PushInterval = 10 * time.Millisecond
	server := testserver.New(serverConfig)
	server.Start()
	defer server.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)

	const numClients = 10
	clients := make([]*wsclient.Client, numClients)

	// 创建并连接客户端
	for i := 0; i < numClients; i++ {
		config := wsclient.DefaultClientConfig("ws://127.0.0.1:18091/ws",
			fmt.Sprintf("bench-token-%d", i))
		clients[i] = wsclient.New(config)

		if err := clients[i].Connect(context.Background()); err != nil {
			b.Fatalf("Client %d connect failed: %v", i, err)
		}
	}

	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		clientID := int(atomic.AddInt32(&clientCounter, 1)) % numClients
		client := clients[clientID]
		actionSeq := uint64(0)

		for pb.Next() {
			actionSeq++
			action := &gamev1.PlayerAction{
				ActionSeq:       actionSeq,
				PlayerId:        fmt.Sprintf("bench-player-%d", clientID),
				ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
				ClientTimestamp: time.Now().UnixMilli(),
			}

			if err := client.SendAction(action); err != nil {
				b.Errorf("Send action failed: %v", err)
			}
		}
	})
}

// BenchmarkProtobufMarshal 基准测试Protobuf序列化性能
func BenchmarkProtobufMarshal(b *testing.B) {
	message := &gamev1.BattlePush{
		Seq:       12345,
		BattleId:  "battle_benchmark",
		StateHash: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Units: []*gamev1.BattleUnit{
			{
				UnitId:   "unit_001",
				Hp:       100,
				Mp:       50,
				Position: &gamev1.Position{X: 10.5, Y: 20.3, Z: 0.0},
				Status:   gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
			},
			{
				UnitId:   "unit_002",
				Hp:       75,
				Mp:       25,
				Position: &gamev1.Position{X: 15.2, Y: 18.7, Z: 1.0},
				Status:   gamev1.UnitStatus_UNIT_STATUS_MOVING,
			},
		},
		Timestamp: 1699999999000,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := proto.Marshal(message)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}
		b.SetBytes(int64(len(data)))
	}
}

// BenchmarkProtobufUnmarshal 基准测试Protobuf反序列化性能
func BenchmarkProtobufUnmarshal(b *testing.B) {
	message := &gamev1.BattlePush{
		Seq:       12345,
		BattleId:  "battle_benchmark",
		StateHash: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Units: []*gamev1.BattleUnit{
			{
				UnitId:   "unit_001",
				Hp:       100,
				Mp:       50,
				Position: &gamev1.Position{X: 10.5, Y: 20.3, Z: 0.0},
				Status:   gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
			},
		},
		Timestamp: 1699999999000,
	}

	data, err := proto.Marshal(message)
	if err != nil {
		b.Fatalf("Marshal failed: %v", err)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		msg := &gamev1.BattlePush{}
		if err := proto.Unmarshal(data, msg); err != nil {
			b.Fatalf("Unmarshal failed: %v", err)
		}
	}
}

// BenchmarkFrameEncode 基准测试帧编码性能
func BenchmarkFrameEncode(b *testing.B) {
	body := []byte("This is a test message body for frame encoding benchmark")
	opcode := protocol.OpBattlePush

	b.ResetTimer()
	b.SetBytes(int64(len(body) + protocol.FrameHeaderSize))

	for i := 0; i < b.N; i++ {
		frame := protocol.EncodeFrame(opcode, body)
		_ = frame
	}
}

// BenchmarkFrameDecode 基准测试帧解码性能
func BenchmarkFrameDecode(b *testing.B) {
	body := []byte("This is a test message body for frame decoding benchmark")
	frame := protocol.EncodeFrame(protocol.OpBattlePush, body)

	b.ResetTimer()
	b.SetBytes(int64(len(frame)))

	for i := 0; i < b.N; i++ {
		opcode, decodedBody, err := protocol.DecodeFrame(frame)
		if err != nil {
			b.Fatalf("Decode failed: %v", err)
		}
		_ = opcode
		_ = decodedBody
	}
}

// BenchmarkMessageThroughput 基准测试消息吞吐量
func BenchmarkMessageThroughput(b *testing.B) {
	serverConfig := testserver.DefaultServerConfig(":18092")
	serverConfig.PushInterval = time.Millisecond // 高频推送
	server := testserver.New(serverConfig)
	server.Start()
	defer server.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)

	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18092/ws", "throughput-token")
	client := wsclient.New(config)

	var messageCount int64
	client.SetPushHandler(func(opcode uint16, message proto.Message) {
		atomic.AddInt64(&messageCount, 1)
	})

	if err := client.Connect(context.Background()); err != nil {
		b.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	b.ResetTimer()

	// 运行固定时间来测量吞吐量
	duration := time.Second
	start := time.Now()

	for time.Since(start) < duration {
		time.Sleep(10 * time.Millisecond)
	}

	count := atomic.LoadInt64(&messageCount)
	throughput := float64(count) / duration.Seconds()

	b.ReportMetric(throughput, "messages/sec")
	b.ReportMetric(float64(count), "total_messages")
}

// BenchmarkLargeMessageHandling 基准测试大消息处理
func BenchmarkLargeMessageHandling(b *testing.B) {
	// 创建大消息（10KB）
	largeStateHash := make([]byte, 10*1024)
	for i := range largeStateHash {
		largeStateHash[i] = byte(i % 256)
	}

	message := &gamev1.BattlePush{
		Seq:       1,
		BattleId:  "large_battle",
		StateHash: largeStateHash,
		Units:     make([]*gamev1.BattleUnit, 100),
		Timestamp: 1699999999000,
	}

	// 填充战斗单位
	for i := 0; i < 100; i++ {
		message.Units[i] = &gamev1.BattleUnit{
			UnitId: fmt.Sprintf("unit_%03d", i),
			Hp:     int32(100 - i),
			Mp:     int32(50 + i),
			Position: &gamev1.Position{
				X: float32(i) * 2.5,
				Y: float32(i) * 1.8,
				Z: 0,
			},
			Status: gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 序列化
		data, err := proto.Marshal(message)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}

		// 帧编码
		frame := protocol.EncodeFrame(protocol.OpBattlePush, data)

		// 帧解码
		opcode, body, err := protocol.DecodeFrame(frame)
		if err != nil {
			b.Fatalf("Frame decode failed: %v", err)
		}

		// 反序列化
		decoded := &gamev1.BattlePush{}
		if err := proto.Unmarshal(body, decoded); err != nil {
			b.Fatalf("Unmarshal failed: %v", err)
		}

		_ = opcode
		_ = decoded

		b.SetBytes(int64(len(frame)))
	}
}

// BenchmarkMemoryAllocation 基准测试内存分配
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("SmallMessage", func(b *testing.B) {
		message := &gamev1.Heartbeat{
			ClientUnixMs: time.Now().UnixMilli(),
			PingSeq:      1,
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			data, _ := proto.Marshal(message)
			frame := protocol.EncodeFrame(protocol.OpHeartbeat, data)
			_ = frame
		}
	})

	b.Run("MediumMessage", func(b *testing.B) {
		message := &gamev1.PlayerAction{
			ActionSeq:       1,
			PlayerId:        "benchmark_player",
			ActionType:      gamev1.ActionType_ACTION_TYPE_SKILL,
			ClientTimestamp: time.Now().UnixMilli(),
			ActionData: &gamev1.ActionData{
				Data: &gamev1.ActionData_Skill{
					Skill: &gamev1.SkillAction{
						SkillId:       101,
						TargetUnitIds: []string{"unit_1", "unit_2", "unit_3"},
						CastPosition:  &gamev1.Position{X: 1.0, Y: 2.0, Z: 3.0},
					},
				},
			},
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			data, _ := proto.Marshal(message)
			frame := protocol.EncodeFrame(protocol.OpPlayerAction, data)
			_ = frame
		}
	})

	b.Run("LargeMessage", func(b *testing.B) {
		message := &gamev1.BattlePush{
			Seq:       1,
			BattleId:  "benchmark_battle",
			StateHash: make([]byte, 1024),
			Units:     make([]*gamev1.BattleUnit, 50),
			Timestamp: time.Now().UnixMilli(),
		}

		for i := 0; i < 50; i++ {
			message.Units[i] = &gamev1.BattleUnit{
				UnitId:   fmt.Sprintf("unit_%d", i),
				Hp:       100,
				Mp:       50,
				Position: &gamev1.Position{X: float32(i), Y: float32(i), Z: 0},
				Status:   gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			data, _ := proto.Marshal(message)
			frame := protocol.EncodeFrame(protocol.OpBattlePush, data)
			_ = frame
		}
	})
}

// 全局计数器用于并发测试
var clientCounter int32
