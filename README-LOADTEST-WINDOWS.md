# GoSlg 负载测试平台 - Windows 使用指南

## 🎯 项目概述

GoSlg 负载测试平台现已全面支持 **gRPC**、**HTTP API** 和 **WebSocket** 三种协议的压力测试！

### ✨ 新增功能

- **🌐 HTTP API 压测**: 支持 REST API、多端点、权重分配、模板变量
- **⚡ gRPC 服务压测**: 支持 RPC 方法测试、流式接口、连接池管理  
- **📊 详细性能指标**: 延迟分布、吞吐量、错误率、时间序列数据
- **🎛️ Web 管理界面**: 直观的测试创建、监控和结果展示
- **📈 实时监控**: 测试状态、性能曲线、错误统计实时更新

## 🚀 Windows 快速启动

### 方式一：使用 PowerShell 脚本（推荐）

```powershell
# 在项目根目录打开 PowerShell
.\scripts\run-loadtest-platform.ps1
```

**可选参数：**
```powershell
# 仅检查依赖
.\scripts\run-loadtest-platform.ps1 -CheckDeps

# 仅构建项目
.\scripts\run-loadtest-platform.ps1 -BuildOnly

# 生成 protobuf 代码并启动
.\scripts\run-loadtest-platform.ps1 -GenProto

# 显示帮助
.\scripts\run-loadtest-platform.ps1 -Help
```

### 方式二：手动启动

```powershell
# 1. 构建项目
go build -o test-platform-extended.exe cmd/test-platform/main_extended.go

# 2. 启动平台
.\test-platform-extended.exe
```

## 🌟 访问地址

启动成功后，您可以访问：

- **🏠 主页**: http://localhost:8080
- **🧪 负载测试控制台**: http://localhost:8080/loadtest  
- **❤️ 健康检查**: http://localhost:8080/api/v1/health
- **🌐 HTTP 测试服务器**: http://localhost:19000
- **⚡ gRPC 测试服务器**: localhost:19001

## 📋 系统要求

### 必需依赖
- **Go 1.21+**: 从 https://golang.org/dl/ 下载安装
- **Git**: 用于克隆代码仓库

### 可选依赖  
- **Protocol Buffers**: https://github.com/protocolbuffers/protobuf/releases
- **Buf**: https://buf.build/docs/installation （更好的 protobuf 工具）

## 🧪 快速测试示例

### 1. HTTP API 压力测试

启动平台后，在浏览器访问 http://localhost:8080/loadtest，或使用 API：

```powershell
# 创建 HTTP 压力测试
$body = @{
    name = "HTTP API 压力测试"
    type = "http"
    duration = 60
    config = @{
        base_url = "http://localhost:19000"
        concurrent_clients = 20
        target_rps = 200
        endpoints = @(
            @{
                path = "/api/v1/test/fast"
                method = "GET"
                weight = 1
            },
            @{
                path = "/api/v1/test/medium" 
                method = "GET"
                weight = 1
            }
        )
    }
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" -Method POST -Body $body -ContentType "application/json"
```

### 2. gRPC 服务压力测试

```powershell
# 创建 gRPC 压力测试
$body = @{
    name = "gRPC 服务压力测试"
    type = "grpc"
    duration = 60
    config = @{
        server_addr = "localhost:19001"
        concurrent_clients = 10
        target_rps = 100
        test_methods = @("Login", "SendPlayerAction", "GetPlayerStatus")
    }
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" -Method POST -Body $body -ContentType "application/json"
```

### 3. 查看测试结果

```powershell
# 获取所有测试
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" -Method GET

# 获取特定测试状态（替换 test_id）
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests/{test_id}" -Method GET
```

## 📊 性能指标说明

### 核心指标
- **总请求数**: 测试期间发送的请求总数
- **成功率**: 成功请求 / 总请求 × 100%
- **平均延迟**: 所有请求的平均响应时间
- **P95/P99延迟**: 95%/99% 的请求响应时间
- **吞吐量(RPS)**: 每秒处理的请求数

### HTTP 特有指标
- **状态码分布**: 200, 404, 500 等状态码统计
- **字节吞吐量**: 网络传输字节数/秒
- **连接复用率**: Keep-Alive 连接复用情况

### gRPC 特有指标  
- **方法级统计**: 每个 RPC 方法的独立性能数据
- **连接池效率**: 连接创建/复用统计
- **流式RPC性能**: 流式接口的特殊指标

## 🔧 配置文件

主配置文件: `configs/test-config.yaml`

### 修改 HTTP 测试默认配置
```yaml
http_loadtest:
  default_config:
    concurrent_clients: 20        # 并发客户端数
    duration: "120s"             # 测试持续时间
    target_rps: 500              # 目标 RPS
    timeout: "30s"               # 请求超时
    keep_alive: true             # 启用 Keep-Alive
    max_idle_conns: 100          # 最大空闲连接
  performance_thresholds:
    max_avg_latency: "50ms"      # 平均延迟阈值
    max_p99_latency: "200ms"     # P99延迟阈值
    min_success_rate: 0.98       # 最小成功率
```

