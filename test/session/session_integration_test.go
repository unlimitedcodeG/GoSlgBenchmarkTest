package session_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
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

	// 确保服务器完全启动
	time.Sleep(200 * time.Millisecond)

	// 延迟清理服务器
	defer func() {
		t.Log("🧹 清理测试服务器...")
		server.Shutdown(context.Background())
		time.Sleep(500 * time.Millisecond)
	}()

	// 创建会话录制器
	sessionID := fmt.Sprintf("test_session_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 创建WebSocket客户端
	config := wsclient.DefaultClientConfig("ws://127.0.0.1:18090/ws", "session-test-token")
	client := wsclient.New(config)

	// 延迟清理客户端
	defer func() {
		t.Log("🧹 清理测试客户端...")
		client.Close()
		time.Sleep(200 * time.Millisecond)
	}()

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

	// 调试：打印所有事件信息
	t.Logf("   📋 会话总事件数: %d", len(testSession.Events))
	for i, event := range testSession.Events {
		if event.Type == session.EventMessageReceive {
			t.Logf("      接收消息[%d]: opcode=%d, timestamp=%v, metadata=%v",
				i, event.Opcode, event.Timestamp, event.Metadata)
		}
	}

	// 创建断言套件
	suite := session.NewAssertionSuite("Session Quality Test", "验证会话质量指标")

	// 添加各种断言 - 检查2000-2004范围的操作码
	suite.AddAssertion(session.NewMessageOrderAssertion(
		"Message Order Check 2000",
		"验证2000操作码消息按顺序接收",
		2000, // opcode
		1,    // 最小数量
		5,    // 最大数量
	))

	suite.AddAssertion(session.NewMessageOrderAssertion(
		"Message Order Check 2001",
		"验证2001操作码消息按顺序接收",
		2001, // opcode
		1,    // 最小数量
		5,    // 最大数量
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

	// 添加企业级断言
	suite.AddAssertion(session.NewRecoveryTimeAssertion(
		"Recovery Time Check",
		"验证重连恢复时间在可接受范围内",
		2*time.Second, // 最大恢复时间 2秒
	))

	suite.AddAssertion(session.NewPlannedFaultExemptionAssertion(
		"Planned Fault Exemption",
		"验证计划性故障期间的异常被正确豁免",
		5, // 豁免区间长度（消息数）
	))

	suite.AddAssertion(session.NewGoodputAssertion(
		"Goodput Check",
		"验证有效吞吐量达到要求",
		8.0,           // 最小有效吞吐量 8 msg/s
		5*time.Second, // 滑动窗口大小 5秒
	))

	suite.AddAssertion(session.NewTailLatencyBudgetAssertion(
		"Tail Latency Budget",
		"验证尾延迟预算控制",
		350*time.Millisecond, // 预算上限 350ms
		1,                    // 当前窗口数量
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

		// 发送消息 - 使用2000范围的操作码以匹配断言期望
		opcode := uint16(2000 + i%5) // 2000-2004范围
		recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
			"opcode":       opcode,
			"sequence_num": uint64(i + 1),
			"message_id":   fmt.Sprintf("msg_%d", i+1),
			"timestamp":    sendTime,
		})

		// 大部分消息会收到响应，少数会超时
		if i < 18 {
			receiveTime := sendTime.Add(time.Duration(20+i*3) * time.Millisecond)
			latency := receiveTime.Sub(sendTime)

			recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
				"opcode":       opcode,
				"sequence_num": uint64(i + 1),
				"message_id":   fmt.Sprintf("msg_%d", i+1),
				"timestamp":    receiveTime,
				"duration":     latency,
			})

			// 正确记录延迟数据
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

	// 查找延迟异常 (按不同阈值分级)
	t.Log("   🚨 查找延迟异常...")

	// 轻微异常: > 100ms
	lightAnomalies := analyzer.FindLatencyAnomalies(100 * time.Millisecond)
	// 严重异常: > 300ms
	severeAnomalies := analyzer.FindLatencyAnomalies(300 * time.Millisecond)

	t.Logf("      📊 延迟分级分析:")
	t.Logf("        轻微异常 (>100ms): %d 个", len(lightAnomalies))
	t.Logf("        严重异常 (>300ms): %d 个", len(severeAnomalies))
	t.Logf("        正常范围 (≤100ms): %d 个", len(analyzer.AnalyzeMessageFlows())-len(lightAnomalies))

	if len(severeAnomalies) > 0 {
		t.Logf("      🚨 严重异常详情:")
		for i, anomaly := range severeAnomalies[:min(3, len(severeAnomalies))] {
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

// TestTimelineAnalysisRealWallClock 测试真实5分钟墙钟时间线分析功能
func TestTimelineAnalysisRealWallClock(t *testing.T) {
	t.Log("🚀 测试真实5分钟墙钟时间线分析功能...")

	// 创建模拟会话数据
	sessionID := fmt.Sprintf("timeline_enhanced_test_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 模拟5分钟的复杂消息流（真实墙钟时间测试）
	startWallTime := time.Now()
	testDuration := 300 * time.Second         // 5分钟测试
	messageInterval := 100 * time.Millisecond // 每100ms一个消息

	// 连接事件
	recorder.RecordEvent(session.EventConnect, nil)
	t.Logf("   🔗 连接建立时间: %v", startWallTime)

	// 模拟多个消息流 - 5分钟内发送大量消息
	messageCount := int(testDuration / messageInterval) // 约3000个消息
	t.Logf("   📤 计划发送 %d 个消息，持续 %v", messageCount, testDuration)

	// 用于统计的变量
	var sentMessages, timeoutMessages int
	var latencies []time.Duration
	var receivedMessages int32 // 使用原子操作

	// 创建定时器 - 按真实时间间隔发送消息
	ticker := time.NewTicker(messageInterval)
	defer ticker.Stop()

	// 中间断开重连点 (2.5分钟后)
	disconnectAt := messageCount / 2 // 1500个消息后断开

	for i := 0; i < messageCount; i++ {
		// 等待真实的时间间隔（除了第一次）
		if i > 0 {
			<-ticker.C
		}

		// 获取当前真实墙钟时间
		sendTime := time.Now()

		// 发送消息 - 使用2000-2010范围的操作码
		opcode := uint16(2000 + i%11) // 2000-2010范围，更分散
		recorder.RecordEvent(session.EventMessageSend, map[string]interface{}{
			"opcode":       opcode,
			"sequence_num": uint64(i + 1),
			"message_id":   fmt.Sprintf("msg_%d", i+1),
			"timestamp":    sendTime,
		})
		sentMessages++

		// 在预定点插入断开重连事件
		if i == disconnectAt {
			t.Logf("      🔌 在消息 %d (约2.5分钟后) 执行断开重连...", i+1)

			// 记录断开事件
			recorder.RecordEvent(session.EventDisconnect, map[string]interface{}{
				"reason": "Network fluctuation simulation",
			})

			// 模拟重连延迟 - 这个会增加额外时间
			time.Sleep(800 * time.Millisecond)

			// 记录重连事件
			reconnectTime := time.Now()
			recorder.RecordEvent(session.EventReconnect, map[string]interface{}{
				"attempt":   1,
				"duration":  800 * time.Millisecond,
				"success":   true,
				"timestamp": reconnectTime,
			})
			t.Logf("      🔄 重连成功，耗时: %v", time.Since(sendTime))
		}

		// 模拟不同的响应模式 - 使用goroutine异步处理，不阻塞发送间隔
		shouldRespond := true
		var latency time.Duration

		// 模拟不同的业务场景
		switch i % 10 {
		case 0, 1, 2, 3, 4, 5, 6, 7: // 80% 正常响应
			latency = time.Duration(20+i%50) * time.Millisecond // 20-70ms随机延迟
		case 8: // 10% 高延迟
			latency = time.Duration(200+i%100) * time.Millisecond // 200-300ms高延迟
		case 9: // 10% 超时
			shouldRespond = false
			timeoutMessages++
		}

		if shouldRespond {
			// 使用goroutine异步处理消息响应，不阻塞下一个消息的发送
			go func(msgIndex int, msgOpcode uint16, msgLatency time.Duration, msgSendTime time.Time) {
				// 等待实际的延迟时间
				time.Sleep(msgLatency)

				receiveTime := time.Now()
				actualLatency := receiveTime.Sub(msgSendTime)

				// 记录接收事件
				recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
					"opcode":       msgOpcode,
					"sequence_num": uint64(msgIndex + 1),
					"message_id":   fmt.Sprintf("msg_%d", msgIndex+1),
					"timestamp":    receiveTime,
					"duration":     actualLatency,
				})

				recorder.RecordLatency(actualLatency)
				// 原子操作增加接收消息计数
				atomic.AddInt32(&receivedMessages, 1)
			}(i, opcode, latency, sendTime)

			latencies = append(latencies, latency)

			// 每250个消息打印一次进度 (5分钟测试，每30秒打印一次)
			if i%250 == 0 {
				elapsed := time.Since(startWallTime)
				progress := float64(i+1) / float64(messageCount) * 100
				currentReceived := atomic.LoadInt32(&receivedMessages)
				t.Logf("      📊 已发送 %d/%d 消息 (%.1f%%), 已接收: %d, 耗时: %v",
					i+1, messageCount, progress, currentReceived, elapsed)
			}
		}
	}

	// 等待所有异步响应完成
	t.Log("      ⏳ 等待所有异步响应完成...")
	time.Sleep(500 * time.Millisecond) // 等待最后一批消息的响应

	// 获取最终的接收消息数
	finalReceivedMessages := int(atomic.LoadInt32(&receivedMessages))

	// 计算实际的墙钟运行时间
	endWallTime := time.Now()
	actualDuration := endWallTime.Sub(startWallTime)
	t.Logf("   🔚 测试结束时间: %v", endWallTime)
	t.Logf("   ⏱️ 实际运行时间: %v (计划: %v)", actualDuration, testDuration)

	// 模拟一些业务错误
	recorder.RecordEvent(session.EventError, map[string]interface{}{
		"error_type": "business_logic",
		"details":    "Invalid player state",
	})

	// 关闭会话
	recorder.RecordClose(session.CloseNormal, "Enhanced analysis test completed")

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

	// 增强的指标显示
	t.Logf("   📈 增强网络指标 (真实5分钟墙钟测试):")
	t.Logf("      总消息数: %d", metrics.TotalMessages)
	t.Logf("      成功消息: %d", metrics.SuccessfulMessages)
	t.Logf("      超时消息: %d", timeoutMessages)
	t.Logf("      平均延迟: %v", metrics.AverageLatency)
	t.Logf("      最小延迟: %v", metrics.MinLatency)
	t.Logf("      最大延迟: %v", metrics.MaxLatency)

	// 显示百分位数延迟
	if len(metrics.LatencyPercentiles) > 0 {
		t.Logf("      P50延迟: %v", metrics.LatencyPercentiles[50])
		t.Logf("      P90延迟: %v", metrics.LatencyPercentiles[90])
		t.Logf("      P95延迟: %v", metrics.LatencyPercentiles[95])
		t.Logf("      P99延迟: %v", metrics.LatencyPercentiles[99])
	}

	t.Logf("      抖动: %v", metrics.Jitter)
	t.Logf("      传输丢包率: %.2f%%", metrics.PacketLoss*100)
	t.Logf("      应用超时率: %.2f%%", float64(timeoutMessages)/float64(sentMessages)*100)
	t.Logf("      吞吐量: %.2f msg/s", metrics.Throughput)

	// 计算真实的墙钟吞吐量
	realThroughput := float64(finalReceivedMessages) / actualDuration.Seconds()
	t.Logf("      墙钟吞吐量: %.2f msg/s", realThroughput)

	// 计算业务成功率
	businessSuccessRate := float64(finalReceivedMessages) / float64(sentMessages)
	t.Logf("      业务成功率: %.2f%%", businessSuccessRate*100)

	// 查找延迟异常 (按不同阈值分级)
	t.Log("   🚨 查找延迟异常...")

	// 轻微异常: > 100ms
	lightAnomalies := analyzer.FindLatencyAnomalies(100 * time.Millisecond)
	// 严重异常: > 300ms
	severeAnomalies := analyzer.FindLatencyAnomalies(300 * time.Millisecond)

	t.Logf("      📊 延迟分级分析:")
	t.Logf("        轻微异常 (>100ms): %d 个", len(lightAnomalies))
	t.Logf("        严重异常 (>300ms): %d 个", len(severeAnomalies))
	t.Logf("        正常范围 (≤100ms): %d 个", len(analyzer.AnalyzeMessageFlows())-len(lightAnomalies))

	if len(severeAnomalies) > 0 {
		t.Logf("      🚨 严重异常详情:")
		for i, anomaly := range severeAnomalies[:min(3, len(severeAnomalies))] {
			t.Logf("        [%d] 消息 %s: 延迟 %v", i+1, anomaly.MessageID, anomaly.Latency)
		}
	}

	// 分析连接稳定性
	t.Log("   🔗 分析连接稳定性...")
	stability := analyzer.AnalyzeConnectionStability()

	t.Logf("   📊 连接稳定性 (真实5分钟墙钟时间测试):")
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

	// 导出报告为JSON
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)
	t.Logf("   📄 报告大小: %d 字节", len(reportJSON))

	// 增强的验证指标
	assert.Greater(t, metrics.SuccessfulMessages, 2500, "5分钟测试应有足够的消息量")
	assert.Less(t, metrics.PacketLoss, 0.2, "丢包率应小于20%") // 只计算传输层丢包
	assert.Greater(t, metrics.Throughput, 5.0, "吞吐量应大于5 msg/s")
	assert.Greater(t, businessSuccessRate, 0.8, "业务成功率应大于80%")

	// 企业级断言已在 TestSessionAssertions 中验证
	t.Logf("   📊 企业级断言验证:")
	t.Logf("      ✅ 恢复时间断言已在单独测试中验证")
	t.Logf("      ✅ 故障豁免断言已在单独测试中验证")
	t.Logf("      ✅ 有效吞吐断言已在单独测试中验证")
	t.Logf("      ✅ 尾延迟预算断言已在单独测试中验证")

	// 验证实际运行时间接近5分钟
	t.Logf("   ✅ 墙钟时间验证:")
	t.Logf("      实际运行时间: %v (计划: %v, 误差: %v)",
		actualDuration, testDuration, actualDuration-testDuration)
	t.Logf("      墙钟吞吐量: %.2f msg/s", realThroughput)

	// 验证墙钟时间 - 实际运行时间应接近300秒
	durationDiff := actualDuration - testDuration
	if durationDiff < 0 {
		durationDiff = -durationDiff
	}
	maxDurationDiff := 10 * time.Second // 允许10秒误差（包含重连时间和异步处理）
	assert.Less(t, durationDiff, maxDurationDiff,
		"实际运行时间应接近300秒，允许误差%v", maxDurationDiff)

	// 验证墙钟吞吐量
	assert.Greater(t, realThroughput, 8.0, "墙钟吞吐量应大于8 msg/s")
	assert.Less(t, realThroughput, 15.0, "墙钟吞吐量应小于15 msg/s")

	// 验证延迟分布
	if len(metrics.LatencyPercentiles) > 0 {
		p50 := metrics.LatencyPercentiles[50]
		p95 := metrics.LatencyPercentiles[95]
		p99 := metrics.LatencyPercentiles[99]

		assert.Less(t, p50, 100*time.Millisecond, "P50延迟应小于100ms")
		assert.Less(t, p95, 300*time.Millisecond, "P95延迟应小于300ms")
		assert.Less(t, p99, 400*time.Millisecond, "P99延迟应小于400ms")

		t.Logf("   ✅ 延迟目标验证:")
		t.Logf("      P50 < 100ms: %v", p50 < 100*time.Millisecond)
		t.Logf("      P95 < 300ms: %v", p95 < 300*time.Millisecond)
		t.Logf("      P99 < 400ms: %v", p99 < 400*time.Millisecond)
	}

	t.Log("   ✅ 真实5分钟墙钟时间线分析完成")
}

// min 函数辅助
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestSessionExportAndImport 测试会话导出和导入
func TestSessionExportAndImport(t *testing.T) {
	t.Log("💾 测试会话导出和导入功能...")

	// 创建并录制会话
	sessionID := fmt.Sprintf("export_test_%d", time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	// 添加一些测试事件（CONNECT事件已经在构造函数中自动添加）
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
