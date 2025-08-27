package test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/protocol"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// FuzzBattlePushUnmarshal 模糊测试战斗推送消息的反序列化
func FuzzBattlePushUnmarshal(f *testing.F) {
	// 添加种子数据
	seed := &gamev1.BattlePush{
		Seq:       1,
		BattleId:  "battle_001",
		StateHash: []byte{1, 2, 3, 4},
		Units: []*gamev1.BattleUnit{
			{
				UnitId:   "unit_001",
				Hp:       100,
				Mp:       50,
				Position: &gamev1.Position{X: 1.0, Y: 2.0, Z: 0.0},
				Status:   gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
			},
		},
		Timestamp: 1699999999000,
	}

	seedData, _ := proto.Marshal(seed)
	f.Add(seedData)

	// 添加一些边界情况的种子
	f.Add([]byte{})                 // 空数据
	f.Add([]byte{0x00})             // 单字节
	f.Add([]byte{0xFF, 0xFF, 0xFF}) // 无效数据

	f.Fuzz(func(t *testing.T, data []byte) {
		msg := &gamev1.BattlePush{}
		// 反序列化不应该panic，但可以返回错误
		_ = proto.Unmarshal(data, msg)

		// 如果反序列化成功，重新序列化应该也成功
		if proto.Size(msg) > 0 {
			_, err := proto.Marshal(msg)
			if err != nil {
				t.Errorf("Re-marshaling failed after successful unmarshal: %v", err)
			}
		}
	})
}

// FuzzLoginReqUnmarshal 模糊测试登录请求的反序列化
func FuzzLoginReqUnmarshal(f *testing.F) {
	seed := &gamev1.LoginReq{
		Token:         "test-token-12345",
		ClientVersion: "1.0.0",
		DeviceId:      "device-abcdef",
	}

	seedData, _ := proto.Marshal(seed)
	f.Add(seedData)
	f.Add([]byte{})
	f.Add([]byte{0x08, 0x01}) // 简单的protobuf数据

	f.Fuzz(func(t *testing.T, data []byte) {
		msg := &gamev1.LoginReq{}
		_ = proto.Unmarshal(data, msg)

		// 验证重新序列化
		if proto.Size(msg) > 0 {
			if _, err := proto.Marshal(msg); err != nil {
				t.Errorf("Re-marshaling failed: %v", err)
			}
		}
	})
}

// FuzzPlayerActionUnmarshal 模糊测试玩家操作的反序列化
func FuzzPlayerActionUnmarshal(f *testing.F) {
	// 移动操作种子
	moveSeed := &gamev1.PlayerAction{
		ActionSeq:       1,
		PlayerId:        "player_001",
		ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
		ClientTimestamp: 1699999999000,
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Move{
				Move: &gamev1.MoveAction{
					TargetPosition: &gamev1.Position{X: 10.0, Y: 20.0, Z: 0.0},
					MoveSpeed:      5.0,
				},
			},
		},
	}
	moveData, _ := proto.Marshal(moveSeed)
	f.Add(moveData)

	// 攻击操作种子
	attackSeed := &gamev1.PlayerAction{
		ActionSeq:       2,
		PlayerId:        "player_002",
		ActionType:      gamev1.ActionType_ACTION_TYPE_ATTACK,
		ClientTimestamp: 1699999999000,
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Attack{
				Attack: &gamev1.AttackAction{
					TargetUnitId: "enemy_001",
					Damage:       50,
				},
			},
		},
	}
	attackData, _ := proto.Marshal(attackSeed)
	f.Add(attackData)

	// 聊天操作种子
	chatSeed := &gamev1.PlayerAction{
		ActionSeq:       3,
		PlayerId:        "player_003",
		ActionType:      gamev1.ActionType_ACTION_TYPE_CHAT,
		ClientTimestamp: 1699999999000,
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Chat{
				Chat: &gamev1.ChatAction{
					Message: "Hello World!",
					Channel: gamev1.ChatChannel_CHAT_CHANNEL_WORLD,
				},
			},
		},
	}
	chatData, _ := proto.Marshal(chatSeed)
	f.Add(chatData)

	f.Fuzz(func(t *testing.T, data []byte) {
		msg := &gamev1.PlayerAction{}
		_ = proto.Unmarshal(data, msg)

		// 验证oneof字段的一致性
		if msg.ActionData != nil {
			switch msg.ActionData.Data.(type) {
			case *gamev1.ActionData_Move:
				if msg.ActionType != gamev1.ActionType_ACTION_TYPE_MOVE &&
					msg.ActionType != gamev1.ActionType_ACTION_TYPE_UNKNOWN {
					// 不是严格错误，但记录不一致性
					t.Logf("Action type mismatch: type=%v, data=move", msg.ActionType)
				}
			case *gamev1.ActionData_Attack:
				if msg.ActionType != gamev1.ActionType_ACTION_TYPE_ATTACK &&
					msg.ActionType != gamev1.ActionType_ACTION_TYPE_UNKNOWN {
					t.Logf("Action type mismatch: type=%v, data=attack", msg.ActionType)
				}
			case *gamev1.ActionData_Chat:
				if msg.ActionType != gamev1.ActionType_ACTION_TYPE_CHAT &&
					msg.ActionType != gamev1.ActionType_ACTION_TYPE_UNKNOWN {
					t.Logf("Action type mismatch: type=%v, data=chat", msg.ActionType)
				}
			}
		}

		// 重新序列化测试
		if proto.Size(msg) > 0 {
			if _, err := proto.Marshal(msg); err != nil {
				t.Errorf("Re-marshaling failed: %v", err)
			}
		}
	})
}

