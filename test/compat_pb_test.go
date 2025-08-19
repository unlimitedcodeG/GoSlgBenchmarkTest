package test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// 测试数据目录
const testDataDir = "../testdata"

// TestBackwardCompatibility_LoginResp 测试登录响应的向后兼容性
func TestBackwardCompatibility_LoginResp(t *testing.T) {
	// 创建测试数据
	original := &gamev1.LoginResp{
		Ok:         true,
		PlayerId:   "player_12345",
		SessionId:  "session_abcdef",
		ServerTime: 1699999999000,
	}

	// 序列化
	data, err := proto.Marshal(original)
	require.NoError(t, err)

	// 保存为golden文件（如果不存在）
	goldenFile := filepath.Join(testDataDir, "login_resp_v1.bin")
	if err := ensureTestData(goldenFile, data); err != nil {
		t.Skipf("Cannot create test data: %v", err)
	}

	// 读取golden文件
	golden, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	// 新代码应该能解析旧的二进制数据
	parsed := &gamev1.LoginResp{}
	require.NoError(t, proto.Unmarshal(golden, parsed))

	// 验证关键字段
	assert.Equal(t, original.Ok, parsed.Ok)
	assert.Equal(t, original.PlayerId, parsed.PlayerId)
	assert.Equal(t, original.SessionId, parsed.SessionId)
	assert.Equal(t, original.ServerTime, parsed.ServerTime)

	// 重新序列化不应该丢失信息
	reSerialized, err := proto.Marshal(parsed)
	require.NoError(t, err)
	require.NotZero(t, len(reSerialized))
}

// TestBackwardCompatibility_BattlePush 测试战斗推送的向后兼容性
func TestBackwardCompatibility_BattlePush(t *testing.T) {
	original := &gamev1.BattlePush{
		Seq:       12345,
		BattleId:  "battle_001",
		StateHash: []byte{0x01, 0x02, 0x03, 0x04},
		Units: []*gamev1.BattleUnit{
			{
				UnitId: "unit_001",
				Hp:     100,
				Mp:     50,
				Position: &gamev1.Position{
					X: 10.5,
					Y: 20.3,
					Z: 0.0,
				},
				Status: gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
			},
			{
				UnitId: "unit_002",
				Hp:     75,
				Mp:     25,
				Position: &gamev1.Position{
					X: 15.2,
					Y: 18.7,
					Z: 1.0,
				},
				Status: gamev1.UnitStatus_UNIT_STATUS_MOVING,
			},
		},
		Timestamp: 1699999999000,
	}

	data, err := proto.Marshal(original)
	require.NoError(t, err)

	goldenFile := filepath.Join(testDataDir, "battle_push_v1.bin")
	if err := ensureTestData(goldenFile, data); err != nil {
		t.Skipf("Cannot create test data: %v", err)
	}

	golden, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	parsed := &gamev1.BattlePush{}
	require.NoError(t, proto.Unmarshal(golden, parsed))

	// 验证关键字段
	assert.Equal(t, original.Seq, parsed.Seq)
	assert.Equal(t, original.BattleId, parsed.BattleId)
	assert.Equal(t, original.StateHash, parsed.StateHash)
	assert.Equal(t, len(original.Units), len(parsed.Units))
	assert.Equal(t, original.Timestamp, parsed.Timestamp)

	// 验证战斗单位
	for i, originalUnit := range original.Units {
		parsedUnit := parsed.Units[i]
		assert.Equal(t, originalUnit.UnitId, parsedUnit.UnitId)
		assert.Equal(t, originalUnit.Hp, parsedUnit.Hp)
		assert.Equal(t, originalUnit.Mp, parsedUnit.Mp)
		assert.Equal(t, originalUnit.Status, parsedUnit.Status)

		if originalUnit.Position != nil && parsedUnit.Position != nil {
			assert.InDelta(t, originalUnit.Position.X, parsedUnit.Position.X, 0.001)
			assert.InDelta(t, originalUnit.Position.Y, parsedUnit.Position.Y, 0.001)
			assert.InDelta(t, originalUnit.Position.Z, parsedUnit.Position.Z, 0.001)
		}
	}
}

