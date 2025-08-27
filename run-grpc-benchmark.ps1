#!/usr/bin/env pwsh

# gRPC 压测脚本 - 1000并发测试
Write-Host "🎯 gRPC 压测脚本 - 1000并发测试" -ForegroundColor Green
Write-Host "=" * 50 -ForegroundColor Yellow

# 配置参数
$GRPC_PORT = "19001"
$CONCURRENCY = 1000
$TOTAL_REQUESTS = 10000
$DURATION = "60s"
$PROTO_FILE = "proto/game/v1/game_service.proto"
$IMPORT_PATH = "proto/game/v1"
$SERVICE_METHOD = "game.v1.GameService.Login"

# 测试数据文件
$TEST_DATA = @'
{
  "token": "test-token-{concurrency}-{worker}",
  "client_version": "1.0.0",
  "device_id": "device-test-{worker}"
}
'@

Write-Host "📋 压测配置:" -ForegroundColor Cyan
Write-Host "  • 服务地址: localhost:$GRPC_PORT" -ForegroundColor White
Write-Host "  • 并发数: $CONCURRENCY" -ForegroundColor White
Write-Host "  • 总请求数: $TOTAL_REQUESTS" -ForegroundColor White
Write-Host "  • 持续时间: $DURATION" -ForegroundColor White
Write-Host "  • 测试方法: $SERVICE_METHOD" -ForegroundColor White
Write-Host ""

# 检查服务器是否运行
Write-Host "🔍 检查gRPC服务器状态..." -ForegroundColor Yellow
$serverCheck = Test-NetConnection -ComputerName localhost -Port $GRPC_PORT -InformationLevel Quiet
if (-not $serverCheck) {
    Write-Host "❌ gRPC服务器未运行，请先启动服务器" -ForegroundColor Red
    Write-Host "启动命令: go run tools/grpc-server.go grpc" -ForegroundColor Yellow
    exit 1
}
Write-Host "✅ gRPC服务器正在运行" -ForegroundColor Green

# 启动gRPC服务器（如果需要）
Write-Host "🚀 启动gRPC服务器..." -ForegroundColor Yellow
Start-Process -FilePath "go" -ArgumentList "run", "tools/grpc-server.go", "grpc" -NoNewWindow

# 等待服务器启动
Start-Sleep -Seconds 3

Write-Host "🧪 开始执行压测..." -ForegroundColor Green
Write-Host ""

# 执行ghz压测
$ghzCommand = @"
ghz --insecure `
    --proto $PROTO_FILE `
    --import-paths $IMPORT_PATH `
    --call $SERVICE_METHOD `
    --concurrency $CONCURRENCY `
    --total $TOTAL_REQUESTS `
    --duration $DURATION `
    --timeout 10s `
    --dial-timeout 5s `
    --keepalive 30s `
    --enable-compression `
    --data '$TEST_DATA' `
    localhost:$GRPC_PORT
"@

Write-Host "执行命令:" -ForegroundColor Cyan
Write-Host $ghzCommand -ForegroundColor Gray
Write-Host ""

# 执行压测
try {
    Invoke-Expression $ghzCommand
} catch {
    Write-Host "❌ 压测执行失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✅ 压测完成" -ForegroundColor Green