package analyzer

import (
	"fmt"
	"math"
	"time"
)

// IntelligentAnalyzer 智能分析器
type IntelligentAnalyzer struct {
	benchmarkDatabase *BenchmarkDatabase
	ruleEngine        *RuleEngine
	mlPredictor       *MLPredictor
}

// BenchmarkDatabase 基准数据库
type BenchmarkDatabase struct {
	SLGBenchmarks map[string]*SLGBenchmark `json:"slg_benchmarks"`
	History       []*TestResult            `json:"history"`
}

// SLGBenchmark SLG游戏基准数据
type SLGBenchmark struct {
	GameType           string                `json:"game_type"`
	ExpectedLatency    time.Duration         `json:"expected_latency"`
	ExpectedThroughput float64               `json:"expected_throughput"`
	BattleMetrics      *BattleBenchmark      `json:"battle_metrics"`
	ConnectionMetrics  *ConnectionBenchmark  `json:"connection_metrics"`
	Thresholds         map[string]*Threshold `json:"thresholds"`
}

// BattleBenchmark 战斗基准
type BattleBenchmark struct {
	MaxBattleInitTime  time.Duration `json:"max_battle_init_time"`
	MinUpdateFrequency float64       `json:"min_update_frequency"`
	MaxSyncError       float64       `json:"max_sync_error"`
	MaxCommandLatency  time.Duration `json:"max_command_latency"`
}

// ConnectionBenchmark 连接基准
type ConnectionBenchmark struct {
	MaxReconnectRate     float64       `json:"max_reconnect_rate"`
	MinStabilityScore    float64       `json:"min_stability_score"`
	MaxConnectionLatency time.Duration `json:"max_connection_latency"`
}

// Threshold 阈值定义
type Threshold struct {
	Excellent float64 `json:"excellent"` // A+ 阈值
	Good      float64 `json:"good"`      // A 阈值
	Average   float64 `json:"average"`   // B 阈值
	Poor      float64 `json:"poor"`      // C 阈值
	Critical  float64 `json:"critical"`  // D 阈值
}

// TestResult 测试结果
type TestResult struct {
	TestID      string        `json:"test_id"`
	TestType    string        `json:"test_type"`
	Timestamp   time.Time     `json:"timestamp"`
	Score       float64       `json:"score"`
	Grade       string        `json:"grade"`
	Metrics     *TestMetrics  `json:"metrics"`
	Issues      []*Issue      `json:"issues"`
	Suggestions []*Suggestion `json:"suggestions"`
}

// TestMetrics 测试指标
type TestMetrics struct {
	Performance *PerformanceMetrics `json:"performance"`
	Stability   *StabilityMetrics   `json:"stability"`
	SLGSpecific *SLGMetrics         `json:"slg_specific"`
	Quality     *QualityMetrics     `json:"quality"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	AverageLatency time.Duration `json:"average_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	Throughput     float64       `json:"throughput"`
	ErrorRate      float64       `json:"error_rate"`
	MemoryUsage    float64       `json:"memory_usage"`
	CPUUsage       float64       `json:"cpu_usage"`
}

// StabilityMetrics 稳定性指标
type StabilityMetrics struct {
	ReconnectRate    float64       `json:"reconnect_rate"`
	ConnectionUptime time.Duration `json:"connection_uptime"`
	MessageLossRate  float64       `json:"message_loss_rate"`
	StabilityScore   float64       `json:"stability_score"`
}

// SLGMetrics SLG游戏特定指标
type SLGMetrics struct {
	BattleInitLatency    time.Duration `json:"battle_init_latency"`
	StateUpdateFrequency float64       `json:"state_update_frequency"`
	SyncErrorRate        float64       `json:"sync_error_rate"`
	PlayerActionLatency  time.Duration `json:"player_action_latency"`
	UnityFrameRate       float64       `json:"unity_frame_rate"`
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	ProtocolCompliance float64 `json:"protocol_compliance"`
	DataIntegrity      float64 `json:"data_integrity"`
	SecurityScore      float64 `json:"security_score"`
	CompatibilityScore float64 `json:"compatibility_score"`
}

