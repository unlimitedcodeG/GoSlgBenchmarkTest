package protocol

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	// SLG v1.0.0 协议
	v1_0_0_building "GoSlgBenchmarkTest/generated/slg/v1_0_0/building"
	v1_0_0_combat "GoSlgBenchmarkTest/generated/slg/v1_0_0/combat"
	v1_0_0_common "GoSlgBenchmarkTest/generated/slg/v1_0_0/common"

	// SLG v1.1.0 协议
	v1_1_0_building "GoSlgBenchmarkTest/generated/slg/v1_1_0/building"
	v1_1_0_combat "GoSlgBenchmarkTest/generated/slg/v1_1_0/combat"
	v1_1_0_common "GoSlgBenchmarkTest/generated/slg/v1_1_0/common"
	v1_1_0_event "GoSlgBenchmarkTest/generated/slg/v1_1_0/event"
)

// SLG协议操作码定义
const (
	// 战斗相关
	OpSLGBattleRequest  uint16 = 5001
	OpSLGBattleResponse uint16 = 5002
	OpSLGBattleUpdate   uint16 = 5003
	OpSLGBattleEnd      uint16 = 5004

	// 建筑相关
	OpSLGCityUpdate       uint16 = 5101
	OpSLGBuildingUpgrade  uint16 = 5102
	OpSLGBuildingComplete uint16 = 5103

	// 活动相关 (v1.1.0+)
	OpSLGActivityStart  uint16 = 5201
	OpSLGActivityUpdate uint16 = 5202
	OpSLGActivityEnd    uint16 = 5203

	// PVP相关 (v1.1.0+)
	OpSLGPVPRequest  uint16 = 5301
	OpSLGPVPResponse uint16 = 5302
	OpSLGPVPUpdate   uint16 = 5303
)

// SLGMessageAdapter SLG协议消息适配器
type SLGMessageAdapter struct {
	version string
}

// NewSLGMessageAdapter 创建SLG消息适配器
func NewSLGMessageAdapter(version string) *SLGMessageAdapter {
	return &SLGMessageAdapter{
		version: version,
	}
}

// EncodeMessage 编码SLG消息
func (adapter *SLGMessageAdapter) EncodeMessage(opcode uint16, message proto.Message) ([]byte, error) {
	// 序列化消息
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SLG message: %v", err)
	}

	// 使用现有的帧编码
	frame := EncodeFrame(opcode, data)
	return frame, nil
}

// DecodeMessage 解码SLG消息
func (adapter *SLGMessageAdapter) DecodeMessage(frame []byte) (uint16, proto.Message, error) {
	// 解码帧
	opcode, data, err := DecodeFrame(frame)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to decode frame: %v", err)
	}

	// 根据操作码和版本创建对应的消息类型
	message, err := adapter.createMessageByOpcode(opcode)
	if err != nil {
		return 0, nil, err
	}

	// 反序列化消息
	if err := proto.Unmarshal(data, message); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal SLG message: %v", err)
	}

	return opcode, message, nil
}

// createMessageByOpcode 根据操作码创建消息实例
func (adapter *SLGMessageAdapter) createMessageByOpcode(opcode uint16) (proto.Message, error) {
	switch adapter.version {
	case "v1.0.0":
		return adapter.createV1_0_0Message(opcode)
	case "v1.1.0":
		return adapter.createV1_1_0Message(opcode)
	default:
		return nil, fmt.Errorf("unsupported SLG protocol version: %s", adapter.version)
	}
}

// createV1_0_0Message 创建v1.0.0版本消息
func (adapter *SLGMessageAdapter) createV1_0_0Message(opcode uint16) (proto.Message, error) {
	switch opcode {
	case OpSLGBattleRequest:
		return &v1_0_0_combat.BattleRequest{}, nil
	case OpSLGBattleResponse:
		return &v1_0_0_combat.BattleResponse{}, nil
	case OpSLGBattleUpdate:
		// BattleUpdate 不存在，使用 BattleResponse 作为更新
		return &v1_0_0_combat.BattleResponse{}, nil
	case OpSLGCityUpdate:
		// CityUpdate 不存在，使用 CityInfo 作为更新
		return &v1_0_0_building.CityInfo{}, nil
	default:
		return nil, fmt.Errorf("unknown SLG v1.0.0 opcode: %d", opcode)
	}
}

