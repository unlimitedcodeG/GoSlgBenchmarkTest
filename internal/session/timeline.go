package session

import (
	"fmt"
	"sort"
	"time"
)

// TimelineEvent 时间线事件
type TimelineEvent struct {
	Timestamp     time.Time              `json:"timestamp"`
	EventType     EventType              `json:"event_type"`
	Opcode        uint16                 `json:"opcode,omitempty"`
	MessageID     string                 `json:"message_id,omitempty"`
	Direction     string                 `json:"direction,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	RelatedEvents []string               `json:"related_events,omitempty"`
}

// MessageFlow 消息流
type MessageFlow struct {
	MessageID     string        `json:"message_id"`
	Opcode        uint16        `json:"opcode"`
	SendTime      time.Time     `json:"send_time"`
	ReceiveTime   time.Time     `json:"receive_time,omitempty"`
	Latency       time.Duration `json:"latency,omitempty"`
	Status        string        `json:"status"` // "sent", "received", "timeout", "error"
	Error         string        `json:"error,omitempty"`
	RetryCount    int           `json:"retry_count,omitempty"`
	RelatedEvents []string      `json:"related_events,omitempty"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	TotalMessages      int                   `json:"total_messages"`
	SuccessfulMessages int                   `json:"successful_messages"`
	FailedMessages     int                   `json:"failed_messages"`
	TimeoutMessages    int                   `json:"timeout_messages"`
	AverageLatency     time.Duration         `json:"average_latency"`
	MinLatency         time.Duration         `json:"min_latency"`
	MaxLatency         time.Duration         `json:"max_latency"`
	LatencyPercentiles map[int]time.Duration `json:"latency_percentiles"`
	Jitter             time.Duration         `json:"jitter"`
	PacketLoss         float64               `json:"packet_loss"`
	Throughput         float64               `json:"throughput"` // messages per second
}

// TimelineAnalyzer 时间线分析器
type TimelineAnalyzer struct {
	session *Session
}

// NewTimelineAnalyzer 创建时间线分析器
func NewTimelineAnalyzer(session *Session) *TimelineAnalyzer {
	return &TimelineAnalyzer{
		session: session,
	}
}

// AnalyzeTimeline 分析完整时间线
func (a *TimelineAnalyzer) AnalyzeTimeline() []*TimelineEvent {
	var timeline []*TimelineEvent

	// 按时间戳排序所有事件
	events := make([]*SessionEvent, len(a.session.Events))
	copy(events, a.session.Events)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// 转换为时间线事件
	for _, event := range events {
		timelineEvent := &TimelineEvent{
			Timestamp: event.Timestamp,
			EventType: event.Type,
			Opcode:    event.Opcode,
			MessageID: event.ID,
			Duration:  event.Duration,
			Metadata:  event.Metadata,
		}

		// 设置方向
		switch event.Type {
		case EventMessageSend:
			timelineEvent.Direction = "send"
		case EventMessageReceive:
			timelineEvent.Direction = "receive"
		case EventConnect:
			timelineEvent.Direction = "in"
		case EventDisconnect:
			timelineEvent.Direction = "out"
		default:
			timelineEvent.Direction = "internal"
		}

		timeline = append(timeline, timelineEvent)
	}

	return timeline
}

// AnalyzeMessageFlows 分析消息流
func (a *TimelineAnalyzer) AnalyzeMessageFlows() []*MessageFlow {
	var flows []*MessageFlow
	flowMap := make(map[string]*MessageFlow)

	// 遍历所有事件，构建消息流
	for _, event := range a.session.Events {
		if event.Type == EventMessageSend || event.Type == EventMessageReceive {
			messageID := a.extractMessageID(event)

			flow, exists := flowMap[messageID]
			if !exists {
				flow = &MessageFlow{
					MessageID:     messageID,
					Opcode:        event.Opcode,
					RelatedEvents: []string{event.ID},
				}
				flowMap[messageID] = flow
			} else {
				flow.RelatedEvents = append(flow.RelatedEvents, event.ID)
			}

			// 更新消息流状态
			switch event.Type {
			case EventMessageSend:
				flow.SendTime = event.Timestamp
				flow.Status = "sent"
			case EventMessageReceive:
				flow.ReceiveTime = event.Timestamp
				if !flow.SendTime.IsZero() {
					flow.Latency = event.Timestamp.Sub(flow.SendTime)
					flow.Status = "received"
				}
			}
		}
	}

	// 转换为切片并计算统计
	for _, flow := range flowMap {
		// 检查超时
		if flow.Status == "sent" && time.Since(flow.SendTime) > 30*time.Second {
			flow.Status = "timeout"
		}

		flows = append(flows, flow)
	}

	// 按发送时间排序
	sort.Slice(flows, func(i, j int) bool {
		return flows[i].SendTime.Before(flows[j].SendTime)
	})

	return flows
}

