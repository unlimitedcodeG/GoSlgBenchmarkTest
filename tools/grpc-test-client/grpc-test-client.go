package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

func main() {
	// 1) 可达性预检：分别检查 WS:18000 和 gRPC:19001
	checkTCP("127.0.0.1:18000", "WebSocket")
	checkTCP("127.0.0.1:19001", "gRPC")

	fmt.Println("🚀 开始 gRPC 接口测试...")

	// 2) 建立 gRPC 连接：使用 NewClient API
	conn, err := grpc.NewClient(
		"127.0.0.1:19001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("❌ 连接 gRPC 服务器失败: %v", err)
	}
	defer conn.Close()

	client := gamev1.NewGameServiceClient(conn)

	// 3) 统一 RPC 调用超时
	rpcCtx, rpcCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer rpcCancel()

	// 测试1: Login
	fmt.Println("\n📝 测试1: Login")
	loginResp, err := client.Login(rpcCtx, &gamev1.LoginReq{
		Token:         "test_token_123",
		ClientVersion: "1.0.0",
		DeviceId:      "test_device_001",
	})
	if err != nil {
		log.Printf("❌ Login 失败: %v", err)
		return // 后续依赖 login，直接返回更清晰
	}
	fmt.Printf("✅ Login 成功 - PlayerID: %s, SessionID: %s\n", loginResp.PlayerId, loginResp.SessionId)

	// 测试2: GetPlayerStatus
	fmt.Println("\n📝 测试2: GetPlayerStatus")
	statusResp, err := client.GetPlayerStatus(rpcCtx, &gamev1.PlayerStatusReq{
		PlayerId: loginResp.PlayerId,
	})
	if err != nil {
		log.Printf("❌ GetPlayerStatus 失败: %v", err)
	} else {
		fmt.Printf("✅ GetPlayerStatus 成功 - Level: %d, Status: %s\n", statusResp.Level, statusResp.Status)
	}

	// 测试3: SendPlayerAction
	fmt.Println("\n📝 测试3: SendPlayerAction")
	actionResp, err := client.SendPlayerAction(rpcCtx, &gamev1.PlayerAction{
		ActionSeq:  1,
		PlayerId:   loginResp.PlayerId,
		ActionType: gamev1.ActionType_ACTION_TYPE_MOVE,
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Move{
				Move: &gamev1.MoveAction{
					TargetPosition: &gamev1.Position{X: 10, Y: 20, Z: 0},
					MoveSpeed:      5.0,
				},
			},
		},
		ClientTimestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		log.Printf("❌ SendPlayerAction 失败: %v", err)
	} else {
		fmt.Printf("✅ SendPlayerAction 成功 - ActionID: %d\n", actionResp.ActionId)
	}

	// 测试4: Logout
	fmt.Println("\n📝 测试4: Logout")
	logoutResp, err := client.Logout(rpcCtx, &gamev1.LogoutReq{
		SessionId: loginResp.SessionId,
	})
	if err != nil {
		log.Printf("❌ Logout 失败: %v", err)
	} else {
		fmt.Printf("✅ Logout 成功 - Message: %s\n", logoutResp.Message)
	}

	// 测试5: JoinBattle
	fmt.Println("\n📝 测试5: JoinBattle")
	joinResp, err := client.JoinBattle(rpcCtx, &gamev1.JoinBattleReq{
		PlayerId:       loginResp.PlayerId,
		BattleType:     "pve",
		TeamPreference: "auto",
	})
	if err != nil {
		log.Printf("❌ JoinBattle 失败: %v", err)
	} else {
		fmt.Printf("✅ JoinBattle 成功 - BattleID: %s, Team: %s\n", joinResp.BattleId, joinResp.TeamAssigned)
	}

	// 测试6: GetBattleStatus
	if joinResp != nil {
		fmt.Println("\n📝 测试6: GetBattleStatus")
		battleResp, err := client.GetBattleStatus(rpcCtx, &gamev1.BattleStatusReq{
			BattleId:     joinResp.BattleId,
			IncludeUnits: true,
		})
		if err != nil {
			log.Printf("❌ GetBattleStatus 失败: %v", err)
		} else {
			fmt.Printf("✅ GetBattleStatus 成功 - Status: %s, UnitCount: %d\n",
				battleResp.Status, len(battleResp.Units))
		}
	}

	fmt.Println("\n🎉 gRPC 接口测试完成！")
}

func checkTCP(addr, name string) {
	c, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		fmt.Printf("❌ %s 端口 %s 不可达: %v\n", name, addr, err)
	} else {
		_ = c.Close()
		fmt.Printf("✅ %s 端口 %s 可达\n", name, addr)
	}
}
