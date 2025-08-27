package testrunner

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"GoSlgBenchmarkTest/internal/logger"
)

// TestType 测试类型
type TestType string

const (
	TestTypeUnityIntegration TestType = "unity_integration"
	TestTypeStress           TestType = "stress"
	TestTypeBenchmark        TestType = "benchmark"
	TestTypeFuzz             TestType = "fuzz"
)

// TestResult 测试结果
type TestResult struct {
	Type        TestType      `json:"type"`
	Status      string        `json:"status"` // running, completed, failed
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Output      []string      `json:"output"`
	Summary     string        `json:"summary"`
	TestsPassed int           `json:"tests_passed"`
	TestsFailed int           `json:"tests_failed"`
	Coverage    float64       `json:"coverage"`
	Score       float64       `json:"score"`
	Grade       string        `json:"grade"`
}

// TestExecutor 测试执行器
type TestExecutor struct {
	WorkDir string
	mu      sync.RWMutex
	results map[int64]*TestResult
}

// NewTestExecutor 创建测试执行器
func NewTestExecutor(workDir string) *TestExecutor {
	return &TestExecutor{
		WorkDir: workDir,
		results: make(map[int64]*TestResult),
	}
}

// ExecuteTest 执行指定类型的测试
func (te *TestExecutor) ExecuteTest(testID int64, testType TestType, config map[string]interface{}) error {
	te.mu.Lock()
	result := &TestResult{
		Type:      testType,
		Status:    "running",
		StartTime: time.Now(),
		Output:    []string{},
	}
	te.results[testID] = result
	te.mu.Unlock()

	// 创建动态配置，接口参数优先
	dynamicConfig := NewDynamicConfig(config)

	// 设置环境变量，供测试代码读取
	dynamicConfig.SetEnvironmentVariables()

	// 记录最终使用的配置
	summary := dynamicConfig.GetSummary()
	logger.LogInfo("test-runner", fmt.Sprintf("开始执行 %s 测试，配置: %+v", testType, summary), &testID)

	switch testType {
	case TestTypeUnityIntegration:
		return te.executeIntegrationTests(testID, dynamicConfig)
	case TestTypeStress:
		return te.executeStressTests(testID, dynamicConfig)
	case TestTypeBenchmark:
		return te.executeBenchmarkTests(testID, dynamicConfig)
	case TestTypeFuzz:
		return te.executeFuzzTests(testID, dynamicConfig)
	default:
		return fmt.Errorf("unsupported test type: %s", testType)
	}
}

// executeIntegrationTests 执行集成测试
func (te *TestExecutor) executeIntegrationTests(testID int64, config *DynamicConfig) error {
	logger.LogInfo("test-runner", "执行集成测试", &testID)

	// 执行 E2E WebSocket 测试
	err := te.runGoTest(testID, []string{
		"./test",
		"-v",
		"-run", "TestBasicConnection|TestReconnectAndSequenceMonotonic",
		"-timeout", "5m",
	})
	if err != nil {
		return err
	}

	// 执行 SLG 协议集成测试
	err = te.runGoTest(testID, []string{
		"./test/slg",
		"-v",
		"-run", "TestSLGProtocolIntegration|TestSLGMessageSerialization",
		"-timeout", "5m",
	})
	if err != nil {
		return err
	}

	// 执行会话测试
	return te.runGoTest(testID, []string{
		"./test/session",
		"-v",
		"-run", "TestSessionRecordingAndReplay",
		"-timeout", "5m",
	})
}

// executeStressTests 执行压力测试
func (te *TestExecutor) executeStressTests(testID int64, config *DynamicConfig) error {
	logger.LogInfo("test-runner", "执行压力测试", &testID)

	clients := config.GetConcurrentClients()
	duration := config.GetTestDuration()
	logger.LogInfo("test-runner", fmt.Sprintf("执行并发连接测试: %d 客户端, 持续 %s", clients, duration), &testID)

	// 执行并发连接测试
	err := te.runGoTest(testID, []string{
		"./test",
		"-v",
		"-run", "TestConcurrentConnections",
		"-timeout", "10m",
	})
	if err != nil {
		return err
	}

	// 执行SLG负载测试
	logger.LogInfo("test-runner", "执行SLG负载测试", &testID)
	err = te.runGoTest(testID, []string{
		"./test/slg",
		"-v",
		"-run", "TestSLGLoadTest",
		"-timeout", "10m",
	})
	if err != nil {
		return err
	}

	// 执行并发基准测试
	logger.LogInfo("test-runner", "执行并发基准测试", &testID)
	return te.runGoTest(testID, []string{
		"./test",
		"-bench", "BenchmarkConcurrentClients",
		"-benchtime", "30s",
		"-timeout", "10m",
	})
}

