package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"GoSlgBenchmarkTest/internal/config"
)

// TestAssertions 测试断言助手
type TestAssertions struct {
	t      *testing.T
	config *config.TestConfig
}

// NewTestAssertions 创建测试断言助手
func NewTestAssertions(t *testing.T) *TestAssertions {
	return &TestAssertions{
		t:      t,
		config: config.GetTestConfig(),
	}
}

// AssertConnection 断言连接成功
func (ta *TestAssertions) AssertConnection(client *TestClient) {
	err := client.ValidateConnection()
	require.NoError(ta.t, err, "Connection validation failed")
	ta.t.Logf("✅ Connection assertion passed")
}

// AssertMessageCount 断言消息数量
func (ta *TestAssertions) AssertMessageCount(client *TestClient, expectedMin, expectedMax int) {
	messages := client.GetReceivedMessages()
	count := len(messages)

	assert.GreaterOrEqual(ta.t, count, expectedMin, "Message count below minimum")
	assert.LessOrEqual(ta.t, count, expectedMax, "Message count above maximum")
	ta.t.Logf("✅ Message count assertion passed: %d messages (range: %d-%d)", count, expectedMin, expectedMax)
}

// AssertMessageSequence 断言消息序列完整性
func (ta *TestAssertions) AssertMessageSequence(client *TestClient) {
	messages := client.GetReceivedMessages()
	if len(messages) < 2 {
		ta.t.Logf("⚠️  Insufficient messages for sequence validation")
		return
	}

	outOfOrder := 0
	for i := 1; i < len(messages); i++ {
		if messages[i].Timestamp.Before(messages[i-1].Timestamp) {
			outOfOrder++
		}
	}

	maxOutOfOrder := ta.config.Assertions.MessageOrder.MaxOutOfOrder
	assert.LessOrEqual(ta.t, outOfOrder, maxOutOfOrder,
		"Too many out-of-order messages: %d > %d", outOfOrder, maxOutOfOrder)
	ta.t.Logf("✅ Message sequence assertion passed: %d out-of-order messages (max: %d)", outOfOrder, maxOutOfOrder)
}

// AssertLatency 断言延迟性能
func (ta *TestAssertions) AssertLatency(client *TestClient) {
	rtts := client.GetRTTReadings()
	if len(rtts) == 0 {
		ta.t.Logf("⚠️  No RTT readings for latency validation")
		return
	}

	// 计算平均延迟
	var total time.Duration
	var max time.Duration
	for _, rtt := range rtts {
		total += rtt
		if rtt > max {
			max = rtt
		}
	}
	avg := total / time.Duration(len(rtts))

	// 断言平均延迟
	maxAvg := ta.config.Assertions.Latency.MaxAverage
	assert.LessOrEqual(ta.t, avg, maxAvg, "Average latency too high: %v > %v", avg, maxAvg)

	// 断言最大延迟（作为P100）
	maxP99 := ta.config.Assertions.Latency.MaxP99
	assert.LessOrEqual(ta.t, max, maxP99, "Maximum latency too high: %v > %v", max, maxP99)

	ta.t.Logf("✅ Latency assertion passed: avg=%v, max=%v", avg, max)
}

// AssertReconnects 断言重连行为
func (ta *TestAssertions) AssertReconnects(client *TestClient, expectedReconnects int) {
	reconnects := client.GetReconnectCount()
	assert.Equal(ta.t, expectedReconnects, reconnects,
		"Unexpected reconnect count: expected %d, got %d", expectedReconnects, reconnects)
	ta.t.Logf("✅ Reconnect assertion passed: %d reconnects", reconnects)
}

// AssertReconnectTime 断言重连时间
func (ta *TestAssertions) AssertReconnectTime(actualTime time.Duration) {
	maxTime := ta.config.TestScenarios.Reconnect.MaxReconnectTime
	assert.LessOrEqual(ta.t, actualTime, maxTime,
		"Reconnect time too long: %v > %v", actualTime, maxTime)
	ta.t.Logf("✅ Reconnect time assertion passed: %v (max: %v)", actualTime, maxTime)
}

// AssertErrorRate 断言错误率
func (ta *TestAssertions) AssertErrorRate(errorCount, totalCount int) {
	if totalCount == 0 {
		ta.t.Logf("⚠️  No operations for error rate validation")
		return
	}

	errorRate := float64(errorCount) / float64(totalCount)
	maxErrorRate := ta.config.Assertions.ErrorRate.MaxErrorRate

	assert.LessOrEqual(ta.t, errorRate, maxErrorRate,
		"Error rate too high: %.2f%% > %.2f%%", errorRate*100, maxErrorRate*100)
	ta.t.Logf("✅ Error rate assertion passed: %.2f%% (%d/%d)", errorRate*100, errorCount, totalCount)
}