// CalculateNetworkMetrics 计算网络指标
func (a *TimelineAnalyzer) CalculateNetworkMetrics() *NetworkMetrics {
	flows := a.AnalyzeMessageFlows()

	metrics := &NetworkMetrics{
		LatencyPercentiles: make(map[int]time.Duration),
	}

	var latencies []time.Duration

	// 计算会话实际持续时间（从第一条消息到最后一条消息）
	var sessionDuration time.Duration
	if len(flows) > 0 {
		var firstTime, lastTime time.Time

		for _, flow := range flows {
			if !flow.SendTime.IsZero() {
				if firstTime.IsZero() || flow.SendTime.Before(firstTime) {
					firstTime = flow.SendTime
				}

				endTime := flow.SendTime
				if !flow.ReceiveTime.IsZero() {
					endTime = flow.ReceiveTime
				}

				if lastTime.IsZero() || endTime.After(lastTime) {
					lastTime = endTime
				}
			}
		}

		if !firstTime.IsZero() && !lastTime.IsZero() {
			sessionDuration = lastTime.Sub(firstTime)
		}
	}

	for _, flow := range flows {
		metrics.TotalMessages++

		switch flow.Status {
		case "received":
			metrics.SuccessfulMessages++
			if flow.Latency > 0 {
				latencies = append(latencies, flow.Latency)
			}
		case "timeout":
			metrics.TimeoutMessages++
		case "error":
			metrics.FailedMessages++
		}
	}

	// 计算延迟统计
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		// 基本统计
		metrics.MinLatency = latencies[0]
		metrics.MaxLatency = latencies[len(latencies)-1]

		// 平均延迟
		var totalLatency time.Duration
		for _, latency := range latencies {
			totalLatency += latency
		}
		metrics.AverageLatency = totalLatency / time.Duration(len(latencies))

		// 百分位数
		metrics.LatencyPercentiles[50] = latencies[len(latencies)*50/100]
		metrics.LatencyPercentiles[90] = latencies[len(latencies)*90/100]
		metrics.LatencyPercentiles[95] = latencies[len(latencies)*95/100]
		metrics.LatencyPercentiles[99] = latencies[len(latencies)*99/100]

		// 计算抖动（延迟变化的标准差）
		metrics.Jitter = a.calculateJitter(latencies, metrics.AverageLatency)
	}

	// 计算丢包率 - 只计算传输层丢包，不包含应用层超时
	if metrics.TotalMessages > 0 {
		metrics.PacketLoss = float64(metrics.FailedMessages) / float64(metrics.TotalMessages)
	}

	// 计算吞吐量 - 使用会话实际持续时间
	if sessionDuration > 0 {
		metrics.Throughput = float64(metrics.SuccessfulMessages) / sessionDuration.Seconds()
	}

	return metrics
}

// FindLatencyAnomalies 查找延迟异常
func (a *TimelineAnalyzer) FindLatencyAnomalies(threshold time.Duration) []*MessageFlow {
	flows := a.AnalyzeMessageFlows()
	var anomalies []*MessageFlow

	for _, flow := range flows {
		if flow.Latency > threshold {
			anomalies = append(anomalies, flow)
		}
	}

	// 按延迟降序排序
	sort.Slice(anomalies, func(i, j int) bool {
		return anomalies[i].Latency > anomalies[j].Latency
	})

	return anomalies
}