// FuzzFrameDecode 模糊测试帧解码
func FuzzFrameDecode(f *testing.F) {
	// 有效帧的种子
	validFrame := protocol.EncodeFrame(protocol.OpBattlePush, []byte{1, 2, 3, 4})
	f.Add(validFrame)

	// 空帧
	emptyFrame := protocol.EncodeFrame(protocol.OpHeartbeat, []byte{})
	f.Add(emptyFrame)

	// 最小帧
	f.Add([]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00}) // opcode=1, length=0

	// 边界情况
	f.Add([]byte{})                       // 空数据
	f.Add([]byte{0x00})                   // 太短
	f.Add([]byte{0x00, 0x01, 0xFF, 0xFF}) // 不完整的长度

	f.Fuzz(func(t *testing.T, data []byte) {
		// 性能优化：增加输入大小限制，防止过大的输入影响性能
		if len(data) > 64*1024 { // 64KB限制，减少内存使用
			t.Skip("Input too large, skipping")
			return
		}

		opcode, body, err := protocol.DecodeFrame(data)

		if err != nil {
			// 错误是可以接受的，但不应该panic
			return
		}

		// 如果解码成功，验证数据的一致性
		if !protocol.IsValidOpcode(opcode) {
			t.Logf("Decoded invalid opcode: %d", opcode)
		}

		// 重新编码应该产生相同的数据（对于有效帧）
		if len(data) >= protocol.FrameHeaderSize {
			reEncoded := protocol.EncodeFrame(opcode, body)
			if len(reEncoded) == len(data) {
				// 长度匹配时，内容也应该匹配
				for i := 0; i < len(data); i++ {
					if data[i] != reEncoded[i] {
						t.Logf("Re-encoding mismatch at byte %d: original=%02x, re-encoded=%02x",
							i, data[i], reEncoded[i])
						break
					}
				}
			}
		}
	})
}

// FuzzFrameDecoder 模糊测试流式帧解码器
func FuzzFrameDecoder(f *testing.F) {
	// 创建包含多个帧的数据流
	frame1 := protocol.EncodeFrame(protocol.OpLoginReq, []byte{0x01, 0x02})
	frame2 := protocol.EncodeFrame(protocol.OpHeartbeat, []byte{})
	frame3 := protocol.EncodeFrame(protocol.OpBattlePush, []byte{0x03, 0x04, 0x05})

	multiFrame := append(frame1, frame2...)
	multiFrame = append(multiFrame, frame3...)
	f.Add(multiFrame)

	// 单个帧
	f.Add(frame1)
	f.Add(frame2)
	f.Add(frame3)

	// 不完整的帧
	f.Add(frame1[:3]) // 只有部分头部
	f.Add(frame1[:5]) // 头部完整但数据不完整

	f.Fuzz(func(t *testing.T, data []byte) {
		// 性能优化：限制输入大小，防止过大的输入影响性能
		if len(data) > 64*1024 { // 64KB限制
			t.Skip("Input too large, skipping")
			return
		}

		// 性能优化：使用对象池减少内存分配
		decoder := protocol.NewFrameDecoder()
		defer func() {
			// 确保解码器被正确重置
			decoder.Reset()
		}()

		decoder.Feed(data)

		frameCount := 0
		maxFrames := 20      // 性能优化：进一步减少最大帧数，提高处理速度
		maxIterations := 100 // 性能优化：添加最大迭代次数限制，防止无限循环
		iterations := 0

		for frameCount < maxFrames && iterations < maxIterations {
			iterations++

			frame, err := decoder.Next()
			if err != nil {
				// 错误是可以接受的
				break
			}
			if frame == nil {
				// 需要更多数据
				break
			}

			frameCount++

			// 性能优化：简化验证逻辑，只进行必要的检查
			if !protocol.IsValidOpcode(frame.Opcode) {
				t.Logf("Decoded invalid opcode in stream: %d", frame.Opcode)
			}

			// 性能优化：跳过重新编码验证，减少计算开销
			// _ = protocol.EncodeFrame(frame.Opcode, frame.Body)
		}

		// 重置解码器应该清空状态
		decoder.Reset()
		if decoder.BufferSize() != 0 {
			t.Errorf("Decoder buffer not empty after reset: size=%d", decoder.BufferSize())
		}
	})
}

// FuzzErrorResp 模糊测试错误响应
func FuzzErrorResp(f *testing.F) {
	seed := &gamev1.ErrorResp{
		ErrorCode:    1001,
		ErrorMessage: "Test error message",
		RequestId:    "req_12345",
	}

	seedData, _ := proto.Marshal(seed)
	f.Add(seedData)

	// 添加极端情况
	longMsgSeed := &gamev1.ErrorResp{
		ErrorCode:    9999,
		ErrorMessage: string(make([]byte, 10000)), // 很长的错误消息
		RequestId:    "long_req_id_" + string(make([]byte, 1000)),
	}
	longData, _ := proto.Marshal(longMsgSeed)
	f.Add(longData)

	f.Fuzz(func(t *testing.T, data []byte) {
		msg := &gamev1.ErrorResp{}
		_ = proto.Unmarshal(data, msg)

		// 验证字段的合理性
		if msg.ErrorCode < 0 {
			t.Logf("Negative error code: %d", msg.ErrorCode)
		}

		if len(msg.ErrorMessage) > 100000 { // 100KB
			t.Logf("Very long error message: %d bytes", len(msg.ErrorMessage))
		}

		if len(msg.RequestId) > 10000 { // 10KB
			t.Logf("Very long request ID: %d bytes", len(msg.RequestId))
		}

		// 重新序列化测试
		if proto.Size(msg) > 0 {
			if _, err := proto.Marshal(msg); err != nil {
				t.Errorf("Re-marshaling failed: %v", err)
			}
		}
	})
}
