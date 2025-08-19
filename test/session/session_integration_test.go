package session_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/protocol"
	"GoSlgBenchmarkTest/internal/session"
	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/wsclient"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// TestSessionRecordingAndReplay 测试会话录制和回放
func TestSessionRecordingAndReplay(t *testing.T) {
	t.Log("🎬 测试会话录制和回放功能...")

	// 启动测试服务器
	server := testserver.New(testserver.DefaultServerConfig(":18090"))
	require.NoError(t, server.Start())
	defer server.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond)

	// 创建会话录制器
	sessionID := fmt.Sprintf("test_session_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 创建WebSocket客户端
	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18090/ws", "session-test-token")
	client := wsclient.New(config)

	// 设置消息处理器，记录到录制器
	client.SetPushHandler(func(opcode uint16, message proto.Message) {
		// 记录接收到的消息
		recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
			"opcode":  opcode,
			"message": message,
		})
	})

	client.SetStateChangeHandler(func(oldState, newState wsclient.ClientState) {
		// 记录状态变化
		recorder.RecordEvent(session.EventConnect, map[string]interface{}{
			"old_state": oldState.String(),
			"new_state": newState.String(),
		})
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接并记录
	require.NoError(t, client.Connect(ctx))
	recorder.RecordEvent(session.EventLogin, map[string]interface{}{
		"player_id": "test_player",
		"status":    "success",
	})

	// 发送一些消息
	for i := 0; i < 5; i++ {
		action := &gamev1.PlayerAction{
			ActionSeq:       uint64(i + 1),
			PlayerId:        "test_player",
			ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
			ClientTimestamp: time.Now().UnixMilli(),
		}

		// 记录发送消息
		recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
			"opcode":       protocol.OpPlayerAction,
			"sequence_num": uint64(i + 1),
			"action_type":  action.ActionType.String(),
		})

		client.SendAction(action)
		time.Sleep(100 * time.Millisecond)
	}

	// 等待一些推送消息
	time.Sleep(1 * time.Second)

	// 关闭连接
	client.Close()
	recorder.RecordClose(session.CloseNormal, "Test completed")

	// 获取录制的会话
	recordedSession := recorder.GetSession()
	require.NotNil(t, recordedSession)

	t.Logf("   📹 录制完成: %d 个事件, %d 个消息帧",
		len(recordedSession.Events), len(recordedSession.Frames))

	// 验证录制的内容
	assert.Greater(t, len(recordedSession.Events), 0)
	assert.Equal(t, sessionID, recordedSession.ID)
	assert.False(t, recordedSession.StartTime.IsZero())
	assert.False(t, recordedSession.EndTime.IsZero())

	// 测试会话回放
	t.Log("   🔄 开始会话回放测试...")

	replayConfig := &session.ReplayConfig{
		Speed:        session.SpeedFast,
		EnablePause:  true,
		PauseOnError: false,
	}

	replayer := session.NewSessionReplayer(recordedSession, replayConfig)

	// 添加回放回调
	var replayedEvents []*session.ReplayEvent
	replayer.AddCallback(func(event *session.ReplayEvent) error {
		replayedEvents = append(replayedEvents, event)
		return nil
	})

	// 开始回放
	require.NoError(t, replayer.Play())

	// 等待回放完成
	time.Sleep(2 * time.Second)

	// 停止回放
	replayer.Stop()
	// 等待回放协程结束
	replayer.Wait()

	// 验证回放结果
	replayStats := replayer.GetStats()
	assert.Greater(t, replayStats.ReplayedEvents, 0)
	assert.Equal(t, len(recordedSession.Events), replayer.ReplayedEvents())

	t.Logf("   ✅ 回放完成: %d/%d 事件重放成功",
		replayStats.ReplayedEvents, replayStats.TotalEvents)
}

