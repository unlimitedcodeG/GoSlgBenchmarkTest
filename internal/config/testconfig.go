package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// TestConfig 统一测试配置结构
type TestConfig struct {
	Meta              TestMetaConfig          `yaml:"meta"`
	Server            TestServerConfig        `yaml:"server"`
	Client            TestClientConfig        `yaml:"client"`
	TestScenarios     TestScenariosConfig     `yaml:"test_scenarios"`
	StressTest        StressTestConfig        `yaml:"stress_test"`
	Benchmark         BenchmarkConfig         `yaml:"benchmark"`
	SLGProtocol       SLGProtocolConfig       `yaml:"slg_protocol"`
	SessionRecording  SessionRecordingConfig  `yaml:"session_recording"`
	Assertions        AssertionsConfig        `yaml:"assertions"`
	Monitoring        MonitoringConfig        `yaml:"monitoring"`
	Logging           LoggingConfig           `yaml:"logging"`
	NetworkSimulation NetworkSimulationConfig `yaml:"network_simulation"`
	CICD              CICDConfig              `yaml:"ci_cd"`
	// 新增：负载测试配置
	HTTPLoadTest HTTPLoadTestConfig `yaml:"http_loadtest"`
	GRPCLoadTest GRPCLoadTestConfig `yaml:"grpc_loadtest"`
	TestServers  TestServersConfig  `yaml:"test_servers"`
}

type TestMetaConfig struct {
	Project       string `yaml:"project"`
	ConfigVersion string `yaml:"config_version"`
	LastUpdated   string `yaml:"last_updated"`
}

type TestServerConfig struct {
	BaseHost       string          `yaml:"base_host"`
	PortRange      PortRangeConfig `yaml:"port_range"`
	DefaultTimeout time.Duration   `yaml:"default_timeout"`
	WebSocket      WebSocketConfig `yaml:"websocket"`
	Push           PushConfig      `yaml:"push"`
}

type PortRangeConfig struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

type WebSocketConfig struct {
	Path              string        `yaml:"path"`
	HandshakeTimeout  time.Duration `yaml:"handshake_timeout"`
	ReadBufferSize    int           `yaml:"read_buffer_size"`
	WriteBufferSize   int           `yaml:"write_buffer_size"`
	EnableCompression bool          `yaml:"enable_compression"`
}

type PushConfig struct {
	EnableBattlePush bool          `yaml:"enable_battle_push"`
	Interval         time.Duration `yaml:"interval"`
	BatchSize        int           `yaml:"batch_size"`
}

type TestClientConfig struct {
	Connection ConnectionConfig    `yaml:"connection"`
	Heartbeat  HeartbeatConfig     `yaml:"heartbeat"`
	Reconnect  TestReconnectConfig `yaml:"reconnect"`
}

type ConnectionConfig struct {
	Timeout       time.Duration `yaml:"timeout"`
	RetryInterval time.Duration `yaml:"retry_interval"`
	MaxRetries    int           `yaml:"max_retries"`
	KeepAlive     bool          `yaml:"keep_alive"`
}

type HeartbeatConfig struct {
	Interval  time.Duration `yaml:"interval"`
	Timeout   time.Duration `yaml:"timeout"`
	MaxMissed int           `yaml:"max_missed"`
}

type TestReconnectConfig struct {
	Enable          bool          `yaml:"enable"`
	InitialInterval time.Duration `yaml:"initial_interval"`
	MaxInterval     time.Duration `yaml:"max_interval"`
	Multiplier      float64       `yaml:"multiplier"`
	MaxElapsedTime  time.Duration `yaml:"max_elapsed_time"`
}

type TestScenariosConfig struct {
	BasicConnection BasicConnectionConfig `yaml:"basic_connection"`
	Reconnect       ReconnectTestConfig   `yaml:"reconnect"`
	Heartbeat       HeartbeatTestConfig   `yaml:"heartbeat"`
	Messaging       MessagingTestConfig   `yaml:"messaging"`
}

type BasicConnectionConfig struct {
	Timeout         time.Duration `yaml:"timeout"`
	ExpectedState   string        `yaml:"expected_state"`
	ValidationDelay time.Duration `yaml:"validation_delay"`
}

type ReconnectTestConfig struct {
	ForceDisconnectAfter time.Duration `yaml:"force_disconnect_after"`
	MaxReconnectTime     time.Duration `yaml:"max_reconnect_time"`
	ExpectedReconnects   int           `yaml:"expected_reconnects"`
	SequenceValidation   bool          `yaml:"sequence_validation"`
}

