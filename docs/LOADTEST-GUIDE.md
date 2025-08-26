# GoSlg 负载测试平台使用指南

## 📋 概述

GoSlg 负载测试平台是一个综合性的性能测试解决方案，支持 **WebSocket**、**gRPC** 和 **HTTP API** 三种协议的压力测试。平台提供直观的 Web 界面，实时性能监控，以及详细的测试报告。

## 🚀 快速开始

### 1. 启动平台

**Linux/macOS:**

```bash
chmod +x scripts/run-loadtest-platform.sh
./scripts/run-loadtest-platform.sh
```

**Windows:**

```powershell
.\scripts\run-loadtest-platform.ps1
```

### 2. 访问 Web 界面

- **主页**: <http://localhost:8080>
- **负载测试控制台**: <http://localhost:8080/loadtest>
- **健康检查**: <http://localhost:8080/api/v1/health>

### 3. 内置测试服务器

平台会自动启动测试服务器：

- **HTTP 测试服务器**: <http://localhost:19000>
- **gRPC 测试服务器**: localhost:19001

## 🧪 支持的测试类型

### 1. HTTP API 压力测试

#### 特性

- ✅ 多端点并发测试
- ✅ 请求权重分配
- ✅ 模板变量支持
- ✅ 自定义请求头和Body
- ✅ HTTP状态码统计
- ✅ 响应时间分布分析

#### 配置示例

```json
{
  "base_url": "http://localhost:19000",
  "concurrent_clients": 50,
  "target_rps": 500,
  "duration": 120,
  "endpoints": [
    {
      "path": "/api/v1/players",
      "method": "GET",
      "weight": 2,
      "query_params": {
        "page": "{{client_id}}",
        "limit": "10"
      }
    },
    {
      "path": "/api/v1/players",
      "method": "POST",
      "weight": 1,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "name": "player_{{client_id}}_{{request_id}}",
        "level": 1
      }
    }
  ]
}
```

#### 支持的模板变量

- `{{client_id}}`: 客户端ID
- `{{request_id}}`: 请求ID
- `{{timestamp}}`: Unix时间戳
- `{{timestamp_ms}}`: 毫秒时间戳
- `{{random_id}}`: 随机ID

### 2. gRPC 服务压力测试

#### gRPC特性

- ✅ 多种RPC方法测试
- ✅ 流式RPC支持
- ✅ 连接池管理
- ✅ Keep-Alive配置
- ✅ 方法级性能分析

#### 配置示例

```json
{
  "server_addr": "localhost:19001",
  "concurrent_clients": 20,
  "target_rps": 200,
  "duration": 60,
  "test_methods": [
    "Login",
    "SendPlayerAction", 
    "GetPlayerStatus",
    "GetBattleStatus"
  ],
  "request_timeout": "10s",
  "keep_alive_time": "30s",
  "max_connections": 5
}
```

#### 支持的测试方法

- `Login`: 用户登录
- `SendPlayerAction`: 玩家操作
- `GetPlayerStatus`: 获取玩家状态  
- `GetBattleStatus`: 获取战斗状态
- `StreamBattleUpdates`: 流式战斗更新

### 3. WebSocket 长连接测试

#### 特性  

- ✅ 长连接稳定性测试
- ✅ 自动断线重连
- ✅ 心跳延迟统计
- ✅ 消息序列号验证
- ✅ 实时推送测试

## 📊 性能指标

### 核心指标

#### 延迟指标

- **平均响应时间**: 所有请求的平均延迟
- **P50延迟**: 50%请求的响应时间
- **P95延迟**: 95%请求的响应时间  
- **P99延迟**: 99%请求的响应时间
- **最小/最大延迟**: 延迟范围

#### 吞吐量指标

- **请求/秒 (RPS)**: 每秒处理的请求数
- **字节/秒**: 网络吞吐量
- **并发连接数**: 活跃连接数量

#### 可靠性指标

- **成功率**: 成功请求百分比
- **错误率**: 失败请求百分比
- **状态码分布**: HTTP状态码统计
- **错误类型分析**: 详细错误分类

### 时间序列数据

- **实时延迟曲线**: 响应时间趋势
- **吞吐量曲线**: RPS变化趋势
- **错误率曲线**: 错误发生趋势

## 🎛️ API 接口文档

