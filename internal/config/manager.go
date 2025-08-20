package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigManager 统一配置管理器
type ConfigManager struct {
	mu             sync.RWMutex
	testConfig     *TestConfig
	envConfig      *TestEnvironmentConfig
	testViper      *viper.Viper
	testConfigPath string
	envConfigPath  string
	watchEnabled   bool
}

// ConfigManagerOption 配置管理器选项
type ConfigManagerOption func(*ConfigManager)

// WithTestConfigPath 设置测试配置文件路径
func WithTestConfigPath(path string) ConfigManagerOption {
	return func(cm *ConfigManager) {
		cm.testConfigPath = path
	}
}

// WithEnvConfigPath 设置环境配置文件路径
func WithEnvConfigPath(path string) ConfigManagerOption {
	return func(cm *ConfigManager) {
		cm.envConfigPath = path
	}
}

// WithWatchEnabled 启用配置文件监控
func WithWatchEnabled(enabled bool) ConfigManagerOption {
	return func(cm *ConfigManager) {
		cm.watchEnabled = enabled
	}
}

// NewConfigManager 创建配置管理器
func NewConfigManager(opts ...ConfigManagerOption) *ConfigManager {
	cm := &ConfigManager{
		testConfigPath: "",
		envConfigPath:  "",
		watchEnabled:   false,
	}

	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

// LoadTestConfig 加载测试配置
func (cm *ConfigManager) LoadTestConfig() (*TestConfig, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.testConfig != nil {
		return cm.testConfig, nil
	}

	config, viperInstance, err := loadConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("加载测试配置失败: %w", err)
	}

	cm.testConfig = config
	cm.testViper = viperInstance

	// 启用监控
	if cm.watchEnabled {
		cm.watchTestConfig()
	}

	return config, nil
}

// LoadEnvironmentConfig 加载环境配置
func (cm *ConfigManager) LoadEnvironmentConfig() (*TestEnvironmentConfig, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.envConfig != nil {
		return cm.envConfig, nil
	}

	config, err := LoadConfig(cm.envConfigPath)
	if err != nil {
		return nil, fmt.Errorf("加载环境配置失败: %w", err)
	}

	cm.envConfig = config

	// 启用监控
	if cm.watchEnabled {
		cm.watchEnvConfig()
	}

	return config, nil
}

// GetTestConfig 获取测试配置（如果未加载则自动加载）
func (cm *ConfigManager) GetTestConfig() (*TestConfig, error) {
	cm.mu.RLock()
	if cm.testConfig != nil {
		defer cm.mu.RUnlock()
		return cm.testConfig, nil
	}
	cm.mu.RUnlock()

	return cm.LoadTestConfig()
}

// GetEnvironmentConfig 获取环境配置（如果未加载则自动加载）
func (cm *ConfigManager) GetEnvironmentConfig() (*TestEnvironmentConfig, error) {
	cm.mu.RLock()
	if cm.envConfig != nil {
		defer cm.mu.RUnlock()
		return cm.envConfig, nil
	}
	cm.mu.RUnlock()

	return cm.LoadEnvironmentConfig()
}

// GetEnvironment 获取指定环境配置
func (cm *ConfigManager) GetEnvironment(envType EnvironmentType) (*Environment, error) {
	envConfig, err := cm.GetEnvironmentConfig()
	if err != nil {
		return nil, err
	}

	return envConfig.GetEnvironment(envType)
}

// GetDefaultEnvironment 获取默认环境配置
func (cm *ConfigManager) GetDefaultEnvironment() (*Environment, error) {
	envConfig, err := cm.GetEnvironmentConfig()
	if err != nil {
		return nil, err
	}

	return envConfig.GetDefaultEnvironment()
}

// ReloadTestConfig 重新加载测试配置
func (cm *ConfigManager) ReloadTestConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, viperInstance, err := loadConfigFromFile()
	if err != nil {
		return fmt.Errorf("重新加载测试配置失败: %w", err)
	}

	cm.testConfig = config
	cm.testViper = viperInstance

	return nil
}

