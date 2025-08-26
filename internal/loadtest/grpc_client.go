package loadtest

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// GRPCLoadTestConfig gRPC负载测试配置
type GRPCLoadTestConfig struct {
	ServerAddr        string
	TLS               bool
	ConcurrentClients int
	Duration          time.Duration
	TargetRPS         int // 目标每秒请求数

	// 测试方法配置
	TestMethods      []string // 要测试的方法
	RequestTimeout   time.Duration
	KeepAliveTime    time.Duration
	KeepAliveTimeout time.Duration

	// 连接池配置
	MaxConnections  int
	ConnectionReuse bool
}

// GRPCLoadTestResult gRPC负载测试结果
type GRPCLoadTestResult struct {
	// 基础指标
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	Duration           time.Duration

	// 延迟指标 (毫秒)
	MinLatency float64
	MaxLatency float64
	AvgLatency float64
	P50Latency float64
	P95Latency float64
	P99Latency float64

	// 吞吐量指标
	RequestsPerSecond float64
	BytesPerSecond    float64

	// 错误统计
	ErrorsByType map[string]int64

	// 方法级统计
	MethodStats map[string]*MethodStats

	// 时间序列数据
	LatencyTimeSeries    []TimePoint
	ThroughputTimeSeries []TimePoint
	ErrorTimeSeries      []TimePoint
}

type MethodStats struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	AvgLatency      float64
	MinLatency      float64
	MaxLatency      float64
}

type TimePoint struct {
	Timestamp time.Time
	Value     float64
}

// GRPCLoadTester gRPC负载测试器
type GRPCLoadTester struct {
	config *GRPCLoadTestConfig

	// 连接管理
	connections []*grpc.ClientConn
	clients     []gamev1.GameServiceClient

	// 统计数据
	metrics *GRPCTestMetrics

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 结果
	result *GRPCLoadTestResult
	mu     sync.RWMutex
}

// GRPCTestMetrics gRPC测试指标收集器
type GRPCTestMetrics struct {
	totalRequests   atomic.Int64
	successRequests atomic.Int64
	failedRequests  atomic.Int64
	totalBytes      atomic.Int64

	// 延迟统计
	latencies []time.Duration
	latencyMu sync.Mutex

	// 错误统计
	errorCounts map[string]*atomic.Int64
	errorMu     sync.RWMutex

	// 方法统计
	methodMetrics map[string]*MethodMetrics
	methodMu      sync.RWMutex

	// 时间序列
	timeSeriesData []TimeSeries
	timeSeriesMu   sync.Mutex

	startTime time.Time
}

type MethodMetrics struct {
	requests  atomic.Int64
	successes atomic.Int64
	failures  atomic.Int64
	latencies []time.Duration
	mu        sync.Mutex
}

type TimeSeries struct {
	Timestamp time.Time
	Requests  int64
	Latency   float64
	Errors    int64
}

// NewGRPCLoadTester 创建新的gRPC负载测试器
func NewGRPCLoadTester(config *GRPCLoadTestConfig) *GRPCLoadTester {
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+time.Minute)

	return &GRPCLoadTester{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		metrics: &GRPCTestMetrics{
			errorCounts:   make(map[string]*atomic.Int64),
			methodMetrics: make(map[string]*MethodMetrics),
			startTime:     time.Now(),
		},
	}
}

// Start 开始负载测试
func (t *GRPCLoadTester) Start() error {
	log.Printf("Starting gRPC load test: %d clients, %v duration, target %d RPS",
		t.config.ConcurrentClients, t.config.Duration, t.config.TargetRPS)

	// 初始化连接
	if err := t.initConnections(); err != nil {
		return fmt.Errorf("failed to initialize connections: %v", err)
	}

	// 启动指标收集
	go t.collectMetrics()

	// 启动负载生成器
	return t.startLoadGeneration()
}

// initConnections 初始化gRPC连接
func (t *GRPCLoadTester) initConnections() error {
	connectionCount := t.config.MaxConnections
	if connectionCount <= 0 {
		connectionCount = t.config.ConcurrentClients
	}

	t.connections = make([]*grpc.ClientConn, connectionCount)
	t.clients = make([]gamev1.GameServiceClient, connectionCount)

	// gRPC连接选项
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                t.config.KeepAliveTime,
			Timeout:             t.config.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}

	if t.config.TLS {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 创建连接
	for i := 0; i < connectionCount; i++ {
		conn, err := grpc.DialContext(t.ctx, t.config.ServerAddr, opts...)
		if err != nil {
			// 清理已创建的连接
			for j := 0; j < i; j++ {
				t.connections[j].Close()
			}
			return fmt.Errorf("failed to connect to gRPC server: %v", err)
		}

		t.connections[i] = conn
		t.clients[i] = gamev1.NewGameServiceClient(conn)
	}

	log.Printf("Established %d gRPC connections to %s", connectionCount, t.config.ServerAddr)
	return nil
}