// createV1_1_0Message 创建v1.1.0版本消息
func (adapter *SLGMessageAdapter) createV1_1_0Message(opcode uint16) (proto.Message, error) {
	switch opcode {
	case OpSLGBattleRequest:
		return &v1_1_0_combat.BattleRequest{}, nil
	case OpSLGBattleResponse:
		return &v1_1_0_combat.BattleResponse{}, nil
	case OpSLGBattleUpdate:
		// BattleUpdate 不存在，使用 BattleResponse 作为更新
		return &v1_1_0_combat.BattleResponse{}, nil
	case OpSLGCityUpdate:
		// CityUpdate 不存在，使用 CityInfo 作为更新
		return &v1_1_0_building.CityInfo{}, nil
	case OpSLGActivityStart:
		// ActivityStart 不存在，使用 Activity 作为开始事件
		return &v1_1_0_event.Activity{}, nil
	case OpSLGActivityUpdate:
		// ActivityUpdate 不存在，使用 Activity 作为更新
		return &v1_1_0_event.Activity{}, nil
	case OpSLGActivityEnd:
		// ActivityEnd 不存在，使用 Activity 作为结束事件
		return &v1_1_0_event.Activity{}, nil
	case OpSLGPVPRequest:
		return &v1_1_0_combat.PvpMatchRequest{}, nil
	case OpSLGPVPResponse:
		return &v1_1_0_combat.PvpMatchResponse{}, nil
	case OpSLGPVPUpdate:
		// PVPUpdate 不存在，使用 PvpBattleResult 作为更新
		return &v1_1_0_combat.PvpBattleResult{}, nil
	default:
		return nil, fmt.Errorf("unknown SLG v1.1.0 opcode: %d", opcode)
	}
}

// SLGTestDataGenerator SLG测试数据生成器
type SLGTestDataGenerator struct {
	version string
}

// NewSLGTestDataGenerator 创建SLG测试数据生成器
func NewSLGTestDataGenerator(version string) *SLGTestDataGenerator {
	return &SLGTestDataGenerator{
		version: version,
	}
}