type HeartbeatTestConfig struct {
	TestDuration      time.Duration `yaml:"test_duration"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	ExpectedMinRTTs   int           `yaml:"expected_min_rtts"`
	MaxAcceptableRTT  time.Duration `yaml:"max_acceptable_rtt"`
}

type MessagingTestConfig struct {
	WarmupMessages   int           `yaml:"warmup_messages"`
	WarmupDelay      time.Duration `yaml:"warmup_delay"`
	TestMessages     int           `yaml:"test_messages"`
	MessageTimeout   time.Duration `yaml:"message_timeout"`
	ValidateSequence bool          `yaml:"validate_sequence"`
}

type StressTestConfig struct {
	ConcurrentClients ConcurrentClientsConfig `yaml:"concurrent_clients"`
	Throughput        ThroughputConfig        `yaml:"throughput"`
	LargeMessages     LargeMessagesConfig     `yaml:"large_messages"`
}

type ConcurrentClientsConfig struct {
	MinClients     int           `yaml:"min_clients"`
	MaxClients     int           `yaml:"max_clients"`
	DefaultClients int           `yaml:"default_clients"`
	RampUpDuration time.Duration `yaml:"ramp_up_duration"`
	TestDuration   time.Duration `yaml:"test_duration"`
	StatsInterval  time.Duration `yaml:"stats_interval"`
}

type ThroughputConfig struct {
	TestDuration          time.Duration `yaml:"test_duration"`
	MessageInterval       time.Duration `yaml:"message_interval"`
	ExpectedMinThroughput int           `yaml:"expected_min_throughput"`
	MemoryMonitoring      bool          `yaml:"memory_monitoring"`
}

type LargeMessagesConfig struct {
	MessageSizes   map[string]int `yaml:"message_sizes"`
	TestIterations int            `yaml:"test_iterations"`
	MemoryLimit    string         `yaml:"memory_limit"`
}

type BenchmarkConfig struct {
	WarmupIterations      int                         `yaml:"warmup_iterations"`
	MinTestTime           time.Duration               `yaml:"min_test_time"`
	MaxTestTime           time.Duration               `yaml:"max_test_time"`
	SingleClientRoundtrip SingleClientRoundtripConfig `yaml:"single_client_roundtrip"`
	ConcurrentBenchmark   ConcurrentBenchmarkConfig   `yaml:"concurrent_benchmark"`
	ProtocolPerformance   ProtocolPerformanceConfig   `yaml:"protocol_performance"`
}

type SingleClientRoundtripConfig struct {
	MessageCount   int           `yaml:"message_count"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	MeasureLatency bool          `yaml:"measure_latency"`
}

type ConcurrentBenchmarkConfig struct {
	ClientCount       int  `yaml:"client_count"`
	MessagesPerClient int  `yaml:"messages_per_client"`
	ParallelExecution bool `yaml:"parallel_execution"`
}

type ProtocolPerformanceConfig struct {
	FrameSizes        []int `yaml:"frame_sizes"`
	IterationsPerSize int   `yaml:"iterations_per_size"`
	MeasureMemory     bool  `yaml:"measure_memory"`
}

type SLGProtocolConfig struct {
	Versions map[string]VersionConfig `yaml:"versions"`
	Modules  ModulesConfig            `yaml:"modules"`
	TestData TestDataConfig           `yaml:"test_data"`
}

type VersionConfig struct {
	Path               string   `yaml:"path"`
	CompatibleVersions []string `yaml:"compatible_versions"`
}

type ModulesConfig struct {
	Combat   CombatModuleConfig   `yaml:"combat"`
	Building BuildingModuleConfig `yaml:"building"`
	Alliance AllianceModuleConfig `yaml:"alliance"`
}

type CombatModuleConfig struct {
	TestBattles   int   `yaml:"test_battles"`
	MaxUnits      int   `yaml:"max_units"`
	PositionRange []int `yaml:"position_range"`
}

type BuildingModuleConfig struct {
	TestCities   int `yaml:"test_cities"`
	MaxBuildings int `yaml:"max_buildings"`
}

type AllianceModuleConfig struct {
	TestAlliances int `yaml:"test_alliances"`
	MaxMembers    int `yaml:"max_members"`
}

type TestDataConfig struct {
	BattleRequest BattleRequestConfig `yaml:"battle_request"`
	Performance   PerformanceConfig   `yaml:"performance"`
}

type BattleRequestConfig struct {
	BattleIDPrefix string   `yaml:"battle_id_prefix"`
	PlayerIDPrefix string   `yaml:"player_id_prefix"`
	DefaultUnits   []string `yaml:"default_units"`
	Formations     []string `yaml:"formations"`
	Skills         []string `yaml:"skills"`
}

type PerformanceConfig struct {
	MessageCount      int           `yaml:"message_count"`
	ConcurrentBattles int           `yaml:"concurrent_battles"`
	StressDuration    time.Duration `yaml:"stress_duration"`
}

type SessionRecordingConfig struct {
	Enable           bool                   `yaml:"enable"`
	OutputDir        string                 `yaml:"output_dir"`
	MaxFileSize      string                 `yaml:"max_file_size"`
	Compression      bool                   `yaml:"compression"`
	RecordingOptions RecordingOptionsConfig `yaml:"recording_options"`
	Replay           ReplayConfig           `yaml:"replay"`
}

type RecordingOptionsConfig struct {
	RecordMessages     bool `yaml:"record_messages"`
	RecordStateChanges bool `yaml:"record_state_changes"`
	RecordErrors       bool `yaml:"record_errors"`
	RecordTiming       bool `yaml:"record_timing"`
	RecordNetworkStats bool `yaml:"record_network_stats"`
}

type ReplayConfig struct {
	SpeedMultiplier    float64       `yaml:"speed_multiplier"`
	PauseBetweenEvents time.Duration `yaml:"pause_between_events"`
	ValidateTiming     bool          `yaml:"validate_timing"`
	ErrorTolerance     float64       `yaml:"error_tolerance"`
}

type AssertionsConfig struct {
	MessageOrder MessageOrderConfig `yaml:"message_order"`
	Latency      LatencyConfig      `yaml:"latency"`
	ErrorRate    ErrorRateConfig    `yaml:"error_rate"`
}

type ReconnectConfig struct {
	MaxAttempts    int           `yaml:"max_attempts"`
	BackoffTime    time.Duration `yaml:"backoff_time"`
	TimeoutPerTest time.Duration `yaml:"timeout_per_test"`
}

type MessageOrderConfig struct {
	MaxOutOfOrder int           `yaml:"max_out_of_order"`
	Timeout       time.Duration `yaml:"timeout"`
}