### 修改 gRPC 测试默认配置
```yaml
grpc_loadtest:
  default_config:
    concurrent_clients: 15
    duration: "90s"
    target_rps: 300
    request_timeout: "10s"
    keep_alive_time: "30s" 
    max_connections: 8
  test_methods:                  # 测试的 RPC 方法
    - "Login"
    - "SendPlayerAction"
    - "GetPlayerStatus"
    - "GetBattleStatus"
```

## 🎛️ 环境变量配置

在 PowerShell 中设置环境变量：

```powershell
# 设置日志级别
$env:TEST_LOG_LEVEL = "debug"

# 设置 Go 运行时参数
$env:GOMAXPROCS = "8"

# 设置服务端口（可选）
$env:TEST_HTTP_PORT = "8080"
$env:TEST_HTTP_SERVER_PORT = "19000"
$env:TEST_GRPC_SERVER_PORT = "19001"

# 启动平台
.\test-platform-extended.exe
```

## 🐛 Windows 特定故障排除

### 1. 端口被占用
```powershell
# 查看端口占用
netstat -ano | findstr :8080
netstat -ano | findstr :19000
netstat -ano | findstr :19001

# 杀死占用进程（替换 PID）
taskkill /PID <PID> /F
```

### 2. 防火墙问题
```powershell
# 临时关闭 Windows 防火墙（管理员权限）
netsh advfirewall set allprofiles state off

# 添加防火墙规则（管理员权限）
netsh advfirewall firewall add rule name="GoSlg LoadTest" dir=in action=allow protocol=TCP localport=8080,19000,19001
```

### 3. 权限问题
确保在有写入权限的目录运行，或以管理员身份运行 PowerShell。

### 4. Go 模块问题
```powershell
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 更新依赖
go mod tidy
```

## 📈 高级使用场景

### 场景1: 电商网站压力测试

```json
{
  "name": "电商网站全链路压测",
  "type": "http", 
  "duration": 300,
  "config": {
    "base_url": "https://your-ecommerce-api.com",
    "concurrent_clients": 100,
    "target_rps": 1000,
    "auth_type": "bearer",
    "auth_token": "your-jwt-token",
    "endpoints": [
      {
        "path": "/api/v1/products",
        "method": "GET",
        "weight": 40,
        "query_params": {
          "category": "electronics",
          "page": "{{random_id}}"
        }
      },
      {
        "path": "/api/v1/cart/add",
        "method": "POST", 
        "weight": 30,
        "body": {
          "product_id": "{{random_id}}",
          "quantity": 1,
          "user_id": "user_{{client_id}}"
        }
      },
      {
        "path": "/api/v1/orders",
        "method": "POST",
        "weight": 30,
        "body": {
          "cart_id": "cart_{{client_id}}",
          "payment_method": "credit_card"
        }
      }
    ]
  }
}
```

### 场景2: 微服务架构 gRPC 测试

```json
{
  "name": "微服务gRPC压力测试",
  "type": "grpc",
  "duration": 180,
  "config": {
    "server_addr": "microservice.example.com:9090",
    "concurrent_clients": 50,
    "target_rps": 800,
    "tls": true,
    "test_methods": [
      "UserService.GetUser",
      "UserService.UpdateUser", 
      "OrderService.CreateOrder",
      "OrderService.GetOrderStatus"
    ],
    "request_timeout": "5s",
    "keep_alive_time": "60s",
    "max_connections": 10
  }
}
```

## 🔄 CI/CD 集成

### GitHub Actions Windows 环境

```yaml
name: 性能测试 (Windows)
on: [push, pull_request]

jobs:
  loadtest-windows:
    runs-on: windows-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
        
    - name: 构建测试平台
      run: |
        go build -o test-platform-extended.exe cmd/test-platform/main_extended.go
        
    - name: 运行性能测试
      run: |
        # 后台启动平台
        Start-Process -FilePath ".\test-platform-extended.exe" -NoNewWindow
        Start-Sleep -Seconds 15
        
        # 执行自动化测试
        $testConfig = @{
          name = "CI性能测试"
          type = "http"
          duration = 30
          config = @{
            base_url = "http://localhost:19000"
            concurrent_clients = 10
            target_rps = 100
          }
        } | ConvertTo-Json -Depth 5
        
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" -Method POST -Body $testConfig -ContentType "application/json"
        
        # 启动测试
        Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests/$($response.data.test_id)/start" -Method POST
        
        # 等待测试完成
        Start-Sleep -Seconds 35
        
        # 获取结果
        $result = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests/$($response.data.test_id)" -Method GET
        Write-Output "测试结果: $($result | ConvertTo-Json -Depth 5)"
      shell: powershell
```

## 📞 技术支持

如果您在 Windows 环境下遇到问题：

1. **查看日志**: 启动时添加 `-v` 参数查看详细日志
2. **检查端口**: 确保 8080、19000、19001 端口未被占用
3. **验证Go环境**: 运行 `go version` 确认 Go 已正确安装
4. **网络检查**: 确保防火墙允许相关端口访问

## 🎉 开始您的性能测试之旅

现在您已经在 Windows 环境下成功配置了 GoSlg 负载测试平台！

**下一步建议：**
1. 访问 http://localhost:8080/loadtest 体验 Web 界面
2. 尝试创建您的第一个 HTTP API 压力测试  
3. 探索 gRPC 服务压力测试功能
4. 查看详细的性能指标和报告

Happy Testing on Windows! 🚀🪟