// executeBenchmarkTests 执行基准测试
func (te *TestExecutor) executeBenchmarkTests(testID int64, config *DynamicConfig) error {
	logger.LogInfo("test-runner", "执行基准测试", &testID)

	duration := config.GetTestDuration()
	iterations := config.GetIterations()
	clients := config.GetBenchmarkClients()

	logger.LogInfo("test-runner", fmt.Sprintf("基准测试配置: %d 客户端, %d 次迭代, 持续 %s", clients, iterations, duration), &testID)

	// 执行基准测试 - 使用固定迭代次数而不是时间
	return te.runGoTest(testID, []string{
		"./test",
		"-bench", ".",
		"-benchmem",
		"-benchtime", "100x", // 运行100次迭代，避免无限长时间运行
		"-count", "1",
		"-timeout", "5m", // 5分钟超时足够
	})
}

// executeFuzzTests 执行模糊测试
func (te *TestExecutor) executeFuzzTests(testID int64, config *DynamicConfig) error {
	logger.LogInfo("test-runner", "执行模糊测试", &testID)

	fuzzTime := config.GetFuzzTime()
	iterations := config.GetIterations()

	logger.LogInfo("test-runner", fmt.Sprintf("模糊测试配置: %d 次迭代, 持续 %s", iterations, fuzzTime), &testID)

	// 执行所有模糊测试
	fuzzTests := []string{
		"FuzzBattlePushUnmarshal",
		"FuzzLoginReqUnmarshal",
		"FuzzPlayerActionUnmarshal",
		"FuzzFrameDecode",
		"FuzzFrameDecoder",
		"FuzzErrorResp",
	}

	for _, fuzzTest := range fuzzTests {
		logger.LogInfo("test-runner", fmt.Sprintf("运行模糊测试: %s", fuzzTest), &testID)

		// 性能优化：增加fuzz测试的超时时间，给框架足够的清理时间
		// 超时时间 = fuzz时间 + 清理缓冲时间（15秒）
		timeoutDuration := fuzzTime + 15*time.Second

		err := te.runGoTest(testID, []string{
			"./test",
			"-run", "^$",
			"-fuzz", fmt.Sprintf("^%s$", fuzzTest),
			"-fuzztime", fuzzTime.String(),
			"-timeout", timeoutDuration.String(),
			// Go 1.25优化：增加并行度控制
			"-parallel", "1",
		})
		if err != nil {
			logger.LogError("test-runner", fmt.Sprintf("模糊测试 %s 失败: %v", fuzzTest), &testID)
			// 继续执行其他测试，不中断
		}
	}

	return nil
}

// runGoTest 运行go test命令
func (te *TestExecutor) runGoTest(testID int64, args []string) error {
	te.mu.RLock()
	result := te.results[testID]
	te.mu.RUnlock()

	if result == nil {
		return fmt.Errorf("test result not found for ID %d", testID)
	}

	// 构建完整的go test命令
	cmd := exec.Command("go", append([]string{"test"}, args...)...)
	cmd.Dir = te.WorkDir

	// 创建管道来捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// 合并stdout和stderr
	output := make(chan string, 1000)
	var wg sync.WaitGroup

	// 读取stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output <- line
		}
	}()

	// 读取stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			output <- line
		}
	}()

	// 关闭output channel当所有读取完成
	go func() {
		wg.Wait()
		close(output)
	}()

	// 实时处理输出
	var allOutput []string
	for line := range output {
		allOutput = append(allOutput, line)
		logger.LogInfo("test-output", line, &testID)

		// 更新结果
		te.mu.Lock()
		result.Output = append(result.Output, line)
		te.mu.Unlock()
	}

	// 等待命令完成
	err = cmd.Wait()

	// 更新最终结果
	te.mu.Lock()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Status = "failed"
		result.Summary = fmt.Sprintf("测试失败: %v", err)
		logger.LogError("test-runner", result.Summary, &testID)
	} else {
		result.Status = "completed"
		result.Summary = "测试成功完成"
		logger.LogSuccess("test-runner", result.Summary, &testID)
	}

	// 解析测试结果
	te.parseTestResults(result, allOutput)
	te.mu.Unlock()

	return err
}

