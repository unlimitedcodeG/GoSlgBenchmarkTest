package loadtest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// HTTPLoadTestConfig HTTP负载测试配置
type HTTPLoadTestConfig struct {
	BaseURL           string
	Endpoints         []EndpointConfig
	ConcurrentClients int
	Duration          time.Duration
	TargetRPS         int // 目标每秒请求数

	// HTTP配置
	Timeout         time.Duration
	KeepAlive       bool
	MaxIdleConns    int
	MaxConnsPerHost int
	TLSConfig       *tls.Config

	// 认证配置
	AuthType     string // "none", "bearer", "basic"
	AuthToken    string
	AuthUser     string
	AuthPassword string
	Headers      map[string]string

	// 请求配置
	FollowRedirects   bool
	EnableCompression bool
}

// EndpointConfig 端点配置
type EndpointConfig struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	Weight      int               `json:"weight"` // 权重，用于分配请求比例
	QueryParams map[string]string `json:"query_params"`
}

// HTTPLoadTestResult HTTP负载测试结果
type HTTPLoadTestResult struct {
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
	BytesRead         int64
	BytesWritten      int64

	// HTTP状态码统计
	StatusCodes map[int]int64

	// 错误统计
	ErrorsByType map[string]int64

	// 端点级统计
	EndpointStats map[string]*EndpointStats

	// 时间序列数据
	LatencyTimeSeries    []TimePoint
	ThroughputTimeSeries []TimePoint
	ErrorTimeSeries      []TimePoint
}

type EndpointStats struct {
	Path            string
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	AvgLatency      float64
	MinLatency      float64
	MaxLatency      float64
	StatusCodes     map[int]int64
	BytesRead       int64
	BytesWritten    int64
}

// HTTPLoadTester HTTP负载测试器
type HTTPLoadTester struct {
	config *HTTPLoadTestConfig
	client *http.Client

	// 统计数据
	metrics *HTTPTestMetrics

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 结果
	result *HTTPLoadTestResult
	mu     sync.RWMutex
}

// HTTPTestMetrics HTTP测试指标收集器
type HTTPTestMetrics struct {
	totalRequests   atomic.Int64
	successRequests atomic.Int64
	failedRequests  atomic.Int64
	bytesRead       atomic.Int64
	bytesWritten    atomic.Int64

	// 延迟统计
	latencies []time.Duration
	latencyMu sync.Mutex

	// 状态码统计
	statusCodes map[int]*atomic.Int64
	statusMu    sync.RWMutex

	// 错误统计
	errorCounts map[string]*atomic.Int64
	errorMu     sync.RWMutex

	// 端点统计
	endpointMetrics map[string]*EndpointMetrics
	endpointMu      sync.RWMutex

	// 时间序列
	timeSeriesData []HTTPTimeSeries
	timeSeriesMu   sync.Mutex

	startTime time.Time
}

type EndpointMetrics struct {
	requests     atomic.Int64
	successes    atomic.Int64
	failures     atomic.Int64
	bytesRead    atomic.Int64
	bytesWritten atomic.Int64
	latencies    []time.Duration
	statusCodes  map[int]*atomic.Int64
	mu           sync.Mutex
}

type HTTPTimeSeries struct {
	Timestamp   time.Time
	Requests    int64
	Latency     float64
	Errors      int64
	BytesPerSec float64
}

// NewHTTPLoadTester 创建新的HTTP负载测试器
func NewHTTPLoadTester(config *HTTPLoadTestConfig) *HTTPLoadTester {
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+time.Minute)

	// 创建HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:       config.MaxIdleConns,
		MaxConnsPerHost:    config.MaxConnsPerHost,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: !config.EnableCompression,
		TLSClientConfig:    config.TLSConfig,
	}

	if !config.KeepAlive {
		transport.DisableKeepAlives = true
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	if !config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &HTTPLoadTester{
		config: config,
		client: client,
		ctx:    ctx,
		cancel: cancel,
		metrics: &HTTPTestMetrics{
			statusCodes:     make(map[int]*atomic.Int64),
			errorCounts:     make(map[string]*atomic.Int64),
			endpointMetrics: make(map[string]*EndpointMetrics),
			startTime:       time.Now(),
		},
	}
}