// TestBackwardCompatibility_PlayerAction 测试玩家操作的向后兼容性
func TestBackwardCompatibility_PlayerAction(t *testing.T) {
	original := &gamev1.PlayerAction{
		ActionSeq:       98765,
		PlayerId:        "player_67890",
		ActionType:      gamev1.ActionType_ACTION_TYPE_SKILL,
		ClientTimestamp: 1699999999000,
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Skill{
				Skill: &gamev1.SkillAction{
					SkillId:       101,
					TargetUnitIds: []string{"unit_001", "unit_002", "unit_003"},
					CastPosition: &gamev1.Position{
						X: 25.0,
						Y: 30.0,
						Z: 2.0,
					},
				},
			},
		},
	}

	data, err := proto.Marshal(original)
	require.NoError(t, err)

	goldenFile := filepath.Join(testDataDir, "player_action_v1.bin")
	if err := ensureTestData(goldenFile, data); err != nil {
		t.Skipf("Cannot create test data: %v", err)
	}

	golden, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	parsed := &gamev1.PlayerAction{}
	require.NoError(t, proto.Unmarshal(golden, parsed))

	assert.Equal(t, original.ActionSeq, parsed.ActionSeq)
	assert.Equal(t, original.PlayerId, parsed.PlayerId)
	assert.Equal(t, original.ActionType, parsed.ActionType)
	assert.Equal(t, original.ClientTimestamp, parsed.ClientTimestamp)

	// 验证技能操作数据
	if originalSkill := original.ActionData.GetSkill(); originalSkill != nil {
		parsedSkill := parsed.ActionData.GetSkill()
		require.NotNil(t, parsedSkill)

		assert.Equal(t, originalSkill.SkillId, parsedSkill.SkillId)
		assert.Equal(t, originalSkill.TargetUnitIds, parsedSkill.TargetUnitIds)

		if originalSkill.CastPosition != nil && parsedSkill.CastPosition != nil {
			assert.InDelta(t, originalSkill.CastPosition.X, parsedSkill.CastPosition.X, 0.001)
			assert.InDelta(t, originalSkill.CastPosition.Y, parsedSkill.CastPosition.Y, 0.001)
			assert.InDelta(t, originalSkill.CastPosition.Z, parsedSkill.CastPosition.Z, 0.001)
		}
	}
}

// TestFieldEvolution 测试字段演进（新增字段）
func TestFieldEvolution(t *testing.T) {
	// 模拟旧版本的消息（没有某些新字段）
	oldMessage := &gamev1.LoginResp{
		Ok:       true,
		PlayerId: "old_player",
		// 注意：故意不设置 SessionId 和 ServerTime
	}

	oldData, err := proto.Marshal(oldMessage)
	require.NoError(t, err)

	// 新版本代码解析旧版本数据
	newMessage := &gamev1.LoginResp{}
	require.NoError(t, proto.Unmarshal(oldData, newMessage))

	// 验证旧字段保留
	assert.Equal(t, oldMessage.Ok, newMessage.Ok)
	assert.Equal(t, oldMessage.PlayerId, newMessage.PlayerId)

	// 新字段应该是默认值
	assert.Equal(t, "", newMessage.SessionId)
	assert.Equal(t, int64(0), newMessage.ServerTime)

	// 设置新字段后重新序列化
	newMessage.SessionId = "new_session"
	newMessage.ServerTime = 1699999999000

	newData, err := proto.Marshal(newMessage)
	require.NoError(t, err)

	// 验证新数据可以正确解析
	finalMessage := &gamev1.LoginResp{}
	require.NoError(t, proto.Unmarshal(newData, finalMessage))

	assert.Equal(t, newMessage.Ok, finalMessage.Ok)
	assert.Equal(t, newMessage.PlayerId, finalMessage.PlayerId)
	assert.Equal(t, newMessage.SessionId, finalMessage.SessionId)
	assert.Equal(t, newMessage.ServerTime, finalMessage.ServerTime)
}

