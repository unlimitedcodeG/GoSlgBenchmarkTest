package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"GoSlgBenchmarkTest/internal/protocol"
	"GoSlgBenchmarkTest/internal/session"
	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/wsclient"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"

	"google.golang.org/protobuf/proto"
)

func main() {
	fmt.Println("🎯 真实长连接会话测试框架演示")
	fmt.Println("==================================")
	fmt.Println()

	// 1. 启动测试服务器
	fmt.Println("🚀 启动测试服务器...")
	server := testserver.New(testserver.DefaultServerConfig(":18090"))
	if err := server.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)
	fmt.Println("✅ 测试服务器已启动")

	// 2. 创建会话录制器
	fmt.Println("\n📹 创建会话录制器...")
	sessionID := fmt.Sprintf("demo_session_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)
	fmt.Printf("✅ 会话录制器已创建: %s\n", sessionID)

	// 3. 创建WebSocket客户端并录制会话
	fmt.Println("\n🔗 建立WebSocket连接并录制会话...")

	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18090/ws", "demo-token")
	client := wsclient.New(config)

	// 设置消息处理器
	client.SetPushHandler(func(opcode uint16, message proto.Message) {
		recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
			"opcode":       opcode,
			"message_type": fmt.Sprintf("%T", message),
		})
	})

	client.SetStateChangeHandler(func(oldState, newState wsclient.ClientState) {
		recorder.RecordEvent(session.EventConnect, map[string]interface{}{
			"old_state": oldState.String(),
			"new_state": newState.String(),
		})
	})

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("连接失败: %v", err)
	}

	// 记录登录
	recorder.RecordEvent(session.EventLogin, map[string]interface{}{
		"player_id": "demo_player",
		"status":    "success",
	})

	// 发送一些消息
	fmt.Println("📤 发送测试消息...")
	for i := 0; i < 10; i++ {
		action := &gamev1.PlayerAction{
			ActionSeq:       uint64(i + 1),
			PlayerId:        "demo_player",
			ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
			ClientTimestamp: time.Now().UnixMilli(),
		}

		// 记录发送
		recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
			"opcode":       protocol.OpPlayerAction,
			"sequence_num": uint64(i + 1),
			"action_type":  action.ActionType.String(),
		})

		client.SendAction(action)
		time.Sleep(200 * time.Millisecond)

		// 模拟延迟
		latency := time.Duration(50+i*10) * time.Millisecond
		recorder.RecordLatency(latency)
	}

	// 模拟一些网络问题
	fmt.Println("🌐 模拟网络问题...")
	recorder.RecordEvent(session.EventError, map[string]interface{}{
		"error_type": "network_jitter",
		"details":    "Simulated network jitter",
	})

	recorder.RecordEvent(session.EventReconnect, map[string]interface{}{
		"attempt":  1,
		"duration": 1500 * time.Millisecond,
		"success":  true,
	})

	// 等待更多推送消息
	time.Sleep(2 * time.Second)

	// 关闭连接
	client.Close()
	recorder.RecordClose(session.CloseNormal, "Demo completed")

	// 4. 获取录制的会话
	fmt.Println("\n📋 获取录制的会话数据...")
	recordedSession := recorder.GetSession()
	fmt.Printf("✅ 录制完成: %d 个事件\n", len(recordedSession.Events))

	// 5. 会话回放演示
	fmt.Println("\n🔄 演示会话回放功能...")

	replayConfig := &session.ReplayConfig{
		Speed:        session.SpeedFast,
		EnablePause:  true,
		PauseOnError: false,
	}

	replayer := session.NewSessionReplayer(recordedSession, replayConfig)

	// 添加回放回调
	var replayedCount int
	replayer.AddCallback(func(event *session.ReplayEvent) error {
		replayedCount++
		if replayedCount%5 == 0 {
			fmt.Printf("   📺 已回放 %d 个事件\n", replayedCount)
		}
		return nil
	})

	// 开始回放
	if err := replayer.Play(); err != nil {
		log.Printf("回放失败: %v", err)
	}

	// 等待回放完成
	time.Sleep(3 * time.Second)
	replayer.Stop()

	replayStats := replayer.GetStats()
	fmt.Printf("✅ 回放完成: %d/%d 事件重放成功\n",
		replayStats.ReplayedEvents, replayStats.TotalEvents)

	// 6. 断言测试演示
	fmt.Println("\n🧪 演示断言测试功能...")

	suite := session.NewAssertionSuite("Demo Quality Test", "验证演示会话质量")

	// 添加各种断言
	suite.AddAssertion(session.NewMessageOrderAssertion(
		"Message Order Check",
		"验证消息按顺序接收",
		protocol.OpPlayerAction,
		5,  // 最小数量
		15, // 最大数量
	))

	suite.AddAssertion(session.NewLatencyAssertion(
		"Latency Check",
		"验证延迟在可接受范围内",
		200*time.Millisecond, // 最大延迟
		90,                   // 90%分位数
	))

	suite.AddAssertion(session.NewReconnectAssertion(
		"Reconnect Check",
		"验证重连次数和耗时",
		2,             // 最大重连次数
		5*time.Second, // 最大重连耗时
	))

	suite.AddAssertion(session.NewErrorRateAssertion(
		"Error Rate Check",
		"验证错误率在可接受范围内",
		0.2, // 最大错误率 20%
	))

	// 运行断言
	suite.RunAssertions(recordedSession)

	passedCount := suite.GetPassedCount()
	failedCount := suite.GetFailedCount()
	successRate := suite.GetSuccessRate()

	fmt.Printf("📊 断言结果: %d 通过, %d 失败, 成功率: %.1f%%\n",
		passedCount, failedCount, successRate*100)

	// 7. 时间线分析演示
	fmt.Println("\n📊 演示时间线分析功能...")

	analyzer := session.NewTimelineAnalyzer(recordedSession)

	// 分析时间线
	timeline := analyzer.AnalyzeTimeline()
	fmt.Printf("📈 时间线分析: %d 个事件\n", len(timeline))

	// 分析消息流
	flows := analyzer.AnalyzeMessageFlows()
	fmt.Printf("🔄 消息流分析: %d 个消息流\n", len(flows))

	// 计算网络指标
	metrics := analyzer.CalculateNetworkMetrics()
	fmt.Printf("📊 网络指标:\n")
	fmt.Printf("   总消息数: %d\n", metrics.TotalMessages)
	fmt.Printf("   成功消息: %d\n", metrics.SuccessfulMessages)
	fmt.Printf("   平均延迟: %v\n", metrics.AverageLatency)
	fmt.Printf("   丢包率: %.2f%%\n", metrics.PacketLoss*100)
	fmt.Printf("   吞吐量: %.2f msg/s\n", metrics.Throughput)

	// 分析连接稳定性
	stability := analyzer.AnalyzeConnectionStability()
	fmt.Printf("🔗 连接稳定性:\n")
	fmt.Printf("   连接次数: %v\n", stability["total_connections"])
	fmt.Printf("   重连次数: %v\n", stability["reconnect_count"])
	fmt.Printf("   重连率: %.2f%%\n", stability["reconnect_rate"].(float64)*100)

	// 8. 导出会话数据
	fmt.Println("\n💾 导出会话数据...")

	exportedData, err := recorder.ExportJSON()
	if err != nil {
		log.Printf("导出失败: %v", err)
	} else {
		// 保存到文件
		filename := fmt.Sprintf("session_%s.json", sessionID)
		if err := os.WriteFile(filename, exportedData, 0644); err != nil {
			log.Printf("保存文件失败: %v", err)
		} else {
			fmt.Printf("✅ 会话数据已保存到: %s (%d 字节)\n", filename, len(exportedData))
		}
	}

	// 9. 生成时间线报告
	fmt.Println("\n📋 生成时间线报告...")
	report := analyzer.GenerateTimelineReport()

	reportFilename := fmt.Sprintf("timeline_report_%s.json", sessionID)
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("生成报告失败: %v", err)
	} else {
		if err := os.WriteFile(reportFilename, reportJSON, 0644); err != nil {
			log.Printf("保存报告失败: %v", err)
		} else {
			fmt.Printf("✅ 时间线报告已保存到: %s (%d 字节)\n", reportFilename, len(reportJSON))
		}
	}

	fmt.Println("\n🎉 演示完成！")
	fmt.Println("\n📁 生成的文件:")
	fmt.Printf("   - %s (会话数据)\n", fmt.Sprintf("session_%s.json", sessionID))
	fmt.Printf("   - %s (时间线报告)\n", fmt.Sprintf("timeline_report_%s.json", sessionID))

	fmt.Println("\n🔍 功能特性:")
	fmt.Println("   ✅ 完整会话录制 (握手→认证→心跳→业务收发→异常/关闭)")
	fmt.Println("   ✅ 自动回放与断言 (顺序、去重、延迟、重连耗时)")
	fmt.Println("   ✅ 问题定位 (消息时间线、原始帧、网络指标)")
	fmt.Println("   ✅ 性能分析 (延迟分布、抖动、吞吐量、连接稳定性)")
}