// GenerateBattleRequest 生成战斗请求
func (gen *SLGTestDataGenerator) GenerateBattleRequest(battleID, playerID string, battleType int32) (proto.Message, error) {
	switch gen.version {
	case "v1.0.0":
		return &v1_0_0_combat.BattleRequest{
			BattleId:   battleID,
			PlayerId:   playerID,
			BattleType: v1_0_0_combat.BattleType(battleType),
			UnitIds:    []string{"unit_1", "unit_2", "unit_3"},
			TargetPos: &v1_0_0_common.Position{
				X: 100.0,
				Y: 200.0,
				Z: 0.0,
			},
		}, nil
	case "v1.1.0":
		return &v1_1_0_combat.BattleRequest{
			BattleId:   battleID,
			PlayerId:   playerID,
			BattleType: v1_1_0_combat.BattleType(battleType),
			UnitIds:    []string{"unit_1", "unit_2", "unit_3"},
			TargetPos: &v1_1_0_common.Position{
				X: 100.0,
				Y: 200.0,
				Z: 0.0,
			},
			// v1.1.0 新增字段
			FormationId: "triangle",
			BattleSettings: map[string]int32{
				"auto_battle": 1,
				"fast_mode":   1,
			},
			PresetSkills: []string{"fireball", "heal", "shield"},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported version for battle request: %s", gen.version)
	}
}

// GenerateCityUpdate 生成城市更新
func (gen *SLGTestDataGenerator) GenerateCityUpdate(cityID, playerID string) (proto.Message, error) {
	switch gen.version {
	case "v1.0.0":
		return &v1_0_0_building.CityInfo{
			CityId:    cityID,
			PlayerId:  playerID,
			CityLevel: 5,
			Buildings: []*v1_0_0_building.BuildingInfo{
				{
					BuildingId:    "building_1",
					BuildingType:  v1_0_0_building.BuildingType_BUILDING_TYPE_BARRACKS,
					BuildingLevel: 3,
					Position: &v1_0_0_common.Position{
						X: 10.0,
						Y: 20.0,
						Z: 0.0,
					},
				},
			},
		}, nil
	case "v1.1.0":
		return &v1_1_0_building.CityInfo{
			CityId:    cityID,
			PlayerId:  playerID,
			CityLevel: 5,
			Buildings: []*v1_1_0_building.BuildingInfo{
				{
					BuildingId:    "building_1",
					BuildingType:  v1_1_0_building.BuildingType_BUILDING_TYPE_BARRACKS,
					BuildingLevel: 3,
					Position: &v1_1_0_common.Position{
						X: 10.0,
						Y: 20.0,
						Z: 0.0,
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported version for city update: %s", gen.version)
	}
}

// GenerateActivityStart 生成活动开始（仅v1.1.0+）
func (gen *SLGTestDataGenerator) GenerateActivityStart(activityID string) (proto.Message, error) {
	if gen.version != "v1.1.0" {
		return nil, fmt.Errorf("activity events are only supported in v1.1.0+")
	}

	return &v1_1_0_event.Activity{
		ActivityId:   activityID,
		ActivityName: "Test Battle Activity",
		Description:  "A test battle activity",
		ActivityType: v1_1_0_event.ActivityType_ACTIVITY_TYPE_BATTLE,
		StartTime:    1699999999000,
		EndTime:      1699999999000 + 3600*1000, // 1小时后结束
		Rewards: []*v1_1_0_event.ActivityReward{
			{
				RewardId:   "reward_1",
				RewardName: "Gold Reward",
				RewardType: v1_1_0_event.ActivityRewardType_ACTIVITY_REWARD_TYPE_COMPLETION,
				Gold:       1000,
				Resources: map[string]int32{
					"food": 500,
				},
			},
		},
		Condition: &v1_1_0_event.ActivityCondition{
			MinLevel:       10,
			RequiredQuests: []string{"tutorial_complete"},
			VipOnly:        false,
		},
		Config: map[string]string{
			"min_level": "10",
		},
	}, nil
}

// GeneratePVPRequest 生成PVP请求（仅v1.1.0+）
func (gen *SLGTestDataGenerator) GeneratePVPRequest(playerID, targetID string) (proto.Message, error) {
	if gen.version != "v1.1.0" {
		return nil, fmt.Errorf("PVP features are only supported in v1.1.0+")
	}

	return &v1_1_0_combat.PvpMatchRequest{
		PlayerId:      playerID,
		Mode:          v1_1_0_combat.PvpMode_PVP_MODE_RANKED_1V1, // 使用排位1v1模式
		RatingRange:   100,
		PreferredMaps: []string{"battle_arena_1", "battle_arena_2"},
		TeamId:        targetID, // 对手作为目标队伍
	}, nil
}

// IsVersionSupported 检查版本是否支持
func IsVersionSupported(version string) bool {
	switch version {
	case "v1.0.0", "v1.1.0":
		return true
	default:
		return false
	}
}

// GetSupportedOpcodes 获取版本支持的操作码
func GetSupportedOpcodes(version string) []uint16 {
	switch version {
	case "v1.0.0":
		return []uint16{
			OpSLGBattleRequest,
			OpSLGBattleResponse,
			OpSLGBattleUpdate,
			OpSLGBattleEnd,
			OpSLGCityUpdate,
			OpSLGBuildingUpgrade,
			OpSLGBuildingComplete,
		}
	case "v1.1.0":
		return []uint16{
			OpSLGBattleRequest,
			OpSLGBattleResponse,
			OpSLGBattleUpdate,
			OpSLGBattleEnd,
			OpSLGCityUpdate,
			OpSLGBuildingUpgrade,
			OpSLGBuildingComplete,
			OpSLGActivityStart,
			OpSLGActivityUpdate,
			OpSLGActivityEnd,
			OpSLGPVPRequest,
			OpSLGPVPResponse,
			OpSLGPVPUpdate,
		}
	default:
		return nil
	}
}

// ValidateCompatibility 验证协议版本兼容性
func ValidateCompatibility(fromVersion, toVersion string) error {
	if !IsVersionSupported(fromVersion) {
		return fmt.Errorf("unsupported source version: %s", fromVersion)
	}
	if !IsVersionSupported(toVersion) {
		return fmt.Errorf("unsupported target version: %s", toVersion)
	}

	// v1.0.0 可以向上兼容到 v1.1.0
	if fromVersion == "v1.0.0" && toVersion == "v1.1.0" {
		return nil
	}

	// 同版本兼容
	if fromVersion == toVersion {
		return nil
	}

	return fmt.Errorf("incompatible version transition: %s -> %s", fromVersion, toVersion)
}