// startLoadGeneration 开始负载生成
func (t *GRPCLoadTester) startLoadGeneration() error {
	// 计算每个客户端的目标RPS
	rpsPerClient := float64(t.config.TargetRPS) / float64(t.config.ConcurrentClients)
	if rpsPerClient < 1 {
		rpsPerClient = 1
	}

	// 启动客户端goroutines
	for i := 0; i < t.config.ConcurrentClients; i++ {
		t.wg.Add(1)
		go t.clientWorker(i, rpsPerClient)
	}

	// 等待测试完成
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	// 等待超时或完成
	select {
	case <-done:
		log.Printf("gRPC load test completed")
	case <-time.After(t.config.Duration + time.Minute):
		log.Printf("gRPC load test timeout")
		t.cancel()
	}

	// 生成结果
	t.generateResult()

	// 清理连接
	t.cleanup()

	return nil
}

// clientWorker 客户端工作器
func (t *GRPCLoadTester) clientWorker(clientID int, targetRPS float64) {
	defer t.wg.Done()

	// 选择连接（负载均衡）
	connIndex := clientID % len(t.connections)
	client := t.clients[connIndex]

	// 计算请求间隔
	interval := time.Duration(float64(time.Second) / targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	requestID := 0

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			// 执行请求
			t.executeRequest(client, clientID, requestID)
			requestID++
		}
	}
}

// executeRequest 执行单个请求
func (t *GRPCLoadTester) executeRequest(client gamev1.GameServiceClient, clientID, requestID int) {
	// 选择测试方法
	method := t.selectTestMethod(requestID)

	start := time.Now()
	var err error
	var success bool

	// 创建请求上下文
	ctx, cancel := context.WithTimeout(t.ctx, t.config.RequestTimeout)
	defer cancel()

	// 执行对应的方法
	switch method {
	case "Login":
		success, err = t.testLogin(ctx, client, clientID, requestID)
	case "SendPlayerAction":
		success, err = t.testSendPlayerAction(ctx, client, clientID, requestID)
	case "GetPlayerStatus":
		success, err = t.testGetPlayerStatus(ctx, client, clientID, requestID)
	case "GetBattleStatus":
		success, err = t.testGetBattleStatus(ctx, client, clientID, requestID)
	case "StreamBattleUpdates":
		success, err = t.testStreamBattleUpdates(ctx, client, clientID, requestID)
	default:
		success, err = t.testLogin(ctx, client, clientID, requestID) // 默认测试登录
	}

	latency := time.Since(start)

	// 记录指标
	t.recordMetrics(method, latency, success, err)
}

