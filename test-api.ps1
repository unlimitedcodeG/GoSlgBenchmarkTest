# GoSlg 负载测试平台 - 快速测试脚本

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " GoSlg 负载测试平台 - API 测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 健康检查
Write-Host "[1] 检查平台健康状态..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET -ErrorAction Stop
    Write-Host "✓ 平台运行正常" -ForegroundColor Green
} catch {
    Write-Host "✗ 平台未启动或健康检查失败" -ForegroundColor Red
    Write-Host "  请先运行: .\test-platform-extended.exe" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# 2. 创建 HTTP 压力测试
Write-Host "[2] 创建 HTTP API 压力测试..." -ForegroundColor Yellow
$httpTest = @{
    name = "快速HTTP测试"
    type = "http"
    duration = 10
    config = @{
        base_url = "http://localhost:19000"
        concurrent_clients = 5
        target_rps = 50
        endpoints = @(
            @{
                path = "/api/v1/test/fast"
                method = "GET"
                weight = 1
            }
        )
    }
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" `
        -Method POST -Body $httpTest -ContentType "application/json" -ErrorAction Stop
    
    Write-Host "✓ HTTP 测试创建成功" -ForegroundColor Green
    Write-Host "  测试ID: $($response.test_id)" -ForegroundColor Gray
    $httpTestId = $response.test_id
} catch {
    Write-Host "✗ 创建 HTTP 测试失败: $_" -ForegroundColor Red
}

Write-Host ""

# 3. 创建 gRPC 压力测试  
Write-Host "[3] 创建 gRPC 服务压力测试..." -ForegroundColor Yellow
$grpcTest = @{
    name = "快速gRPC测试"
    type = "grpc"
    duration = 10
    config = @{
        server_addr = "localhost:19001"
        concurrent_clients = 3
        target_rps = 30
        test_methods = @("Login", "SendPlayerAction")
    }
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" `
        -Method POST -Body $grpcTest -ContentType "application/json" -ErrorAction Stop
    
    Write-Host "✓ gRPC 测试创建成功" -ForegroundColor Green
    Write-Host "  测试ID: $($response.test_id)" -ForegroundColor Gray
    $grpcTestId = $response.test_id
} catch {
    Write-Host "✗ 创建 gRPC 测试失败: $_" -ForegroundColor Red
}

Write-Host ""

# 4. 查看测试状态
Write-Host "[4] 查看所有测试状态..." -ForegroundColor Yellow
try {
    $tests = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/tests" -Method GET -ErrorAction Stop
    
    if ($tests.data -and $tests.data.Count -gt 0) {
        Write-Host "✓ 找到 $($tests.data.Count) 个测试" -ForegroundColor Green
        foreach ($test in $tests.data) {
            Write-Host "  - $($test.name) [$($test.type)] - 状态: $($test.status)" -ForegroundColor Gray
        }
    } else {
        Write-Host "  暂无测试" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ 获取测试列表失败: $_" -ForegroundColor Red
}

Write-Host ""

# 5. 检查服务器状态
Write-Host "[5] 检查测试服务器状态..." -ForegroundColor Yellow
try {
    $servers = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/loadtest/servers" -Method GET -ErrorAction Stop
    Write-Host "✓ 测试服务器状态:" -ForegroundColor Green
    Write-Host "  - HTTP 服务器: 端口 $($servers.data.servers.http.port) - $($servers.data.servers.http.status)" -ForegroundColor Gray
    Write-Host "  - gRPC 服务器: 端口 $($servers.data.servers.grpc.port) - $($servers.data.servers.grpc.status)" -ForegroundColor Gray
} catch {
    Write-Host "✗ 获取服务器状态失败: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host " 测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示：" -ForegroundColor Yellow
Write-Host "  1. 访问 http://localhost:8080 查看Web界面" -ForegroundColor Gray
Write-Host "  2. 访问 http://localhost:8080/loadtest 查看负载测试控制台" -ForegroundColor Gray
Write-Host "  3. 查看 README-LOADTEST-WINDOWS.md 了解更多功能" -ForegroundColor Gray
Write-Host ""