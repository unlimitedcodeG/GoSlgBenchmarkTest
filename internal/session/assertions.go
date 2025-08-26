package session

import (
	"fmt"
	"sort"
	"time"
)

// AssertionResult 断言结果
type AssertionResult struct {
	Passed    bool          `json:"passed"`
	Message   string        `json:"message"`
	Expected  interface{}   `json:"expected,omitempty"`
	Actual    interface{}   `json:"actual,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration,omitempty"`
}

// Assertion 断言接口
type Assertion interface {
	Assert(session *Session) *AssertionResult
	GetName() string
	GetDescription() string
}

// MessageOrderAssertion 消息顺序断言
type MessageOrderAssertion struct {
	Name        string
	Description string
	Opcode      uint16
	MinCount    int
	MaxCount    int
}

// NewMessageOrderAssertion 创建消息顺序断言
func NewMessageOrderAssertion(name, description string, opcode uint16, minCount, maxCount int) *MessageOrderAssertion {
	return &MessageOrderAssertion{
		Name:        name,
		Description: description,
		Opcode:      opcode,
		MinCount:    minCount,
		MaxCount:    maxCount,
	}
}

// Assert 执行断言
func (a *MessageOrderAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 过滤指定操作码的事件
	var events []*SessionEvent
	for _, event := range session.Events {
		if event.Type == EventMessageReceive && event.Opcode == a.Opcode {
			events = append(events, event)
		}
	}

	// 检查消息数量
	if len(events) < a.MinCount {
		return &AssertionResult{
			Passed:    false,
			Message:   fmt.Sprintf("Expected at least %d messages with opcode %d, got %d", a.MinCount, a.Opcode, len(events)),
			Expected:  a.MinCount,
			Actual:    len(events),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	if a.MaxCount > 0 && len(events) > a.MaxCount {
		return &AssertionResult{
			Passed:    false,
			Message:   fmt.Sprintf("Expected at most %d messages with opcode %d, got %d", a.MaxCount, a.Opcode, len(events)),
			Expected:  a.MaxCount,
			Actual:    len(events),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 检查消息顺序（按时间戳）
	sortedEvents := make([]*SessionEvent, len(events))
	copy(sortedEvents, events)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Timestamp.Before(sortedEvents[j].Timestamp)
	})

	// 验证顺序一致性
	for i := 0; i < len(sortedEvents)-1; i++ {
		if !sortedEvents[i].Timestamp.Before(sortedEvents[i+1].Timestamp) {
			return &AssertionResult{
				Passed:    false,
				Message:   fmt.Sprintf("Message order violation at index %d: timestamp %v >= %v", i, sortedEvents[i].Timestamp, sortedEvents[i+1].Timestamp),
				Expected:  "strictly increasing timestamps",
				Actual:    fmt.Sprintf("timestamp[%d]=%v, timestamp[%d]=%v", i, sortedEvents[i].Timestamp, i+1, sortedEvents[i+1].Timestamp),
				Timestamp: time.Now(),
				Duration:  time.Since(start),
			}
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   fmt.Sprintf("Message order assertion passed: %d messages with opcode %d in correct order", len(events), a.Opcode),
		Expected:  fmt.Sprintf("%d-%d messages in order", a.MinCount, a.MaxCount),
		Actual:    len(events),
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *MessageOrderAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *MessageOrderAssertion) GetDescription() string {
	return a.Description
}

// LatencyAssertion 延迟断言
type LatencyAssertion struct {
	Name        string
	Description string
	MaxLatency  time.Duration
	Percentile  int // 百分位数 (50, 90, 95, 99)
}

// NewLatencyAssertion 创建延迟断言
func NewLatencyAssertion(name, description string, maxLatency time.Duration, percentile int) *LatencyAssertion {
	return &LatencyAssertion{
		Name:        name,
		Description: description,
		MaxLatency:  maxLatency,
		Percentile:  percentile,
	}
}

// Assert 执行断言
func (a *LatencyAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 收集所有延迟数据
	var latencies []time.Duration
	for _, event := range session.Events {
		if event.Type == EventMessageReceive {
			// 优先使用event.Duration字段
			if event.Duration > 0 {
				latencies = append(latencies, event.Duration)
			} else if event.Metadata != nil {
				// 从元数据中获取延迟信息
				if duration, ok := event.Metadata["duration"]; ok {
					if d, ok := duration.(time.Duration); ok && d > 0 {
						latencies = append(latencies, d)
					}
				}
			}
		}
	}

	if len(latencies) == 0 {
		return &AssertionResult{
			Passed:    false,
			Message:   "No latency data available for assertion",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 排序延迟数据
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// 计算指定百分位数的延迟
	index := (a.Percentile * len(latencies)) / 100
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	percentileLatency := latencies[index]

	if percentileLatency > a.MaxLatency {
		return &AssertionResult{
			Passed:    false,
			Message:   fmt.Sprintf("Latency assertion failed: %dth percentile latency %v exceeds maximum %v", a.Percentile, percentileLatency, a.MaxLatency),
			Expected:  fmt.Sprintf("≤ %v", a.MaxLatency),
			Actual:    percentileLatency,
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   fmt.Sprintf("Latency assertion passed: %dth percentile latency %v within limit %v", a.Percentile, percentileLatency, a.MaxLatency),
		Expected:  fmt.Sprintf("≤ %v", a.MaxLatency),
		Actual:    percentileLatency,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *LatencyAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *LatencyAssertion) GetDescription() string {
	return a.Description
}

// ReconnectAssertion 重连断言
type ReconnectAssertion struct {
	Name        string
	Description string
	MaxCount    int
	MaxDuration time.Duration
}

// NewReconnectAssertion 创建重连断言
func NewReconnectAssertion(name, description string, maxCount int, maxDuration time.Duration) *ReconnectAssertion {
	return &ReconnectAssertion{
		Name:        name,
		Description: description,
		MaxCount:    maxCount,
		MaxDuration: maxDuration,
	}
}

// Assert 执行断言
func (a *ReconnectAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 统计重连事件
	var reconnectEvents []*SessionEvent
	for _, event := range session.Events {
		if event.Type == EventReconnect {
			reconnectEvents = append(reconnectEvents, event)
		}
	}

	// 检查重连次数
	if len(reconnectEvents) > a.MaxCount {
		return &AssertionResult{
			Passed:    false,
			Message:   fmt.Sprintf("Reconnect count assertion failed: %d reconnects exceed maximum %d", len(reconnectEvents), a.MaxCount),
			Expected:  fmt.Sprintf("≤ %d", a.MaxCount),
			Actual:    len(reconnectEvents),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 检查重连耗时
	for _, event := range reconnectEvents {
		if duration, ok := event.Metadata["duration"].(time.Duration); ok {
			if duration > a.MaxDuration {
				return &AssertionResult{
					Passed:    false,
					Message:   fmt.Sprintf("Reconnect duration assertion failed: duration %v exceeds maximum %v", duration, a.MaxDuration),
					Expected:  fmt.Sprintf("≤ %v", a.MaxDuration),
					Actual:    duration,
					Timestamp: time.Now(),
					Duration:  time.Since(start),
				}
			}
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   fmt.Sprintf("Reconnect assertion passed: %d reconnects within limits", len(reconnectEvents)),
		Expected:  fmt.Sprintf("count ≤ %d, duration ≤ %v", a.MaxCount, a.MaxDuration),
		Actual:    fmt.Sprintf("count: %d", len(reconnectEvents)),
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *ReconnectAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *ReconnectAssertion) GetDescription() string {
	return a.Description
}

// ErrorRateAssertion 错误率断言
type ErrorRateAssertion struct {
	Name        string
	Description string
	MaxRate     float64 // 最大错误率 (0.0-1.0)
}

// NewErrorRateAssertion 创建错误率断言
func NewErrorRateAssertion(name, description string, maxRate float64) *ErrorRateAssertion {
	return &ErrorRateAssertion{
		Name:        name,
		Description: description,
		MaxRate:     maxRate,
	}
}

// Assert 执行断言
func (a *ErrorRateAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 统计错误事件
	var errorEvents []*SessionEvent
	for _, event := range session.Events {
		if event.Type == EventError {
			errorEvents = append(errorEvents, event)
		}
	}

	// 计算错误率
	totalEvents := len(session.Events)
	if totalEvents == 0 {
		return &AssertionResult{
			Passed:    false,
			Message:   "No events available for error rate calculation",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	errorRate := float64(len(errorEvents)) / float64(totalEvents)

	if errorRate > a.MaxRate {
		return &AssertionResult{
			Passed:    false,
			Message:   fmt.Sprintf("Error rate assertion failed: %.2f%% exceeds maximum %.2f%%", errorRate*100, a.MaxRate*100),
			Expected:  fmt.Sprintf("≤ %.2f%%", a.MaxRate*100),
			Actual:    fmt.Sprintf("%.2f%%", errorRate*100),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   fmt.Sprintf("Error rate assertion passed: %.2f%% within limit %.2f%%", errorRate*100, a.MaxRate*100),
		Expected:  fmt.Sprintf("≤ %.2f%%", a.MaxRate*100),
		Actual:    fmt.Sprintf("%.2f%%", errorRate*100),
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *ErrorRateAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *ErrorRateAssertion) GetDescription() string {
	return a.Description
}

// AssertionSuite 断言套件
type AssertionSuite struct {
	Name        string
	Description string
	Assertions  []Assertion
	Results     []*AssertionResult
}

// NewAssertionSuite 创建断言套件
func NewAssertionSuite(name, description string) *AssertionSuite {
	return &AssertionSuite{
		Name:        name,
		Description: description,
		Assertions:  make([]Assertion, 0),
		Results:     make([]*AssertionResult, 0),
	}
}

// AddAssertion 添加断言
func (s *AssertionSuite) AddAssertion(assertion Assertion) {
	s.Assertions = append(s.Assertions, assertion)
}

// RunAssertions 运行所有断言
func (s *AssertionSuite) RunAssertions(session *Session) []*AssertionResult {
	s.Results = make([]*AssertionResult, 0, len(s.Assertions))

	for _, assertion := range s.Assertions {
		result := assertion.Assert(session)
		s.Results = append(s.Results, result)
	}

	return s.Results
}

// GetPassedCount 获取通过的断言数量
func (s *AssertionSuite) GetPassedCount() int {
	count := 0
	for _, result := range s.Results {
		if result.Passed {
			count++
		}
	}
	return count
}

// GetFailedCount 获取失败的断言数量
func (s *AssertionSuite) GetFailedCount() int {
	count := 0
	for _, result := range s.Results {
		if !result.Passed {
			count++
		}
	}
	return count
}

// GetSuccessRate 获取成功率
func (s *AssertionSuite) GetSuccessRate() float64 {
	if len(s.Results) == 0 {
		return 0.0
	}
	return float64(s.GetPassedCount()) / float64(len(s.Results))
}

// GetSummary 获取断言套件摘要
func (s *AssertionSuite) GetSummary() string {
	passed := s.GetPassedCount()
	total := len(s.Results)
	successRate := s.GetSuccessRate() * 100

	return fmt.Sprintf("Assertion Suite '%s': %d/%d passed (%.1f%% success rate)",
		s.Name, passed, total, successRate)
}

// ========================================
// 企业级断言 - 生产环境质量保障
// ========================================

// RecoveryTimeAssertion 恢复时间断言 - 验证重连恢复速度
type RecoveryTimeAssertion struct {
	Name            string
	Description     string
	MaxRecoveryTime time.Duration // 最大恢复时间
}

// NewRecoveryTimeAssertion 创建恢复时间断言
func NewRecoveryTimeAssertion(name, description string, maxRecoveryTime time.Duration) *RecoveryTimeAssertion {
	return &RecoveryTimeAssertion{
		Name:            name,
		Description:     description,
		MaxRecoveryTime: maxRecoveryTime,
	}
}

// Assert 执行恢复时间断言
func (a *RecoveryTimeAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	var disconnectTime, reconnectTime, firstMessageTime time.Time
	var foundDisconnect, foundReconnect, foundFirstMessage bool

	// 查找关键时间点
	for _, event := range session.Events {
		switch event.Type {
		case EventDisconnect:
			disconnectTime = event.Timestamp
			foundDisconnect = true
		case EventReconnect:
			reconnectTime = event.Timestamp
			foundReconnect = true
		case EventMessageReceive:
			if !foundFirstMessage && reconnectTime.IsZero() == false {
				// 重连后的第一条成功消息
				firstMessageTime = event.Timestamp
				foundFirstMessage = true
			}
		}
	}

	// 验证重连恢复时间
	if foundDisconnect && foundReconnect {
		recoveryTime := reconnectTime.Sub(disconnectTime)
		if recoveryTime > a.MaxRecoveryTime {
			return &AssertionResult{
				Passed:    false,
				Message:   fmt.Sprintf("Recovery time assertion failed: %v exceeds maximum %v", recoveryTime, a.MaxRecoveryTime),
				Expected:  fmt.Sprintf("≤ %v", a.MaxRecoveryTime),
				Actual:    recoveryTime,
				Timestamp: time.Now(),
				Duration:  time.Since(start),
			}
		}
	}

	// 验证服务恢复时间（重连后第一条消息的时间）
	if foundReconnect && foundFirstMessage {
		serviceRecoveryTime := firstMessageTime.Sub(reconnectTime)
		if serviceRecoveryTime > time.Second {
			return &AssertionResult{
				Passed:    false,
				Message:   fmt.Sprintf("Service recovery time assertion failed: %v exceeds maximum 1s", serviceRecoveryTime),
				Expected:  "≤ 1s",
				Actual:    serviceRecoveryTime,
				Timestamp: time.Now(),
				Duration:  time.Since(start),
			}
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   "Recovery time assertion passed",
		Expected:  fmt.Sprintf("reconnect ≤ %v, service ≤ 1s", a.MaxRecoveryTime),
		Actual:    "within limits",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *RecoveryTimeAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *RecoveryTimeAssertion) GetDescription() string {
	return a.Description
}

// PlannedFaultExemptionAssertion 计划性故障豁免断言
type PlannedFaultExemptionAssertion struct {
	Name             string
	Description      string
	ExemptionZoneLen int // 豁免区间长度（消息数）
}

// NewPlannedFaultExemptionAssertion 创建计划性故障豁免断言
func NewPlannedFaultExemptionAssertion(name, description string, exemptionZoneLen int) *PlannedFaultExemptionAssertion {
	return &PlannedFaultExemptionAssertion{
		Name:             name,
		Description:      description,
		ExemptionZoneLen: exemptionZoneLen,
	}
}

// Assert 执行计划性故障豁免断言
func (a *PlannedFaultExemptionAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 查找故障注入点（重连事件）
	var faultPoints []time.Time
	for _, event := range session.Events {
		if event.Type == EventReconnect {
			faultPoints = append(faultPoints, event.Timestamp)
		}
	}

	// 验证故障点前后消息的异常率
	for _, faultTime := range faultPoints {
		beforeFault := 0
		afterFault := 0
		beforeTimeouts := 0
		afterTimeouts := 0

		for _, event := range session.Events {
			if event.Type == EventMessageSend {
				timeDiff := faultTime.Sub(event.Timestamp)

				// 故障前后的豁免区间
				if timeDiff > 0 && timeDiff < time.Duration(a.ExemptionZoneLen)*100*time.Millisecond {
					beforeFault++
					// 检查对应的接收事件是否超时
					if a.isTimeoutInZone(event, session) {
						beforeTimeouts++
					}
				} else if timeDiff < 0 && -timeDiff < time.Duration(a.ExemptionZoneLen)*100*time.Millisecond {
					afterFault++
					if a.isTimeoutInZone(event, session) {
						afterTimeouts++
					}
				}
			}
		}

		// 验证豁免区间是否控制了异常率
		beforeRate := float64(beforeTimeouts) / float64(beforeFault+1)
		afterRate := float64(afterTimeouts) / float64(afterFault+1)

		if beforeRate > 0.5 || afterRate > 0.5 {
			return &AssertionResult{
				Passed:    false,
				Message:   fmt.Sprintf("Fault exemption assertion failed: high error rate in exemption zone (before: %.1f%%, after: %.1f%%)", beforeRate*100, afterRate*100),
				Expected:  "< 50% in exemption zone",
				Actual:    fmt.Sprintf("%.1f%% / %.1f%%", beforeRate*100, afterRate*100),
				Timestamp: time.Now(),
				Duration:  time.Since(start),
			}
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   "Planned fault exemption assertion passed",
		Expected:  "< 50% error rate in exemption zones",
		Actual:    "within limits",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// isTimeoutInZone 检查发送事件对应的接收事件是否超时
func (a *PlannedFaultExemptionAssertion) isTimeoutInZone(sendEvent *SessionEvent, session *Session) bool {
	// 简化的实现：检查是否存在对应的接收事件
	for _, event := range session.Events {
		if event.Type == EventMessageReceive {
			if seq1, ok1 := sendEvent.Metadata["sequence_num"]; ok1 {
				if seq2, ok2 := event.Metadata["sequence_num"]; ok2 {
					if seq1 == seq2 {
						// 找到了对应的接收事件，检查延迟
						if latency, ok := event.Metadata["duration"]; ok {
							if d, ok := latency.(time.Duration); ok && d > 300*time.Millisecond {
								return true // 延迟 > 300ms 认为超时
							}
						}
					}
				}
			}
		}
	}
	return false
}

// GetName 获取断言名称
func (a *PlannedFaultExemptionAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *PlannedFaultExemptionAssertion) GetDescription() string {
	return a.Description
}

// GoodputAssertion 有效吞吐断言 - 区分传输吞吐和有效吞吐
type GoodputAssertion struct {
	Name        string
	Description string
	MinGoodput  float64       // 最小有效吞吐量 (msg/s)
	WindowSize  time.Duration // 滑动窗口大小
}

// NewGoodputAssertion 创建有效吞吐断言
func NewGoodputAssertion(name, description string, minGoodput float64, windowSize time.Duration) *GoodputAssertion {
	return &GoodputAssertion{
		Name:        name,
		Description: description,
		MinGoodput:  minGoodput,
		WindowSize:  windowSize,
	}
}

// Assert 执行有效吞吐断言
func (a *GoodputAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 计算P95有效吞吐量（滑动窗口）
	var goodputs []float64

	// 收集接收事件的时间戳
	var receiveTimes []time.Time
	for _, event := range session.Events {
		if event.Type == EventMessageReceive {
			receiveTimes = append(receiveTimes, event.Timestamp)
		}
	}

	// 使用5秒滑动窗口计算吞吐量
	if len(receiveTimes) > 0 {
		windowStart := receiveTimes[0]
		windowEnd := windowStart.Add(a.WindowSize)

		for windowEnd.Before(receiveTimes[len(receiveTimes)-1]) {
			// 统计窗口内的消息数
			windowCount := 0
			for _, t := range receiveTimes {
				if t.After(windowStart) && t.Before(windowEnd) {
					windowCount++
				}
			}

			// 计算窗口吞吐量
			if windowCount > 0 {
				windowGoodput := float64(windowCount) / a.WindowSize.Seconds()
				goodputs = append(goodputs, windowGoodput)
			}

			// 滑动窗口
			windowStart = windowStart.Add(a.WindowSize / 2)
			windowEnd = windowStart.Add(a.WindowSize)
		}
	}

	// 计算P95有效吞吐量
	if len(goodputs) > 0 {
		sort.Float64s(goodputs)
		p95Index := int(float64(len(goodputs)) * 0.95)
		if p95Index >= len(goodputs) {
			p95Index = len(goodputs) - 1
		}
		p95Goodput := goodputs[p95Index]

		if p95Goodput < a.MinGoodput {
			return &AssertionResult{
				Passed:    false,
				Message:   fmt.Sprintf("Goodput assertion failed: P95 goodput %.2f msg/s below minimum %.2f msg/s", p95Goodput, a.MinGoodput),
				Expected:  fmt.Sprintf("≥ %.2f msg/s", a.MinGoodput),
				Actual:    fmt.Sprintf("%.2f msg/s", p95Goodput),
				Timestamp: time.Now(),
				Duration:  time.Since(start),
			}
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   "Goodput assertion passed",
		Expected:  fmt.Sprintf("≥ %.2f msg/s", a.MinGoodput),
		Actual:    "within limits",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *GoodputAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *GoodputAssertion) GetDescription() string {
	return a.Description
}

// TailLatencyBudgetAssertion 尾延迟预算断言
type TailLatencyBudgetAssertion struct {
	Name        string
	Description string
	BudgetLimit time.Duration // 预算上限
	WindowCount int           // 窗口数量
}

// NewTailLatencyBudgetAssertion 创建尾延迟预算断言
func NewTailLatencyBudgetAssertion(name, description string, budgetLimit time.Duration, windowCount int) *TailLatencyBudgetAssertion {
	return &TailLatencyBudgetAssertion{
		Name:        name,
		Description: description,
		BudgetLimit: budgetLimit,
		WindowCount: windowCount,
	}
}

// Assert 执行尾延迟预算断言
func (a *TailLatencyBudgetAssertion) Assert(session *Session) *AssertionResult {
	start := time.Now()

	// 收集延迟数据
	var latencies []time.Duration
	for _, event := range session.Events {
		if event.Type == EventMessageReceive {
			if latency, ok := event.Metadata["duration"]; ok {
				if d, ok := latency.(time.Duration); ok {
					latencies = append(latencies, d)
				}
			}
		}
	}

	if len(latencies) == 0 {
		return &AssertionResult{
			Passed:    true,
			Message:   "Tail latency budget assertion passed: no latency data",
			Expected:  fmt.Sprintf("≤ %v", a.BudgetLimit),
			Actual:    "no data",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 计算当前P99延迟
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p99Index := int(float64(len(latencies)) * 0.99)
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}
	currentP99 := latencies[p99Index]

	// 在CI环境中，我们只检查当前窗口
	// 在生产环境中，可以检查多个连续窗口
	if currentP99 > a.BudgetLimit {
		return &AssertionResult{
			Passed:    false,
			Message:   fmt.Sprintf("Tail latency budget assertion failed: P99 latency %v exceeds budget %v", currentP99, a.BudgetLimit),
			Expected:  fmt.Sprintf("≤ %v", a.BudgetLimit),
			Actual:    currentP99,
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	return &AssertionResult{
		Passed:    true,
		Message:   fmt.Sprintf("Tail latency budget assertion passed: P99 latency %v within budget %v", currentP99, a.BudgetLimit),
		Expected:  fmt.Sprintf("≤ %v", a.BudgetLimit),
		Actual:    currentP99,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// GetName 获取断言名称
func (a *TailLatencyBudgetAssertion) GetName() string {
	return a.Name
}

// GetDescription 获取断言描述
func (a *TailLatencyBudgetAssertion) GetDescription() string {
	return a.Description
}
