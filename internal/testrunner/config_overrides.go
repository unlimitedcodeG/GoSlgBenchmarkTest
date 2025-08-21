package testrunner

import (
	"os"
	"strconv"
	"time"

	"GoSlgBenchmarkTest/internal/config"
)

// DynamicConfig 动态配置结构，支持接口参数覆盖配置文件
type DynamicConfig struct {
	baseConfig *config.TestConfig
	overrides  map[string]interface{}
}

// NewDynamicConfig 创建动态配置
func NewDynamicConfig(overrides map[string]interface{}) *DynamicConfig {
	return &DynamicConfig{
		baseConfig: config.GetTestConfig(),
		overrides:  overrides,
	}
}

// GetConcurrentClients 获取并发客户端数量 (接口参数优先)
func (dc *DynamicConfig) GetConcurrentClients() int {
	// 1. 检查接口传入的参数
	if v, ok := dc.overrides["concurrent_clients"]; ok {
		if clients, ok := v.(float64); ok { // JSON数字解析为float64
			return int(clients)
		}
		if clients, ok := v.(int); ok {
			return clients
		}
	}

	// 2. 检查环境变量覆盖
	if envVal := os.Getenv("TEST_STRESS_CONCURRENT_CLIENTS"); envVal != "" {
		if clients, err := strconv.Atoi(envVal); err == nil && clients > 0 {
			return clients
		}
	}

	// 3. 使用配置文件默认值
	if dc.baseConfig.StressTest.ConcurrentClients.DefaultClients > 0 {
		return dc.baseConfig.StressTest.ConcurrentClients.DefaultClients
	}

	// 4. 最后的安全默认值
	return 10
}

// GetBenchmarkClients 获取基准测试客户端数量
func (dc *DynamicConfig) GetBenchmarkClients() int {
	// 1. 检查接口传入的参数
	if v, ok := dc.overrides["benchmark_clients"]; ok {
		if clients, ok := v.(float64); ok {
			return int(clients)
		}
		if clients, ok := v.(int); ok {
			return clients
		}
	}

	// 2. 检查环境变量
	if envVal := os.Getenv("TEST_BENCHMARK_CLIENTS"); envVal != "" {
		if clients, err := strconv.Atoi(envVal); err == nil && clients > 0 {
			return clients
		}
	}

	// 3. 使用配置文件默认值
	if dc.baseConfig.Benchmark.ConcurrentBenchmark.ClientCount > 0 {
		return dc.baseConfig.Benchmark.ConcurrentBenchmark.ClientCount
	}

	// 4. 安全默认值
	return 5
}

// GetTestDuration 获取测试持续时间
func (dc *DynamicConfig) GetTestDuration() time.Duration {
	// 1. 检查接口传入的参数
	if v, ok := dc.overrides["duration"]; ok {
		if duration, ok := v.(string); ok {
			if d, err := time.ParseDuration(duration); err == nil {
				return d
			}
		}
	}

	// 2. 检查环境变量
	if envVal := os.Getenv("TEST_DURATION"); envVal != "" {
		if d, err := time.ParseDuration(envVal); err == nil {
			return d
		}
	}

	// 3. 使用配置文件默认值 (已经是time.Duration类型)
	if dc.baseConfig.StressTest.ConcurrentClients.TestDuration > 0 {
		return dc.baseConfig.StressTest.ConcurrentClients.TestDuration
	}

	// 4. 默认值 - 基准测试用较短时间
	return 30 * time.Second
}

// GetIterations 获取迭代次数
func (dc *DynamicConfig) GetIterations() int {
	// 1. 检查接口传入的参数
	if v, ok := dc.overrides["iterations"]; ok {
		if iterations, ok := v.(float64); ok {
			return int(iterations)
		}
		if iterations, ok := v.(int); ok {
			return iterations
		}
	}

	// 2. 检查环境变量
	if envVal := os.Getenv("TEST_ITERATIONS"); envVal != "" {
		if iterations, err := strconv.Atoi(envVal); err == nil && iterations > 0 {
			return iterations
		}
	}

	// 3. 使用配置文件默认值
	if dc.baseConfig.Benchmark.SingleClientRoundtrip.MessageCount > 0 {
		return dc.baseConfig.Benchmark.SingleClientRoundtrip.MessageCount
	}

	// 4. 默认值
	return 1000
}

// GetFuzzTime 获取模糊测试时间
func (dc *DynamicConfig) GetFuzzTime() time.Duration {
	// 1. 检查接口传入的参数
	if v, ok := dc.overrides["fuzz_time"]; ok {
		if fuzzTime, ok := v.(string); ok {
			if d, err := time.ParseDuration(fuzzTime); err == nil {
				return d
			}
		}
	}

	// 2. 检查环境变量
	if envVal := os.Getenv("TEST_FUZZ_TIME"); envVal != "" {
		if d, err := time.ParseDuration(envVal); err == nil {
			return d
		}
	}

	// 3. 默认值
	return 30 * time.Second
}

// SetEnvironmentVariables 将接口参数设置为环境变量，供测试代码读取
func (dc *DynamicConfig) SetEnvironmentVariables() {
	// 设置并发客户端数量
	clients := dc.GetConcurrentClients()
	os.Setenv("TEST_STRESS_CONCURRENT_CLIENTS", strconv.Itoa(clients))

	// 设置基准测试客户端数量
	benchClients := dc.GetBenchmarkClients()
	os.Setenv("TEST_BENCHMARK_CLIENTS", strconv.Itoa(benchClients))

	// 设置测试持续时间
	duration := dc.GetTestDuration()
	os.Setenv("TEST_DURATION", duration.String())

	// 设置迭代次数
	iterations := dc.GetIterations()
	os.Setenv("TEST_ITERATIONS", strconv.Itoa(iterations))

	// 设置模糊测试时间
	fuzzTime := dc.GetFuzzTime()
	os.Setenv("TEST_FUZZ_TIME", fuzzTime.String())
}

// GetSummary 获取配置摘要用于日志
func (dc *DynamicConfig) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"concurrent_clients": dc.GetConcurrentClients(),
		"benchmark_clients":  dc.GetBenchmarkClients(),
		"duration":           dc.GetTestDuration().String(),
		"iterations":         dc.GetIterations(),
		"fuzz_time":          dc.GetFuzzTime().String(),
		"override_source":    "interface_params_priority",
	}
}