type LatencyConfig struct {
	MaxAverage time.Duration `yaml:"max_average"`
	MaxP95     time.Duration `yaml:"max_p95"`
	MaxP99     time.Duration `yaml:"max_p99"`
}

type ErrorRateConfig struct {
	MaxErrorRate float64 `yaml:"max_error_rate"`
	SampleSize   int     `yaml:"sample_size"`
}

type MonitoringConfig struct {
	Performance PerformanceMonitoringConfig `yaml:"performance"`
	HealthCheck HealthCheckConfig           `yaml:"health_check"`
}

type PerformanceMonitoringConfig struct {
	Enable             bool          `yaml:"enable"`
	CollectionInterval time.Duration `yaml:"collection_interval"`
	Metrics            []string      `yaml:"metrics"`
}

type HealthCheckConfig struct {
	Enable    bool          `yaml:"enable"`
	Interval  time.Duration `yaml:"interval"`
	Timeout   time.Duration `yaml:"timeout"`
	Endpoints []string      `yaml:"endpoints"`
}

type LoggingConfig struct {
	Level        string             `yaml:"level"`
	Format       string             `yaml:"format"`
	Output       string             `yaml:"output"`
	FileRotation FileRotationConfig `yaml:"file_rotation"`
}

type FileRotationConfig struct {
	MaxSize    string `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
}

type NetworkSimulationConfig struct {
	Enable         bool   `yaml:"enable"`
	Latency        string `yaml:"latency"`
	Jitter         string `yaml:"jitter"`
	PacketLoss     string `yaml:"packet_loss"`
	BandwidthLimit string `yaml:"bandwidth_limit"`
}

type CICDConfig struct {
	Timeouts    TimeoutsConfig    `yaml:"timeouts"`
	Parallelism ParallelismConfig `yaml:"parallelism"`
	Reporting   ReportingConfig   `yaml:"reporting"`
}

type TimeoutsConfig struct {
	UnitTests        time.Duration `yaml:"unit_tests"`
	IntegrationTests time.Duration `yaml:"integration_tests"`
	BenchmarkTests   time.Duration `yaml:"benchmark_tests"`
	StressTests      time.Duration `yaml:"stress_tests"`
}

type ParallelismConfig struct {
	MaxParallelTests   int `yaml:"max_parallel_tests"`
	MaxParallelClients int `yaml:"max_parallel_clients"`
}

type ReportingConfig struct {
	EnableCoverage   bool     `yaml:"enable_coverage"`
	EnableBenchmarks bool     `yaml:"enable_benchmarks"`
	EnableProfiling  bool     `yaml:"enable_profiling"`
	OutputFormats    []string `yaml:"output_formats"`
}

// HTTPLoadTestConfig HTTP负载测试配置
type HTTPLoadTestConfig struct {
	DefaultConfig         HTTPDefaultConfig         `yaml:"default_config"`
	PerformanceThresholds HTTPPerformanceThresholds `yaml:"performance_thresholds"`
}

type HTTPDefaultConfig struct {
	ConcurrentClients int           `yaml:"concurrent_clients"`
	Duration          time.Duration `yaml:"duration"`
	TargetRPS         int           `yaml:"target_rps"`
	Timeout           time.Duration `yaml:"timeout"`
	KeepAlive         bool          `yaml:"keep_alive"`
	MaxIdleConns      int           `yaml:"max_idle_conns"`
}

type HTTPPerformanceThresholds struct {
	MaxAvgLatency  time.Duration `yaml:"max_avg_latency"`
	MaxP99Latency  time.Duration `yaml:"max_p99_latency"`
	MinSuccessRate float64       `yaml:"min_success_rate"`
}

// GRPCLoadTestConfig gRPC负载测试配置
type GRPCLoadTestConfig struct {
	DefaultConfig         GRPCDefaultConfig         `yaml:"default_config"`
	TestMethods           []string                  `yaml:"test_methods"`
	PerformanceThresholds GRPCPerformanceThresholds `yaml:"performance_thresholds"`
}

type GRPCDefaultConfig struct {
	ConcurrentClients int           `yaml:"concurrent_clients"`
	Duration          time.Duration `yaml:"duration"`
	TargetRPS         int           `yaml:"target_rps"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
	KeepAliveTime     time.Duration `yaml:"keep_alive_time"`
	MaxConnections    int           `yaml:"max_connections"`
}

type GRPCPerformanceThresholds struct {
	MaxAvgLatency  time.Duration `yaml:"max_avg_latency"`
	MaxP99Latency  time.Duration `yaml:"max_p99_latency"`
	MinSuccessRate float64       `yaml:"min_success_rate"`
}

// TestServersConfig 测试服务器配置
type TestServersConfig struct {
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	GRPCServer GRPCServerConfig `yaml:"grpc_server"`
}

type HTTPServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	ErrorRate    float64       `yaml:"error_rate"`
}

type GRPCServerConfig struct {
	Port           int           `yaml:"port"`
	KeepAliveTime  time.Duration `yaml:"keep_alive_time"`
	MaxRecvMsgSize int           `yaml:"max_recv_msg_size"`
	ErrorRate      float64       `yaml:"error_rate"`
}

// 全局配置实例
var (
	globalConfig  *TestConfig
	configOnce    sync.Once
	portManager   *PortManager
	viperInstance *viper.Viper
)

// PortManager 端口管理器
type PortManager struct {
	mu        sync.Mutex
	usedPorts map[int]bool
	start     int
	end       int
}

