package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"GoSlgBenchmarkTest/internal/grpcserver"
	"GoSlgBenchmarkTest/internal/httpserver"
	"GoSlgBenchmarkTest/internal/loadtest"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// LoadTestHandler 负载测试处理器
type LoadTestHandler struct {
	// 测试实例管理
	httpTests map[string]*HTTPTestInstance
	grpcTests map[string]*GRPCTestInstance
	wsTests   map[string]*WSTestInstance
	testMutex sync.RWMutex

	// 服务器实例
	httpServer *httpserver.APIServer
	grpcServer *grpcserver.GameServer
	grpcConn   *grpc.Server

	// 配置
	testServerPorts map[string]int
	portMutex       sync.Mutex
	nextPort        int
}

// TestInstance 测试实例基础接口
type TestInstance interface {
	GetID() string
	GetType() string
	GetStatus() string
	GetResult() interface{}
	Stop() error
}

// HTTPTestInstance HTTP测试实例
type HTTPTestInstance struct {
	ID        string
	Config    *loadtest.HTTPLoadTestConfig
	Tester    *loadtest.HTTPLoadTester
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Result    *loadtest.HTTPLoadTestResult
	mu        sync.RWMutex
}

// GRPCTestInstance gRPC测试实例
type GRPCTestInstance struct {
	ID        string
	Config    *loadtest.GRPCLoadTestConfig
	Tester    *loadtest.GRPCLoadTester
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Result    *loadtest.GRPCLoadTestResult
	mu        sync.RWMutex
}

// WSTestInstance WebSocket测试实例
type WSTestInstance struct {
	ID        string
	Config    interface{} // TODO: WebSocket配置
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Result    interface{} // TODO: WebSocket结果
	mu        sync.RWMutex
}

// TestRequest 测试请求结构
type TestRequest struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "http", "grpc", "websocket"
	Config   interface{} `json:"config"`
	Duration int         `json:"duration"` // 秒
}