// TestEnumEvolution 测试枚举演进
func TestEnumEvolution(t *testing.T) {
	// 测试未知枚举值的处理
	originalAction := &gamev1.PlayerAction{
		ActionSeq:       1,
		PlayerId:        "test_player",
		ActionType:      999, // 假设这是一个未来版本的新枚举值
		ClientTimestamp: 1699999999000,
	}

	data, err := proto.Marshal(originalAction)
	require.NoError(t, err)

	parsedAction := &gamev1.PlayerAction{}
	require.NoError(t, proto.Unmarshal(data, parsedAction))

	// 未知枚举值应该被保留
	assert.Equal(t, originalAction.ActionSeq, parsedAction.ActionSeq)
	assert.Equal(t, originalAction.PlayerId, parsedAction.PlayerId)
	assert.Equal(t, gamev1.ActionType(999), parsedAction.ActionType)
	assert.Equal(t, originalAction.ClientTimestamp, parsedAction.ClientTimestamp)
}

// TestMessageSizeGrowth 测试消息大小增长的兼容性
func TestMessageSizeGrowth(t *testing.T) {
	// 创建包含大量数据的消息
	largeMessage := &gamev1.BattlePush{
		Seq:       1,
		BattleId:  "large_battle",
		StateHash: make([]byte, 1024),              // 1KB的状态哈希
		Units:     make([]*gamev1.BattleUnit, 100), // 100个战斗单位
		Timestamp: 1699999999000,
	}

	// 填充随机状态哈希
	rand.Read(largeMessage.StateHash)

	// 填充战斗单位
	for i := 0; i < 100; i++ {
		largeMessage.Units[i] = &gamev1.BattleUnit{
			UnitId: fmt.Sprintf("unit_%03d", i),
			Hp:     int32(100 - i),
			Mp:     int32(50 + i),
			Position: &gamev1.Position{
				X: float32(i) * 2.5,
				Y: float32(i) * 1.8,
				Z: float32(i) * 0.1,
			},
			Status: gamev1.UnitStatus(i%4 + 1),
		}
	}

	// 序列化大消息
	data, err := proto.Marshal(largeMessage)
	require.NoError(t, err)

	t.Logf("Large message size: %d bytes", len(data))

	// 反序列化
	parsedMessage := &gamev1.BattlePush{}
	require.NoError(t, proto.Unmarshal(data, parsedMessage))

	// 验证数据完整性
	assert.Equal(t, largeMessage.Seq, parsedMessage.Seq)
	assert.Equal(t, largeMessage.BattleId, parsedMessage.BattleId)
	assert.True(t, bytes.Equal(largeMessage.StateHash, parsedMessage.StateHash))
	assert.Equal(t, len(largeMessage.Units), len(parsedMessage.Units))
	assert.Equal(t, largeMessage.Timestamp, parsedMessage.Timestamp)

	// 抽样验证战斗单位
	for i := 0; i < 10; i++ {
		idx := i * 10 // 每10个检查一个
		original := largeMessage.Units[idx]
		parsed := parsedMessage.Units[idx]

		assert.Equal(t, original.UnitId, parsed.UnitId)
		assert.Equal(t, original.Hp, parsed.Hp)
		assert.Equal(t, original.Mp, parsed.Mp)
		assert.Equal(t, original.Status, parsed.Status)
	}
}

// ensureTestData 确保测试数据文件存在
func ensureTestData(filename string, data []byte) error {
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// 如果文件不存在，创建它
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.WriteFile(filename, data, 0644)
	}

	return nil
}