// NewPortManager 创建端口管理器
func NewPortManager(start, end int) *PortManager {
	return &PortManager{
		usedPorts: make(map[int]bool),
		start:     start,
		end:       end,
	}
}

// AllocatePort 分配可用端口
func (pm *PortManager) AllocatePort() (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for port := pm.start; port <= pm.end; port++ {
		if !pm.usedPorts[port] && pm.isPortAvailable(port) {
			pm.usedPorts[port] = true
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", pm.start, pm.end)
}

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.usedPorts, port)
}

// isPortAvailable 检查端口是否可用
func (pm *PortManager) isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// LoadTestConfig 加载测试配置
func LoadTestConfig() (*TestConfig, error) {
	var err error
	configOnce.Do(func() {
		globalConfig, viperInstance, err = loadConfigFromFile()
		if err == nil && portManager == nil {
			// 初始化端口管理器（只初始化一次）
			portManager = NewPortManager(
				globalConfig.Server.PortRange.Start,
				globalConfig.Server.PortRange.End,
			)
		}
	})
	return globalConfig, err
}

// GetTestConfig 获取测试配置（已加载的）
func GetTestConfig() *TestConfig {
	if globalConfig == nil {
		config, err := LoadTestConfig()
		if err != nil {
			// 如果加载失败，创建一个基本的测试配置
			fmt.Printf("Warning: Failed to load config file, using minimal config: %v\n", err)
			globalConfig = createMinimalTestConfig()
		} else {
			globalConfig = config
		}

		// 初始化端口管理器（只初始化一次）
		if portManager == nil {
			portManager = NewPortManager(
				globalConfig.Server.PortRange.Start,
				globalConfig.Server.PortRange.End,
			)
		}
	}
	return globalConfig
}

// createMinimalTestConfig 创建最小测试配置
func createMinimalTestConfig() *TestConfig {
	config := &TestConfig{
		Meta: TestMetaConfig{
			Project:       "GoSlgBenchmarkTest",
			ConfigVersion: "1.0.0",
			LastUpdated:   "2025-01-27",
		},
		Server: TestServerConfig{
			BaseHost: "127.0.0.1",
			PortRange: PortRangeConfig{
				Start: 18000,
				End:   18999,
			},
			DefaultTimeout: 10 * time.Second,
			WebSocket: WebSocketConfig{
				Path:              "/ws",
				HandshakeTimeout:  15 * time.Second,
				ReadBufferSize:    4096,
				WriteBufferSize:   4096,
				EnableCompression: true,
			},
			Push: PushConfig{
				EnableBattlePush: true,
				Interval:         100 * time.Millisecond,
				BatchSize:        10,
			},
		},
		Client: TestClientConfig{
			Connection: ConnectionConfig{
				Timeout:       10 * time.Second,
				RetryInterval: 1 * time.Second,
				MaxRetries:    5,
				KeepAlive:     true,
			},
			Heartbeat: HeartbeatConfig{
				Interval:  30 * time.Second,
				Timeout:   10 * time.Second,
				MaxMissed: 3,
			},
			Reconnect: TestReconnectConfig{
				Enable:          true,
				InitialInterval: 1 * time.Second,
				MaxInterval:     30 * time.Second,
				Multiplier:      2.0,
				MaxElapsedTime:  5 * time.Minute,
			},
		},
		StressTest: StressTestConfig{
			ConcurrentClients: ConcurrentClientsConfig{
				MinClients:     1,
				MaxClients:     100,
				DefaultClients: 10,
				RampUpDuration: 5 * time.Second,
				TestDuration:   60 * time.Second,
				StatsInterval:  5 * time.Second,
			},
		},
		Assertions: AssertionsConfig{
			ErrorRate: ErrorRateConfig{
				MaxErrorRate: 0.05,
				SampleSize:   100,
			},
		},
		SessionRecording: SessionRecordingConfig{
			Enable:      true,
			OutputDir:   "./recordings",
			Compression: true,
		},
		HTTPLoadTest: HTTPLoadTestConfig{
			DefaultConfig: HTTPDefaultConfig{
				ConcurrentClients: 20,
				Duration:          120 * time.Second,
				TargetRPS:         500,
				Timeout:           30 * time.Second,
				KeepAlive:         true,
				MaxIdleConns:      100,
			},
			PerformanceThresholds: HTTPPerformanceThresholds{
				MaxAvgLatency:  50 * time.Millisecond,
				MaxP99Latency:  200 * time.Millisecond,
				MinSuccessRate: 0.98,
			},
		},
		GRPCLoadTest: GRPCLoadTestConfig{
			DefaultConfig: GRPCDefaultConfig{
				ConcurrentClients: 15,
				Duration:          90 * time.Second,
				TargetRPS:         300,
				RequestTimeout:    10 * time.Second,
				KeepAliveTime:     30 * time.Second,
				MaxConnections:    8,
			},
			TestMethods: []string{"Login", "SendPlayerAction", "GetPlayerStatus", "GetBattleStatus"},
			PerformanceThresholds: GRPCPerformanceThresholds{
				MaxAvgLatency:  30 * time.Millisecond,
				MaxP99Latency:  100 * time.Millisecond,
				MinSuccessRate: 0.99,
			},
		},
		TestServers: TestServersConfig{
			HTTPServer: HTTPServerConfig{
				Port:         19000,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				ErrorRate:    0.02,
			},
			GRPCServer: GRPCServerConfig{
				Port:           19001,
				KeepAliveTime:  60 * time.Second,
				MaxRecvMsgSize: 4 * 1024 * 1024, // 4MB
				ErrorRate:      0.01,
			},
		},
		TestScenarios: TestScenariosConfig{
			BasicConnection: BasicConnectionConfig{
				Timeout:         10 * time.Second,
				ExpectedState:   "CONNECTED",
				ValidationDelay: 100 * time.Millisecond,
			},
			Reconnect: ReconnectTestConfig{
				ForceDisconnectAfter: 2 * time.Second,
				MaxReconnectTime:     10 * time.Second,
				ExpectedReconnects:   2,
				SequenceValidation:   true,
			},
			Heartbeat: HeartbeatTestConfig{
				TestDuration:      5 * time.Second,
				HeartbeatInterval: 1 * time.Second,
				ExpectedMinRTTs:   3,
				MaxAcceptableRTT:  100 * time.Millisecond,
			},
			Messaging: MessagingTestConfig{
				WarmupMessages:   5,
				WarmupDelay:      100 * time.Millisecond,
				TestMessages:     10,
				MessageTimeout:   5 * time.Second,
				ValidateSequence: true,
			},
		},
	}

	return config
}