// AssertThroughput 断言吞吐量
func (ta *TestAssertions) AssertThroughput(messageCount int, duration time.Duration) {
	if duration == 0 {
		ta.t.Logf("⚠️  Zero duration for throughput validation")
		return
	}

	throughput := float64(messageCount) / duration.Seconds()
	minThroughput := float64(ta.config.StressTest.Throughput.ExpectedMinThroughput)

	assert.GreaterOrEqual(ta.t, throughput, minThroughput,
		"Throughput too low: %.2f < %.2f messages/sec", throughput, minThroughput)
	ta.t.Logf("✅ Throughput assertion passed: %.2f messages/sec (min: %.2f)", throughput, minThroughput)
}

// AssertMemoryUsage 断言内存使用
func (ta *TestAssertions) AssertMemoryUsage(beforeBytes, afterBytes uint64) {
	if beforeBytes == 0 {
		ta.t.Logf("⚠️  Invalid memory baseline for validation")
		return
	}

	growth := float64(afterBytes-beforeBytes) / float64(beforeBytes)
	maxGrowth := 2.0 // 最大增长200%

	assert.LessOrEqual(ta.t, growth, maxGrowth,
		"Memory growth too high: %.2f%% > %.2f%%", growth*100, maxGrowth*100)
	ta.t.Logf("✅ Memory assertion passed: growth=%.2f%% (before=%d, after=%d)",
		growth*100, beforeBytes, afterBytes)
}

// AssertConcurrentClients 断言并发客户端
func (ta *TestAssertions) AssertConcurrentClients(clients []*TestClient, expectedConnected int) {
	connected := 0
	for _, client := range clients {
		if client.ValidateConnection() == nil {
			connected++
		}
	}

	assert.Equal(ta.t, expectedConnected, connected,
		"Unexpected connected client count: expected %d, got %d", expectedConnected, connected)
	ta.t.Logf("✅ Concurrent clients assertion passed: %d/%d connected", connected, len(clients))
}

// AssertStressTestResults 断言压力测试结果
func (ta *TestAssertions) AssertStressTestResults(
	clients []*TestClient,
	duration time.Duration,
	expectedMinThroughput float64,
) {
	totalMessages := 0
	totalErrors := 0
	var totalRTT time.Duration
	rttCount := 0

	for _, client := range clients {
		messages := client.GetReceivedMessages()
		totalMessages += len(messages)

		rtts := client.GetRTTReadings()
		for _, rtt := range rtts {
			totalRTT += rtt
			rttCount++
		}

		// 简单错误检测：检查连接状态
		if client.ValidateConnection() != nil {
			totalErrors++
		}
	}

	// 断言吞吐量
	if duration > 0 {
		throughput := float64(totalMessages) / duration.Seconds()
		assert.GreaterOrEqual(ta.t, throughput, expectedMinThroughput,
			"Stress test throughput too low: %.2f < %.2f messages/sec", throughput, expectedMinThroughput)
	}

	// 断言错误率
	if len(clients) > 0 {
		errorRate := float64(totalErrors) / float64(len(clients))
		maxErrorRate := ta.config.Assertions.ErrorRate.MaxErrorRate
		assert.LessOrEqual(ta.t, errorRate, maxErrorRate,
			"Stress test error rate too high: %.2f%% > %.2f%%", errorRate*100, maxErrorRate*100)
	}

	// 断言平均RTT
	if rttCount > 0 {
		avgRTT := totalRTT / time.Duration(rttCount)
		maxAvgRTT := ta.config.Assertions.Latency.MaxAverage
		assert.LessOrEqual(ta.t, avgRTT, maxAvgRTT,
			"Average RTT too high: %v > %v", avgRTT, maxAvgRTT)
	}

	ta.t.Logf("✅ Stress test assertions passed: %d messages, %d errors, avg_rtt=%v",
		totalMessages, totalErrors, totalRTT/time.Duration(max(rttCount, 1)))
}

// AssertBenchmarkResults 断言基准测试结果
func (ta *TestAssertions) AssertBenchmarkResults(b *testing.B, operations int, duration time.Duration) {
	if duration == 0 {
		return
	}

	opsPerSec := float64(operations) / duration.Seconds()
	avgLatency := duration / time.Duration(operations)

	// 报告性能指标
	b.ReportMetric(opsPerSec, "ops/sec")
	b.ReportMetric(float64(avgLatency.Nanoseconds()), "ns/op")
	b.ReportMetric(float64(duration.Nanoseconds()), "total_ns")

	// 基本性能断言
	minOpsPerSec := 100.0 // 最小100 ops/sec
	assert.GreaterOrEqual(ta.t, opsPerSec, minOpsPerSec,
		"Benchmark performance too low: %.2f < %.2f ops/sec", opsPerSec, minOpsPerSec)
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
