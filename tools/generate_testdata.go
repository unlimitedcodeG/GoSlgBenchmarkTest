package main

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"

	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

func generateTestDataMain() {
	// 创建testdata目录
	testdataDir := "testdata"
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		fmt.Printf("创建testdata目录失败: %v\n", err)
		os.Exit(1)
	}

	// 生成LoginResp测试数据
	loginResp := &gamev1.LoginResp{
		Ok:         true,
		PlayerId:   "player_12345",
		SessionId:  "session_abcdef",
		ServerTime: 1699999999000,
	}

	if err := saveProtobufData(filepath.Join(testdataDir, "login_resp_v1.bin"), loginResp); err != nil {
		fmt.Printf("保存LoginResp失败: %v\n", err)
		os.Exit(1)
	}

	// 生成BattlePush测试数据
	battlePush := &gamev1.BattlePush{
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

	if err := saveProtobufData(filepath.Join(testdataDir, "battle_push_v1.bin"), battlePush); err != nil {
		fmt.Printf("保存BattlePush失败: %v\n", err)
		os.Exit(1)
	}

	// 生成PlayerAction测试数据
	playerAction := &gamev1.PlayerAction{
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

	if err := saveProtobufData(filepath.Join(testdataDir, "player_action_v1.bin"), playerAction); err != nil {
		fmt.Printf("保存PlayerAction失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 测试数据生成完成!")
	fmt.Printf("   生成文件: %s/login_resp_v1.bin\n", testdataDir)
	fmt.Printf("   生成文件: %s/battle_push_v1.bin\n", testdataDir)
	fmt.Printf("   生成文件: %s/player_action_v1.bin\n", testdataDir)
}

func saveProtobufData(filename string, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