// GetPortManager 获取端口管理器
func GetPortManager() *PortManager {
	if portManager == nil {
		GetTestConfig() // 确保配置已加载
	}
	return portManager
}

// loadConfigFromFile 使用Viper从文件加载配置
func loadConfigFromFile() (*TestConfig, *viper.Viper, error) {
	v := viper.New()

	// 配置文件搜索路径
	v.SetConfigName("test-config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AddConfigPath(".")

	// 设置环境变量前缀
	v.SetEnvPrefix("TEST")
	v.AutomaticEnv()

	// 设置默认值
	setDefaultValues(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, nil, fmt.Errorf("failed to read config file: %v", err)
		}
	}

	// 打印详细的调试信息
	fmt.Printf("Debug: Before unmarshal - Raw viper values - PortRange start: %v, end: %v\n",
		v.Get("server.port_range.start"), v.Get("server.port_range.end"))

	// 解析到结构体
	var config TestConfig
	if err := v.Unmarshal(&config); err != nil {
		// 如果unmarshal失败，返回最小配置而不是错误
		fmt.Printf("Warning: Failed to unmarshal config, using minimal config: %v\n", err)
		config = *createMinimalTestConfig()
		return &config, v, nil
	}

	// PortRange - 修复端口范围解析问题
	config.Server.PortRange.Start = v.GetInt("server.port_range.start")
	config.Server.PortRange.End = v.GetInt("server.port_range.end")

	// DefaultTimeout - 智能解析：优先尝试duration，失败则尝试int秒
	if config.Server.DefaultTimeout == 0 {
		d := v.GetDuration("server.default_timeout")
		if d == 0 {
			// 如果duration解析失败，尝试解析为int秒数
			if seconds := v.GetInt("server.default_timeout"); seconds > 0 {
				config.Server.DefaultTimeout = time.Duration(seconds) * time.Second
			} else {
				// 最后兜底使用默认值
				config.Server.DefaultTimeout = 10 * time.Second
			}
		} else {
			config.Server.DefaultTimeout = d
		}
	}

	// StressTest - 修复压力测试配置解析问题
	if config.StressTest.ConcurrentClients.DefaultClients == 0 {
		config.StressTest.ConcurrentClients.DefaultClients = v.GetInt("stress_test.concurrent_clients.default_clients")
	}

	// 打印详细的调试信息
	fmt.Printf("Debug: After unmarshal - Raw viper values - PortRange start: %v, end: %v\n",
		v.Get("server.port_range.start"), v.Get("server.port_range.end"))
	fmt.Printf("Debug: Config struct - PortRange: %d-%d\n", config.Server.PortRange.Start, config.Server.PortRange.End)
	fmt.Printf("Debug: Default values check - start: %v, end: %v\n",
		v.GetString("server.port_range.start"), v.GetString("server.port_range.end"))

	// 验证配置
	if err := validateConfig(&config); err != nil {
		// 如果验证失败，返回最小配置而不是错误
		fmt.Printf("Warning: Config validation failed, using minimal config: %v\n", err)
		config = *createMinimalTestConfig()
		return &config, v, nil
	}

	return &config, v, nil
}

