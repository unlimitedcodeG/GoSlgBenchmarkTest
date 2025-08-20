package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
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
	Username string `yaml:"username" mapstructure:"username"`
	Token    string `yaml:"token" mapstructure:"token"`
	PlayerID string `yaml:"player_id" mapstructure:"player_id"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	WsURL   string `yaml:"ws_url" mapstructure:"ws_url"`
	HttpURL string `yaml:"http_url" mapstructure:"http_url"`
	Region  string `yaml:"region" mapstructure:"region"`
	Version string `yaml:"version" mapstructure:"version"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	TokenType    string        `yaml:"token_type" mapstructure:"token_type"`
	DefaultToken string        `yaml:"default_token" mapstructure:"default_token"`
	TestAccounts []TestAccount `yaml:"test_accounts" mapstructure:"test_accounts"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	HandshakeTimeout  time.Duration `yaml:"handshake_timeout" mapstructure:"handshake_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" mapstructure:"heartbeat_interval"`
	PingTimeout       time.Duration `yaml:"ping_timeout" mapstructure:"ping_timeout"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval" mapstructure:"reconnect_interval"`
	MaxReconnectTries int           `yaml:"max_reconnect_tries" mapstructure:"max_reconnect_tries"`
	EnableCompression bool          `yaml:"enable_compression" mapstructure:"enable_compression"`
}

// TestingConfig 测试配置
type TestingConfig struct {
	EnableRecording       bool   `yaml:"enable_recording" mapstructure:"enable_recording"`
	SessionPrefix         string `yaml:"session_prefix" mapstructure:"session_prefix"`
	AutoAssertions        bool   `yaml:"auto_assertions" mapstructure:"auto_assertions"`
	PerformanceMonitoring bool   `yaml:"performance_monitoring" mapstructure:"performance_monitoring"`
}

// Environment 环境配置
type Environment struct {
	Name        string        `yaml:"name" mapstructure:"name"`
	Description string        `yaml:"description" mapstructure:"description"`
	Active      bool          `yaml:"active" mapstructure:"active"`
	Server      ServerConfig  `yaml:"server" mapstructure:"server"`
	Auth        AuthConfig    `yaml:"auth" mapstructure:"auth"`
	Network     NetworkConfig `yaml:"network" mapstructure:"network"`
	Testing     TestingConfig `yaml:"testing" mapstructure:"testing"`
}