### 测试管理 API

#### 创建测试

```http
POST /api/v1/loadtest/tests
Content-Type: application/json

{
  "name": "测试名称",
  "type": "http|grpc|websocket", 
  "config": {...},
  "duration": 60
}
```

#### 启动测试

```http
POST /api/v1/loadtest/tests/{test_id}/start
```

#### 获取测试状态

```http
GET /api/v1/loadtest/tests/{test_id}
```

#### 停止测试

```http
POST /api/v1/loadtest/tests/{test_id}/stop
```

#### 列出测试

```http
GET /api/v1/loadtest/tests?type=http&status=running&page=1&page_size=20
```

#### 删除测试

```http
DELETE /api/v1/loadtest/tests/{test_id}
```

### 响应格式

```json
{
  "success": true,
  "data": {
    "test_id": "http_1640995200000",
    "type": "http",
    "status": "running",
    "start_time": 1640995200000,
    "metrics": {
      "total_requests": 1500,
      "successful_requests": 1485,
      "failed_requests": 15,
      "avg_latency": 45.2,
      "requests_per_second": 125.5,
      "success_rate": 0.99
    }
  },
  "timestamp": 1640995260000
}
```

## 🔧 配置说明

### 全局配置文件

配置文件位置: `configs/test-config.yaml`

#### HTTP 负载测试配置

```yaml
http_loadtest:
  default_config:
    concurrent_clients: 10
    duration: "60s"
    target_rps: 100
    timeout: "30s"
    keep_alive: true
    max_idle_conns: 100
    max_conns_per_host: 50
  performance_thresholds:
    max_avg_latency: "50ms"
    max_p95_latency: "100ms"  
    max_p99_latency: "200ms"
    min_success_rate: 0.98
```

#### gRPC 负载测试配置

```yaml
grpc_loadtest:
  default_config:
    concurrent_clients: 10
    duration: "60s" 
    target_rps: 100
    request_timeout: "10s"
    keep_alive_time: "30s"
    max_connections: 5
  performance_thresholds:
    max_avg_latency: "100ms"
    max_p95_latency: "200ms"
    min_success_rate: 0.95
```

### 环境变量

- `TEST_LOG_LEVEL`: 日志级别 (debug, info, warn, error)
- `GOMAXPROCS`: Go 运行时处理器数量
- `TEST_HTTP_PORT`: HTTP 服务端口 (默认: 8080)
- `TEST_HTTP_SERVER_PORT`: HTTP 测试服务器端口 (默认: 19000)
- `TEST_GRPC_SERVER_PORT`: gRPC 测试服务器端口 (默认: 19001)

## 📈 测试场景示例

### 场景1: 电商API压力测试

```json
{
  "name": "电商API压力测试",
  "type": "http",
  "duration": 300,
  "config": {
    "base_url": "https://api.shop.com",
    "concurrent_clients": 100,
    "target_rps": 1000,
    "auth_type": "bearer",
    "auth_token": "your-jwt-token",
    "endpoints": [
      {
        "path": "/api/v1/products",
        "method": "GET", 
        "weight": 50,
        "query_params": {
          "category": "electronics",
          "page": "{{random_id}}"
        }
      },
      {
        "path": "/api/v1/cart/items",
        "method": "POST",
        "weight": 30,
        "body": {
          "product_id": "{{random_id}}",
          "quantity": 1
        }
      },
      {
        "path": "/api/v1/orders",
        "method": "POST", 
        "weight": 20,
        "body": {
          "user_id": "user_{{client_id}}",
          "items": [
            {
              "product_id": "{{random_id}}",
              "quantity": 1
            }
          ]
        }
      }
    ]
  }
}
```

### 场景2: 游戏服务gRPC测试

```json
{
  "name": "游戏服务压力测试",
  "type": "grpc",
  "duration": 180,
  "config": {
    "server_addr": "game.example.com:9090",
    "concurrent_clients": 50,
    "target_rps": 500,
    "tls": true,
    "test_methods": [
      "Login",
      "SendPlayerAction",
      "GetPlayerStatus", 
      "JoinBattle"
    ],
    "request_timeout": "5s",
    "keep_alive_time": "30s"
  }
}
```

### 场景3: 混合协议压力测试

可以同时运行多个不同协议的测试：