// setDefaultValues 设置默认配置值
func setDefaultValues(v *viper.Viper) {
	// Meta默认值
	v.SetDefault("meta.project", "GoSlgBenchmarkTest")
	v.SetDefault("meta.config_version", "1.0.0")

	// Server默认值
	v.SetDefault("server.base_host", "127.0.0.1")
	v.SetDefault("server.port_range.start", 18000)
	v.SetDefault("server.port_range.end", 18999)
	v.SetDefault("server.default_timeout", "10s")
	v.SetDefault("server.websocket.path", "/ws")
	v.SetDefault("server.websocket.handshake_timeout", "15s")
	v.SetDefault("server.websocket.read_buffer_size", 4096)
	v.SetDefault("server.websocket.write_buffer_size", 4096)
	v.SetDefault("server.websocket.enable_compression", true)
	v.SetDefault("server.push.enable_battle_push", true)
	v.SetDefault("server.push.interval", "100ms")
	v.SetDefault("server.push.batch_size", 10)

	// Client默认值
	v.SetDefault("client.connection.timeout", "10s")
	v.SetDefault("client.connection.retry_interval", "1s")
	v.SetDefault("client.connection.max_retries", 5)
	v.SetDefault("client.connection.keep_alive", true)
	v.SetDefault("client.heartbeat.interval", "30s")
	v.SetDefault("client.heartbeat.timeout", "10s")
	v.SetDefault("client.heartbeat.max_missed", 3)
	v.SetDefault("client.reconnect.enable", true)
	v.SetDefault("client.reconnect.initial_interval", "1s")
	v.SetDefault("client.reconnect.max_interval", "30s")
	v.SetDefault("client.reconnect.multiplier", 2.0)
	v.SetDefault("client.reconnect.max_elapsed_time", "5m")

	// 测试场景默认值
	v.SetDefault("test_scenarios.basic_connection.timeout", "10s")
	v.SetDefault("test_scenarios.basic_connection.expected_state", "CONNECTED")
	v.SetDefault("test_scenarios.basic_connection.validation_delay", "100ms")
	v.SetDefault("test_scenarios.reconnect.force_disconnect_after", "2s")
	v.SetDefault("test_scenarios.reconnect.max_reconnect_time", "10s")
	v.SetDefault("test_scenarios.reconnect.expected_reconnects", 2)
	v.SetDefault("test_scenarios.reconnect.sequence_validation", true)
	v.SetDefault("test_scenarios.heartbeat.test_duration", "5s")
	v.SetDefault("test_scenarios.heartbeat.heartbeat_interval", "200ms")
	v.SetDefault("test_scenarios.heartbeat.expected_min_rtts", 1)
	v.SetDefault("test_scenarios.heartbeat.max_acceptable_rtt", "1s")
	v.SetDefault("test_scenarios.messaging.warmup_messages", 10)
	v.SetDefault("test_scenarios.messaging.warmup_delay", "1ms")
	v.SetDefault("test_scenarios.messaging.test_messages", 100)
	v.SetDefault("test_scenarios.messaging.message_timeout", "5s")
	v.SetDefault("test_scenarios.messaging.validate_sequence", true)

	// 压力测试默认值
	v.SetDefault("stress_test.concurrent_clients.min_clients", 1)
	v.SetDefault("stress_test.concurrent_clients.max_clients", 100)
	v.SetDefault("stress_test.concurrent_clients.default_clients", 10)
	v.SetDefault("stress_test.concurrent_clients.ramp_up_duration", "5s")
	v.SetDefault("stress_test.concurrent_clients.test_duration", "60s")
	v.SetDefault("stress_test.concurrent_clients.stats_interval", "5s")
	v.SetDefault("stress_test.throughput.test_duration", "10s")
	v.SetDefault("stress_test.throughput.message_interval", "1ms")
	v.SetDefault("stress_test.throughput.expected_min_throughput", 100)
	v.SetDefault("stress_test.throughput.memory_monitoring", true)
	v.SetDefault("stress_test.large_messages.message_sizes.small", 1024)
	v.SetDefault("stress_test.large_messages.message_sizes.medium", 10240)
	v.SetDefault("stress_test.large_messages.message_sizes.large", 102400)
	v.SetDefault("stress_test.large_messages.message_sizes.huge", 1048576)
	v.SetDefault("stress_test.large_messages.test_iterations", 100)
	v.SetDefault("stress_test.large_messages.memory_limit", "100MB")

	// 基准测试默认值
	v.SetDefault("benchmark.warmup_iterations", 10)
	v.SetDefault("benchmark.min_test_time", "1s")
	v.SetDefault("benchmark.max_test_time", "10s")
	v.SetDefault("benchmark.single_client_roundtrip.message_count", 1000)
	v.SetDefault("benchmark.single_client_roundtrip.request_timeout", "5s")
	v.SetDefault("benchmark.single_client_roundtrip.measure_latency", true)
	v.SetDefault("benchmark.concurrent_benchmark.client_count", 10)
	v.SetDefault("benchmark.concurrent_benchmark.messages_per_client", 100)
	v.SetDefault("benchmark.concurrent_benchmark.parallel_execution", true)
	v.SetDefault("benchmark.protocol_performance.frame_sizes", []int{64, 256, 1024, 4096, 16384})
	v.SetDefault("benchmark.protocol_performance.iterations_per_size", 10000)
	v.SetDefault("benchmark.protocol_performance.measure_memory", true)

	// SLG协议默认值
	v.SetDefault("slg_protocol.test_data.battle_request.battle_id_prefix", "test_battle_")
	v.SetDefault("slg_protocol.test_data.battle_request.player_id_prefix", "test_player_")
	v.SetDefault("slg_protocol.test_data.battle_request.default_units", []string{"unit_1", "unit_2", "unit_3"})
	v.SetDefault("slg_protocol.test_data.battle_request.formations", []string{"triangle", "line", "circle"})
	v.SetDefault("slg_protocol.test_data.battle_request.skills", []string{"fireball", "heal", "shield", "lightning"})
	v.SetDefault("slg_protocol.test_data.performance.message_count", 1000)
	v.SetDefault("slg_protocol.test_data.performance.concurrent_battles", 10)
	v.SetDefault("slg_protocol.test_data.performance.stress_duration", "30s")

	// 断言默认值
	v.SetDefault("assertions.message_order.max_out_of_order", 0)
	v.SetDefault("assertions.message_order.timeout", "5s")
	v.SetDefault("assertions.latency.max_average", "100ms")
	v.SetDefault("assertions.latency.max_p95", "200ms")
	v.SetDefault("assertions.latency.max_p99", "500ms")
	v.SetDefault("assertions.error_rate.max_error_rate", 0.05)
	v.SetDefault("assertions.error_rate.sample_size", 100)

	// 监控默认值
	v.SetDefault("monitoring.performance.enable", true)
	v.SetDefault("monitoring.performance.collection_interval", "1s")
	v.SetDefault("monitoring.performance.metrics", []string{"cpu_usage", "memory_usage", "goroutine_count", "connection_count", "message_rate"})
	v.SetDefault("monitoring.health_check.enable", true)
	v.SetDefault("monitoring.health_check.interval", "30s")
	v.SetDefault("monitoring.health_check.timeout", "5s")
	v.SetDefault("monitoring.health_check.endpoints", []string{"/health", "/stats"})

	// 日志默认值
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.file_rotation.max_size", "100MB")
	v.SetDefault("logging.file_rotation.max_backups", 5)
	v.SetDefault("logging.file_rotation.max_age_days", 7)

	// HTTP负载测试默认值
	v.SetDefault("http_loadtest.default_config.concurrent_clients", 20)
	v.SetDefault("http_loadtest.default_config.duration", "120s")
	v.SetDefault("http_loadtest.default_config.target_rps", 500)
	v.SetDefault("http_loadtest.default_config.timeout", "30s")
	v.SetDefault("http_loadtest.default_config.keep_alive", true)
	v.SetDefault("http_loadtest.default_config.max_idle_conns", 100)
	v.SetDefault("http_loadtest.performance_thresholds.max_avg_latency", "50ms")
	v.SetDefault("http_loadtest.performance_thresholds.max_p99_latency", "200ms")
	v.SetDefault("http_loadtest.performance_thresholds.min_success_rate", 0.98)

	// gRPC负载测试默认值
	v.SetDefault("grpc_loadtest.default_config.concurrent_clients", 15)
	v.SetDefault("grpc_loadtest.default_config.duration", "90s")
	v.SetDefault("grpc_loadtest.default_config.target_rps", 300)
	v.SetDefault("grpc_loadtest.default_config.request_timeout", "10s")
	v.SetDefault("grpc_loadtest.default_config.keep_alive_time", "30s")
	v.SetDefault("grpc_loadtest.default_config.max_connections", 8)
	v.SetDefault("grpc_loadtest.test_methods", []string{"Login", "SendPlayerAction", "GetPlayerStatus", "GetBattleStatus"})
	v.SetDefault("grpc_loadtest.performance_thresholds.max_avg_latency", "30ms")
	v.SetDefault("grpc_loadtest.performance_thresholds.max_p99_latency", "100ms")
	v.SetDefault("grpc_loadtest.performance_thresholds.min_success_rate", 0.99)

	// 测试服务器默认值
	v.SetDefault("test_servers.http_server.port", 19000)
	v.SetDefault("test_servers.http_server.read_timeout", "30s")
	v.SetDefault("test_servers.http_server.write_timeout", "30s")
	v.SetDefault("test_servers.http_server.error_rate", 0.02)
	v.SetDefault("test_servers.grpc_server.port", 19001)
	v.SetDefault("test_servers.grpc_server.keep_alive_time", "60s")
	v.SetDefault("test_servers.grpc_server.max_recv_msg_size", 4*1024*1024)
	v.SetDefault("test_servers.grpc_server.error_rate", 0.01)
}

