package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"GoSlgBenchmarkTest/internal/database"
	"GoSlgBenchmarkTest/internal/db"
)

// SystemHandlerPgx 使用pgx的系统处理器
type SystemHandlerPgx struct {
	queries   *db.Queries
	startTime time.Time
}

// NewSystemHandlerPgx 创建系统处理器
func NewSystemHandlerPgx() *SystemHandlerPgx {
	return &SystemHandlerPgx{
		queries:   database.GetQueries(),
		startTime: time.Now(),
	}
}

// SystemStatus 系统状态响应
type SystemStatus struct {
	Platform     string              `json:"platform"`
	Version      string              `json:"version"`
	Uptime       string              `json:"uptime"`
	StartTime    time.Time           `json:"start_time"`
	Database     DatabaseStatusPgx   `json:"database"`
	Statistics   SystemStatisticsPgx `json:"statistics"`
	SystemHealth string              `json:"system_health"`
	Memory       MemoryStatsPgx      `json:"memory"`
	Features     []string            `json:"features"`
	Endpoints    map[string]string   `json:"endpoints"`
}

// DatabaseStatusPgx 数据库状态
type DatabaseStatusPgx struct {
	Status    string                 `json:"status"`
	Connected bool                   `json:"connected"`
	Driver    string                 `json:"driver"`
	Message   string                 `json:"message"`
	PoolStats map[string]interface{} `json:"pool_stats"`
}

// SystemStatisticsPgx 系统统计
type SystemStatisticsPgx struct {
	TotalTests     int64 `json:"total_tests"`
	RunningTests   int64 `json:"running_tests"`
	CompletedTests int64 `json:"completed_tests"`
	FailedTests    int64 `json:"failed_tests"`
}

// MemoryStatsPgx 内存统计
type MemoryStatsPgx struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// HealthCheckPgx 健康检查响应
type HealthCheckPgx struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]string      `json:"checks"`
	Details   map[string]interface{} `json:"details"`
}

// GetSystemStatus 获取系统状态
// GET /api/v1/status
func (h *SystemHandlerPgx) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	// 检查数据库状态
	dbStatus := DatabaseStatusPgx{
		Status:    "unknown",
		Connected: false,
		Driver:    "pgx/v5",
		Message:   "Database check failed",
		PoolStats: make(map[string]interface{}),
	}

	if pool := database.GetPgxPool(); pool != nil {
		if err := database.TestConnectionPgx(); err == nil {
			dbStatus.Status = "healthy"
			dbStatus.Connected = true
			dbStatus.Message = "PostgreSQL connection is healthy"

			// 获取连接池统计
			if stats := database.GetPoolStats(); stats != nil {
				dbStatus.PoolStats = map[string]interface{}{
					"total_conns":    stats.TotalConns(),
					"idle_conns":     stats.IdleConns(),
					"acquired_conns": stats.AcquiredConns(),
					"max_conns":      stats.MaxConns(),
					"new_conns":      stats.NewConnsCount(),
					"acquire_count":  stats.AcquireCount(),
					"cancel_count":   stats.CanceledAcquireCount(),
				}
			}
		} else {
			dbStatus.Status = "error"
			dbStatus.Message = err.Error()
		}
	}

	// 获取系统统计
	stats := h.getSystemStatistics()

	// 获取内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memStats := MemoryStatsPgx{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}

	// 确定系统健康状态
	systemHealth := "healthy"
	if !dbStatus.Connected {
		systemHealth = "degraded"
	}

	status := SystemStatus{
		Platform:     "Unity SLG Test Platform",
		Version:      "3.0.0 (PGX + SQLC)",
		Uptime:       time.Since(h.startTime).String(),
		StartTime:    h.startTime,
		Database:     dbStatus,
		Statistics:   stats,
		SystemHealth: systemHealth,
		Memory:       memStats,
		Features: []string{
			"High-performance PGX driver",
			"Type-safe SQLC queries",
			"Connection pooling",
			"Real-time monitoring",
			"Unity integration",
			"SLG game metrics",
			"Automated testing",
			"Performance analysis",
		},
		Endpoints: map[string]string{
			"api_status": "/api/v1/status",
			"health":     "/api/v1/health",
			"tests":      "/api/v1/tests",
			"api_docs":   "/api/docs",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HealthCheck 健康检查
// GET /api/v1/health
func (h *SystemHandlerPgx) HealthCheck(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	details := make(map[string]interface{})
	overallStatus := "healthy"

	// 数据库检查
	if pool := database.GetPgxPool(); pool != nil {
		if err := database.TestConnectionPgx(); err == nil {
			checks["database"] = "healthy"
			if stats := database.GetPoolStats(); stats != nil {
				details["database"] = map[string]interface{}{
					"connected":      true,
					"driver":         "pgx/v5",
					"total_conns":    stats.TotalConns(),
					"idle_conns":     stats.IdleConns(),
					"acquired_conns": stats.AcquiredConns(),
					"max_conns":      stats.MaxConns(),
				}
			}
		} else {
			checks["database"] = "unhealthy"
			details["database"] = map[string]interface{}{
				"connected": false,
				"error":     err.Error(),
				"driver":    "pgx/v5",
			}
			overallStatus = "unhealthy"
		}
	} else {
		checks["database"] = "unavailable"
		details["database"] = map[string]interface{}{
			"connected": false,
			"error":     "Database pool not initialized",
			"driver":    "pgx/v5",
		}
		overallStatus = "unhealthy"
	}

	// 内存检查
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 检查内存使用是否过高 (超过1GB认为不健康)
	if m.Alloc > 1024*1024*1024 {
		checks["memory"] = "warning"
		if overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	} else {
		checks["memory"] = "healthy"
	}

	details["memory"] = map[string]interface{}{
		"alloc_mb":   float64(m.Alloc) / 1024 / 1024,
		"sys_mb":     float64(m.Sys) / 1024 / 1024,
		"num_gc":     m.NumGC,
		"goroutines": runtime.NumGoroutine(),
	}

	// 系统资源检查
	checks["system"] = "healthy"
	details["system"] = map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"cgo_calls":  runtime.NumCgoCall(),
		"uptime":     time.Since(h.startTime).String(),
	}

	health := HealthCheckPgx{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    checks,
		Details:   details,
	}

	// 根据健康状态设置HTTP状态码
	if overallStatus == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if overallStatus == "degraded" {
		w.WriteHeader(http.StatusOK) // 仍然返回200，但状态是degraded
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// getSystemStatistics 获取系统统计
func (h *SystemHandlerPgx) getSystemStatistics() SystemStatisticsPgx {
	ctx := context.Background()

	stats, err := h.queries.GetTestStats(ctx)
	if err != nil {
		return SystemStatisticsPgx{}
	}

	return SystemStatisticsPgx{
		TotalTests:     stats.TotalTests,
		RunningTests:   stats.RunningTests,
		CompletedTests: stats.CompletedTests,
		FailedTests:    stats.FailedTests,
	}
}
