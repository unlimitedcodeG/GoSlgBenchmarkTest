package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// EnvironmentType 环境类型枚举
type EnvironmentType string

const (
	EnvDevelopment EnvironmentType = "development"
	EnvTesting     EnvironmentType = "testing"
	EnvStaging     EnvironmentType = "staging"
	EnvLocal       EnvironmentType = "local"
)

// String 实现字符串接口
func (e EnvironmentType) String() string {
	return string(e)
}

// IsValid 检查环境类型是否有效
func (e EnvironmentType) IsValid() bool {
	switch e {
	case EnvDevelopment, EnvTesting, EnvStaging, EnvLocal:
		return true
	default:
		return false
	}
}

// TestAccount 测试账号配置
type TestAccount struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	PlayerID string `yaml:"player_id"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	WsURL   string `yaml:"ws_url"`
	HttpURL string `yaml:"http_url"`
	Region  string `yaml:"region"`
	Version string `yaml:"version"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	TokenType    string        `yaml:"token_type"`
	DefaultToken string        `yaml:"default_token"`
	TestAccounts []TestAccount `yaml:"test_accounts"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	HandshakeTimeout  time.Duration `yaml:"handshake_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	PingTimeout       time.Duration `yaml:"ping_timeout"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`
	MaxReconnectTries int           `yaml:"max_reconnect_tries"`
	EnableCompression bool          `yaml:"enable_compression"`
}

// TestingConfig 测试配置
type TestingConfig struct {
	EnableRecording       bool   `yaml:"enable_recording"`
	SessionPrefix         string `yaml:"session_prefix"`
	AutoAssertions        bool   `yaml:"auto_assertions"`
	PerformanceMonitoring bool   `yaml:"performance_monitoring"`
}

// Environment 环境配置
type Environment struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Active      bool          `yaml:"active"`
	Server      ServerConfig  `yaml:"server"`
	Auth        AuthConfig    `yaml:"auth"`
	Network     NetworkConfig `yaml:"network"`
	Testing     TestingConfig `yaml:"testing"`
}

// RecordingConfig 录制配置
type RecordingConfig struct {
	OutputDir     string   `yaml:"output_dir"`
	AutoExport    bool     `yaml:"auto_export"`
	ExportFormat  []string `yaml:"export_format"`
	Compression   bool     `yaml:"compression"`
	RetentionDays int      `yaml:"retention_days"`
}

// AssertionConfig 断言配置
type AssertionConfig struct {
	MessageOrder MessageOrderAssertionConfig `yaml:"message_order"`
	Latency      LatencyAssertionConfig      `yaml:"latency"`
	Reconnect    ReconnectAssertionConfig    `yaml:"reconnect"`
	ErrorRate    ErrorRateAssertionConfig    `yaml:"error_rate"`
}

// MessageOrderAssertionConfig 消息顺序断言配置
type MessageOrderAssertionConfig struct {
	Enabled     bool `yaml:"enabled"`
	MinMessages int  `yaml:"min_messages"`
	MaxMessages int  `yaml:"max_messages"`
}

// LatencyAssertionConfig 延迟断言配置
type LatencyAssertionConfig struct {
	Enabled    bool          `yaml:"enabled"`
	MaxLatency time.Duration `yaml:"max_latency"`
	Percentile int           `yaml:"percentile"`
}

// ReconnectAssertionConfig 重连断言配置
type ReconnectAssertionConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxCount    int           `yaml:"max_count"`
	MaxDuration time.Duration `yaml:"max_duration"`
}

// ErrorRateAssertionConfig 错误率断言配置
type ErrorRateAssertionConfig struct {
	Enabled bool    `yaml:"enabled"`
	MaxRate float64 `yaml:"max_rate"`
}

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	EnableCPUMonitoring     bool                       `yaml:"enable_cpu_monitoring"`
	EnableMemoryMonitoring  bool                       `yaml:"enable_memory_monitoring"`
	EnableNetworkMonitoring bool                       `yaml:"enable_network_monitoring"`
	SampleInterval          time.Duration              `yaml:"sample_interval"`
	AlertThresholds         PerformanceAlertThresholds `yaml:"alert_thresholds"`
}