// TestSessionAssertions 测试会话断言
func TestSessionAssertions(t *testing.T) {
	t.Log("🧪 测试会话断言功能...")

	// 创建模拟会话数据
	sessionID := fmt.Sprintf("assertion_test_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 模拟一些事件
	baseTime := time.Now()

	// 连接事件
	recorder.RecordEvent(session.EventConnect, map[string]interface{}{
		"client_ip": "127.0.0.1",
	})

	// 登录事件
	recorder.RecordEvent(session.EventLogin, map[string]interface{}{
		"player_id": "test_player",
		"status":    "success",
	})

	// 模拟消息发送和接收（带延迟）
	for i := 0; i < 10; i++ {
		sendTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
		receiveTime := sendTime.Add(time.Duration(50+i*5) * time.Millisecond)

		// 发送事件
		recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
			"opcode":       uint16(2000 + i),
			"sequence_num": uint64(i + 1),
			"timestamp":    sendTime,
		})

		// 接收事件（带延迟）
		latency := receiveTime.Sub(sendTime)
		recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
			"opcode":       uint16(2000 + i),
			"sequence_num": uint64(i + 1),
			"timestamp":    receiveTime,
			"duration":     latency,
		})

		recorder.RecordLatency(latency)
	}

	// 模拟重连事件
	recorder.RecordEvent(session.EventReconnect, map[string]interface{}{
		"attempt":  1,
		"duration": 2 * time.Second,
		"success":  true,
	})

	// 模拟错误事件
	recorder.RecordEvent(session.EventError, map[string]interface{}{
		"error_type": "network_timeout",
		"details":    "Connection timeout after 30 seconds",
	})

	// 关闭事件
	recorder.RecordClose(session.CloseNormal, "Test completed")

	// 获取会话
	testSession := recorder.GetSession()

	// 创建断言套件
	suite := session.NewAssertionSuite("Session Quality Test", "验证会话质量指标")

	// 添加各种断言
	suite.AddAssertion(session.NewMessageOrderAssertion(
		"Message Order Check",
		"验证消息按顺序接收",
		2000, // opcode
		5,    // 最小数量
		15,   // 最大数量
	))

	suite.AddAssertion(session.NewLatencyAssertion(
		"Latency Check",
		"验证延迟在可接受范围内",
		200*time.Millisecond, // 最大延迟
		95,                   // 95%分位数
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
		0.1, // 最大错误率 10%
	))

	// 运行断言
	t.Log("   🔍 运行断言套件...")
	results := suite.RunAssertions(testSession)

	// 验证断言结果
	passedCount := suite.GetPassedCount()
	failedCount := suite.GetFailedCount()
	successRate := suite.GetSuccessRate()

	t.Logf("   📊 断言结果: %d 通过, %d 失败, 成功率: %.1f%%",
		passedCount, failedCount, successRate*100)

	// 打印详细结果
	for i, result := range results {
		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		t.Logf("   %s [%d] %s: %s", status, i+1, result.Message, result.Duration)
	}

	// 验证大部分断言应该通过
	assert.GreaterOrEqual(t, successRate, 0.5, "至少50%的断言应该通过")
}