// validateConfig 验证配置有效性
func validateConfig(config *TestConfig) error {
	// 验证端口范围
	if config.Server.PortRange.Start >= config.Server.PortRange.End {
		return fmt.Errorf("invalid port range: start(%d) >= end(%d)",
			config.Server.PortRange.Start, config.Server.PortRange.End)
	}

	// 验证超时时间
	if config.Server.DefaultTimeout <= 0 {
		return fmt.Errorf("invalid server timeout: %v", config.Server.DefaultTimeout)
	}

	if config.Client.Connection.Timeout <= 0 {
		return fmt.Errorf("invalid client connection timeout: %v", config.Client.Connection.Timeout)
	}

	// 验证压力测试参数
	if config.StressTest.ConcurrentClients.DefaultClients < 1 {
		return fmt.Errorf("invalid default clients count: %d", config.StressTest.ConcurrentClients.DefaultClients)
	}

	if config.StressTest.Throughput.ExpectedMinThroughput < 0 {
		return fmt.Errorf("invalid min throughput: %d", config.StressTest.Throughput.ExpectedMinThroughput)
	}

	// 验证断言参数
	if config.Assertions.ErrorRate.MaxErrorRate < 0 || config.Assertions.ErrorRate.MaxErrorRate > 1 {
		return fmt.Errorf("invalid max error rate: %f (must be between 0 and 1)",
			config.Assertions.ErrorRate.MaxErrorRate)
	}

	return nil
}

// GetConfigValue 获取配置值（支持环境变量覆盖）
func GetConfigValue(key string) interface{} {
	if viperInstance != nil {
		return viperInstance.Get(key)
	}
	return nil
}

// GetConfigString 获取字符串配置值
func GetConfigString(key string) string {
	if viperInstance != nil {
		return viperInstance.GetString(key)
	}
	return ""
}

// GetConfigInt 获取整数配置值
func GetConfigInt(key string) int {
	if viperInstance != nil {
		return viperInstance.GetInt(key)
	}
	return 0
}

// GetConfigDuration 获取时间配置值
func GetConfigDuration(key string) time.Duration {
	if viperInstance != nil {
		return viperInstance.GetDuration(key)
	}
	return 0
}

// GetConfigBool 获取布尔配置值
func GetConfigBool(key string) bool {
	if viperInstance != nil {
		return viperInstance.GetBool(key)
	}
	return false
}

// SetConfigValue 设置配置值（运行时动态配置）
func SetConfigValue(key string, value interface{}) {
	if viperInstance != nil {
		viperInstance.Set(key, value)
	}
}

