package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"GoSlgBenchmarkTest/api/handlers"
	"GoSlgBenchmarkTest/internal/database"
	"GoSlgBenchmarkTest/internal/logger"
	"GoSlgBenchmarkTest/internal/testrunner"
)

var (
	port   = flag.String("port", "8080", "服务器端口")
	dbHost = flag.String("db-host", "localhost", "数据库主机")
	dbPort = flag.Int("db-port", 5432, "数据库端口")
	dbUser = flag.String("db-user", "postgres", "数据库用户名")
	dbPass = flag.String("db-pass", "2020", "数据库密码")
	dbName = flag.String("db-name", "postgres", "数据库名称")
	debug  = flag.Bool("debug", false, "启用调试模式")
)

func main() {
	flag.Parse()

	fmt.Println("🎮 Unity SLG 测试平台 v3.0 (PGX + SQLC)")
	fmt.Println("==========================================")
	fmt.Println()

	// 初始化WebSocket日志器
	logger.InitGlobalLogger()
	logger.LogInfo("system", "Unity SLG 测试平台启动中...", nil)

	// 初始化测试执行器
	workDir, _ := os.Getwd()
	testrunner.InitGlobalExecutor(workDir)
	logger.LogInfo("system", "测试执行器初始化完成", nil)

	// 连接数据库
	dbConfig := &database.Config{
		Host:     *dbHost,
		Port:     *dbPort,
		User:     *dbUser,
		Password: *dbPass,
		DBName:   *dbName,
		SSLMode:  "disable",
	}

	fmt.Printf("🔗 连接数据库: %s:%d/%s\n", dbConfig.Host, dbConfig.Port, dbConfig.DBName)
	logger.LogInfo("database", fmt.Sprintf("连接数据库: %s:%d/%s", dbConfig.Host, dbConfig.Port, dbConfig.DBName), nil)

	if err := database.ConnectPgx(dbConfig); err != nil {
		logger.LogError("database", fmt.Sprintf("数据库连接失败: %v", err), nil)
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 测试连接
	if err := database.TestConnectionPgx(); err != nil {
		logger.LogError("database", fmt.Sprintf("数据库连接测试失败: %v", err), nil)
		log.Fatalf("❌ 数据库连接测试失败: %v", err)
	}

	logger.LogSuccess("database", "PostgreSQL连接池创建成功", nil)

	// 初始化处理器
	testHandler := handlers.NewTestHandlerPgx()
	systemHandler := handlers.NewSystemHandlerPgx()

	// 创建路由器
	router := mux.NewRouter()

	// 注册API路由
	registerAPIRoutes(router, testHandler, systemHandler)

	// 注册静态文件和页面
	registerStaticRoutes(router)

	// 配置CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动服务器
	go func() {
		fmt.Printf("🚀 服务器启动在端口 %s\n", *port)
		fmt.Printf("📊 访问API: http://localhost:%s/api/v1/status\n", *port)
		fmt.Printf("🔧 API文档: http://localhost:%s/api/docs\n", *port)
		fmt.Printf("💾 数据库: PostgreSQL (PGX) %s:%d\n", *dbHost, *dbPort)
		fmt.Println()
		fmt.Println("🎮 Unity SLG 测试平台功能:")
		fmt.Println("  ✅ PGX高性能数据库驱动")
		fmt.Println("  ✅ SQLC强类型SQL查询")
		fmt.Println("  ✅ RESTful API接口")
		fmt.Println("  ✅ 测试管理和执行")
		fmt.Println("  ✅ Unity客户端集成")
		fmt.Println("  ✅ SLG游戏指标收集")
		fmt.Println("  ✅ 实时数据监控")
		fmt.Println("  ✅ 测试报告生成")
		fmt.Println()

		// 打印连接池统计
		if stats := database.GetPoolStats(); stats != nil {
			fmt.Printf("📊 连接池状态: 总连接=%d, 空闲=%d, 使用中=%d\n",
				stats.TotalConns(), stats.IdleConns(), stats.AcquiredConns())
		}
		fmt.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🔄 正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ 服务器关闭错误: %v", err)
	}

	// 关闭数据库连接
	database.ClosePgx()

	fmt.Println("✅ 服务器已关闭")
}

// registerAPIRoutes 注册API路由
func registerAPIRoutes(router *mux.Router, testHandler *handlers.TestHandlerPgx, systemHandler *handlers.SystemHandlerPgx) {
	// v1 API路由
	api := router.PathPrefix("/api/v1").Subrouter()

	// 系统相关
	api.HandleFunc("/status", systemHandler.GetSystemStatus).Methods("GET")
	api.HandleFunc("/health", systemHandler.HealthCheck).Methods("GET")

	// 测试管理
	api.HandleFunc("/tests", testHandler.CreateTest).Methods("POST")
	api.HandleFunc("/tests", testHandler.GetTestList).Methods("GET")
	api.HandleFunc("/tests/{id:[0-9]+}", testHandler.GetTest).Methods("GET")
	api.HandleFunc("/tests/{id:[0-9]+}/start", testHandler.StartTest).Methods("POST")
	api.HandleFunc("/tests/{id:[0-9]+}/stop", testHandler.StopTest).Methods("POST")
	api.HandleFunc("/tests/{id:[0-9]+}", testHandler.DeleteTest).Methods("DELETE")
	api.HandleFunc("/tests/{id:[0-9]+}/report", testHandler.GetTestReport).Methods("GET")

	// API文档
	router.HandleFunc("/api/docs", serveAPIDocs).Methods("GET")

	// WebSocket日志流
	router.HandleFunc("/api/v1/logs/ws", logger.GlobalLogger.HandleWebSocket)
}

// registerStaticRoutes 注册静态文件路由
func registerStaticRoutes(router *mux.Router) {
	// Dashboard页面
	router.HandleFunc("/dashboard", serveDashboard).Methods("GET")
	router.HandleFunc("/", serveDashboard).Methods("GET")

	// API信息
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"platform": "Unity SLG Test Platform",
			"version":  "3.0.0",
			"message":  "Unity SLG 测试平台 API 服务 (PGX + SQLC)",
			"database": "PostgreSQL (PGX Driver + SQLC)",
			"endpoints": map[string]string{
				"dashboard":  "/dashboard",
				"api_status": "/api/v1/status",
				"health":     "/api/v1/health",
				"tests":      "/api/v1/tests",
				"api_docs":   "/api/docs",
			},
			"features": []string{
				"High-performance PGX driver",
				"Type-safe SQLC queries",
				"RESTful API",
				"Test management",
				"Real-time monitoring",
				"Performance benchmarking",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")
}

// serveAPIDocs 提供API文档
func serveAPIDocs(w http.ResponseWriter, r *http.Request) {
	docs := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unity SLG 测试平台 API 文档 v3.0</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 2rem; line-height: 1.6; }
        .header { text-align: center; margin-bottom: 2rem; }
        .badge { background: #2ecc71; color: white; padding: 0.2rem 0.5rem; border-radius: 3px; font-size: 0.8rem; }
        .endpoint { margin: 1rem 0; padding: 1rem; border: 1px solid #ddd; border-radius: 5px; }
        .method { font-weight: bold; padding: 0.2rem 0.5rem; border-radius: 3px; color: white; margin-right: 0.5rem; }
        .get { background-color: #61affe; }
        .post { background-color: #49cc90; }
        .put { background-color: #fca130; }
        .delete { background-color: #f93e3e; }
        .path { font-family: monospace; background: #f8f9fa; padding: 0.2rem 0.5rem; border-radius: 3px; }
        .description { margin: 0.5rem 0; color: #666; }
        .section { margin: 2rem 0; }
        .toc { background: #f8f9fa; padding: 1rem; border-radius: 5px; margin-bottom: 2rem; }
        .toc ul { margin: 0; padding-left: 1.5rem; }
        .example { background: #f4f4f4; padding: 1rem; border-radius: 5px; margin: 1rem 0; }
        .example pre { margin: 0; }
        .tech-stack { background: #e8f5e8; padding: 1rem; border-radius: 5px; margin: 1rem 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎮 Unity SLG 测试平台 API 文档</h1>
        <p><span class="badge">v3.0</span> 基于PGX + SQLC的高性能版本</p>
    </div>

    <div class="tech-stack">
        <h3>🚀 技术栈升级</h3>
        <ul>
            <li><strong>PGX Driver</strong>: PostgreSQL的高性能原生驱动</li>
            <li><strong>SQLC</strong>: 从SQL生成类型安全的Go代码</li>
            <li><strong>连接池</strong>: pgxpool自动管理连接生命周期</li>
            <li><strong>性能优化</strong>: 零反射, 预编译查询, 批处理支持</li>
        </ul>
    </div>

    <div class="toc">
        <h3>目录</h3>
        <ul>
            <li><a href="#system">系统管理 API</a></li>
            <li><a href="#tests">测试管理 API</a></li>
            <li><a href="#examples">使用示例</a></li>
            <li><a href="#database">数据库架构</a></li>
        </ul>
    </div>

    <div id="system" class="section">
        <h2>系统管理 API</h2>
        
        <div class="endpoint">
            <div><span class="method get">GET</span> <span class="path">/api/v1/status</span></div>
            <div class="description">获取系统状态、连接池统计和数据库信息</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method get">GET</span> <span class="path">/api/v1/health</span></div>
            <div class="description">健康检查接口，包含PGX连接池状态</div>
        </div>
    </div>

    <div id="tests" class="section">
        <h2>测试管理 API</h2>
        
        <div class="endpoint">
            <div><span class="method post">POST</span> <span class="path">/api/v1/tests</span></div>
            <div class="description">创建新测试任务 (使用类型安全的SQLC查询)</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method get">GET</span> <span class="path">/api/v1/tests</span></div>
            <div class="description">获取测试列表，支持高效分页和过滤</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method get">GET</span> <span class="path">/api/v1/tests/{id}</span></div>
            <div class="description">获取指定测试的详细信息</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method post">POST</span> <span class="path">/api/v1/tests/{id}/start</span></div>
            <div class="description">启动指定测试</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method post">POST</span> <span class="path">/api/v1/tests/{id}/stop</span></div>
            <div class="description">停止指定测试</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method get">GET</span> <span class="path">/api/v1/tests/{id}/report</span></div>
            <div class="description">获取测试详细报告 (包含关联的metrics和sessions)</div>
        </div>
        
        <div class="endpoint">
            <div><span class="method delete">DELETE</span> <span class="path">/api/v1/tests/{id}</span></div>
            <div class="description">软删除指定测试</div>
        </div>
    </div>

    <div id="examples" class="section">
        <h2>使用示例</h2>
        
        <h3>1. 创建Unity SLG测试</h3>
        <div class="example">
            <pre><code>POST /api/v1/tests
Content-Type: application/json

{
  "name": "Unity SLG 高性能测试",
  "type": "unity_integration",
  "slg_mode": true,
  "config": {
    "duration": "10m",
    "clients": 10,
    "battle_scenarios": ["1v1", "team", "guild_war"]
  }
}</code></pre>
        </div>

        <h3>2. 获取测试列表 (分页)</h3>
        <div class="example">
            <pre><code>GET /api/v1/tests?page=1&limit=20&type=unity_integration&status=completed</code></pre>
        </div>

        <h3>3. 获取详细报告</h3>
        <div class="example">
            <pre><code>GET /api/v1/tests/1/report</code></pre>
        </div>
    </div>

    <div id="database" class="section">
        <h2>数据库架构</h2>
        <p>本平台使用PostgreSQL + PGX + SQLC，具有以下优势：</p>
        <ul>
            <li><strong>高性能</strong>: PGX原生驱动，零拷贝，批处理优化</li>
            <li><strong>类型安全</strong>: SQLC生成的代码在编译时检查SQL正确性</li>
            <li><strong>连接池</strong>: 自动管理连接生命周期，支持健康检查</li>
            <li><strong>事务支持</strong>: 完整的事务和上下文支持</li>
            <li><strong>可维护性</strong>: SQL语句独立管理，易于审查和优化</li>
        </ul>

        <h3>主要数据表</h3>
        <ul>
            <li><strong>tests</strong> - 测试主表 (支持软删除)</li>
            <li><strong>test_reports</strong> - 测试报告 (JSONB存储)</li>
            <li><strong>test_metrics</strong> - 测试指标 (高精度时间序列)</li>
            <li><strong>test_sessions</strong> - 测试会话</li>
            <li><strong>session_events</strong> - 会话事件 (JSONB存储)</li>
            <li><strong>session_messages</strong> - 会话消息 (二进制数据)</li>
            <li><strong>slg_battle_records</strong> - SLG战斗记录</li>
            <li><strong>unity_client_records</strong> - Unity客户端记录</li>
        </ul>
    </div>

    <div style="text-align: center; margin: 2rem 0; padding-top: 2rem; border-top: 1px solid #ddd;">
        <p>Unity SLG 测试平台 v3.0 - PGX + SQLC 高性能版本</p>
        <p><small>🚀 更快的查询 • 🔒 类型安全 • 💪 生产就绪</small></p>
    </div>
</body>
</html>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(docs))
}

// serveDashboard 提供Dashboard页面
func serveDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardHTML := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unity SLG 测试平台 v3.0 (PGX + SQLC)</title>
    
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    
    <style>
        .test-card { transition: transform 0.2s; }
        .test-card:hover { transform: translateY(-2px); }
        .status-badge { font-size: 0.8rem; }
        .metric-item { background: #f8f9fa; border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
        .log-container { background: #1e1e1e; color: #00ff00; font-family: 'Courier New', monospace; height: 300px; overflow-y: auto; padding: 1rem; border-radius: 5px; }
        .test-progress { height: 4px; }
        .test-running { animation: pulse 2s infinite; }
        @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }
    </style>
</head>
<body class="bg-light">
    <!-- 导航栏 -->
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="#"><i class="fas fa-gamepad me-2"></i>Unity SLG 测试平台 v3.0</a>
            <span class="badge bg-success ms-auto"><i class="fas fa-database me-1"></i>PGX + SQLC</span>
        </div>
    </nav>

    <div class="container mt-4">
        <!-- 系统状态面板 -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h5 class="mb-0"><i class="fas fa-heartbeat me-2"></i>系统状态</h5>
                        <button class="btn btn-outline-primary btn-sm" onclick="refreshSystemStatus()">
                            <i class="fas fa-sync-alt"></i> 刷新
                        </button>
                    </div>
                    <div class="card-body">
                        <div class="row" id="systemStatus">
                            <div class="col-md-3">
                                <div class="metric-item text-center">
                                    <div class="h4 text-info" id="totalConns">-</div>
                                    <small class="text-muted">数据库连接池</small>
                                </div>
                            </div>
                            <div class="col-md-3">
                                <div class="metric-item text-center">
                                    <div class="h4 text-success" id="totalTests">-</div>
                                    <small class="text-muted">总测试数</small>
                                </div>
                            </div>
                            <div class="col-md-3">
                                <div class="metric-item text-center">
                                    <div class="h4 text-warning" id="runningTests">-</div>
                                    <small class="text-muted">运行中测试</small>
                                </div>
                            </div>
                            <div class="col-md-3">
                                <div class="metric-item text-center">
                                    <div class="h4 text-primary" id="completedTests">-</div>
                                    <small class="text-muted">已完成测试</small>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 快速创建测试 -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0"><i class="fas fa-rocket me-2"></i>快速创建测试</h5>
                    </div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-3">
                                <button class="btn btn-primary w-100 mb-2" onclick="createQuickTest('unity_integration')">
                                    <i class="fas fa-cube me-2"></i>Unity SLG集成测试
                                </button>
                            </div>
                            <div class="col-md-3">
                                <button class="btn btn-info w-100 mb-2" onclick="createQuickTest('stress')">
                                    <i class="fas fa-tachometer-alt me-2"></i>WebSocket压力测试
                                </button>
                            </div>
                            <div class="col-md-3">
                                <button class="btn btn-warning w-100 mb-2" onclick="createQuickTest('fuzz')">
                                    <i class="fas fa-random me-2"></i>协议模糊测试
                                </button>
                            </div>
                            <div class="col-md-3">
                                <button class="btn btn-success w-100 mb-2" onclick="createQuickTest('benchmark')">
                                    <i class="fas fa-chart-bar me-2"></i>性能基准测试
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 测试列表 -->
        <div class="row">
            <div class="col-lg-8">
                <div class="card">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h5 class="mb-0"><i class="fas fa-list me-2"></i>测试列表</h5>
                        <button class="btn btn-outline-primary btn-sm" onclick="refreshTestList()">
                            <i class="fas fa-sync-alt"></i> 刷新
                        </button>
                    </div>
                    <div class="card-body">
                        <div id="testList"><!-- 测试列表将在这里动态加载 --></div>
                    </div>
                </div>
            </div>
            
            <div class="col-lg-4">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0"><i class="fas fa-terminal me-2"></i>实时日志</h5>
                    </div>
                    <div class="card-body p-0">
                        <div class="log-container" id="logContainer">
                            <div>[INFO] Unity SLG 测试平台已启动...</div>
                            <div>[INFO] 数据库连接池: PGX + SQLC</div>
                            <div>[INFO] 等待测试任务...</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 测试详情模态框 -->
        <div class="modal fade" id="testDetailModal" tabindex="-1">
            <div class="modal-dialog modal-xl">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">测试详情</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <div id="testDetailContent"><!-- 测试详情将在这里显示 --></div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    
    <script>
        // WebSocket连接
        let logSocket = null;
        
        // 页面加载时初始化
        document.addEventListener('DOMContentLoaded', function() {
            refreshSystemStatus();
            refreshTestList();
            connectLogWebSocket();
            
            // 每5秒刷新一次数据
            setInterval(() => {
                refreshSystemStatus();
                refreshTestList();
            }, 5000);
        });
        
        // 连接WebSocket日志流
        function connectLogWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/api/v1/logs/ws';
            
            logSocket = new WebSocket(wsUrl);
            
            logSocket.onopen = function(event) {
                console.log('WebSocket日志连接已建立');
            };
            
            logSocket.onmessage = function(event) {
                try {
                    const logMsg = JSON.parse(event.data);
                    addRealLog(logMsg);
                } catch (e) {
                    console.error('解析日志消息失败:', e);
                }
            };
            
            logSocket.onclose = function(event) {
                console.log('WebSocket日志连接已关闭，3秒后重连...');
                setTimeout(connectLogWebSocket, 3000);
            };
            
            logSocket.onerror = function(error) {
                console.error('WebSocket日志连接错误:', error);
            };
        }
        
        // 添加真实日志消息
        function addRealLog(logMsg) {
            const logContainer = document.getElementById('logContainer');
            const time = new Date(logMsg.timestamp).toLocaleTimeString();
            const logEntry = document.createElement('div');
            
            // 根据日志级别设置颜色
            let color = '#00ff00'; // 默认绿色
            switch (logMsg.level) {
                case 'ERROR':
                    color = '#ff5555';
                    break;
                case 'WARNING':
                    color = '#ffff55';
                    break;
                case 'SUCCESS':
                    color = '#55ff55';
                    break;
                case 'INFO':
                    color = '#55ffff';
                    break;
            }
            
            logEntry.style.color = color;
            
            let prefix = '[' + time + '] [' + logMsg.level + ']';
            if (logMsg.test_id) {
                prefix += ' [Test-' + logMsg.test_id + ']';
            }
            prefix += ' ' + logMsg.module + ': ';
            
            logEntry.textContent = prefix + logMsg.message;
            logContainer.appendChild(logEntry);
            logContainer.scrollTop = logContainer.scrollHeight;
            
            // 保持最多100条日志
            while (logContainer.children.length > 100) {
                logContainer.removeChild(logContainer.firstChild);
            }
        }
        
        // 刷新系统状态
        async function refreshSystemStatus() {
            try {
                const response = await fetch('/api/v1/status');
                const data = await response.json();
                
                // 更新连接池状态
                const poolStats = data.database.pool_stats || {};
                document.getElementById('totalConns').textContent = 
                    (poolStats.acquired_conns || 0) + '/' + (poolStats.total_conns || 0);
                
                // 更新测试统计
                const stats = data.statistics || {};
                document.getElementById('totalTests').textContent = stats.total_tests || 0;
                document.getElementById('runningTests').textContent = stats.running_tests || 0;
                document.getElementById('completedTests').textContent = stats.completed_tests || 0;
                
                addLog('系统状态: ' + data.system_health + ', 数据库: ' + data.database.status);
            } catch (error) {
                console.error('Failed to fetch system status:', error);
                addLog('获取系统状态失败: ' + error.message);
            }
        }
        
        // 刷新测试列表
        async function refreshTestList() {
            try {
                const response = await fetch('/api/v1/tests?limit=20');
                const data = await response.json();
                
                const testListContainer = document.getElementById('testList');
                
                if (data.tests && data.tests.length > 0) {
                    testListContainer.innerHTML = data.tests.map(test => {
                        return '<div class="test-card card mb-3 ' + (test.status === 'running' ? 'test-running' : '') + '">' +
                            '<div class="card-body">' +
                                '<div class="d-flex justify-content-between align-items-start">' +
                                    '<div>' +
                                        '<h6 class="card-title mb-1">' + test.name + '</h6>' +
                                        '<p class="card-text text-muted mb-2">' +
                                            '<span class="badge bg-secondary me-2">' + test.type + '</span>' +
                                            '<small>ID: ' + test.id + '</small>' +
                                        '</p>' +
                                    '</div>' +
                                    '<div class="text-end">' +
                                        '<span class="badge status-badge ' + getStatusBadgeClass(test.status) + '">' + getStatusText(test.status) + '</span>' +
                                        (test.score ? '<div class="mt-1"><small class="text-muted">得分: <strong>' + test.score.toFixed(1) + '</strong></small></div>' : '') +
                                    '</div>' +
                                '</div>' +
                                
                                (test.status === 'running' && test.start_time ? 
                                    '<div class="progress test-progress mt-2">' +
                                        '<div class="progress-bar progress-bar-striped progress-bar-animated" style="width: 100%"></div>' +
                                    '</div>' +
                                    '<small class="text-muted">运行时间: ' + getRunningTime(test.start_time) + '</small>'
                                : '') +
                                
                                '<div class="mt-2">' +
                                    '<button class="btn btn-sm btn-outline-primary me-2" onclick="viewTestDetail(' + test.id + ')">' +
                                        '<i class="fas fa-eye"></i> 详情' +
                                    '</button>' +
                                    
                                    (test.status === 'pending' ? 
                                        '<button class="btn btn-sm btn-success me-2" onclick="startTest(' + test.id + ')">' +
                                            '<i class="fas fa-play"></i> 启动' +
                                        '</button>'
                                    : '') +
                                    
                                    (test.status === 'running' ? 
                                        '<button class="btn btn-sm btn-warning me-2" onclick="stopTest(' + test.id + ')">' +
                                            '<i class="fas fa-stop"></i> 停止' +
                                        '</button>'
                                    : '') +
                                    
                                    (test.status === 'completed' || test.status === 'failed' ? 
                                        '<button class="btn btn-sm btn-info me-2" onclick="viewTestReport(' + test.id + ')">' +
                                            '<i class="fas fa-chart-line"></i> 报告' +
                                        '</button>'
                                    : '') +
                                    
                                    '<button class="btn btn-sm btn-outline-danger" onclick="deleteTest(' + test.id + ')">' +
                                        '<i class="fas fa-trash"></i>' +
                                    '</button>' +
                                '</div>' +
                            '</div>' +
                        '</div>';
                    }).join('');
                } else {
                    testListContainer.innerHTML = 
                        '<div class="text-center text-muted py-4">' +
                            '<i class="fas fa-flask fa-3x mb-3"></i>' +
                            '<p>暂无测试任务，点击上方按钮创建新测试</p>' +
                        '</div>';
                }
            } catch (error) {
                console.error('Failed to fetch test list:', error);
                addLog('获取测试列表失败: ' + error.message);
            }
        }
        
        // 创建快速测试
        async function createQuickTest(type) {
            const testConfigs = {
                'unity_integration': {
                    name: 'Unity SLG集成测试 - ' + new Date().toLocaleTimeString(),
                    type: 'unity_integration',
                    slg_mode: true,
                    unity_url: 'ws://localhost:7777/ws',
                    game_url: 'ws://localhost:8080/ws',
                    config: { duration: '30s', clients: 3, battle_scenarios: ['1v1', 'team'], unity_version: '2022.3.12f1' }
                },
                'stress': {
                    name: 'WebSocket压力测试 - ' + new Date().toLocaleTimeString(),
                    type: 'stress',
                    config: { duration: '60s', concurrent_clients: 50, message_rate: 10, target_url: 'ws://localhost:8080/ws' }
                },
                'fuzz': {
                    name: '协议模糊测试 - ' + new Date().toLocaleTimeString(),
                    type: 'fuzz',
                    config: { iterations: 1000, test_cases: ['decode', 'encode', 'frame'], timeout: '30s' }
                },
                'benchmark': {
                    name: '性能基准测试 - ' + new Date().toLocaleTimeString(),
                    type: 'benchmark',
                    config: { duration: '60s', benchmarks: ['serialization', 'websocket', 'database'], iterations: 10000 }
                }
            };
            
            try {
                addLog('创建' + type + '测试...');
                
                const response = await fetch('/api/v1/tests', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(testConfigs[type])
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    addLog('测试创建成功，ID: ' + result.test_id);
                    refreshTestList();
                    setTimeout(() => startTest(result.test_id), 1000);
                } else {
                    addLog('测试创建失败: ' + (result.message || response.statusText));
                }
            } catch (error) {
                addLog('创建测试失败: ' + error.message);
            }
        }
        
        // 启动测试
        async function startTest(testId) {
            try {
                addLog('启动测试 ' + testId + '...');
                const response = await fetch('/api/v1/tests/' + testId + '/start', { method: 'POST' });
                const result = await response.json();
                
                if (response.ok) {
                    addLog('测试 ' + testId + ' 启动成功');
                    refreshTestList();
                } else {
                    addLog('启动测试失败: ' + (result.message || response.statusText));
                }
            } catch (error) {
                addLog('启动测试失败: ' + error.message);
            }
        }
        
        // 停止测试
        async function stopTest(testId) {
            try {
                addLog('停止测试 ' + testId + '...');
                const response = await fetch('/api/v1/tests/' + testId + '/stop', { method: 'POST' });
                const result = await response.json();
                
                if (response.ok) {
                    addLog('测试 ' + testId + ' 停止成功');
                    refreshTestList();
                } else {
                    addLog('停止测试失败: ' + (result.message || response.statusText));
                }
            } catch (error) {
                addLog('停止测试失败: ' + error.message);
            }
        }
        
        // 查看测试详情
        async function viewTestDetail(testId) {
            try {
                const response = await fetch('/api/v1/tests/' + testId);
                const test = await response.json();
                
                document.getElementById('testDetailContent').innerHTML = 
                    '<div class="row">' +
                        '<div class="col-md-6">' +
                            '<h6>基本信息</h6>' +
                            '<table class="table table-sm">' +
                                '<tr><td>ID</td><td>' + test.id + '</td></tr>' +
                                '<tr><td>名称</td><td>' + test.name + '</td></tr>' +
                                '<tr><td>类型</td><td><span class="badge bg-secondary">' + test.type + '</span></td></tr>' +
                                '<tr><td>状态</td><td><span class="badge ' + getStatusBadgeClass(test.status) + '">' + getStatusText(test.status) + '</span></td></tr>' +
                                '<tr><td>创建时间</td><td>' + new Date(test.created_at).toLocaleString() + '</td></tr>' +
                                (test.start_time ? '<tr><td>开始时间</td><td>' + new Date(test.start_time).toLocaleString() + '</td></tr>' : '') +
                                (test.end_time ? '<tr><td>结束时间</td><td>' + new Date(test.end_time).toLocaleString() + '</td></tr>' : '') +
                                (test.duration ? '<tr><td>运行时长</td><td>' + (test.duration / 1000).toFixed(1) + '秒</td></tr>' : '') +
                                (test.score ? '<tr><td>得分</td><td><strong>' + test.score.toFixed(1) + '</strong> (' + test.grade + ')</td></tr>' : '') +
                            '</table>' +
                        '</div>' +
                        '<div class="col-md-6">' +
                            '<h6>配置信息</h6>' +
                            '<pre class="bg-light p-3 rounded"><code>' + JSON.stringify(test.config || {}, null, 2) + '</code></pre>' +
                        '</div>' +
                    '</div>';
                
                new bootstrap.Modal(document.getElementById('testDetailModal')).show();
            } catch (error) {
                addLog('获取测试详情失败: ' + error.message);
            }
        }
        
        // 查看测试报告
        async function viewTestReport(testId) {
            try {
                addLog('加载测试报告 ' + testId + '...');
                const response = await fetch('/api/v1/tests/' + testId + '/report');
                const report = await response.json();
                
                if (response.ok) {
                    document.getElementById('testDetailContent').innerHTML = 
                        '<div class="row">' +
                            '<div class="col-12">' +
                                '<h6>测试报告 - ' + report.name + '</h6>' +
                                '<div class="alert alert-info">' +
                                    '<strong>总体得分:</strong> ' + (report.score ? report.score.toFixed(1) : 'N/A') + ' (' + (report.grade || 'N/A') + ')' +
                                '</div>' +
                            '</div>' +
                        '</div>';
                    
                    new bootstrap.Modal(document.getElementById('testDetailModal')).show();
                    addLog('测试报告加载完成');
                } else {
                    addLog('获取测试报告失败: ' + (report.message || response.statusText));
                }
            } catch (error) {
                addLog('获取测试报告失败: ' + error.message);
            }
        }
        
        // 删除测试
        async function deleteTest(testId) {
            if (!confirm('确定要删除这个测试吗？')) return;
            
            try {
                addLog('删除测试 ' + testId + '...');
                const response = await fetch('/api/v1/tests/' + testId, { method: 'DELETE' });
                const result = await response.json();
                
                if (response.ok) {
                    addLog('测试 ' + testId + ' 删除成功');
                    refreshTestList();
                } else {
                    addLog('删除测试失败: ' + (result.message || response.statusText));
                }
            } catch (error) {
                addLog('删除测试失败: ' + error.message);
            }
        }
        
        // 工具函数
        function getStatusBadgeClass(status) {
            const classes = { 'pending': 'bg-secondary', 'running': 'bg-warning', 'completed': 'bg-success', 'failed': 'bg-danger', 'stopped': 'bg-dark' };
            return classes[status] || 'bg-secondary';
        }
        
        function getStatusText(status) {
            const texts = { 'pending': '等待中', 'running': '运行中', 'completed': '已完成', 'failed': '失败', 'stopped': '已停止' };
            return texts[status] || status;
        }
        
        function getRunningTime(startTime) {
            const diff = Date.now() - new Date(startTime).getTime();
            return Math.floor(diff / 1000) + '秒';
        }
        
        function addLog(message) {
            const logContainer = document.getElementById('logContainer');
            const time = new Date().toLocaleTimeString();
            const logEntry = document.createElement('div');
            logEntry.textContent = '[' + time + '] ' + message;
            logContainer.appendChild(logEntry);
            logContainer.scrollTop = logContainer.scrollHeight;
            
            // 保持最多100条日志
            while (logContainer.children.length > 100) {
                logContainer.removeChild(logContainer.firstChild);
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}