// Start 开始负载测试
func (t *HTTPLoadTester) Start() error {
	log.Printf("Starting HTTP load test: %d clients, %v duration, target %d RPS",
		t.config.ConcurrentClients, t.config.Duration, t.config.TargetRPS)

	// 验证配置
	if err := t.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	// 启动指标收集
	go t.collectMetrics()

	// 启动负载生成器
	return t.startLoadGeneration()
}

// validateConfig 验证配置
func (t *HTTPLoadTester) validateConfig() error {
	if len(t.config.Endpoints) == 0 {
		return fmt.Errorf("no endpoints configured")
	}

	for i, ep := range t.config.Endpoints {
		if ep.Path == "" {
			return fmt.Errorf("endpoint %d: path is required", i)
		}
		if ep.Method == "" {
			t.config.Endpoints[i].Method = "GET"
		}
		if ep.Weight <= 0 {
			t.config.Endpoints[i].Weight = 1
		}
	}

	return nil
}

// startLoadGeneration 开始负载生成
func (t *HTTPLoadTester) startLoadGeneration() error {
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
		log.Printf("HTTP load test completed")
	case <-time.After(t.config.Duration + time.Minute):
		log.Printf("HTTP load test timeout")
		t.cancel()
	}

	// 生成结果
	t.generateResult()

	return nil
}

// clientWorker 客户端工作器
func (t *HTTPLoadTester) clientWorker(clientID int, targetRPS float64) {
	defer t.wg.Done()

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
			// 选择端点并执行请求
			endpoint := t.selectEndpoint(requestID)
			t.executeRequest(endpoint, clientID, requestID)
			requestID++
		}
	}
}

// selectEndpoint 根据权重选择端点
func (t *HTTPLoadTester) selectEndpoint(requestID int) EndpointConfig {
	if len(t.config.Endpoints) == 1 {
		return t.config.Endpoints[0]
	}

	// 计算总权重
	totalWeight := 0
	for _, ep := range t.config.Endpoints {
		totalWeight += ep.Weight
	}

	// 根据权重选择
	target := requestID % totalWeight
	current := 0

	for _, ep := range t.config.Endpoints {
		current += ep.Weight
		if target < current {
			return ep
		}
	}

	// 默认返回第一个
	return t.config.Endpoints[0]
}