```bash
# 启动 HTTP 测试
curl -X POST http://localhost:8080/api/v1/loadtest/tests \
  -H "Content-Type: application/json" \
  -d '{"name":"HTTP测试","type":"http","config":{...},"duration":300}'

# 启动 gRPC 测试  
curl -X POST http://localhost:8080/api/v1/loadtest/tests \
  -H "Content-Type: application/json" \
  -d '{"name":"gRPC测试","type":"grpc","config":{...},"duration":300}'
```

## 🐛 故障排除

### 常见问题

#### 1. 连接超时

**问题**: HTTP/gRPC 连接超时
**解决**:

- 检查目标服务器是否可达
- 增加 `timeout` 配置
- 检查防火墙设置

#### 2. 高延迟

**问题**: 响应时间过高
**解决**:

- 减少并发客户端数量
- 降低目标 RPS
- 检查网络带宽
- 优化目标服务器性能

#### 3. 内存不足

**问题**: 测试过程中内存占用过高
**解决**:

- 减少 `max_idle_conns` 配置
- 缩短测试持续时间
- 调整 `GOMAXPROCS` 环境变量

#### 4. gRPC 连接失败

**问题**: gRPC 服务连接失败
**解决**:

- 检查服务器地址和端口
- 确认是否需要 TLS 配置
- 验证 protobuf 定义是否匹配

### 性能调优建议

#### 系统级优化

```bash
# 增加文件描述符限制
ulimit -n 65536

# 调整TCP参数
echo 'net.core.somaxconn = 65536' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65536' >> /etc/sysctl.conf
sysctl -p
```

#### 应用级优化

- 合理设置并发客户端数量 (CPU核心数 × 10-50)
- 优化连接池大小
- 启用 HTTP Keep-Alive
- 使用连接复用

## 📚 进阶用法

### 自定义测试端点

#### 创建自定义 HTTP 测试服务器

```go
// 示例：创建自定义测试端点
func customHandler(w http.ResponseWriter, r *http.Request) {
    // 模拟不同的响应时间
    delay := time.Duration(rand.Intn(100)) * time.Millisecond
    time.Sleep(delay)
    
    // 返回JSON响应
    response := map[string]interface{}{
        "status": "success",
        "timestamp": time.Now().UnixMilli(),
        "data": generateTestData(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

#### 扩展 gRPC 测试方法

```go
// 在 internal/grpcserver/game_server.go 中添加新方法
func (s *GameServer) CustomTestMethod(ctx context.Context, req *CustomRequest) (*CustomResponse, error) {
    // 自定义测试逻辑
    return &CustomResponse{
        Success: true,
        Data: "test data",
    }, nil
}
```

### 集成到 CI/CD

#### GitHub Actions 示例

```yaml
name: 性能测试
on: [push, pull_request]

jobs:
  loadtest:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
        
    - name: 运行负载测试
      run: |
        ./scripts/run-loadtest-platform.sh --build-only
        ./test-platform-extended &
        sleep 10
        
        # 运行自动化测试
        curl -X POST http://localhost:8080/api/v1/loadtest/tests \
          -H "Content-Type: application/json" \
          -d '{"name":"CI测试","type":"http","config":{"target_rps":100},"duration":30}'
```

### 监控集成

#### Prometheus 指标导出

```go
// 添加 Prometheus 指标收集
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "loadtest_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"test_type", "endpoint"},
    )
)
```

## 🤝 贡献指南

### 开发环境设置

```bash
# 克隆项目
git clone <repository-url>
cd GoSlgBenchmarkTest

# 安装依赖
go mod download

# 生成 protobuf 代码
buf generate

# 运行测试
go test ./...

# 构建项目
go build -o test-platform-extended cmd/test-platform/main_extended.go
```

### 添加新的测试协议

1. 在 `internal/loadtest/` 中创建新的客户端实现
2. 在 `api/handlers/loadtest_handlers.go` 中添加处理逻辑
3. 更新配置文件和文档
4. 添加相应的测试用例

## 📞 支持与反馈

如果您在使用过程中遇到问题或有改进建议，请：

1. 查看本文档的故障排除部分
2. 检查项目的 Issue 列表  
3. 创建新的 Issue 描述问题
4. 提交 Pull Request 贡献代码

---

**Happy Testing! 🚀**