// parseTestResults 解析测试输出，提取有用信息
func (te *TestExecutor) parseTestResults(result *TestResult, output []string) {
	var passed, failed int
	var coverage float64
	var benchmarkResults []string

	// 正确解析go test输出格式
	testPassRegex := regexp.MustCompile(`^--- (PASS|FAIL): (\w+)`)
	coverageRegex := regexp.MustCompile(`coverage:\s+(\d+\.?\d*)%`)
	benchRegex := regexp.MustCompile(`^Benchmark\w+`)

	for _, line := range output {
		// 解析具体测试的通过/失败
		if matches := testPassRegex.FindStringSubmatch(line); len(matches) > 1 {
			if matches[1] == "PASS" {
				passed++
			} else if matches[1] == "FAIL" {
				failed++
			}
		}

		// 解析代码覆盖率
		if matches := coverageRegex.FindStringSubmatch(line); len(matches) > 1 {
			if c, err := strconv.ParseFloat(matches[1], 64); err == nil {
				coverage = c
			}
		}

		// 收集基准测试结果
		if benchRegex.MatchString(line) {
			benchmarkResults = append(benchmarkResults, line)
		}
	}

	result.TestsPassed = passed
	result.TestsFailed = failed
	result.Coverage = coverage

	// 如果是基准测试，调整评分逻辑
	if len(benchmarkResults) > 0 {
		result.Score = te.calculateBenchmarkScore(result, benchmarkResults)
	} else {
		result.Score = te.calculateTestScore(result)
	}
	result.Grade = te.calculateGrade(result.Score)
}

// calculateTestScore 计算普通测试得分
func (te *TestExecutor) calculateTestScore(result *TestResult) float64 {
	if result.TestsPassed == 0 && result.TestsFailed == 0 {
		return 85.0 // 没有测试但编译通过，基础分
	}

	total := result.TestsPassed + result.TestsFailed
	if total == 0 {
		return 85.0
	}

	successRate := float64(result.TestsPassed) / float64(total)

	// 重新设计评分系统：
	// 1. 测试通过率是核心（70分）
	// 2. 代码覆盖率加分（最多20分）
	// 3. 无失败测试奖励（最多10分）

	baseScore := successRate * 70.0                   // 通过率：0-70分
	coverageBonus := (result.Coverage / 100.0) * 20.0 // 覆盖率：0-20分
	perfectBonus := 0.0
	if result.TestsFailed == 0 && result.TestsPassed > 0 {
		perfectBonus = 10.0 // 完美通过奖励：10分
	}

	score := baseScore + coverageBonus + perfectBonus

	// 额外奖励：多测试通过
	if result.TestsPassed >= 5 {
		score += 5.0 // 5个以上测试额外奖励
	}

	if score > 100.0 {
		score = 100.0
	}

	return score
}

// calculateBenchmarkScore 计算基准测试得分
func (te *TestExecutor) calculateBenchmarkScore(result *TestResult, benchmarkResults []string) float64 {
	// 基准测试以完成度为主
	baseScore := 80.0 // 基础分

	if len(benchmarkResults) > 0 {
		baseScore += float64(len(benchmarkResults)) * 2.0 // 每个基准测试+2分
	}

	if result.TestsFailed == 0 {
		baseScore += 15.0 // 无失败奖励
	}

	if baseScore > 100.0 {
		baseScore = 100.0
	}

	return baseScore
}

// calculateGrade 计算等级
func (te *TestExecutor) calculateGrade(score float64) string {
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

// GetResult 获取测试结果
func (te *TestExecutor) GetResult(testID int64) (*TestResult, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	result, exists := te.results[testID]
	if !exists {
		return nil, fmt.Errorf("test result not found for ID %d", testID)
	}

	return result, nil
}

// GetAllResults 获取所有测试结果
func (te *TestExecutor) GetAllResults() map[int64]*TestResult {
	te.mu.RLock()
	defer te.mu.RUnlock()

	results := make(map[int64]*TestResult)
	for id, result := range te.results {
		results[id] = result
	}

	return results
}

// 全局测试执行器实例
var GlobalExecutor *TestExecutor

// InitGlobalExecutor 初始化全局测试执行器
func InitGlobalExecutor(workDir string) {
	GlobalExecutor = NewTestExecutor(workDir)
}