// Issue 发现的问题
type Issue struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"` // "critical", "high", "medium", "low"
	Category    string    `json:"category"` // "performance", "stability", "security", "slg_specific"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Evidence    []string  `json:"evidence"`
	Timestamp   time.Time `json:"timestamp"`
}

// Suggestion 优化建议
type Suggestion struct {
	ID             string   `json:"id"`
	Priority       string   `json:"priority"` // "high", "medium", "low"
	Category       string   `json:"category"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Actions        []string `json:"actions"`
	ExpectedImpact string   `json:"expected_impact"`
}

// ComparisonResult 对比结果
type ComparisonResult struct {
	BaselineID    string                   `json:"baseline_id"`
	CurrentID     string                   `json:"current_id"`
	OverallChange float64                  `json:"overall_change"` // 正数表示改善，负数表示恶化
	MetricChanges map[string]*MetricChange `json:"metric_changes"`
	Summary       string                   `json:"summary"`
	Highlights    []string                 `json:"highlights"`
}

// MetricChange 指标变化
type MetricChange struct {
	MetricName    string  `json:"metric_name"`
	BaselineValue float64 `json:"baseline_value"`
	CurrentValue  float64 `json:"current_value"`
	ChangePercent float64 `json:"change_percent"`
	Trend         string  `json:"trend"`  // "improved", "degraded", "stable"
	Impact        string  `json:"impact"` // "critical", "significant", "minor", "negligible"
}

// NewIntelligentAnalyzer 创建智能分析器
func NewIntelligentAnalyzer() *IntelligentAnalyzer {
	return &IntelligentAnalyzer{
		benchmarkDatabase: NewBenchmarkDatabase(),
		ruleEngine:        NewRuleEngine(),
		mlPredictor:       NewMLPredictor(),
	}
}

// AnalyzeTestResults 分析测试结果
func (ia *IntelligentAnalyzer) AnalyzeTestResults(testPlan interface{}) *TestResult {
	// 这里需要根据实际的testPlan结构进行调整
	result := &TestResult{
		TestID:    "test_123", // 从testPlan中获取
		TestType:  "slg_integration",
		Timestamp: time.Now(),
		Metrics:   ia.calculateMetrics(testPlan),
	}

	// 计算总分
	result.Score = ia.calculateOverallScore(result.Metrics)
	result.Grade = ia.assignGrade(result.Score)

	// 识别问题
	result.Issues = ia.identifyIssues(result.Metrics)

	// 生成建议
	result.Suggestions = ia.generateSuggestions(result.Metrics, result.Issues)

	return result
}

// calculateMetrics 计算各类指标
func (ia *IntelligentAnalyzer) calculateMetrics(testPlan interface{}) *TestMetrics {
	// 模拟指标计算 - 实际应该从testPlan和session数据中提取
	return &TestMetrics{
		Performance: &PerformanceMetrics{
			AverageLatency: 45 * time.Millisecond,
			P95Latency:     95 * time.Millisecond,
			P99Latency:     150 * time.Millisecond,
			Throughput:     1250.5,
			ErrorRate:      0.02,
			MemoryUsage:    78.5,
			CPUUsage:       45.2,
		},
		Stability: &StabilityMetrics{
			ReconnectRate:    0.05,
			ConnectionUptime: 58 * time.Minute,
			MessageLossRate:  0.001,
			StabilityScore:   95.5,
		},
		SLGSpecific: &SLGMetrics{
			BattleInitLatency:    850 * time.Millisecond,
			StateUpdateFrequency: 20.0,
			SyncErrorRate:        0.003,
			PlayerActionLatency:  25 * time.Millisecond,
			UnityFrameRate:       58.5,
		},
		Quality: &QualityMetrics{
			ProtocolCompliance: 98.5,
			DataIntegrity:      99.2,
			SecurityScore:      95.0,
			CompatibilityScore: 96.8,
		},
	}
}

// calculateOverallScore 计算总分 (0-100)
func (ia *IntelligentAnalyzer) calculateOverallScore(metrics *TestMetrics) float64 {
	// 权重分配
	weights := map[string]float64{
		"performance":  0.30, // 性能 30%
		"stability":    0.25, // 稳定性 25%
		"slg_specific": 0.25, // SLG特定指标 25%
		"quality":      0.20, // 质量 20%
	}

	// 计算各项得分
	perfScore := ia.calculatePerformanceScore(metrics.Performance)
	stabilityScore := ia.calculateStabilityScore(metrics.Stability)
	slgScore := ia.calculateSLGScore(metrics.SLGSpecific)
	qualityScore := ia.calculateQualityScore(metrics.Quality)

	// 加权平均
	totalScore := perfScore*weights["performance"] +
		stabilityScore*weights["stability"] +
		slgScore*weights["slg_specific"] +
		qualityScore*weights["quality"]

	return math.Round(totalScore*100) / 100
}

// calculatePerformanceScore 计算性能得分
func (ia *IntelligentAnalyzer) calculatePerformanceScore(perf *PerformanceMetrics) float64 {
	// 延迟评分 (权重 40%)
	latencyScore := ia.evaluateLatency(perf.AverageLatency, perf.P95Latency, perf.P99Latency)

	// 吞吐量评分 (权重 30%)
	throughputScore := ia.evaluateThroughput(perf.Throughput)

	// 错误率评分 (权重 20%)
	errorScore := ia.evaluateErrorRate(perf.ErrorRate)

	// 资源使用评分 (权重 10%)
	resourceScore := ia.evaluateResourceUsage(perf.MemoryUsage, perf.CPUUsage)

	finalScore := latencyScore*0.4 + throughputScore*0.3 + errorScore*0.2 + resourceScore*0.1

	return math.Max(0, math.Min(100, finalScore))
}

// calculateStabilityScore 计算稳定性得分
func (ia *IntelligentAnalyzer) calculateStabilityScore(stability *StabilityMetrics) float64 {
	score := 100.0

	// 重连率评分 (权重 40%)
	if stability.ReconnectRate > 0.1 {
		score -= 30
	} else if stability.ReconnectRate > 0.05 {
		score -= 15
	} else if stability.ReconnectRate > 0.02 {
		score -= 5
	}

	// 消息丢失率评分 (权重 30%)
	if stability.MessageLossRate > 0.01 {
		score -= 25
	} else if stability.MessageLossRate > 0.005 {
		score -= 10
	} else if stability.MessageLossRate > 0.001 {
		score -= 3
	}

	// 连接正常运行时间评分 (权重 30%)
	if stability.ConnectionUptime < 30*time.Minute {
		score -= 20
	} else if stability.ConnectionUptime < 45*time.Minute {
		score -= 10
	}

	return math.Max(0, math.Min(100, score))
}

// calculateSLGScore 计算SLG特定得分
func (ia *IntelligentAnalyzer) calculateSLGScore(slg *SLGMetrics) float64 {
	score := 100.0

	// 战斗初始化延迟评分 (权重 25%)
	if slg.BattleInitLatency > 2*time.Second {
		score -= 30
	} else if slg.BattleInitLatency > 1*time.Second {
		score -= 15
	} else if slg.BattleInitLatency > 500*time.Millisecond {
		score -= 5
	}

	// 状态更新频率评分 (权重 25%)
	if slg.StateUpdateFrequency < 10 {
		score -= 25
	} else if slg.StateUpdateFrequency < 15 {
		score -= 10
	} else if slg.StateUpdateFrequency < 20 {
		score -= 3
	}

	// 同步错误率评分 (权重 25%)
	if slg.SyncErrorRate > 0.01 {
		score -= 25
	} else if slg.SyncErrorRate > 0.005 {
		score -= 10
	} else if slg.SyncErrorRate > 0.001 {
		score -= 3
	}

	// Unity帧率评分 (权重 25%)
	if slg.UnityFrameRate < 30 {
		score -= 30
	} else if slg.UnityFrameRate < 45 {
		score -= 15
	} else if slg.UnityFrameRate < 55 {
		score -= 5
	}

	return math.Max(0, math.Min(100, score))
}

// calculateQualityScore 计算质量得分
func (ia *IntelligentAnalyzer) calculateQualityScore(quality *QualityMetrics) float64 {
	// 简单的加权平均
	return (quality.ProtocolCompliance*0.3 +
		quality.DataIntegrity*0.3 +
		quality.SecurityScore*0.2 +
		quality.CompatibilityScore*0.2)
}

// assignGrade 分配等级
func (ia *IntelligentAnalyzer) assignGrade(score float64) string {
	switch {
	case score >= 95:
		return "A+"
	case score >= 90:
		return "A"
	case score >= 85:
		return "B+"
	case score >= 80:
		return "B"
	case score >= 75:
		return "C+"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// identifyIssues 识别问题
func (ia *IntelligentAnalyzer) identifyIssues(metrics *TestMetrics) []*Issue {
	var issues []*Issue

	// 性能问题检测
	if metrics.Performance.AverageLatency > 100*time.Millisecond {
		issues = append(issues, &Issue{
			ID:          "PERF_001",
			Severity:    "high",
			Category:    "performance",
			Title:       "平均延迟过高",
			Description: fmt.Sprintf("平均延迟 %.2fms 超过推荐阈值 100ms", float64(metrics.Performance.AverageLatency)/float64(time.Millisecond)),
			Impact:      "用户体验下降，游戏响应感差",
			Evidence:    []string{"平均延迟: " + metrics.Performance.AverageLatency.String()},
			Timestamp:   time.Now(),
		})
	}

	// SLG特定问题检测
	if metrics.SLGSpecific.BattleInitLatency > 1*time.Second {
		issues = append(issues, &Issue{
			ID:          "SLG_001",
			Severity:    "critical",
			Category:    "slg_specific",
			Title:       "战斗初始化耗时过长",
			Description: fmt.Sprintf("战斗初始化耗时 %.2fs 影响游戏流畅度", metrics.SLGSpecific.BattleInitLatency.Seconds()),
			Impact:      "玩家进入战斗等待时间过长，影响游戏体验",
			Evidence:    []string{"战斗初始化延迟: " + metrics.SLGSpecific.BattleInitLatency.String()},
			Timestamp:   time.Now(),
		})
	}

	// 稳定性问题检测
	if metrics.Stability.ReconnectRate > 0.05 {
		issues = append(issues, &Issue{
			ID:          "STAB_001",
			Severity:    "medium",
			Category:    "stability",
			Title:       "重连率偏高",
			Description: fmt.Sprintf("重连率 %.2f%% 表明网络连接不稳定", metrics.Stability.ReconnectRate*100),
			Impact:      "可能导致游戏中断和数据丢失",
			Evidence:    []string{fmt.Sprintf("重连率: %.2f%%", metrics.Stability.ReconnectRate*100)},
			Timestamp:   time.Now(),
		})
	}

	return issues
}

// generateSuggestions 生成优化建议
func (ia *IntelligentAnalyzer) generateSuggestions(metrics *TestMetrics, issues []*Issue) []*Suggestion {
	var suggestions []*Suggestion

	// 基于问题生成建议
	for _, issue := range issues {
		switch issue.Category {
		case "performance":
			if issue.ID == "PERF_001" {
				suggestions = append(suggestions, &Suggestion{
					ID:          "SUGG_PERF_001",
					Priority:    "high",
					Category:    "performance",
					Title:       "优化网络延迟",
					Description: "建议优化网络通信和协议处理逻辑",
					Actions: []string{
						"检查网络配置和带宽",
						"优化消息序列化方式",
						"启用消息压缩",
						"调整心跳间隔",
					},
					ExpectedImpact: "延迟降低 20-30%",
				})
			}
		case "slg_specific":
			if issue.ID == "SLG_001" {
				suggestions = append(suggestions, &Suggestion{
					ID:          "SUGG_SLG_001",
					Priority:    "critical",
					Category:    "slg_specific",
					Title:       "优化战斗初始化流程",
					Description: "建议优化战斗数据加载和Unity渲染流程",
					Actions: []string{
						"预加载战斗资源",
						"优化Unity场景切换",
						"异步加载战斗单位数据",
						"减少不必要的网络往返",
					},
					ExpectedImpact: "战斗初始化时间减少 40-50%",
				})
			}
		}
	}

	// 通用优化建议
	if metrics.Performance.Throughput < 1000 {
		suggestions = append(suggestions, &Suggestion{
			ID:          "SUGG_GEN_001",
			Priority:    "medium",
			Category:    "performance",
			Title:       "提升消息吞吐量",
			Description: "当前吞吐量偏低，建议优化消息处理流程",
			Actions: []string{
				"增加消息批处理",
				"优化连接池配置",
				"调整缓冲区大小",
			},
			ExpectedImpact: "吞吐量提升 15-25%",
		})
	}

	return suggestions
}

// CompareBenchmarks 对比基准测试
func (ia *IntelligentAnalyzer) CompareBenchmarks(baselineID, currentID string) *ComparisonResult {
	// 模拟对比结果
	return &ComparisonResult{
		BaselineID:    baselineID,
		CurrentID:     currentID,
		OverallChange: 5.2, // 5.2% 改善
		MetricChanges: map[string]*MetricChange{
			"average_latency": {
				MetricName:    "平均延迟",
				BaselineValue: 55.0,
				CurrentValue:  45.0,
				ChangePercent: -18.2,
				Trend:         "improved",
				Impact:        "significant",
			},
			"throughput": {
				MetricName:    "吞吐量",
				BaselineValue: 1150.0,
				CurrentValue:  1250.5,
				ChangePercent: 8.7,
				Trend:         "improved",
				Impact:        "significant",
			},
		},
		Summary: "整体性能有显著提升，延迟降低18.2%，吞吐量提升8.7%",
		Highlights: []string{
			"网络延迟显著改善",
			"消息处理效率提升",
			"连接稳定性保持良好",
		},
	}
}

// 辅助函数
func (ia *IntelligentAnalyzer) evaluateLatency(avg, p95, p99 time.Duration) float64 {
	score := 100.0

	// 基于平均延迟评分
	avgMs := float64(avg) / float64(time.Millisecond)
	if avgMs > 100 {
		score -= 30
	} else if avgMs > 50 {
		score -= 15
	} else if avgMs > 25 {
		score -= 5
	}

	return score
}

func (ia *IntelligentAnalyzer) evaluateThroughput(throughput float64) float64 {
	if throughput >= 1500 {
		return 100
	} else if throughput >= 1000 {
		return 85
	} else if throughput >= 500 {
		return 70
	} else {
		return 50
	}
}

func (ia *IntelligentAnalyzer) evaluateErrorRate(errorRate float64) float64 {
	if errorRate <= 0.001 {
		return 100
	} else if errorRate <= 0.01 {
		return 90
	} else if errorRate <= 0.05 {
		return 75
	} else {
		return 50
	}
}

func (ia *IntelligentAnalyzer) evaluateResourceUsage(memory, cpu float64) float64 {
	score := 100.0

	if memory > 90 || cpu > 80 {
		score -= 30
	} else if memory > 80 || cpu > 70 {
		score -= 15
	} else if memory > 70 || cpu > 60 {
		score -= 5
	}

	return score
}

// NewBenchmarkDatabase 创建基准数据库
func NewBenchmarkDatabase() *BenchmarkDatabase {
	return &BenchmarkDatabase{
		SLGBenchmarks: make(map[string]*SLGBenchmark),
		History:       make([]*TestResult, 0),
	}
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// NewMLPredictor 创建ML预测器
func NewMLPredictor() *MLPredictor {
	return &MLPredictor{}
}

// 占位符类型
type RuleEngine struct{}
type MLPredictor struct{}