// TestTimelineAnalysis 测试时间线分析
func TestTimelineAnalysis(t *testing.T) {
	t.Log("📊 测试时间线分析功能...")

	// 创建模拟会话数据
	sessionID := fmt.Sprintf("timeline_test_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 模拟复杂的消息流
	baseTime := time.Now()

	// 连接
	recorder.RecordEvent(session.EventConnect, nil)

	// 模拟多个消息流
	for i := 0; i < 20; i++ {
		sendTime := baseTime.Add(time.Duration(i) * 50 * time.Millisecond)

		// 发送消息
		recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
			"opcode":       uint16(1000 + i%5),
			"sequence_num": uint64(i + 1),
			"message_id":   fmt.Sprintf("msg_%d", i+1),
			"timestamp":    sendTime,
		})

		// 大部分消息会收到响应，少数会超时
		if i < 18 {
			receiveTime := sendTime.Add(time.Duration(20+i*3) * time.Millisecond)
			latency := receiveTime.Sub(sendTime)

			recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
				"opcode":       uint16(1000 + i%5),
				"sequence_num": uint64(i + 1),
				"message_id":   fmt.Sprintf("msg_%d", i+1),
				"timestamp":    receiveTime,
				"duration":     latency,
			})

			recorder.RecordLatency(latency)
		}
	}

	// 模拟一些网络问题
	recorder.RecordEvent(session.EventError, map[string]interface{}{
		"error_type": "network_jitter",
		"details":    "High latency variation detected",
	})

	recorder.RecordEvent(session.EventReconnect, map[string]interface{}{
		"attempt":  1,
		"duration": 1500 * time.Millisecond,
		"success":  true,
	})

	// 关闭
	recorder.RecordClose(session.CloseNormal, "Analysis test completed")

	// 获取会话
	testSession := recorder.GetSession()

	// 创建时间线分析器
	analyzer := session.NewTimelineAnalyzer(testSession)

	// 分析时间线
	t.Log("   📈 分析时间线...")
	timeline := analyzer.AnalyzeTimeline()
	assert.Equal(t, len(testSession.Events), len(timeline))

	// 分析消息流
	t.Log("   🔄 分析消息流...")
	flows := analyzer.AnalyzeMessageFlows()
	assert.Greater(t, len(flows), 0)

	// 计算网络指标
	t.Log("   📊 计算网络指标...")
	metrics := analyzer.CalculateNetworkMetrics()
	assert.NotNil(t, metrics)

	t.Logf("   📈 网络指标:")
	t.Logf("      总消息数: %d", metrics.TotalMessages)
	t.Logf("      成功消息: %d", metrics.SuccessfulMessages)
	t.Logf("      平均延迟: %v", metrics.AverageLatency)
	t.Logf("      最小延迟: %v", metrics.MinLatency)
	t.Logf("      最大延迟: %v", metrics.MaxLatency)
	t.Logf("      抖动: %v", metrics.Jitter)
	t.Logf("      丢包率: %.2f%%", metrics.PacketLoss*100)
	t.Logf("      吞吐量: %.2f msg/s", metrics.Throughput)

	// 查找延迟异常
	t.Log("   🚨 查找延迟异常...")
	anomalies := analyzer.FindLatencyAnomalies(100 * time.Millisecond)
	if len(anomalies) > 0 {
		t.Logf("      发现 %d 个延迟异常:", len(anomalies))
		for i, anomaly := range anomalies[:3] { // 只显示前3个
			t.Logf("        [%d] 消息 %s: 延迟 %v", i+1, anomaly.MessageID, anomaly.Latency)
		}
	}

	// 分析连接稳定性
	t.Log("   🔗 分析连接稳定性...")
	stability := analyzer.AnalyzeConnectionStability()

	t.Logf("   📊 连接稳定性:")
	t.Logf("      连接次数: %v", stability["total_connections"])
	t.Logf("      断开次数: %v", stability["total_disconnections"])
	t.Logf("      重连次数: %v", stability["reconnect_count"])
	t.Logf("      平均连接时长: %v", stability["avg_connection_duration"])
	t.Logf("      重连率: %.2f%%", stability["reconnect_rate"].(float64)*100)

	// 生成完整报告
	t.Log("   📋 生成时间线报告...")
	report := analyzer.GenerateTimelineReport()

	// 验证报告结构
	assert.Contains(t, report, "session_info")
	assert.Contains(t, report, "timeline")
	assert.Contains(t, report, "message_flows")
	assert.Contains(t, report, "network_metrics")
	assert.Contains(t, report, "connection_stability")

	// 导出报告为JSON（用于调试）
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)

	t.Logf("   📄 报告大小: %d 字节", len(reportJSON))

	// 验证关键指标
	assert.Greater(t, metrics.SuccessfulMessages, 0)
	assert.Less(t, metrics.PacketLoss, 0.5) // 丢包率应小于50%
	assert.Greater(t, metrics.Throughput, 0.0)

	t.Log("   ✅ 时间线分析完成")
}

// TestSessionExportAndImport 测试会话导出和导入
func TestSessionExportAndImport(t *testing.T) {
	t.Log("💾 测试会话导出和导入功能...")

	// 创建并录制会话
	sessionID := fmt.Sprintf("export_test_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 添加一些测试事件
	recorder.RecordEvent(session.EventConnect, map[string]interface{}{
		"test": "export_import",
	})

	recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
		"opcode": 1001,
		"data":   "test_message",
	})

	recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
		"opcode": 1001,
		"data":   "test_response",
	})

	recorder.RecordClose(session.CloseNormal, "Export test completed")

	// 导出为JSON
	exportedData, err := recorder.ExportJSON()
	require.NoError(t, err)
	require.Greater(t, len(exportedData), 0)

	t.Logf("   📤 导出完成: %d 字节", len(exportedData))

	// 解析导出的数据
	var importedSession session.Session
	err = json.Unmarshal(exportedData, &importedSession)
	require.NoError(t, err)

	// 验证导入的数据
	assert.Equal(t, sessionID, importedSession.ID)
	assert.Equal(t, 4, len(importedSession.Events)) // Connect + Send + Receive + Close

	// 验证事件类型
	eventTypes := make(map[session.EventType]int)
	for _, event := range importedSession.Events {
		eventTypes[event.Type]++
	}

	assert.Equal(t, 1, eventTypes[session.EventConnect])
	assert.Equal(t, 1, eventTypes[session.EventMessageSend])
	assert.Equal(t, 1, eventTypes[session.EventMessageReceive])
	assert.Equal(t, 1, eventTypes[session.EventClose])

	t.Log("   ✅ 导出导入测试完成")
}
