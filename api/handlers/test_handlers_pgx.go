package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"GoSlgBenchmarkTest/internal/database"
	"GoSlgBenchmarkTest/internal/db"
	"GoSlgBenchmarkTest/internal/logger"
	"GoSlgBenchmarkTest/internal/testrunner"
)

// TestHandlerPgx 使用pgx+sqlc的测试处理器
type TestHandlerPgx struct {
	queries *db.Queries
}

// NewTestHandlerPgx 创建测试处理器
func NewTestHandlerPgx() *TestHandlerPgx {
	return &TestHandlerPgx{
		queries: database.GetQueries(),
	}
}

// CreateTestRequest 创建测试请求
type CreateTestRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Type     string                 `json:"type" binding:"required"` // unity_integration, stress, benchmark, fuzz
	Config   map[string]interface{} `json:"config"`
	SLGMode  bool                   `json:"slg_mode,omitempty"`
	UnityURL string                 `json:"unity_url,omitempty"`
	GameURL  string                 `json:"game_url,omitempty"`
}

// TestResponse 测试响应
type TestResponse struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Duration     *int64     `json:"duration"`
	Score        *float64   `json:"score"`
	Grade        *string    `json:"grade"`
	Config       *string    `json:"config"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateTest 创建测试
// POST /api/v1/tests
func (h *TestHandlerPgx) CreateTest(w http.ResponseWriter, r *http.Request) {
	var req CreateTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证测试类型
	validTypes := map[string]bool{
		"unity_integration": true,
		"stress":            true,
		"benchmark":         true,
		"fuzz":              true,
	}
	if !validTypes[req.Type] {
		http.Error(w, "Invalid test type", http.StatusBadRequest)
		return
	}

	// 构建配置JSON
	configJSON := "{}"
	if req.Config != nil {
		if req.SLGMode {
			req.Config["slg_mode"] = true
			req.Config["unity_url"] = req.UnityURL
			req.Config["game_url"] = req.GameURL
		}
		if configBytes, err := json.Marshal(req.Config); err == nil {
			configJSON = string(configBytes)
		}
	}

	// 创建测试记录
	ctx := context.Background()
	test, err := h.queries.CreateTest(ctx, db.CreateTestParams{
		Name:   req.Name,
		Type:   req.Type,
		Status: "pending",
		Config: pgtype.Text{String: configJSON, Valid: true},
	})

	if err != nil {
		logger.LogError("test", "Failed to create test", nil)
		http.Error(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	logger.LogSuccess("test", fmt.Sprintf("测试创建成功: %s (ID: %d)", req.Name, test.ID), &test.ID)

	// 返回结果
	response := map[string]interface{}{
		"test_id": test.ID,
		"status":  "created",
		"message": "Test created successfully",
		"test":    h.convertToTestResponse(&test),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTestList 获取测试列表
// GET /api/v1/tests?page=1&limit=20&type=unity_integration&status=completed
func (h *TestHandlerPgx) GetTestList(w http.ResponseWriter, r *http.Request) {
	// 分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := int32((page - 1) * limit)

	// 过滤参数
	testType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	// 构建查询参数
	var typeParam, statusParam, searchParam string
	typeParam = testType
	statusParam = status
	searchParam = search

	ctx := context.Background()

	// 获取总数
	total, err := h.queries.CountTests(ctx, db.CountTestsParams{
		Column1: typeParam,
		Column2: statusParam,
		Column3: searchParam,
	})
	if err != nil {
		http.Error(w, "Failed to count tests", http.StatusInternalServerError)
		return
	}

	// 获取数据
	tests, err := h.queries.GetTestList(ctx, db.GetTestListParams{
		Column1: typeParam,
		Column2: statusParam,
		Column3: searchParam,
		Limit:   int32(limit),
		Offset:  offset,
	})

	if err != nil {
		http.Error(w, "Failed to fetch tests", http.StatusInternalServerError)
		return
	}

	// 转换响应
	testResponses := make([]TestResponse, len(tests))
	for i, test := range tests {
		testResponses[i] = *h.convertToTestResponse(&test)
	}

	response := map[string]interface{}{
		"tests":       testResponses,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTest 获取单个测试详情
// GET /api/v1/tests/{id}
func (h *TestHandlerPgx) GetTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	test, err := h.queries.GetTest(ctx, testID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Test not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch test", http.StatusInternalServerError)
		}
		return
	}

	response := h.convertToTestResponse(&test)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartTest 启动测试
// POST /api/v1/tests/{id}/start
func (h *TestHandlerPgx) StartTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	test, err := h.queries.GetTest(ctx, testID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Test not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch test", http.StatusInternalServerError)
		}
		return
	}

	if test.Status != "pending" {
		http.Error(w, "Test is not in pending status", http.StatusBadRequest)
		return
	}

	// 更新测试状态
	now := time.Now()
	updatedTest, err := h.queries.UpdateTest(ctx, db.UpdateTestParams{
		ID:     testID,
		Status: "running",
		StartTime: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		EndTime:      pgtype.Timestamptz{Valid: false},
		Duration:     pgtype.Int8{Valid: false},
		Score:        pgtype.Numeric{Valid: false},
		Grade:        pgtype.Text{Valid: false},
		ErrorMessage: pgtype.Text{Valid: false},
	})

	if err != nil {
		logger.LogError("test", fmt.Sprintf("启动测试失败: %v", err), &testID)
		http.Error(w, "Failed to start test", http.StatusInternalServerError)
		return
	}

	logger.LogInfo("test", fmt.Sprintf("测试启动: %s", test.Name), &testID)

	// 异步执行测试
	go h.executeTest(testID)

	response := map[string]interface{}{
		"test_id": testID,
		"status":  "running",
		"message": "Test started successfully",
		"test":    h.convertToTestResponse(&updatedTest),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopTest 停止测试
// POST /api/v1/tests/{id}/stop
func (h *TestHandlerPgx) StopTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	test, err := h.queries.GetTest(ctx, testID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Test not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch test", http.StatusInternalServerError)
		}
		return
	}

	if test.Status != "running" {
		http.Error(w, "Test is not running", http.StatusBadRequest)
		return
	}

	// 更新测试状态
	now := time.Now()
	var duration int64
	if test.StartTime.Valid {
		duration = now.Sub(test.StartTime.Time).Milliseconds()
	}

	updatedTest, err := h.queries.UpdateTest(ctx, db.UpdateTestParams{
		ID:     testID,
		Status: "stopped",
		StartTime: pgtype.Timestamptz{
			Time:  test.StartTime.Time,
			Valid: test.StartTime.Valid,
		},
		EndTime: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		Duration: pgtype.Int8{
			Int64: duration,
			Valid: true,
		},
		Score:        pgtype.Numeric{Valid: false},
		Grade:        pgtype.Text{Valid: false},
		ErrorMessage: pgtype.Text{Valid: false},
	})

	if err != nil {
		http.Error(w, "Failed to stop test", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"test_id": testID,
		"status":  "stopped",
		"message": "Test stopped successfully",
		"test":    h.convertToTestResponse(&updatedTest),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteTest 删除测试
// DELETE /api/v1/tests/{id}
func (h *TestHandlerPgx) DeleteTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = h.queries.DeleteTest(ctx, testID)
	if err != nil {
		http.Error(w, "Failed to delete test", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"test_id": testID,
		"status":  "deleted",
		"message": "Test deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTestReport 获取测试报告
// GET /api/v1/tests/{id}/report
func (h *TestHandlerPgx) GetTestReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	test, err := h.queries.GetTest(ctx, testID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Test not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch test", http.StatusInternalServerError)
		}
		return
	}

	if test.Status != "completed" && test.Status != "failed" {
		http.Error(w, "Test report not ready", http.StatusBadRequest)
		return
	}

	// 获取测试报告
	reports, err := h.queries.GetTestReports(ctx, testID)
	if err != nil {
		http.Error(w, "Failed to fetch test reports", http.StatusInternalServerError)
		return
	}

	// 获取测试指标
	metrics, err := h.queries.GetTestMetrics(ctx, testID)
	if err != nil {
		http.Error(w, "Failed to fetch test metrics", http.StatusInternalServerError)
		return
	}

	// 获取测试会话
	sessions, err := h.queries.GetTestSessions(ctx, testID)
	if err != nil {
		http.Error(w, "Failed to fetch test sessions", http.StatusInternalServerError)
		return
	}

	// 构建详细报告
	report := map[string]interface{}{
		"test_id":      test.ID,
		"name":         test.Name,
		"type":         test.Type,
		"status":       test.Status,
		"start_time":   test.StartTime,
		"end_time":     test.EndTime,
		"duration":     test.Duration,
		"score":        test.Score,
		"grade":        test.Grade,
		"config":       h.parseConfig(test.Config),
		"reports":      reports,
		"metrics":      h.groupMetricsByType(metrics),
		"sessions":     sessions,
		"generated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// executeTest 执行真实测试
func (h *TestHandlerPgx) executeTest(testID int64) {
	logger.LogInfo("test", "开始执行真实测试...", &testID)

	// 获取测试信息
	ctx := context.Background()
	test, err := h.queries.GetTest(ctx, testID)
	if err != nil {
		logger.LogError("test", fmt.Sprintf("获取测试信息失败: %v", err), &testID)
		return
	}

	// 解析配置
	config := h.parseConfig(test.Config)

	// 确定测试类型
	var testType testrunner.TestType
	switch test.Type {
	case "unity_integration":
		testType = testrunner.TestTypeUnityIntegration
	case "stress":
		testType = testrunner.TestTypeStress
	case "benchmark":
		testType = testrunner.TestTypeBenchmark
	case "fuzz":
		testType = testrunner.TestTypeFuzz
	default:
		logger.LogError("test", fmt.Sprintf("不支持的测试类型: %s", test.Type), &testID)
		return
	}

	// 执行真实测试
	err = testrunner.GlobalExecutor.ExecuteTest(testID, testType, config)
	if err != nil {
		logger.LogError("test", fmt.Sprintf("测试执行失败: %v", err), &testID)
		h.updateTestStatus(testID, "failed", 0.0, "F")
		return
	}

	// 获取测试结果
	result, err := testrunner.GlobalExecutor.GetResult(testID)
	if err != nil {
		logger.LogError("test", fmt.Sprintf("获取测试结果失败: %v", err), &testID)
		return
	}

	// 更新数据库中的测试状态
	h.updateTestStatus(testID, result.Status, result.Score, result.Grade)

	// 创建测试报告
	h.createTestReport(testID, result)
}

// updateTestStatus 更新测试状态
func (h *TestHandlerPgx) updateTestStatus(testID int64, status string, score float64, grade string) {
	ctx := context.Background()
	test, err := h.queries.GetTest(ctx, testID)
	if err != nil {
		logger.LogError("test", fmt.Sprintf("获取测试信息失败: %v", err), &testID)
		return
	}

	now := time.Now()
	var duration int64
	if test.StartTime.Valid {
		duration = now.Sub(test.StartTime.Time).Milliseconds()
	}

	_, err = h.queries.UpdateTest(ctx, db.UpdateTestParams{
		ID:     testID,
		Status: status,
		StartTime: pgtype.Timestamptz{
			Time:  test.StartTime.Time,
			Valid: test.StartTime.Valid,
		},
		EndTime: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		Duration: pgtype.Int8{
			Int64: duration,
			Valid: true,
		},
		Score: func() pgtype.Numeric {
			var n pgtype.Numeric
			n.ScanScientific(fmt.Sprintf("%.2f", score))
			return n
		}(),
		Grade: pgtype.Text{
			String: grade,
			Valid:  true,
		},
		ErrorMessage: pgtype.Text{Valid: false},
	})

	if err != nil {
		logger.LogError("test", fmt.Sprintf("更新测试状态失败: %v", err), &testID)
		return
	}

	logger.LogSuccess("test", fmt.Sprintf("测试完成! 得分: %.1f (%s)", score, grade), &testID)
}

// createTestReport 创建测试报告
func (h *TestHandlerPgx) createTestReport(testID int64, result *testrunner.TestResult) {
	ctx := context.Background()

	// 创建测试报告
	summary := fmt.Sprintf("真实测试执行完成，得分 %.1f，通过 %d 项，失败 %d 项，覆盖率 %.1f%%",
		result.Score, result.TestsPassed, result.TestsFailed, result.Coverage)

	_, err := h.queries.CreateTestReport(ctx, db.CreateTestReportParams{
		TestID:      testID,
		Summary:     pgtype.Text{String: summary, Valid: true},
		Issues:      []byte(fmt.Sprintf(`[{"severity": "info", "category": "test", "description": "执行了 %d 个测试，%d 通过，%d 失败", "impact": "测试结果"}]`, result.TestsPassed+result.TestsFailed, result.TestsPassed, result.TestsFailed)),
		Suggestions: []byte(`[{"priority": "info", "category": "test", "description": "这是真实的go test执行结果", "expected_impact": "提供准确的测试反馈"}]`),
		RawData: pgtype.Text{
			String: fmt.Sprintf(`{"tests_passed": %d, "tests_failed": %d, "coverage": %.2f, "duration_ms": %d}`,
				result.TestsPassed, result.TestsFailed, result.Coverage, result.Duration.Milliseconds()),
			Valid: true,
		},
	})
	if err != nil {
		logger.LogError("test", fmt.Sprintf("创建测试报告失败: %v", err), &testID)
		return
	}

	// 创建测试指标
	h.createTestMetrics(testID, result)
}

// createTestMetrics 创建测试指标
func (h *TestHandlerPgx) createTestMetrics(testID int64, result *testrunner.TestResult) {
	ctx := context.Background()
	now := time.Now()

	createMetric := func(value float64) pgtype.Numeric {
		var n pgtype.Numeric
		n.ScanScientific(fmt.Sprintf("%.2f", value))
		return n
	}

	metrics := []db.CreateTestMetricParams{
		{TestID: testID, MetricType: "test_results", MetricName: "tests_passed", MetricValue: createMetric(float64(result.TestsPassed)), MetricUnit: pgtype.Text{String: "count", Valid: true}, Timestamp: pgtype.Timestamptz{Time: now, Valid: true}},
		{TestID: testID, MetricType: "test_results", MetricName: "tests_failed", MetricValue: createMetric(float64(result.TestsFailed)), MetricUnit: pgtype.Text{String: "count", Valid: true}, Timestamp: pgtype.Timestamptz{Time: now, Valid: true}},
		{TestID: testID, MetricType: "test_results", MetricName: "coverage", MetricValue: createMetric(result.Coverage), MetricUnit: pgtype.Text{String: "%", Valid: true}, Timestamp: pgtype.Timestamptz{Time: now, Valid: true}},
		{TestID: testID, MetricType: "performance", MetricName: "duration", MetricValue: createMetric(float64(result.Duration.Milliseconds())), MetricUnit: pgtype.Text{String: "ms", Valid: true}, Timestamp: pgtype.Timestamptz{Time: now, Valid: true}},
	}

	for _, metric := range metrics {
		_, err := h.queries.CreateTestMetric(ctx, metric)
		if err != nil {
			logger.LogError("test", fmt.Sprintf("创建测试指标失败: %v", err), &testID)
		}
	}
}

// calculateGrade 计算等级
func (h *TestHandlerPgx) calculateGrade(score float64) string {
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

// convertToTestResponse 转换为响应格式
func (h *TestHandlerPgx) convertToTestResponse(test *db.Test) *TestResponse {
	var score *float64
	if test.Score.Valid {
		if f64, err := test.Score.Float64Value(); err == nil && f64.Valid {
			s := f64.Float64
			score = &s
		}
	}

	var grade *string
	if test.Grade.Valid {
		grade = &test.Grade.String
	}

	var config *string
	if test.Config.Valid {
		config = &test.Config.String
	}

	var errorMessage *string
	if test.ErrorMessage.Valid {
		errorMessage = &test.ErrorMessage.String
	}

	var startTime *time.Time
	if test.StartTime.Valid {
		startTime = &test.StartTime.Time
	}

	var endTime *time.Time
	if test.EndTime.Valid {
		endTime = &test.EndTime.Time
	}

	var duration *int64
	if test.Duration.Valid {
		duration = &test.Duration.Int64
	}

	return &TestResponse{
		ID:           test.ID,
		Name:         test.Name,
		Type:         test.Type,
		Status:       test.Status,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     duration,
		Score:        score,
		Grade:        grade,
		Config:       config,
		ErrorMessage: errorMessage,
		CreatedAt:    test.CreatedAt.Time,
		UpdatedAt:    test.UpdatedAt.Time,
	}
}

// parseConfig 解析配置JSON
func (h *TestHandlerPgx) parseConfig(configText pgtype.Text) map[string]interface{} {
	if !configText.Valid {
		return map[string]interface{}{}
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configText.String), &config); err != nil {
		return map[string]interface{}{}
	}
	return config
}

// groupMetricsByType 按类型分组指标
func (h *TestHandlerPgx) groupMetricsByType(metrics []db.TestMetric) map[string][]db.TestMetric {
	groups := make(map[string][]db.TestMetric)
	for _, metric := range metrics {
		groups[metric.MetricType] = append(groups[metric.MetricType], metric)
	}
	return groups
}