// LoadTestResponse 负载测试响应结构
type LoadTestResponse struct {
	Success   bool        `json:"success"`
	TestID    string      `json:"test_id,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewLoadTestHandler 创建负载测试处理器
func NewLoadTestHandler() *LoadTestHandler {
	// 从环境变量获取起始端口
	startPort := getEnvInt("TEST_HTTP_SERVER_PORT", 19000)

	return &LoadTestHandler{
		httpTests:       make(map[string]*HTTPTestInstance),
		grpcTests:       make(map[string]*GRPCTestInstance),
		wsTests:         make(map[string]*WSTestInstance),
		testServerPorts: make(map[string]int),
		nextPort:        startPort, // 从配置的端口开始分配
	}
}

// StartTestServers 启动测试服务器
func (h *LoadTestHandler) StartTestServers() error {
	// 启动HTTP测试服务器
	httpPort := h.allocatePort("http")
	httpPort = h.findAvailablePort(httpPort, 10) // 最多尝试10次
	if httpPort == 0 {
		return fmt.Errorf("failed to find available port for HTTP server")
	}

	log.Printf("Starting HTTP test server on port %d", httpPort)
	h.httpServer = httpserver.NewAPIServer(fmt.Sprintf(":%d", httpPort))
	go func() {
		if err := h.httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP test server error: %v", err)
		}
	}()

	// 启动gRPC测试服务器
	grpcPort := h.allocatePort("grpc")
	grpcPort = h.findAvailablePort(grpcPort, 10) // 最多尝试10次
	if grpcPort == 0 {
		return fmt.Errorf("failed to find available port for gRPC server")
	}

	log.Printf("Starting gRPC test server on port %d", grpcPort)
	h.grpcServer = grpcserver.NewGameServer()
	h.grpcConn = grpc.NewServer()
	gamev1.RegisterGameServiceServer(h.grpcConn, h.grpcServer)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			log.Printf("gRPC server listen error: %v", err)
			return
		}
		if err := h.grpcConn.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	log.Printf("Test servers started - HTTP: %d, gRPC: %d", httpPort, grpcPort)
	return nil
}

// allocatePort 分配端口
func (h *LoadTestHandler) allocatePort(serverType string) int {
	h.portMutex.Lock()
	defer h.portMutex.Unlock()

	port := h.nextPort
	h.testServerPorts[serverType] = port
	h.nextPort++
	return port
}

// findAvailablePort 查找可用的端口
func (h *LoadTestHandler) findAvailablePort(startPort int, maxAttempts int) int {
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if h.isPortAvailable(port) {
			return port
		}
		log.Printf("Port %d is in use, trying next port...", port)
	}
	return 0 // 没有找到可用端口
}

// isPortAvailable 检查端口是否可用
func (h *LoadTestHandler) isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// getEnvInt 从环境变量获取整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default %d", key, value, defaultValue)
	}
	return defaultValue
}

// CreateTest 创建测试
func (h *LoadTestHandler) CreateTest(w http.ResponseWriter, r *http.Request) {
	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证请求
	if req.Name == "" || req.Type == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Name and type are required")
		return
	}

	testID := fmt.Sprintf("%s_%d", req.Type, time.Now().UnixNano())

	switch req.Type {
	case "http":
		err := h.createHTTPTest(testID, req)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	case "grpc":
		err := h.createGRPCTest(testID, req)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	case "websocket":
		err := h.createWSTest(testID, req)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	default:
		h.writeErrorResponse(w, http.StatusBadRequest, "Unsupported test type")
		return
	}

	h.writeSuccessResponse(w, map[string]interface{}{
		"test_id": testID,
		"status":  "created",
		"type":    req.Type,
		"name":    req.Name,
	})
}

// createHTTPTest 创建HTTP测试
func (h *LoadTestHandler) createHTTPTest(testID string, req TestRequest) error {
	// 解析HTTP配置
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	var config loadtest.HTTPLoadTestConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		// 如果解析失败，使用默认配置
		httpPort := h.testServerPorts["http"]
		baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
		config = *loadtest.DefaultHTTPLoadTestConfig(baseURL)
	}

	// 设置测试持续时间
	if req.Duration > 0 {
		config.Duration = time.Duration(req.Duration) * time.Second
	}

	// 如果没有指定BaseURL，使用测试服务器
	if config.BaseURL == "" {
		httpPort := h.testServerPorts["http"]
		config.BaseURL = fmt.Sprintf("http://localhost:%d", httpPort)
	}

	instance := &HTTPTestInstance{
		ID:     testID,
		Config: &config,
		Status: "created",
	}

	h.testMutex.Lock()
	h.httpTests[testID] = instance
	h.testMutex.Unlock()

	return nil
}

// createGRPCTest 创建gRPC测试
func (h *LoadTestHandler) createGRPCTest(testID string, req TestRequest) error {
	// 解析gRPC配置
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	var config loadtest.GRPCLoadTestConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		// 如果解析失败，使用默认配置
		grpcPort := h.testServerPorts["grpc"]
		serverAddr := fmt.Sprintf("localhost:%d", grpcPort)
		config = *loadtest.DefaultGRPCLoadTestConfig(serverAddr)
	}

	// 设置测试持续时间
	if req.Duration > 0 {
		config.Duration = time.Duration(req.Duration) * time.Second
	}

	// 如果没有指定ServerAddr，使用测试服务器
	if config.ServerAddr == "" {
		grpcPort := h.testServerPorts["grpc"]
		config.ServerAddr = fmt.Sprintf("localhost:%d", grpcPort)
	}

	instance := &GRPCTestInstance{
		ID:     testID,
		Config: &config,
		Status: "created",
	}

	h.testMutex.Lock()
	h.grpcTests[testID] = instance
	h.testMutex.Unlock()

	return nil
}

// createWSTest 创建WebSocket测试
func (h *LoadTestHandler) createWSTest(testID string, req TestRequest) error {
	// TODO: 实现WebSocket测试创建
	instance := &WSTestInstance{
		ID:     testID,
		Status: "created",
	}

	h.testMutex.Lock()
	h.wsTests[testID] = instance
	h.testMutex.Unlock()

	return nil
}

// StartTest 启动测试
func (h *LoadTestHandler) StartTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["id"]

	if testID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Test ID is required")
		return
	}

	// 查找并启动测试
	if err := h.startTestByID(testID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, map[string]interface{}{
		"test_id":    testID,
		"status":     "running",
		"started_at": time.Now().UnixMilli(),
	})
}

// startTestByID 根据ID启动测试
func (h *LoadTestHandler) startTestByID(testID string) error {
	h.testMutex.Lock()
	defer h.testMutex.Unlock()

	// 尝试HTTP测试
	if httpTest, exists := h.httpTests[testID]; exists {
		if httpTest.Status != "created" {
			return fmt.Errorf("test is not in created state: %s", httpTest.Status)
		}

		httpTest.Tester = loadtest.NewHTTPLoadTester(httpTest.Config)
		httpTest.Status = "running"
		httpTest.StartTime = time.Now()

		go func() {
			if err := httpTest.Tester.Start(); err != nil {
				log.Printf("HTTP test %s failed: %v", testID, err)
			}

			httpTest.mu.Lock()
			httpTest.Status = "completed"
			now := time.Now()
			httpTest.EndTime = &now
			httpTest.Result = httpTest.Tester.GetResult()
			httpTest.mu.Unlock()
		}()

		return nil
	}

	// 尝试gRPC测试
	if grpcTest, exists := h.grpcTests[testID]; exists {
		if grpcTest.Status != "created" {
			return fmt.Errorf("test is not in created state: %s", grpcTest.Status)
		}

		grpcTest.Tester = loadtest.NewGRPCLoadTester(grpcTest.Config)
		grpcTest.Status = "running"
		grpcTest.StartTime = time.Now()

		go func() {
			if err := grpcTest.Tester.Start(); err != nil {
				log.Printf("gRPC test %s failed: %v", testID, err)
			}

			grpcTest.mu.Lock()
			grpcTest.Status = "completed"
			now := time.Now()
			grpcTest.EndTime = &now
			grpcTest.Result = grpcTest.Tester.GetResult()
			grpcTest.mu.Unlock()
		}()

		return nil
	}

	// 尝试WebSocket测试
	if wsTest, exists := h.wsTests[testID]; exists {
		if wsTest.Status != "created" {
			return fmt.Errorf("test is not in created state: %s", wsTest.Status)
		}

		wsTest.Status = "running"
		wsTest.StartTime = time.Now()

		// TODO: 启动WebSocket测试

		return nil
	}

	return fmt.Errorf("test not found: %s", testID)
}

// StopTest 停止测试
func (h *LoadTestHandler) StopTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["id"]

	if testID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Test ID is required")
		return
	}

	if err := h.stopTestByID(testID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, map[string]interface{}{
		"test_id":    testID,
		"status":     "stopped",
		"stopped_at": time.Now().UnixMilli(),
	})
}

// stopTestByID 根据ID停止测试
func (h *LoadTestHandler) stopTestByID(testID string) error {
	h.testMutex.Lock()
	defer h.testMutex.Unlock()

	// 尝试HTTP测试
	if httpTest, exists := h.httpTests[testID]; exists {
		if httpTest.Tester != nil {
			httpTest.Tester.Stop()
		}
		httpTest.Status = "stopped"
		now := time.Now()
		httpTest.EndTime = &now
		return nil
	}

	// 尝试gRPC测试
	if grpcTest, exists := h.grpcTests[testID]; exists {
		if grpcTest.Tester != nil {
			grpcTest.Tester.Stop()
		}
		grpcTest.Status = "stopped"
		now := time.Now()
		grpcTest.EndTime = &now
		return nil
	}

	// 尝试WebSocket测试
	if wsTest, exists := h.wsTests[testID]; exists {
		wsTest.Status = "stopped"
		now := time.Now()
		wsTest.EndTime = &now
		return nil
	}

	return fmt.Errorf("test not found: %s", testID)
}

// GetTestStatus 获取测试状态
func (h *LoadTestHandler) GetTestStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["id"]

	if testID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Test ID is required")
		return
	}

	status, err := h.getTestStatusByID(testID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, status)
}

// getTestStatusByID 根据ID获取测试状态
func (h *LoadTestHandler) getTestStatusByID(testID string) (interface{}, error) {
	h.testMutex.RLock()
	defer h.testMutex.RUnlock()

	// 尝试HTTP测试
	if httpTest, exists := h.httpTests[testID]; exists {
		httpTest.mu.RLock()
		defer httpTest.mu.RUnlock()

		status := map[string]interface{}{
			"test_id":    testID,
			"type":       "http",
			"status":     httpTest.Status,
			"start_time": httpTest.StartTime.UnixMilli(),
		}

		if httpTest.EndTime != nil {
			status["end_time"] = httpTest.EndTime.UnixMilli()
			status["duration"] = httpTest.EndTime.Sub(httpTest.StartTime).Milliseconds()
		}

		if httpTest.Result != nil {
			status["metrics"] = httpTest.Result
		}

		return status, nil
	}

	// 尝试gRPC测试
	if grpcTest, exists := h.grpcTests[testID]; exists {
		grpcTest.mu.RLock()
		defer grpcTest.mu.RUnlock()

		status := map[string]interface{}{
			"test_id":    testID,
			"type":       "grpc",
			"status":     grpcTest.Status,
			"start_time": grpcTest.StartTime.UnixMilli(),
		}

		if grpcTest.EndTime != nil {
			status["end_time"] = grpcTest.EndTime.UnixMilli()
			status["duration"] = grpcTest.EndTime.Sub(grpcTest.StartTime).Milliseconds()
		}

		if grpcTest.Result != nil {
			status["metrics"] = grpcTest.Result
		}

		return status, nil
	}

	// 尝试WebSocket测试
	if wsTest, exists := h.wsTests[testID]; exists {
		wsTest.mu.RLock()
		defer wsTest.mu.RUnlock()

		status := map[string]interface{}{
			"test_id":    testID,
			"type":       "websocket",
			"status":     wsTest.Status,
			"start_time": wsTest.StartTime.UnixMilli(),
		}

		if wsTest.EndTime != nil {
			status["end_time"] = wsTest.EndTime.UnixMilli()
			status["duration"] = wsTest.EndTime.Sub(wsTest.StartTime).Milliseconds()
		}

		if wsTest.Result != nil {
			status["metrics"] = wsTest.Result
		}

		return status, nil
	}

	return nil, fmt.Errorf("test not found: %s", testID)
}

// ListTests 列出所有测试
func (h *LoadTestHandler) ListTests(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	testType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tests := h.collectAllTests(testType, status)

	// 分页
	total := len(tests)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var pagedTests []interface{}
	if start < total {
		pagedTests = tests[start:end]
	}

	response := map[string]interface{}{
		"tests": pagedTests,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	}

	h.writeSuccessResponse(w, response)
}

// collectAllTests 收集所有测试
func (h *LoadTestHandler) collectAllTests(testType, status string) []interface{} {
	h.testMutex.RLock()
	defer h.testMutex.RUnlock()

	var tests []interface{}

	// 收集HTTP测试
	if testType == "" || testType == "http" {
		for _, test := range h.httpTests {
			if status == "" || test.Status == status {
				testInfo := map[string]interface{}{
					"test_id":    test.ID,
					"type":       "http",
					"status":     test.Status,
					"start_time": test.StartTime.UnixMilli(),
				}
				if test.EndTime != nil {
					testInfo["end_time"] = test.EndTime.UnixMilli()
					testInfo["duration"] = test.EndTime.Sub(test.StartTime).Milliseconds()
				}
				tests = append(tests, testInfo)
			}
		}
	}

	// 收集gRPC测试
	if testType == "" || testType == "grpc" {
		for _, test := range h.grpcTests {
			if status == "" || test.Status == status {
				testInfo := map[string]interface{}{
					"test_id":    test.ID,
					"type":       "grpc",
					"status":     test.Status,
					"start_time": test.StartTime.UnixMilli(),
				}
				if test.EndTime != nil {
					testInfo["end_time"] = test.EndTime.UnixMilli()
					testInfo["duration"] = test.EndTime.Sub(test.StartTime).Milliseconds()
				}
				tests = append(tests, testInfo)
			}
		}
	}

	// 收集WebSocket测试
	if testType == "" || testType == "websocket" {
		for _, test := range h.wsTests {
			if status == "" || test.Status == status {
				testInfo := map[string]interface{}{
					"test_id":    test.ID,
					"type":       "websocket",
					"status":     test.Status,
					"start_time": test.StartTime.UnixMilli(),
				}
				if test.EndTime != nil {
					testInfo["end_time"] = test.EndTime.UnixMilli()
					testInfo["duration"] = test.EndTime.Sub(test.StartTime).Milliseconds()
				}
				tests = append(tests, testInfo)
			}
		}
	}

	return tests
}

// DeleteTest 删除测试
func (h *LoadTestHandler) DeleteTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["id"]

	if testID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Test ID is required")
		return
	}

	if err := h.deleteTestByID(testID); err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, map[string]interface{}{
		"test_id": testID,
		"message": "Test deleted successfully",
	})
}

// deleteTestByID 根据ID删除测试
func (h *LoadTestHandler) deleteTestByID(testID string) error {
	h.testMutex.Lock()
	defer h.testMutex.Unlock()

	// 尝试删除HTTP测试
	if _, exists := h.httpTests[testID]; exists {
		delete(h.httpTests, testID)
		return nil
	}

	// 尝试删除gRPC测试
	if _, exists := h.grpcTests[testID]; exists {
		delete(h.grpcTests, testID)
		return nil
	}

	// 尝试删除WebSocket测试
	if _, exists := h.wsTests[testID]; exists {
		delete(h.wsTests, testID)
		return nil
	}

	return fmt.Errorf("test not found: %s", testID)
}

// GetTestServersInfo 获取测试服务器信息
func (h *LoadTestHandler) GetTestServersInfo(w http.ResponseWriter, r *http.Request) {
	h.portMutex.Lock()
	serverInfo := make(map[string]interface{})
	for serverType, port := range h.testServerPorts {
		serverInfo[serverType] = map[string]interface{}{
			"port": port,
			"url":  fmt.Sprintf("localhost:%d", port),
		}
	}
	h.portMutex.Unlock()

	h.writeSuccessResponse(w, map[string]interface{}{
		"servers": serverInfo,
	})
}

// 辅助方法
func (h *LoadTestHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	response := LoadTestResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *LoadTestHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := LoadTestResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
	h.writeJSONResponse(w, statusCode, response)
}

func (h *LoadTestHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Stop 停止处理器并清理资源
func (h *LoadTestHandler) Stop() error {
	// 停止所有运行中的测试
	h.testMutex.Lock()
	for _, test := range h.httpTests {
		if test.Tester != nil {
			test.Tester.Stop()
		}
	}
	for _, test := range h.grpcTests {
		if test.Tester != nil {
			test.Tester.Stop()
		}
	}
	h.testMutex.Unlock()

	// 停止测试服务器
	if h.httpServer != nil {
		h.httpServer.Stop()
	}
	if h.grpcConn != nil {
		h.grpcConn.GracefulStop()
	}

	return nil
}