// 测试方法实现
func (t *GRPCLoadTester) testLogin(ctx context.Context, client gamev1.GameServiceClient, clientID, requestID int) (bool, error) {
	req := &gamev1.LoginReq{
		Token:         fmt.Sprintf("test_token_%d_%d", clientID, requestID),
		ClientVersion: "1.0.0",
		DeviceId:      fmt.Sprintf("device_%d", clientID),
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.Ok, nil
}

func (t *GRPCLoadTester) testSendPlayerAction(ctx context.Context, client gamev1.GameServiceClient, clientID, requestID int) (bool, error) {
	req := &gamev1.PlayerAction{
		ActionSeq:       uint64(requestID),
		PlayerId:        fmt.Sprintf("player_%d", clientID),
		ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
		ClientTimestamp: time.Now().UnixMilli(),
		ActionData: &gamev1.ActionData{
			Data: &gamev1.ActionData_Move{
				Move: &gamev1.MoveAction{
					TargetPosition: &gamev1.Position{
						X: float32(requestID % 100),
						Y: float32((requestID * 2) % 100),
						Z: 0,
					},
					MoveSpeed: 5.0,
				},
			},
		},
	}

	resp, err := client.SendPlayerAction(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

func (t *GRPCLoadTester) testGetPlayerStatus(ctx context.Context, client gamev1.GameServiceClient, clientID, requestID int) (bool, error) {
	req := &gamev1.PlayerStatusReq{
		PlayerId: fmt.Sprintf("player_%d", clientID),
	}

	resp, err := client.GetPlayerStatus(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.PlayerId != "", nil
}

func (t *GRPCLoadTester) testGetBattleStatus(ctx context.Context, client gamev1.GameServiceClient, clientID, requestID int) (bool, error) {
	req := &gamev1.BattleStatusReq{
		BattleId:     fmt.Sprintf("battle_%d", requestID%10), // 循环使用10个战斗ID
		IncludeUnits: true,
	}

	resp, err := client.GetBattleStatus(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.BattleId != "", nil
}

func (t *GRPCLoadTester) testStreamBattleUpdates(ctx context.Context, client gamev1.GameServiceClient, clientID, requestID int) (bool, error) {
	req := &gamev1.BattleStreamReq{
		BattleId:             fmt.Sprintf("battle_%d", requestID%5),
		PlayerId:             fmt.Sprintf("player_%d", clientID),
		IncludeDetailedStats: false,
	}

	stream, err := client.StreamBattleUpdates(ctx, req)
	if err != nil {
		return false, err
	}

	// 接收几个消息后关闭
	for i := 0; i < 3; i++ {
		_, err := stream.Recv()
		if err != nil {
			break
		}
	}

	return true, nil
}

// selectTestMethod 选择测试方法
func (t *GRPCLoadTester) selectTestMethod(requestID int) string {
	if len(t.config.TestMethods) == 0 {
		return "Login" // 默认方法
	}

	return t.config.TestMethods[requestID%len(t.config.TestMethods)]
}

// recordMetrics 记录指标
func (t *GRPCLoadTester) recordMetrics(method string, latency time.Duration, success bool, err error) {
	// 基础计数
	t.metrics.totalRequests.Add(1)
	if success {
		t.metrics.successRequests.Add(1)
	} else {
		t.metrics.failedRequests.Add(1)
	}

	// 记录延迟
	t.metrics.latencyMu.Lock()
	t.metrics.latencies = append(t.metrics.latencies, latency)
	// 保持最近10000个请求的延迟数据
	if len(t.metrics.latencies) > 10000 {
		t.metrics.latencies = t.metrics.latencies[1:]
	}
	t.metrics.latencyMu.Unlock()

	// 记录错误
	if err != nil {
		errorType := "unknown"
		if err != nil {
			errorType = err.Error()
			if len(errorType) > 50 {
				errorType = errorType[:50] // 截断长错误消息
			}
		}

		t.metrics.errorMu.Lock()
		if counter, exists := t.metrics.errorCounts[errorType]; exists {
			counter.Add(1)
		} else {
			counter := &atomic.Int64{}
			counter.Add(1)
			t.metrics.errorCounts[errorType] = counter
		}
		t.metrics.errorMu.Unlock()
	}

	// 记录方法级指标
	t.metrics.methodMu.Lock()
	if methodMetric, exists := t.metrics.methodMetrics[method]; exists {
		methodMetric.requests.Add(1)
		if success {
			methodMetric.successes.Add(1)
		} else {
			methodMetric.failures.Add(1)
		}
		methodMetric.mu.Lock()
		methodMetric.latencies = append(methodMetric.latencies, latency)
		if len(methodMetric.latencies) > 1000 {
			methodMetric.latencies = methodMetric.latencies[1:]
		}
		methodMetric.mu.Unlock()
	} else {
		methodMetric := &MethodMetrics{
			latencies: []time.Duration{latency},
		}
		methodMetric.requests.Add(1)
		if success {
			methodMetric.successes.Add(1)
		} else {
			methodMetric.failures.Add(1)
		}
		t.metrics.methodMetrics[method] = methodMetric
	}
	t.metrics.methodMu.Unlock()
}

// collectMetrics 定期收集时间序列指标
func (t *GRPCLoadTester) collectMetrics() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lastRequests := int64(0)
	lastErrors := int64(0)

	for {
		select {
		case <-t.ctx.Done():
			return
		case now := <-ticker.C:
			currentRequests := t.metrics.totalRequests.Load()
			currentErrors := t.metrics.failedRequests.Load()

			// 计算当前延迟
			var currentLatency float64
			t.metrics.latencyMu.Lock()
			if len(t.metrics.latencies) > 0 {
				var total time.Duration
				for _, lat := range t.metrics.latencies {
					total += lat
				}
				currentLatency = float64(total.Nanoseconds()) / float64(len(t.metrics.latencies)) / 1e6 // ms
			}
			t.metrics.latencyMu.Unlock()

			// 记录时间序列
			t.metrics.timeSeriesMu.Lock()
			t.metrics.timeSeriesData = append(t.metrics.timeSeriesData, TimeSeries{
				Timestamp: now,
				Requests:  currentRequests - lastRequests,
				Latency:   currentLatency,
				Errors:    currentErrors - lastErrors,
			})
			// 保持最近1000个数据点
			if len(t.metrics.timeSeriesData) > 1000 {
				t.metrics.timeSeriesData = t.metrics.timeSeriesData[1:]
			}
			t.metrics.timeSeriesMu.Unlock()

			lastRequests = currentRequests
			lastErrors = currentErrors
		}
	}
}

// generateResult 生成测试结果
func (t *GRPCLoadTester) generateResult() {
	t.mu.Lock()
	defer t.mu.Unlock()

	duration := time.Since(t.metrics.startTime)
	totalRequests := t.metrics.totalRequests.Load()
	successRequests := t.metrics.successRequests.Load()
	failedRequests := t.metrics.failedRequests.Load()

	result := &GRPCLoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successRequests,
		FailedRequests:     failedRequests,
		Duration:           duration,
		RequestsPerSecond:  float64(totalRequests) / duration.Seconds(),
		ErrorsByType:       make(map[string]int64),
		MethodStats:        make(map[string]*MethodStats),
	}

	// 计算延迟统计
	t.metrics.latencyMu.Lock()
	if len(t.metrics.latencies) > 0 {
		latencies := make([]time.Duration, len(t.metrics.latencies))
		copy(latencies, t.metrics.latencies)

		// 排序计算百分位数
		for i := 0; i < len(latencies); i++ {
			for j := i + 1; j < len(latencies); j++ {
				if latencies[i] > latencies[j] {
					latencies[i], latencies[j] = latencies[j], latencies[i]
				}
			}
		}

		result.MinLatency = float64(latencies[0].Nanoseconds()) / 1e6
		result.MaxLatency = float64(latencies[len(latencies)-1].Nanoseconds()) / 1e6
		result.P50Latency = float64(latencies[len(latencies)/2].Nanoseconds()) / 1e6
		result.P95Latency = float64(latencies[int(float64(len(latencies))*0.95)].Nanoseconds()) / 1e6
		result.P99Latency = float64(latencies[int(float64(len(latencies))*0.99)].Nanoseconds()) / 1e6

		var total time.Duration
		for _, lat := range latencies {
			total += lat
		}
		result.AvgLatency = float64(total.Nanoseconds()) / float64(len(latencies)) / 1e6
	}
	t.metrics.latencyMu.Unlock()

	// 错误统计
	t.metrics.errorMu.RLock()
	for errorType, counter := range t.metrics.errorCounts {
		result.ErrorsByType[errorType] = counter.Load()
	}
	t.metrics.errorMu.RUnlock()

	// 方法统计
	t.metrics.methodMu.RLock()
	for method, metric := range t.metrics.methodMetrics {
		methodStat := &MethodStats{
			TotalRequests:   metric.requests.Load(),
			SuccessRequests: metric.successes.Load(),
			FailedRequests:  metric.failures.Load(),
		}

		metric.mu.Lock()
		if len(metric.latencies) > 0 {
			var total time.Duration
			min := metric.latencies[0]
			max := metric.latencies[0]

			for _, lat := range metric.latencies {
				total += lat
				if lat < min {
					min = lat
				}
				if lat > max {
					max = lat
				}
			}

			methodStat.AvgLatency = float64(total.Nanoseconds()) / float64(len(metric.latencies)) / 1e6
			methodStat.MinLatency = float64(min.Nanoseconds()) / 1e6
			methodStat.MaxLatency = float64(max.Nanoseconds()) / 1e6
		}
		metric.mu.Unlock()

		result.MethodStats[method] = methodStat
	}
	t.metrics.methodMu.RUnlock()

	// 时间序列数据
	t.metrics.timeSeriesMu.Lock()
	for _, ts := range t.metrics.timeSeriesData {
		result.LatencyTimeSeries = append(result.LatencyTimeSeries, TimePoint{
			Timestamp: ts.Timestamp,
			Value:     ts.Latency,
		})
		result.ThroughputTimeSeries = append(result.ThroughputTimeSeries, TimePoint{
			Timestamp: ts.Timestamp,
			Value:     float64(ts.Requests),
		})
		result.ErrorTimeSeries = append(result.ErrorTimeSeries, TimePoint{
			Timestamp: ts.Timestamp,
			Value:     float64(ts.Errors),
		})
	}
	t.metrics.timeSeriesMu.Unlock()

	t.result = result
}

// GetResult 获取测试结果
func (t *GRPCLoadTester) GetResult() *GRPCLoadTestResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result
}

// Stop 停止测试
func (t *GRPCLoadTester) Stop() {
	t.cancel()
}

// cleanup 清理资源
func (t *GRPCLoadTester) cleanup() {
	for _, conn := range t.connections {
		if conn != nil {
			conn.Close()
		}
	}
}

// DefaultGRPCLoadTestConfig 返回默认的gRPC负载测试配置
func DefaultGRPCLoadTestConfig(serverAddr string) *GRPCLoadTestConfig {
	return &GRPCLoadTestConfig{
		ServerAddr:        serverAddr,
		TLS:               false,
		ConcurrentClients: 10,
		Duration:          time.Minute,
		TargetRPS:         100,
		TestMethods:       []string{"Login", "SendPlayerAction", "GetPlayerStatus"},
		RequestTimeout:    time.Second * 10,
		KeepAliveTime:     time.Second * 30,
		KeepAliveTimeout:  time.Second * 5,
		MaxConnections:    5,
		ConnectionReuse:   true,
	}
}
