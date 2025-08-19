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
		if event.Type == EventMessageReceive && event.Duration > 0 {
			latencies = append(latencies, event.Duration)
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