// ReloadEnvironmentConfig 重新加载环境配置
func (cm *ConfigManager) ReloadEnvironmentConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, err := LoadConfig(cm.envConfigPath)
	if err != nil {
		return fmt.Errorf("重新加载环境配置失败: %w", err)
	}

	cm.envConfig = config

	return nil
}

// watchTestConfig 监控测试配置文件变化
func (cm *ConfigManager) watchTestConfig() {
	if cm.testViper == nil {
		return
	}

	cm.testViper.WatchConfig()
	cm.testViper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件变化时重新加载
		cm.ReloadTestConfig()
	})
}

// watchEnvConfig 监控环境配置文件变化
func (cm *ConfigManager) watchEnvConfig() {
	// TODO: 为环境配置实现文件监控
	// 需要创建新的viper实例来监控environment配置
}

// ValidateConfigs 验证所有配置
func (cm *ConfigManager) ValidateConfigs() error {
	// 验证测试配置
	testConfig, err := cm.GetTestConfig()
	if err != nil {
		return fmt.Errorf("测试配置验证失败: %w", err)
	}

	if err := validateConfig(testConfig); err != nil {
		return fmt.Errorf("测试配置验证失败: %w", err)
	}

	// 验证环境配置
	envConfig, err := cm.GetEnvironmentConfig()
	if err != nil {
		return fmt.Errorf("环境配置验证失败: %w", err)
	}

	if err := envConfig.Validate(); err != nil {
		return fmt.Errorf("环境配置验证失败: %w", err)
	}

	return nil
}

// GetConfigSummary 获取配置摘要信息
func (cm *ConfigManager) GetConfigSummary() (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 测试配置摘要
	if testConfig, err := cm.GetTestConfig(); err == nil {
		summary["test_config"] = map[string]interface{}{
			"project":        testConfig.Meta.Project,
			"config_version": testConfig.Meta.ConfigVersion,
			"last_updated":   testConfig.Meta.LastUpdated,
			"server_host":    testConfig.Server.BaseHost,
		}
	}

	// 环境配置摘要
	if envConfig, err := cm.GetEnvironmentConfig(); err == nil {
		summary["environment_config"] = map[string]interface{}{
			"project":             envConfig.Meta.Project,
			"config_version":      envConfig.Meta.ConfigVersion,
			"default_environment": envConfig.DefaultEnvironment,
			"environments":        envConfig.GetEnvironmentNames(),
		}
	}

	return summary, nil
}

// 全局配置管理器实例
var (
	globalConfigManager *ConfigManager
	configManagerOnce   sync.Once
)

// GetGlobalConfigManager 获取全局配置管理器
func GetGlobalConfigManager() *ConfigManager {
	configManagerOnce.Do(func() {
		globalConfigManager = NewConfigManager(
			WithWatchEnabled(true),
		)
	})
	return globalConfigManager
}

// 便捷函数，使用全局配置管理器

// GetGlobalTestConfig 获取全局测试配置
func GetGlobalTestConfig() (*TestConfig, error) {
	return GetGlobalConfigManager().GetTestConfig()
}

// GetGlobalEnvironmentConfig 获取全局环境配置
func GetGlobalEnvironmentConfig() (*TestEnvironmentConfig, error) {
	return GetGlobalConfigManager().GetEnvironmentConfig()
}

// GetGlobalEnvironment 获取全局指定环境配置
func GetGlobalEnvironment(envType EnvironmentType) (*Environment, error) {
	return GetGlobalConfigManager().GetEnvironment(envType)
}

// GetGlobalDefaultEnvironment 获取全局默认环境配置
func GetGlobalDefaultEnvironment() (*Environment, error) {
	return GetGlobalConfigManager().GetDefaultEnvironment()
}
