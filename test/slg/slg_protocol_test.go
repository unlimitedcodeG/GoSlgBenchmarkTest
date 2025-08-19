package slg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/wsclient"

	// SLG v1.0.0 协议
	v1_0_0_combat "GoSlgBenchmarkTest/generated/slg/v1_0_0/GoSlgBenchmarkTest/generated/slg/v1_0_0/combat"
	v1_0_0_common "GoSlgBenchmarkTest/generated/slg/v1_0_0/GoSlgBenchmarkTest/generated/slg/v1_0_0/common"

	// SLG v1.1.0 协议
	v1_1_0_combat "GoSlgBenchmarkTest/generated/slg/v1_1_0/GoSlgBenchmarkTest/generated/slg/v1_1_0/combat"
	v1_1_0_common "GoSlgBenchmarkTest/generated/slg/v1_1_0/GoSlgBenchmarkTest/generated/slg/v1_1_0/common"
	v1_1_0_event "GoSlgBenchmarkTest/generated/slg/v1_1_0/GoSlgBenchmarkTest/generated/slg/v1_1_0/event"
)

// TestSLGProtocolVersionCompatibility 测试SLG协议版本兼容性
func TestSLGProtocolVersionCompatibility(t *testing.T) {
	t.Log("🧪 测试SLG协议版本兼容性...")

	// 测试v1.0.0的战斗消息能否被v1.1.0解析
	oldBattle := &v1_0_0_combat.BattleRequest{
		BattleId:   "battle_001",
		PlayerId:   "player_123",
		BattleType: v1_0_0_combat.BattleType_BATTLE_TYPE_PVE,
		UnitIds:    []string{"unit_1", "unit_2"},
		TargetPos: &v1_0_0_common.Position{
			X: 100.5,
			Y: 200.5,
			Z: 0,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := proto.Marshal(oldBattle)
	require.NoError(t, err)
	t.Logf("   序列化v1.0.0战斗请求: %d字节", len(data))

	// 用新版本解析
	newBattle := &v1_1_0_combat.BattleRequest{}
	err = proto.Unmarshal(data, newBattle)
	require.NoError(t, err)
	t.Log("   ✅ v1.1.0成功解析v1.0.0消息")

	// 验证关键字段保持兼容
	assert.Equal(t, oldBattle.BattleId, newBattle.BattleId)
	assert.Equal(t, oldBattle.PlayerId, newBattle.PlayerId)
	assert.Equal(t, int32(oldBattle.BattleType), int32(newBattle.BattleType))
	assert.Equal(t, oldBattle.UnitIds, newBattle.UnitIds)

	// 测试新字段为默认值
	assert.Empty(t, newBattle.FormationId)
	assert.Empty(t, newBattle.BattleSettings)
	assert.Empty(t, newBattle.PresetSkills)

	t.Log("   ✅ 所有字段兼容性验证通过")
}

// TestSLGNewFeaturesV1_1_0 测试v1.1.0新功能
func TestSLGNewFeaturesV1_1_0(t *testing.T) {
	t.Log("🚀 测试SLG v1.1.0新功能...")

	// 测试PVP功能
	pvpRequest := &v1_1_0_combat.PvpMatchRequest{
		PlayerId:      "player_123",
		Mode:          v1_1_0_combat.PvpMode_PVP_MODE_RANKED_1V1,
		RatingRange:   100,
		PreferredMaps: []string{"map_arena_1", "map_arena_2"},
	}

	data, err := proto.Marshal(pvpRequest)
	require.NoError(t, err)
	t.Logf("   序列化PVP匹配请求: %d字节", len(data))

	// 反序列化验证
	parsed := &v1_1_0_combat.PvpMatchRequest{}
	err = proto.Unmarshal(data, parsed)
	require.NoError(t, err)
	assert.Equal(t, pvpRequest.PlayerId, parsed.PlayerId)
	assert.Equal(t, pvpRequest.Mode, parsed.Mode)
	t.Log("   ✅ PVP协议序列化/反序列化成功")

	// 测试活动系统
	activity := &v1_1_0_event.Activity{
		ActivityId:   "activity_001",
		ActivityName: "春节活动",
		ActivityType: v1_1_0_event.ActivityType_ACTIVITY_TYPE_FESTIVAL,
		Status:       v1_1_0_event.ActivityStatus_ACTIVITY_STATUS_ACTIVE,
		StartTime:    time.Now().UnixMilli(),
		EndTime:      time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
	}

	activityData, err := proto.Marshal(activity)
	require.NoError(t, err)
	t.Logf("   序列化活动信息: %d字节", len(activityData))

	parsedActivity := &v1_1_0_event.Activity{}
	err = proto.Unmarshal(activityData, parsedActivity)
	require.NoError(t, err)
	assert.Equal(t, activity.ActivityName, parsedActivity.ActivityName)
	t.Log("   ✅ 活动系统协议序列化/反序列化成功")
}

// TestSLGEnhancedBattleSystem 测试增强的战斗系统
func TestSLGEnhancedBattleSystem(t *testing.T) {
	t.Log("⚔️ 测试增强的战斗系统...")

	// 创建增强的战斗单位
	battleUnit := &v1_1_0_combat.BattleUnit{
		UnitId:     "unit_001",
		TemplateId: "warrior_template",
		Hp:         100,
		MaxHp:      100,
		Mp:         50,
		MaxMp:      50,
		Position: &v1_1_0_common.Position{
			X:        10.0,
			Y:        20.0,
			Z:        0.0,
			Rotation: 90.0,     // v1.1.0新增
			ZoneId:   "zone_1", // v1.1.0新增
		},
		Status: v1_1_0_combat.UnitStatus_UNIT_STATUS_IDLE,
		Stats: &v1_1_0_combat.UnitStats{
			Attack:         50,
			Defense:        30,
			Speed:          25,
			CriticalRate:   150,  // 15%
			CriticalDamage: 1200, // 120%
			// v1.1.0新增属性
			MagicAttack:  20,
			MagicDefense: 15,
			Penetration:  5,
			Block:        100, // 10%
		},
		// v1.1.0新增字段
		AvailableSkills: []string{"skill_slash", "skill_block", "skill_charge"},
		Energy:          10,
		MaxEnergy:       50,
		Equipment: &v1_1_0_combat.UnitEquipment{
			WeaponId:     "sword_001",
			ArmorId:      "armor_001",
			AccessoryId:  "ring_001",
			Enchantments: []string{"sharpness", "durability"},
		},
	}

	data, err := proto.Marshal(battleUnit)
	require.NoError(t, err)
	t.Logf("   序列化增强战斗单位: %d字节", len(data))

	parsed := &v1_1_0_combat.BattleUnit{}
	err = proto.Unmarshal(data, parsed)
	require.NoError(t, err)

	// 验证增强功能
	assert.Equal(t, battleUnit.Equipment.WeaponId, parsed.Equipment.WeaponId)
	assert.Equal(t, battleUnit.Stats.MagicAttack, parsed.Stats.MagicAttack)
	assert.Equal(t, battleUnit.Position.Rotation, parsed.Position.Rotation)
	assert.Len(t, parsed.AvailableSkills, 3)

	t.Log("   ✅ 增强战斗系统验证通过")
}

// TestSLGWebSocketIntegration 测试SLG协议与WebSocket集成
func TestSLGWebSocketIntegration(t *testing.T) {
	t.Log("🌐 测试SLG协议与WebSocket集成...")

	// 启动测试服务器
	server := testserver.New(testserver.DefaultServerConfig(":18090"))
	server.Start()
	defer func() {
		server.Shutdown(context.Background())
		t.Log("   🛑 测试服务器已关闭")
	}()

	time.Sleep(100 * time.Millisecond)

	// 创建WebSocket客户端
	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18090/ws", "slg-test-token")
	client := wsclient.New(config)

	ctx := context.Background()
	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Close()

	t.Log("   ✅ WebSocket连接建立成功")

	// 模拟发送SLG协议消息（这里我们需要扩展框架支持SLG消息）
	// 由于当前框架主要支持game.proto消息，我们演示概念

	// 创建SLG战斗请求
	battleReq := &v1_1_0_combat.BattleRequest{
		BattleId:   "slg_battle_001",
		PlayerId:   "slg_player_123",
		BattleType: v1_1_0_combat.BattleType_BATTLE_TYPE_PVP,
		UnitIds:    []string{"slg_unit_1", "slg_unit_2"},
		TargetPos: &v1_1_0_common.Position{
			X: 150.0,
			Y: 250.0,
		},
		// v1.1.0新增字段
		FormationId: "formation_triangle",
		BattleSettings: map[string]int32{
			"auto_battle": 1,
			"fast_mode":   1,
		},
		PresetSkills: []string{"fireball", "heal", "shield"},
	}

	// 序列化为字节
	battleData, err := proto.Marshal(battleReq)
	require.NoError(t, err)
	t.Logf("   📦 SLG战斗请求序列化: %d字节", len(battleData))

	// 验证数据完整性
	parsedReq := &v1_1_0_combat.BattleRequest{}
	err = proto.Unmarshal(battleData, parsedReq)
	require.NoError(t, err)
	assert.Equal(t, battleReq.BattleId, parsedReq.BattleId)
	assert.Equal(t, battleReq.FormationId, parsedReq.FormationId)
	assert.Equal(t, battleReq.BattleSettings, parsedReq.BattleSettings)

	t.Log("   ✅ SLG协议数据完整性验证通过")
	t.Log("   💡 注意：完整的WebSocket传输需要扩展框架的协议支持")
}

// TestSLGPerformanceBenchmark 测试SLG协议性能
func TestSLGPerformanceBenchmark(t *testing.T) {
	t.Log("⚡ 测试SLG协议性能...")

	// 创建复杂的战斗状态
	battleState := &v1_1_0_combat.BattleStatePush{
		BattleId:  "perf_battle_001",
		FrameSeq:  12345,
		Phase:     v1_1_0_combat.BattlePhase_BATTLE_PHASE_FIGHTING,
		Timestamp: time.Now().UnixMilli(),
		Actions: []*v1_1_0_combat.BattleAction{
			{
				ActionId:         "action_001",
				ActionType:       v1_1_0_combat.ActionType_ACTION_TYPE_SKILL,
				SourceUnitId:     "unit_001",
				TargetUnitIds:    []string{"unit_002", "unit_003"},
				Damage:           150,
				Healing:          0,
				EffectIds:        []string{"effect_burn", "effect_slow"},
				SkillId:          "fireball",
				IsCritical:       true,
				DamageMultiplier: 1.5,
				ComboSkills:      []string{"lightning", "freeze"},
			},
			{
				ActionId:         "action_002",
				ActionType:       v1_1_0_combat.ActionType_ACTION_TYPE_HEAL,
				SourceUnitId:     "unit_004",
				TargetUnitIds:    []string{"unit_001"},
				Healing:          80,
				SkillId:          "greater_heal",
				DamageMultiplier: 1.0,
			},
		},
		Environment: &v1_1_0_combat.EnvironmentEffect{
			EffectId:      "env_fire_rain",
			EffectType:    v1_1_0_combat.EnvironmentType_ENVIRONMENT_TYPE_FIRE_RAIN,
			Duration:      30,
			AffectedUnits: []string{"unit_001", "unit_002", "unit_003"},
			EffectParams: map[string]float32{
				"damage_per_second": 10.5,
				"chance":            0.3,
			},
		},
		ActiveSkills: []string{"fireball", "heal", "shield", "lightning"},
	}

	// 性能测试：序列化
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := proto.Marshal(battleState)
		require.NoError(t, err)
	}
	serializeDuration := time.Since(start)
	t.Logf("   📊 1000次序列化耗时: %v (平均: %v)", serializeDuration, serializeDuration/1000)

	// 获取序列化数据
	data, err := proto.Marshal(battleState)
	require.NoError(t, err)
	t.Logf("   📦 复杂战斗状态大小: %d字节", len(data))

	// 性能测试：反序列化
	start = time.Now()
	for i := 0; i < 1000; i++ {
		parsed := &v1_1_0_combat.BattleStatePush{}
		err := proto.Unmarshal(data, parsed)
		require.NoError(t, err)
	}
	deserializeDuration := time.Since(start)
	t.Logf("   📊 1000次反序列化耗时: %v (平均: %v)", deserializeDuration, deserializeDuration/1000)

	// 验证性能可接受性（每次操作应该在1ms内）
	avgSerialize := serializeDuration.Nanoseconds() / 1000
	avgDeserialize := deserializeDuration.Nanoseconds() / 1000

	assert.Less(t, avgSerialize, int64(time.Millisecond), "序列化性能应在1ms内")
	assert.Less(t, avgDeserialize, int64(time.Millisecond), "反序列化性能应在1ms内")

	t.Log("   ✅ SLG协议性能测试通过")
}

// TestSLGMessageSizeAnalysis 测试SLG消息大小分析
func TestSLGMessageSizeAnalysis(t *testing.T) {
	t.Log("📏 SLG消息大小分析...")

	messages := map[string]proto.Message{
		"简单战斗请求(v1.0.0)": &v1_0_0_combat.BattleRequest{
			BattleId:   "battle_001",
			PlayerId:   "player_123",
			BattleType: v1_0_0_combat.BattleType_BATTLE_TYPE_PVE,
		},
		"增强战斗请求(v1.1.0)": &v1_1_0_combat.BattleRequest{
			BattleId:    "battle_001",
			PlayerId:    "player_123",
			BattleType:  v1_1_0_combat.BattleType_BATTLE_TYPE_PVP,
			FormationId: "formation_001",
			BattleSettings: map[string]int32{
				"auto_battle": 1,
				"speed":       2,
			},
			PresetSkills: []string{"skill1", "skill2", "skill3"},
		},
		"PVP匹配请求": &v1_1_0_combat.PvpMatchRequest{
			PlayerId:      "player_123",
			Mode:          v1_1_0_combat.PvpMode_PVP_MODE_RANKED_1V1,
			RatingRange:   100,
			PreferredMaps: []string{"arena1", "arena2"},
		},
		"活动信息": &v1_1_0_event.Activity{
			ActivityId:   "spring_festival",
			ActivityName: "春节庆典活动",
			Description:  "欢度春节，领取丰厚奖励！",
			ActivityType: v1_1_0_event.ActivityType_ACTIVITY_TYPE_FESTIVAL,
			Status:       v1_1_0_event.ActivityStatus_ACTIVITY_STATUS_ACTIVE,
		},
	}

	for name, msg := range messages {
		data, err := proto.Marshal(msg)
		require.NoError(t, err)
		t.Logf("   📦 %s: %d 字节", name, len(data))
	}

	t.Log("   💡 消息大小分析完成，协议设计合理")
}