// executeRequest 执行HTTP请求
func (t *HTTPLoadTester) executeRequest(endpoint EndpointConfig, clientID, requestID int) {
	start := time.Now()

	// 构建URL
	fullURL := t.config.BaseURL + endpoint.Path
	if len(endpoint.QueryParams) > 0 {
		u, _ := url.Parse(fullURL)
		q := u.Query()
		for k, v := range endpoint.QueryParams {
			// 支持模板变量
			value := t.replaceVariables(v, clientID, requestID)
			q.Set(k, value)
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	// 准备请求体
	var body io.Reader
	var bodyBytes []byte
	if endpoint.Body != nil {
		switch v := endpoint.Body.(type) {
		case string:
			bodyContent := t.replaceVariables(v, clientID, requestID)
			bodyBytes = []byte(bodyContent)
		case map[string]interface{}:
			// 替换JSON中的变量
			bodyContent := t.replaceJSONVariables(v, clientID, requestID)
			bodyBytes, _ = json.Marshal(bodyContent)
		default:
			bodyBytes, _ = json.Marshal(v)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(t.ctx, endpoint.Method, fullURL, body)
	if err != nil {
		t.recordError(endpoint.Path, fmt.Errorf("create request failed: %v", err))
		return
	}

	// 设置头部
	t.setRequestHeaders(req, endpoint.Headers)

	// 记录发送字节数
	if bodyBytes != nil {
		t.metrics.bytesWritten.Add(int64(len(bodyBytes)))
	}

	// 执行请求
	resp, err := t.client.Do(req)
	if err != nil {
		t.recordError(endpoint.Path, err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.recordError(endpoint.Path, fmt.Errorf("read response failed: %v", err))
		return
	}

	latency := time.Since(start)

	// 记录指标
	t.recordMetrics(endpoint.Path, resp.StatusCode, latency, len(respBody), len(bodyBytes))
}

// setRequestHeaders 设置请求头
func (t *HTTPLoadTester) setRequestHeaders(req *http.Request, endpointHeaders map[string]string) {
	// 设置全局头部
	for k, v := range t.config.Headers {
		req.Header.Set(k, v)
	}

	// 设置端点特定头部
	for k, v := range endpointHeaders {
		req.Header.Set(k, v)
	}

	// 设置认证
	switch t.config.AuthType {
	case "bearer":
		if t.config.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+t.config.AuthToken)
		}
	case "basic":
		if t.config.AuthUser != "" && t.config.AuthPassword != "" {
			req.SetBasicAuth(t.config.AuthUser, t.config.AuthPassword)
		}
	}

	// 设置默认Content-Type
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
}

// replaceVariables 替换模板变量
func (t *HTTPLoadTester) replaceVariables(template string, clientID, requestID int) string {
	result := template

	// 支持的变量
	variables := map[string]string{
		"{{client_id}}":    fmt.Sprintf("%d", clientID),
		"{{request_id}}":   fmt.Sprintf("%d", requestID),
		"{{timestamp}}":    fmt.Sprintf("%d", time.Now().Unix()),
		"{{timestamp_ms}}": fmt.Sprintf("%d", time.Now().UnixMilli()),
		"{{random_id}}":    fmt.Sprintf("%d", time.Now().UnixNano()%10000),
	}

	for placeholder, value := range variables {
		result = bytes.NewBuffer(bytes.ReplaceAll([]byte(result), []byte(placeholder), []byte(value))).String()
	}

	return result
}

// replaceJSONVariables 替换JSON中的变量
func (t *HTTPLoadTester) replaceJSONVariables(data map[string]interface{}, clientID, requestID int) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range data {
		switch val := v.(type) {
		case string:
			result[k] = t.replaceVariables(val, clientID, requestID)
		case map[string]interface{}:
			result[k] = t.replaceJSONVariables(val, clientID, requestID)
		default:
			result[k] = v
		}
	}

	return result
}

// recordMetrics 记录指标
func (t *HTTPLoadTester) recordMetrics(endpoint string, statusCode int, latency time.Duration, respSize, reqSize int) {
	// 基础计数
	t.metrics.totalRequests.Add(1)
	if statusCode >= 200 && statusCode < 400 {
		t.metrics.successRequests.Add(1)
	} else {
		t.metrics.failedRequests.Add(1)
	}

	// 字节计数
	t.metrics.bytesRead.Add(int64(respSize))
	if reqSize > 0 {
		t.metrics.bytesWritten.Add(int64(reqSize))
	}

	// 记录延迟
	t.metrics.latencyMu.Lock()
	t.metrics.latencies = append(t.metrics.latencies, latency)
	if len(t.metrics.latencies) > 10000 {
		t.metrics.latencies = t.metrics.latencies[1:]
	}
	t.metrics.latencyMu.Unlock()

	// 记录状态码
	t.metrics.statusMu.Lock()
	if counter, exists := t.metrics.statusCodes[statusCode]; exists {
		counter.Add(1)
	} else {
		counter := &atomic.Int64{}
		counter.Add(1)
		t.metrics.statusCodes[statusCode] = counter
	}
	t.metrics.statusMu.Unlock()

	// 记录端点级指标
	t.metrics.endpointMu.Lock()
	if endpointMetric, exists := t.metrics.endpointMetrics[endpoint]; exists {
		endpointMetric.requests.Add(1)
		if statusCode >= 200 && statusCode < 400 {
			endpointMetric.successes.Add(1)
		} else {
			endpointMetric.failures.Add(1)
		}
		endpointMetric.bytesRead.Add(int64(respSize))
		endpointMetric.bytesWritten.Add(int64(reqSize))

		endpointMetric.mu.Lock()
		endpointMetric.latencies = append(endpointMetric.latencies, latency)
		if len(endpointMetric.latencies) > 1000 {
			endpointMetric.latencies = endpointMetric.latencies[1:]
		}

		if counter, exists := endpointMetric.statusCodes[statusCode]; exists {
			counter.Add(1)
		} else {
			counter := &atomic.Int64{}
			counter.Add(1)
			endpointMetric.statusCodes[statusCode] = counter
		}
		endpointMetric.mu.Unlock()
	} else {
		endpointMetric := &EndpointMetrics{
			latencies:   []time.Duration{latency},
			statusCodes: make(map[int]*atomic.Int64),
		}
		endpointMetric.requests.Add(1)
		if statusCode >= 200 && statusCode < 400 {
			endpointMetric.successes.Add(1)
		} else {
			endpointMetric.failures.Add(1)
		}
		endpointMetric.bytesRead.Add(int64(respSize))
		endpointMetric.bytesWritten.Add(int64(reqSize))

		counter := &atomic.Int64{}
		counter.Add(1)
		endpointMetric.statusCodes[statusCode] = counter

		t.metrics.endpointMetrics[endpoint] = endpointMetric
	}
	t.metrics.endpointMu.Unlock()
}

// recordError 记录错误
func (t *HTTPLoadTester) recordError(endpoint string, err error) {
	t.metrics.totalRequests.Add(1)
	t.metrics.failedRequests.Add(1)

	errorType := err.Error()
	if len(errorType) > 50 {
		errorType = errorType[:50]
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

	// 端点错误计数
	t.metrics.endpointMu.Lock()
	if endpointMetric, exists := t.metrics.endpointMetrics[endpoint]; exists {
		endpointMetric.requests.Add(1)
		endpointMetric.failures.Add(1)
	} else {
		endpointMetric := &EndpointMetrics{
			latencies:   []time.Duration{},
			statusCodes: make(map[int]*atomic.Int64),
		}
		endpointMetric.requests.Add(1)
		endpointMetric.failures.Add(1)
		t.metrics.endpointMetrics[endpoint] = endpointMetric
	}
	t.metrics.endpointMu.Unlock()
}

// collectMetrics 定期收集时间序列指标
func (t *HTTPLoadTester) collectMetrics() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lastRequests := int64(0)
	lastErrors := int64(0)
	lastBytes := int64(0)

	for {
		select {
		case <-t.ctx.Done():
			return
		case now := <-ticker.C:
			currentRequests := t.metrics.totalRequests.Load()
			currentErrors := t.metrics.failedRequests.Load()
			currentBytes := t.metrics.bytesRead.Load()

			// 计算当前延迟
			var currentLatency float64
			t.metrics.latencyMu.Lock()
			if len(t.metrics.latencies) > 0 {
				var total time.Duration
				for _, lat := range t.metrics.latencies {
					total += lat
				}
				currentLatency = float64(total.Nanoseconds()) / float64(len(t.metrics.latencies)) / 1e6
			}
			t.metrics.latencyMu.Unlock()

			// 记录时间序列
			t.metrics.timeSeriesMu.Lock()
			t.metrics.timeSeriesData = append(t.metrics.timeSeriesData, HTTPTimeSeries{
				Timestamp:   now,
				Requests:    currentRequests - lastRequests,
				Latency:     currentLatency,
				Errors:      currentErrors - lastErrors,
				BytesPerSec: float64(currentBytes - lastBytes),
			})
			if len(t.metrics.timeSeriesData) > 1000 {
				t.metrics.timeSeriesData = t.metrics.timeSeriesData[1:]
			}
			t.metrics.timeSeriesMu.Unlock()

			lastRequests = currentRequests
			lastErrors = currentErrors
			lastBytes = currentBytes
		}
	}
}

// generateResult 生成测试结果
func (t *HTTPLoadTester) generateResult() {
	t.mu.Lock()
	defer t.mu.Unlock()

	duration := time.Since(t.metrics.startTime)
	totalRequests := t.metrics.totalRequests.Load()
	successRequests := t.metrics.successRequests.Load()
	failedRequests := t.metrics.failedRequests.Load()
	bytesRead := t.metrics.bytesRead.Load()
	bytesWritten := t.metrics.bytesWritten.Load()

	result := &HTTPLoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successRequests,
		FailedRequests:     failedRequests,
		Duration:           duration,
		RequestsPerSecond:  float64(totalRequests) / duration.Seconds(),
		BytesPerSecond:     float64(bytesRead) / duration.Seconds(),
		BytesRead:          bytesRead,
		BytesWritten:       bytesWritten,
		StatusCodes:        make(map[int]int64),
		ErrorsByType:       make(map[string]int64),
		EndpointStats:      make(map[string]*EndpointStats),
	}

	// 计算延迟统计
	t.metrics.latencyMu.Lock()
	if len(t.metrics.latencies) > 0 {
		latencies := make([]time.Duration, len(t.metrics.latencies))
		copy(latencies, t.metrics.latencies)

		// 排序计算百分位数
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

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

	// 状态码统计
	t.metrics.statusMu.RLock()
	for code, counter := range t.metrics.statusCodes {
		result.StatusCodes[code] = counter.Load()
	}
	t.metrics.statusMu.RUnlock()

	// 错误统计
	t.metrics.errorMu.RLock()
	for errorType, counter := range t.metrics.errorCounts {
		result.ErrorsByType[errorType] = counter.Load()
	}
	t.metrics.errorMu.RUnlock()

	// 端点统计
	t.metrics.endpointMu.RLock()
	for endpoint, metric := range t.metrics.endpointMetrics {
		endpointStat := &EndpointStats{
			Path:            endpoint,
			TotalRequests:   metric.requests.Load(),
			SuccessRequests: metric.successes.Load(),
			FailedRequests:  metric.failures.Load(),
			BytesRead:       metric.bytesRead.Load(),
			BytesWritten:    metric.bytesWritten.Load(),
			StatusCodes:     make(map[int]int64),
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

			endpointStat.AvgLatency = float64(total.Nanoseconds()) / float64(len(metric.latencies)) / 1e6
			endpointStat.MinLatency = float64(min.Nanoseconds()) / 1e6
			endpointStat.MaxLatency = float64(max.Nanoseconds()) / 1e6
		}

		for code, counter := range metric.statusCodes {
			endpointStat.StatusCodes[code] = counter.Load()
		}
		metric.mu.Unlock()

		result.EndpointStats[endpoint] = endpointStat
	}
	t.metrics.endpointMu.RUnlock()

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
func (t *HTTPLoadTester) GetResult() *HTTPLoadTestResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result
}

// Stop 停止测试
func (t *HTTPLoadTester) Stop() {
	t.cancel()
}

// DefaultHTTPLoadTestConfig 返回默认的HTTP负载测试配置
func DefaultHTTPLoadTestConfig(baseURL string) *HTTPLoadTestConfig {
	return &HTTPLoadTestConfig{
		BaseURL:           baseURL,
		ConcurrentClients: 10,
		Duration:          time.Minute,
		TargetRPS:         100,
		Timeout:           time.Second * 30,
		KeepAlive:         true,
		MaxIdleConns:      100,
		MaxConnsPerHost:   50,
		AuthType:          "none",
		Headers:           make(map[string]string),
		FollowRedirects:   true,
		EnableCompression: true,
		Endpoints: []EndpointConfig{
			{
				Path:   "/api/v1/health",
				Method: "GET",
				Weight: 1,
			},
		},
	}
}
