package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"GoSlgBenchmarkTest/api/handlers"
	"GoSlgBenchmarkTest/internal/config"
	"GoSlgBenchmarkTest/internal/logger"
)

// ExtendedTestPlatform 扩展的测试平台
type ExtendedTestPlatform struct {
	router          *mux.Router
	server          *http.Server
	loadTestHandler *handlers.LoadTestHandler
	cfg             *config.TestConfig
}

// findAvailablePort 查找可用的端口
func findAvailablePort(startPort int, maxAttempts int) int {
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if isPortAvailable(port) {
			return port
		}
		log.Printf("Port %d is in use, trying next port...", port)
	}
	return 0 // 没有找到可用端口
}

// isPortAvailable 检查端口是否可用
func isPortAvailable(port int) bool {
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

// NewExtendedTestPlatform 创建扩展测试平台
func NewExtendedTestPlatform() *ExtendedTestPlatform {
	cfg := config.GetTestConfig()

	platform := &ExtendedTestPlatform{
		router: mux.NewRouter(),
		cfg:    cfg,
	}

	// 创建负载测试处理器
	platform.loadTestHandler = handlers.NewLoadTestHandler()

	// 设置路由
	platform.setupRoutes()

	// 配置CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// 从环境变量获取端口配置
	mainPort := getEnvInt("TEST_PLATFORM_PORT", 8080)
	mainPort = findAvailablePort(mainPort, 10)
	if mainPort == 0 {
		log.Fatalf("Failed to find available port for main server")
	}

	log.Printf("Main server will use port %d", mainPort)
	platform.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", mainPort),
		Handler:      c.Handler(platform.router),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return platform
}

// setupRoutes 设置路由
func (p *ExtendedTestPlatform) setupRoutes() {
	// API路由
	api := p.router.PathPrefix("/api/v1").Subrouter()

	// 负载测试API - 直接注册到/api/v1路径下
	api.HandleFunc("/tests", p.loadTestHandler.CreateTest).Methods("POST")
	api.HandleFunc("/tests", p.loadTestHandler.ListTests).Methods("GET")
	api.HandleFunc("/tests/{id}", p.loadTestHandler.GetTestStatus).Methods("GET")
	api.HandleFunc("/tests/{id}/start", p.loadTestHandler.StartTest).Methods("POST")
	api.HandleFunc("/tests/{id}/stop", p.loadTestHandler.StopTest).Methods("POST")
	api.HandleFunc("/tests/{id}", p.loadTestHandler.DeleteTest).Methods("DELETE")

	// 静态文件路由
	p.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// 前端页面路由
	p.router.HandleFunc("/", p.serveIndexPage).Methods("GET")
	p.router.HandleFunc("/loadtest", p.serveLoadTestPage).Methods("GET")

	// 健康检查
	api.HandleFunc("/health", p.healthCheckHandler).Methods("GET")
}

// Start 启动平台
func (p *ExtendedTestPlatform) Start() error {
	log.Printf("Starting Extended Test Platform on %s", p.server.Addr)

	// 启动负载测试服务器
	if err := p.loadTestHandler.StartTestServers(); err != nil {
		return fmt.Errorf("failed to start test servers: %v", err)
	}

	// 启动HTTP服务器
	return p.server.ListenAndServe()
}

// Stop 停止平台
func (p *ExtendedTestPlatform) Stop(ctx context.Context) error {
	log.Println("Stopping Extended Test Platform...")

	// 停止负载测试处理器
	if err := p.loadTestHandler.Stop(); err != nil {
		log.Printf("Error stopping load test handler: %v", err)
	}

	// 停止HTTP服务器
	return p.server.Shutdown(ctx)
}

// 页面处理器
func (p *ExtendedTestPlatform) serveIndexPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoSlg 性能测试平台</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 40px; }
        .header h1 { color: #333; margin-bottom: 10px; }
        .header p { color: #666; }
        .feature-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 40px; }
        .feature-card { padding: 20px; border: 1px solid #ddd; border-radius: 8px; }
        .feature-card h3 { color: #2c5aa0; margin-top: 0; }
        .feature-card ul { margin: 0; padding-left: 20px; }
        .feature-card li { margin: 5px 0; }
        .nav-buttons { text-align: center; }
        .btn { display: inline-block; padding: 12px 24px; margin: 0 10px; text-decoration: none; border-radius: 5px; font-weight: bold; }
        .btn-primary { background-color: #2c5aa0; color: white; }
        .btn-secondary { background-color: #6c757d; color: white; }
        .btn:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 GoSlg 性能测试平台</h1>
            <p>支持 WebSocket、gRPC、HTTP API 的综合性能测试解决方案</p>
        </div>
        
        <div class="feature-grid">
            <div class="feature-card">
                <h3>📡 WebSocket 测试</h3>
                <ul>
                    <li>长连接稳定性测试</li>
                    <li>消息吞吐量测试</li>
                    <li>断线重连测试</li>
                    <li>心跳延迟统计</li>
                </ul>
            </div>
            
            <div class="feature-card">
                <h3>⚡ gRPC 测试</h3>
                <ul>
                    <li>高性能RPC调用测试</li>
                    <li>流式接口测试</li>
                    <li>并发连接池测试</li>
                    <li>方法级性能分析</li>
                </ul>
            </div>
            
            <div class="feature-card">
                <h3>🌐 HTTP API 测试</h3>
                <ul>
                    <li>REST API 压力测试</li>
                    <li>多端点负载均衡</li>
                    <li>状态码分析</li>
                    <li>响应时间分布</li>
                </ul>
            </div>
            
            <div class="feature-card">
                <h3>📊 性能指标</h3>
                <ul>
                    <li>实时延迟监控</li>
                    <li>吞吐量统计</li>
                    <li>错误率分析</li>
                    <li>资源使用监控</li>
                </ul>
            </div>
        </div>
        
        <div class="nav-buttons">
            <a href="/loadtest" class="btn btn-primary">🧪 开始负载测试</a>
            <a href="/api/v1/health" class="btn btn-secondary">❤️ 健康检查</a>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (p *ExtendedTestPlatform) serveLoadTestPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>负载测试控制台</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .control-panel { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .test-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .test-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input, .form-group select, .form-group textarea { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: bold; }
        .btn-primary { background-color: #007bff; color: white; }
        .btn-success { background-color: #28a745; color: white; }
        .btn-danger { background-color: #dc3545; color: white; }
        .btn:hover { opacity: 0.9; }
        .results { background: #f8f9fa; padding: 15px; border-radius: 4px; margin-top: 15px; }
        .metric { display: inline-block; margin: 5px 10px; padding: 5px 10px; background: #e9ecef; border-radius: 4px; }
        .status-running { color: #28a745; }
        .status-completed { color: #007bff; }
        .status-failed { color: #dc3545; }
        pre { background: #f8f9fa; padding: 10px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 负载测试控制台</h1>
            <p>创建和管理 HTTP、gRPC、WebSocket 负载测试</p>
        </div>
        
        <div class="test-grid">
            <!-- HTTP 测试 -->
            <div class="test-card">
                <h3>🌐 HTTP API 测试</h3>
                <form id="httpTestForm">
                    <div class="form-group">
                        <label>测试名称:</label>
                        <input type="text" id="httpTestName" value="HTTP API Test" required>
                    </div>
                    <div class="form-group">
                        <label>目标URL:</label>
                        <input type="text" id="httpBaseUrl" value="http://localhost:19000" required>
                    </div>
                    <div class="form-group">
                        <label>并发客户端:</label>
                        <input type="number" id="httpClients" value="10" min="1" max="1000">
                    </div>
                    <div class="form-group">
                        <label>测试时长(秒):</label>
                        <input type="number" id="httpDuration" value="60" min="1" max="3600">
                    </div>
                    <div class="form-group">
                        <label>目标RPS:</label>
                        <input type="number" id="httpRPS" value="100" min="1" max="10000">
                    </div>
                    <button type="submit" class="btn btn-primary">开始测试</button>
                </form>
                <div id="httpResults" class="results" style="display:none;"></div>
            </div>
            
            <!-- gRPC 测试 -->
            <div class="test-card">
                <h3>⚡ gRPC 测试</h3>
                <form id="grpcTestForm">
                    <div class="form-group">
                        <label>测试名称:</label>
                        <input type="text" id="grpcTestName" value="gRPC Service Test" required>
                    </div>
                    <div class="form-group">
                        <label>服务地址:</label>
                        <input type="text" id="grpcServerAddr" value="localhost:19001" required>
                    </div>
                    <div class="form-group">
                        <label>并发客户端:</label>
                        <input type="number" id="grpcClients" value="10" min="1" max="1000">
                    </div>
                    <div class="form-group">
                        <label>测试时长(秒):</label>
                        <input type="number" id="grpcDuration" value="60" min="1" max="3600">
                    </div>
                    <div class="form-group">
                        <label>目标RPS:</label>
                        <input type="number" id="grpcRPS" value="100" min="1" max="10000">
                    </div>
                    <button type="submit" class="btn btn-primary">开始测试</button>
                </form>
                <div id="grpcResults" class="results" style="display:none;"></div>
            </div>
        </div>
        
        <!-- 测试结果展示 -->
        <div class="control-panel">
            <h3>📊 测试结果</h3>
            <button onclick="refreshTests()" class="btn btn-success">刷新测试列表</button>
            <div id="testsList"></div>
        </div>
    </div>

    <script>
        // HTTP 测试表单提交
        document.getElementById('httpTestForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const config = {
                base_url: document.getElementById('httpBaseUrl').value,
                concurrent_clients: parseInt(document.getElementById('httpClients').value),
                target_rps: parseInt(document.getElementById('httpRPS').value),
                endpoints: [
                    { path: '/api/v1/test/fast', method: 'GET', weight: 1 },
                    { path: '/api/v1/test/medium', method: 'GET', weight: 1 },
                    { path: '/api/v1/test/slow', method: 'GET', weight: 1 }
                ]
            };
            
            await createTest('http', document.getElementById('httpTestName').value, config, 
                           parseInt(document.getElementById('httpDuration').value), 'httpResults');
        });
        
        // gRPC 测试表单提交
        document.getElementById('grpcTestForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const config = {
                server_addr: document.getElementById('grpcServerAddr').value,
                concurrent_clients: parseInt(document.getElementById('grpcClients').value),
                target_rps: parseInt(document.getElementById('grpcRPS').value),
                test_methods: ['Login', 'SendPlayerAction', 'GetPlayerStatus']
            };
            
            await createTest('grpc', document.getElementById('grpcTestName').value, config,
                           parseInt(document.getElementById('grpcDuration').value), 'grpcResults');
        });
        
        // 创建测试
        async function createTest(type, name, config, duration, resultElementId) {
            try {
                const response = await fetch('/api/v1/loadtest/tests', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, type, config, duration })
                });
                
                const result = await response.json();
                if (result.success) {
                    showResult(resultElementId, '测试创建成功', result.data);
                    // 自动启动测试
                    await startTest(result.data.test_id, resultElementId);
                } else {
                    showResult(resultElementId, '测试创建失败', result);
                }
            } catch (error) {
                showResult(resultElementId, '请求失败', { error: error.message });
            }
        }
        
        // 启动测试
        async function startTest(testId, resultElementId) {
            try {
                const response = await fetch('/api/v1/loadtest/tests/' + testId + '/start', {
                    method: 'POST'
                });
                
                const result = await response.json();
                if (result.success) {
                    showResult(resultElementId, '测试已启动', result.data);
                    // 开始轮询状态
                    pollTestStatus(testId, resultElementId);
                } else {
                    showResult(resultElementId, '测试启动失败', result);
                }
            } catch (error) {
                showResult(resultElementId, '启动请求失败', { error: error.message });
            }
        }
        
        // 轮询测试状态
        async function pollTestStatus(testId, resultElementId) {
            const interval = setInterval(async () => {
                try {
                    const response = await fetch('/api/v1/loadtest/tests/' + testId);
                    const result = await response.json();
                    
                    if (result.success) {
                        showResult(resultElementId, '测试状态', result.data);
                        
                        // 如果测试完成，停止轮询
                        if (result.data.status === 'completed' || result.data.status === 'stopped') {
                            clearInterval(interval);
                        }
                    }
                } catch (error) {
                    console.error('轮询状态失败:', error);
                }
            }, 2000); // 每2秒轮询一次
        }
        
        // 显示结果
        function showResult(elementId, title, data) {
            const element = document.getElementById(elementId);
            element.style.display = 'block';
            element.innerHTML = '<h4>' + title + '</h4><pre>' + JSON.stringify(data, null, 2) + '</pre>';
        }
        
        // 刷新测试列表
        async function refreshTests() {
            try {
                const response = await fetch('/api/v1/loadtest/tests');
                const result = await response.json();
                
                const testsListElement = document.getElementById('testsList');
                if (result.success && result.data.tests) {
                    let html = '<h4>活跃测试</h4>';
                    result.data.tests.forEach(test => {
                        const statusClass = 'status-' + test.status;
                        html += '<div class="metric">';
                        html += '<strong>' + test.test_id + '</strong> ';
                        html += '<span class="' + statusClass + '">' + test.status + '</span> ';
                        html += '(' + test.type + ')';
                        html += '</div>';
                    });
                    testsListElement.innerHTML = html;
                } else {
                    testsListElement.innerHTML = '<p>暂无测试</p>';
                }
            } catch (error) {
                console.error('刷新测试列表失败:', error);
            }
        }
        
        // 页面加载时刷新测试列表
        refreshTests();
        
        // 每10秒自动刷新测试列表
        setInterval(refreshTests, 10000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (p *ExtendedTestPlatform) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UnixMilli(),
		"version":   "1.0.0",
		"services": map[string]string{
			"loadtest": "running",
			"database": "connected",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 初始化日志
	logger.InitLogger()

	// 创建并启动平台
	platform := NewExtendedTestPlatform()

	// 优雅关闭处理
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Received shutdown signal")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := platform.Stop(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// 启动平台
	if err := platform.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start platform: %v", err)
	}
}