// AnalyzeConnectionStability 分析连接稳定性
func (a *TimelineAnalyzer) AnalyzeConnectionStability() map[string]interface{} {
	var connectEvents, disconnectEvents []*SessionEvent
	var reconnectEvents []*SessionEvent

	for _, event := range a.session.Events {
		switch event.Type {
		case EventConnect:
			connectEvents = append(connectEvents, event)
		case EventDisconnect:
			disconnectEvents = append(disconnectEvents, event)
		case EventReconnect:
			reconnectEvents = append(reconnectEvents, event)
		}
	}

	// 计算连接统计 - 重新设计逻辑
	// reconnect事件也算作连接建立，但需要正确计算连接时长
	allConnectionEvents := append(connectEvents, reconnectEvents...)
	connectionCount := len(allConnectionEvents)
	disconnectionCount := len(disconnectEvents)
	reconnectCount := len(reconnectEvents)

	// 计算平均连接时长 - 改进算法
	var totalConnectionTime time.Duration
	var connectionDurations []time.Duration
	var connectionDetails []map[string]interface{} // 用于调试

	// 配对连接事件和断开事件
	for i, connEvent := range allConnectionEvents {
		var endTime time.Time
		var connectionType string

		if i < len(disconnectEvents) {
			endTime = disconnectEvents[i].Timestamp
			connectionType = "primary"
		} else {
			// 最后一个连接使用会话结束时间
			endTime = a.session.EndTime
			connectionType = "final"
		}

		duration := endTime.Sub(connEvent.Timestamp)
		if duration > 0 {
			totalConnectionTime += duration
			connectionDurations = append(connectionDurations, duration)

			// 记录连接详情用于调试
			connectionDetails = append(connectionDetails, map[string]interface{}{
				"index":      i + 1,
				"type":       connectionType,
				"start_time": connEvent.Timestamp,
				"end_time":   endTime,
				"duration":   duration,
				"event_type": connEvent.Type,
			})
		}
	}

	var avgConnectionDuration time.Duration
	if len(connectionDurations) > 0 {
		avgConnectionDuration = totalConnectionTime / time.Duration(len(connectionDurations))
	}

	// 计算连接稳定性指标
	stability := map[string]interface{}{
		"total_connections":       connectionCount,
		"total_disconnections":    disconnectionCount,
		"reconnect_count":         reconnectCount,
		"avg_connection_duration": avgConnectionDuration,
		"connection_ratio":        float64(connectionCount) / float64(disconnectionCount+1),
		"reconnect_rate":          float64(reconnectCount) / float64(connectionCount+1),
		"connection_details":      connectionDetails, // 添加调试信息
	}

	// 添加连接时长分布
	if len(connectionDurations) > 0 {
		sort.Slice(connectionDurations, func(i, j int) bool {
			return connectionDurations[i] < connectionDurations[j]
		})

		stability["min_connection_duration"] = connectionDurations[0]
		stability["max_connection_duration"] = connectionDurations[len(connectionDurations)-1]
		stability["median_connection_duration"] = connectionDurations[len(connectionDurations)/2]
		stability["total_connection_time"] = totalConnectionTime
	}

	return stability
}

// GenerateTimelineReport 生成时间线报告
func (a *TimelineAnalyzer) GenerateTimelineReport() map[string]interface{} {
	timeline := a.AnalyzeTimeline()
	flows := a.AnalyzeMessageFlows()
	metrics := a.CalculateNetworkMetrics()
	stability := a.AnalyzeConnectionStability()

	report := map[string]interface{}{
		"session_info": map[string]interface{}{
			"id":           a.session.ID,
			"start_time":   a.session.StartTime,
			"end_time":     a.session.EndTime,
			"duration":     a.session.EndTime.Sub(a.session.StartTime),
			"total_events": a.session.Stats.TotalEvents,
		},
		"timeline": map[string]interface{}{
			"total_events": len(timeline),
			"events":       timeline,
		},
		"message_flows": map[string]interface{}{
			"total_flows": len(flows),
			"flows":       flows,
		},
		"network_metrics":      metrics,
		"connection_stability": stability,
		"analysis_timestamp":   time.Now(),
	}

	return report
}

// extractMessageID 提取消息ID
func (a *TimelineAnalyzer) extractMessageID(event *SessionEvent) string {
	// 尝试从元数据中提取消息ID
	if event.Metadata != nil {
		if messageID, ok := event.Metadata["message_id"].(string); ok {
			return messageID
		}
		if sequenceNum, ok := event.Metadata["sequence_num"].(uint64); ok {
			return fmt.Sprintf("seq_%d", sequenceNum)
		}
	}

	// 使用事件ID作为后备
	return event.ID
}

// calculateJitter 计算抖动（相邻延迟变化）
func (a *TimelineAnalyzer) calculateJitter(latencies []time.Duration, avgLatency time.Duration) time.Duration {
	if len(latencies) < 2 {
		return 0
	}

	// 计算相邻延迟之间的差异（真正的抖动定义）
	var jitterSum time.Duration
	for i := 1; i < len(latencies); i++ {
		diff := latencies[i] - latencies[i-1]
		if diff < 0 {
			diff = -diff // 取绝对值
		}
		jitterSum += diff
	}

	// 返回平均抖动
	avgJitter := jitterSum / time.Duration(len(latencies)-1)

	// 确保抖动不为负数
	if avgJitter < 0 {
		return 0
	}

	return avgJitter
}