// PerformanceAlertThresholds 性能告警阈值
type PerformanceAlertThresholds struct {
	CPUUsage    float64       `yaml:"cpu_usage"`
	MemoryUsage float64       `yaml:"memory_usage"`
	LatencyP99  time.Duration `yaml:"latency_p99"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	Recording   RecordingConfig   `yaml:"recording"`
	Assertions  AssertionConfig   `yaml:"assertions"`
	Performance PerformanceConfig `yaml:"performance"`
}

// SLGBattleConfig SLG战斗系统配置
type SLGBattleConfig struct {
	EnableBattleRecording bool                     `yaml:"enable_battle_recording"`
	BattleTimeout         time.Duration            `yaml:"battle_timeout"`
	RequiredEvents        []string                 `yaml:"required_events"`
	PerformanceThresholds map[string]time.Duration `yaml:"performance_thresholds"`
}

// SLGBuildingConfig SLG建筑系统配置
type SLGBuildingConfig struct {
	EnableBuildingRecording bool                     `yaml:"enable_building_recording"`
	OperationTimeout        time.Duration            `yaml:"operation_timeout"`
	RequiredEvents          []string                 `yaml:"required_events"`
	PerformanceThresholds   map[string]time.Duration `yaml:"performance_thresholds"`
}

// SLGAllianceConfig SLG联盟系统配置
type SLGAllianceConfig struct {
	EnableAllianceRecording bool                     `yaml:"enable_alliance_recording"`
	OperationTimeout        time.Duration            `yaml:"operation_timeout"`
	RequiredEvents          []string                 `yaml:"required_events"`
	PerformanceThresholds   map[string]time.Duration `yaml:"performance_thresholds"`
}

// SLGGameConfig SLG游戏特定配置
type SLGGameConfig struct {
	Battle   SLGBattleConfig   `yaml:"battle"`
	Building SLGBuildingConfig `yaml:"building"`
	Alliance SLGAllianceConfig `yaml:"alliance"`
}

// UnityClientInfo Unity客户端信息
type UnityClientInfo struct {
	Version        string `yaml:"version"`
	Platform       string `yaml:"platform"`
	DeviceIDPrefix string `yaml:"device_id_prefix"`
	UserAgent      string `yaml:"user_agent"`
}

// UnityAutomationConfig Unity自动化配置
type UnityAutomationConfig struct {
	EnableAutoLogin     bool          `yaml:"enable_auto_login"`
	EnableAutoBattle    bool          `yaml:"enable_auto_battle"`
	EnableAutoBuilding  bool          `yaml:"enable_auto_building"`
	TestScenarioTimeout time.Duration `yaml:"test_scenario_timeout"`
}

// UnityDebugConfig Unity调试配置
type UnityDebugConfig struct {
	EnableVerboseLogging     bool   `yaml:"enable_verbose_logging"`
	LogLevel                 string `yaml:"log_level"`
	EnableFrameLogging       bool   `yaml:"enable_frame_logging"`
	EnablePerformanceLogging bool   `yaml:"enable_performance_logging"`
}

// UnityClientConfig Unity客户端配置
type UnityClientConfig struct {
	ClientInfo UnityClientInfo       `yaml:"client_info"`
	Automation UnityAutomationConfig `yaml:"automation"`
	Debug      UnityDebugConfig      `yaml:"debug"`
}

// MetaConfig 元数据配置
type MetaConfig struct {
	Project       string `yaml:"project"`
	Team          string `yaml:"team"`
	LastUpdated   string `yaml:"last_updated"`
	ConfigVersion string `yaml:"config_version"`
}

// TestEnvironmentConfig 完整的测试环境配置
type TestEnvironmentConfig struct {
	Meta               MetaConfig                      `yaml:"meta"`
	Environments       map[EnvironmentType]Environment `yaml:"environments"`
	DefaultEnvironment EnvironmentType                 `yaml:"default_environment"`
	Global             GlobalConfig                    `yaml:"global"`
	SLGGame            SLGGameConfig                   `yaml:"slg_game"`
	UnityClient        UnityClientConfig               `yaml:"unity_client"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*TestEnvironmentConfig, error) {
	// 如果没有指定路径，使用默认路径
	if configPath == "" {
		configPath = "configs/test-environments.yaml"
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config TestEnvironmentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// GetEnvironment 获取指定环境配置
func (c *TestEnvironmentConfig) GetEnvironment(envType EnvironmentType) (*Environment, error) {
	env, exists := c.Environments[envType]
	if !exists {
		return nil, fmt.Errorf("环境不存在: %s", envType)
	}

	if !env.Active {
		return nil, fmt.Errorf("环境已禁用: %s", envType)
	}

	return &env, nil
}

// GetDefaultEnvironment 获取默认环境配置
func (c *TestEnvironmentConfig) GetDefaultEnvironment() (*Environment, error) {
	return c.GetEnvironment(c.DefaultEnvironment)
}

// GetTestAccount 获取测试账号
func (env *Environment) GetTestAccount(username string) (*TestAccount, error) {
	for _, account := range env.Auth.TestAccounts {
		if account.Username == username {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("测试账号不存在: %s", username)
}

// GetFirstTestAccount 获取第一个可用的测试账号
func (env *Environment) GetFirstTestAccount() (*TestAccount, error) {
	if len(env.Auth.TestAccounts) == 0 {
		return nil, fmt.Errorf("没有可用的测试账号")
	}
	return &env.Auth.TestAccounts[0], nil
}

// Validate 验证配置
func (c *TestEnvironmentConfig) Validate() error {
	// 验证默认环境是否存在
	if !c.DefaultEnvironment.IsValid() {
		return fmt.Errorf("无效的默认环境类型: %s", c.DefaultEnvironment)
	}

	if _, exists := c.Environments[c.DefaultEnvironment]; !exists {
		return fmt.Errorf("默认环境不存在: %s", c.DefaultEnvironment)
	}

	// 验证每个环境配置
	for envType, env := range c.Environments {
		if err := env.Validate(envType); err != nil {
			return fmt.Errorf("环境 %s 配置错误: %w", envType, err)
		}
	}

	return nil
}

// Validate 验证环境配置
func (env *Environment) Validate(envType EnvironmentType) error {
	if env.Name == "" {
		return fmt.Errorf("环境名称不能为空")
	}

	if env.Server.WsURL == "" {
		return fmt.Errorf("WebSocket URL不能为空")
	}

	if env.Auth.DefaultToken == "" {
		return fmt.Errorf("默认Token不能为空")
	}

	if env.Network.HandshakeTimeout <= 0 {
		return fmt.Errorf("握手超时时间必须大于0")
	}

	if env.Network.HeartbeatInterval <= 0 {
		return fmt.Errorf("心跳间隔必须大于0")
	}

	if env.Network.MaxReconnectTries < 0 {
		return fmt.Errorf("最大重连次数不能为负数")
	}

	return nil
}

// SaveConfig 保存配置到文件
func (c *TestEnvironmentConfig) SaveConfig(configPath string) error {
	if configPath == "" {
		configPath = "configs/test-environments.yaml"
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化为YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// ListEnvironments 列出所有可用环境
func (c *TestEnvironmentConfig) ListEnvironments() map[EnvironmentType]Environment {
	activeEnvs := make(map[EnvironmentType]Environment)
	for envType, env := range c.Environments {
		if env.Active {
			activeEnvs[envType] = env
		}
	}
	return activeEnvs
}

// GetEnvironmentNames 获取所有环境名称
func (c *TestEnvironmentConfig) GetEnvironmentNames() []string {
	names := make([]string, 0, len(c.Environments))
	for envType, env := range c.Environments {
		if env.Active {
			names = append(names, fmt.Sprintf("%s (%s)", env.Name, envType))
		}
	}
	return names
}