// WatchConfig 监听配置文件变化（热重载）
func WatchConfig() {
	if viperInstance != nil {
		viperInstance.WatchConfig()
		viperInstance.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("Config file changed: %s\n", e.Name)
			// 重新加载配置
			reloadConfig()
		})
	}
}

// reloadConfig 重新加载配置
func reloadConfig() {
	var newConfig TestConfig
	if err := viperInstance.Unmarshal(&newConfig); err != nil {
		fmt.Printf("Failed to reload config: %v\n", err)
		return
	}

	if err := validateConfig(&newConfig); err != nil {
		fmt.Printf("Config validation failed during reload: %v\n", err)
		return
	}

	// 原子更新全局配置
	globalConfig = &newConfig
	fmt.Println("Configuration reloaded successfully")
}

// PrintConfigInfo 打印配置信息（调试用）
func PrintConfigInfo() {
	if viperInstance == nil {
		fmt.Println("Configuration not loaded")
		return
	}

	fmt.Println("Configuration Info:")
	fmt.Printf("  Config file: %s\n", viperInstance.ConfigFileUsed())
	fmt.Printf("  Environment variables prefix: TEST_\n")
	fmt.Printf("  Search paths: %v\n", viperInstance.Get("_search_paths"))

	// 显示一些关键配置
	fmt.Printf("Key Configuration Values:\n")
	fmt.Printf("  Server Host: %s\n", viperInstance.GetString("server.base_host"))
	fmt.Printf("  Port Range: %d-%d\n",
		viperInstance.GetInt("server.port_range.start"),
		viperInstance.GetInt("server.port_range.end"))
	fmt.Printf("  Default Clients: %d\n", viperInstance.GetInt("stress_test.concurrent_clients.default_clients"))
	fmt.Printf("  Test Duration: %s\n", viperInstance.GetString("stress_test.concurrent_clients.test_duration"))
}

// GetEnvironmentOverrides 获取环境变量覆盖的配置
func GetEnvironmentOverrides() map[string]interface{} {
	overrides := make(map[string]interface{})

	if viperInstance == nil {
		return overrides
	}

	// 常用的环境变量
	envKeys := []string{
		"TEST_SERVER_BASE_HOST",
		"TEST_SERVER_PORT_RANGE_START",
		"TEST_SERVER_PORT_RANGE_END",
		"TEST_STRESS_TEST_CONCURRENT_CLIENTS_DEFAULT_CLIENTS",
		"TEST_STRESS_TEST_CONCURRENT_CLIENTS_TEST_DURATION",
		"TEST_CLIENT_CONNECTION_TIMEOUT",
		"TEST_LOGGING_LEVEL",
	}

	for _, envKey := range envKeys {
		if value := os.Getenv(envKey); value != "" {
			configKey := convertEnvToConfigKey(envKey)
			overrides[configKey] = value
		}
	}

	return overrides
}

// convertEnvToConfigKey 将环境变量名转换为配置键
func convertEnvToConfigKey(envKey string) string {
	// 移除前缀 TEST_
	key := envKey[5:]
	// 转换为小写并替换下划线为点
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "_", ".")
	return key
}

// GetServerAddress 获取服务器地址
func (c *TestConfig) GetServerAddress() (string, error) {
	port, err := GetPortManager().AllocatePort()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", c.Server.BaseHost, port), nil
}

// GetWebSocketURL 获取WebSocket URL
func (c *TestConfig) GetWebSocketURL(addr string) string {
	return fmt.Sprintf("ws://%s%s", addr, c.Server.WebSocket.Path)
}

// ReleaseServerPort 释放服务器端口
func (c *TestConfig) ReleaseServerPort(addr string) {
	if host, portStr, err := net.SplitHostPort(addr); err == nil && host == c.Server.BaseHost {
		if port := parsePort(portStr); port > 0 {
			GetPortManager().ReleasePort(port)
		}
	}
}

// parsePort 解析端口号
func parsePort(portStr string) int {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

// GetTestDataDir 获取测试数据目录
func (c *TestConfig) GetTestDataDir() string {
	if c.SessionRecording.OutputDir == "" {
		return "recordings"
	}
	return c.SessionRecording.OutputDir
}

// EnsureTestDataDir 确保测试数据目录存在
func (c *TestConfig) EnsureTestDataDir() error {
	dir := c.GetTestDataDir()
	return os.MkdirAll(dir, 0755)
}

// GetLogLevel 获取日志级别
func (c *TestConfig) GetLogLevel() string {
	if c.Logging.Level == "" {
		return "info"
	}
	return c.Logging.Level
}

// IsDebugEnabled 是否启用调试模式
func (c *TestConfig) IsDebugEnabled() bool {
	return c.GetLogLevel() == "debug"
}

// GetBenchmarkIterations 获取基准测试迭代次数
func (c *TestConfig) GetBenchmarkIterations() int {
	if c.Benchmark.SingleClientRoundtrip.MessageCount == 0 {
		return 1000
	}
	return c.Benchmark.SingleClientRoundtrip.MessageCount
}

// GetStressTestClients 获取压力测试客户端数量
func (c *TestConfig) GetStressTestClients() int {
	if c.StressTest.ConcurrentClients.DefaultClients == 0 {
		return 10
	}
	return c.StressTest.ConcurrentClients.DefaultClients
}

// GetStressTestDuration 获取压力测试持续时间
func (c *TestConfig) GetStressTestDuration() time.Duration {
	if c.StressTest.ConcurrentClients.TestDuration == 0 {
		return 60 * time.Second
	}
	return c.StressTest.ConcurrentClients.TestDuration
}