// RecordingConfig 录制配置
type RecordingConfig struct {
	OutputDir     string   `yaml:"output_dir" mapstructure:"output_dir"`
	AutoExport    bool     `yaml:"auto_export" mapstructure:"auto_export"`
	ExportFormat  []string `yaml:"export_format" mapstructure:"export_format"`
	Compression   bool     `yaml:"compression" mapstructure:"compression"`
	RetentionDays int      `yaml:"retention_days" mapstructure:"retention_days"`
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

// DetailedPerformanceConfig 详细性能监控配置
type DetailedPerformanceConfig struct {
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
	Recording   RecordingConfig           `yaml:"recording" mapstructure:"recording"`
	Assertions  AssertionConfig           `yaml:"assertions" mapstructure:"assertions"`
	Performance DetailedPerformanceConfig `yaml:"performance" mapstructure:"performance"`
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
	Project       string `yaml:"project" mapstructure:"project"`
	Team          string `yaml:"team" mapstructure:"team"`
	LastUpdated   string `yaml:"last_updated" mapstructure:"last_updated"`
	ConfigVersion string `yaml:"config_version" mapstructure:"config_version"`
}

// TestEnvironmentConfig 完整的测试环境配置
type TestEnvironmentConfig struct {
	Meta               MetaConfig                      `yaml:"meta" mapstructure:"meta"`
	Environments       map[EnvironmentType]Environment `yaml:"environments" mapstructure:"environments"`
	DefaultEnvironment EnvironmentType                 `yaml:"default_environment" mapstructure:"default_environment"`
	Global             GlobalConfig                    `yaml:"global" mapstructure:"global"`
	SLGGame            SLGGameConfig                   `yaml:"slg_game" mapstructure:"slg_game"`
	UnityClient        UnityClientConfig               `yaml:"unity_client" mapstructure:"unity_client"`
}

// LoadConfig 从文件加载配置（使用viper）
func LoadConfig(configPath string) (*TestEnvironmentConfig, error) {
	v := viper.New()

	// 配置文件路径和类型
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("test-environments")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath(".")
	}

	// 设置环境变量前缀
	v.SetEnvPrefix("SLGTEST")
	v.AutomaticEnv()

	// 设置默认配置（在读取配置文件之前，不会覆盖文件中的值）
	setEnvironmentDefaults(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("配置文件不存在: %s", configPath)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析到结构体
	var config TestEnvironmentConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// setEnvironmentDefaults 设置环境配置默认值
func setEnvironmentDefaults(v *viper.Viper) {
	// Meta默认值
	v.SetDefault("meta.project", "SLG Benchmark Test")
	v.SetDefault("meta.config_version", "1.0.0")
	v.SetDefault("meta.last_updated", "2024-01-01")

	// 默认环境（仅在配置文件不存在时使用）
	// v.SetDefault("default_environment", "local") // 注释掉以避免覆盖配置文件中的值

	// 全局配置默认值
	v.SetDefault("global.recording.output_dir", "./recordings")
	v.SetDefault("global.recording.auto_export", true)
	v.SetDefault("global.recording.export_format", []string{"json", "csv"})
	v.SetDefault("global.recording.compression", true)
	v.SetDefault("global.recording.retention_days", 30)

	// 性能监控默认值
	v.SetDefault("global.performance.enable_cpu_monitoring", true)
	v.SetDefault("global.performance.enable_memory_monitoring", true)
	v.SetDefault("global.performance.enable_network_monitoring", true)
	v.SetDefault("global.performance.sample_interval", "1s")
	v.SetDefault("global.performance.alert_thresholds.cpu_usage", 80.0)
	v.SetDefault("global.performance.alert_thresholds.memory_usage", 85.0)
	v.SetDefault("global.performance.alert_thresholds.latency_p99", "500ms")

	// 断言配置默认值
	v.SetDefault("global.assertions.message_order.enabled", true)
	v.SetDefault("global.assertions.message_order.min_messages", 10)
	v.SetDefault("global.assertions.message_order.max_messages", 1000)
	v.SetDefault("global.assertions.latency.enabled", true)
	v.SetDefault("global.assertions.latency.max_latency", "1s")
	v.SetDefault("global.assertions.latency.percentile", 95)
	v.SetDefault("global.assertions.reconnect.enabled", true)
	v.SetDefault("global.assertions.reconnect.max_count", 3)
	v.SetDefault("global.assertions.reconnect.max_duration", "30s")
	v.SetDefault("global.assertions.error_rate.enabled", true)
	v.SetDefault("global.assertions.error_rate.max_rate", 0.05)

	// SLG游戏配置默认值
	v.SetDefault("slg_game.battle.enable_battle_recording", true)
	v.SetDefault("slg_game.battle.battle_timeout", "5m")
	v.SetDefault("slg_game.battle.required_events", []string{"battle_start", "battle_end"})
	v.SetDefault("slg_game.building.enable_building_recording", true)
	v.SetDefault("slg_game.building.operation_timeout", "30s")
	v.SetDefault("slg_game.building.required_events", []string{"building_upgrade", "building_complete"})
	v.SetDefault("slg_game.alliance.enable_alliance_recording", false)
	v.SetDefault("slg_game.alliance.operation_timeout", "1m")

	// Unity客户端配置默认值
	v.SetDefault("unity_client.client_info.version", "1.0.0")
	v.SetDefault("unity_client.client_info.platform", "PC")
	v.SetDefault("unity_client.client_info.device_id_prefix", "test_device")
	v.SetDefault("unity_client.client_info.user_agent", "Unity/2023.2.0f1")
	v.SetDefault("unity_client.automation.enable_auto_login", true)
	v.SetDefault("unity_client.automation.enable_auto_battle", false)
	v.SetDefault("unity_client.automation.enable_auto_building", false)
	v.SetDefault("unity_client.automation.test_scenario_timeout", "10m")
	v.SetDefault("unity_client.debug.enable_verbose_logging", false)
	v.SetDefault("unity_client.debug.log_level", "INFO")
	v.SetDefault("unity_client.debug.enable_frame_logging", false)
	v.SetDefault("unity_client.debug.enable_performance_logging", true)
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

// SaveConfig 保存配置到文件（使用viper）
func (c *TestEnvironmentConfig) SaveConfig(configPath string) error {
	if configPath == "" {
		configPath = "configs/test-environments.yaml"
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 创建viper实例并设置配置
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 直接设置配置到viper中

	// 将配置写入文件
	v.Set("meta", c.Meta)
	v.Set("environments", c.Environments)
	v.Set("default_environment", c.DefaultEnvironment)
	v.Set("global", c.Global)
	v.Set("slg_game", c.SLGGame)
	v.Set("unity_client", c.UnityClient)

	if err := v.WriteConfig(); err != nil {
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
