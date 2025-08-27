#!/bin/bash

# gRPC 压测脚本 - 1000并发测试
echo "🎯 gRPC 压测脚本 - 1000并发测试"
echo "=================================================="

# 配置参数
GRPC_PORT="19001"
CONCURRENCY=1000
TOTAL_REQUESTS=10000
DURATION="60s"
PROTO_FILE="proto/game/v1/game_service.proto"
IMPORT_PATH="proto/game/v1"
SERVICE_METHOD="game.v1.GameService.Login"

# 测试数据
TEST_DATA='{
  "token": "test-token-'$CONCURRENCY'-worker",
  "client_version": "1.0.0",
  "device_id": "device-test-worker"
}'

echo "📋 压测配置:"
echo "  • 服务地址: localhost:$GRPC_PORT"
echo "  • 并发数: $CONCURRENCY"
echo "  • 总请求数: $TOTAL_REQUESTS"
echo "  • 持续时间: $DURATION"
echo "  • 测试方法: $SERVICE_METHOD"
echo ""

# 检查服务器是否运行
echo "🔍 检查gRPC服务器状态..."
if ! nc -z localhost $GRPC_PORT 2>/dev/null; then
    echo "❌ gRPC服务器未运行，请先启动服务器"
    echo "启动命令: go run tools/grpc-server.go grpc"
    exit 1
fi
echo "✅ gRPC服务器正在运行"

# 启动gRPC服务器（后台运行）
echo "🚀 启动gRPC服务器..."
go run tools/grpc-server.go grpc &
SERVER_PID=$!

# 等待服务器启动
sleep 3

echo "🧪 开始执行压测..."
echo ""

# 执行ghz压测
ghz --insecure \
    --proto "$PROTO_FILE" \
    --import-paths "$IMPORT_PATH" \
    --call "$SERVICE_METHOD" \
    --concurrency "$CONCURRENCY" \
    --total "$TOTAL_REQUESTS" \
    --duration "$DURATION" \
    --timeout 10s \
    --dial-timeout 5s \
    --keepalive 30s \
    --enable-compression \
    --data "$TEST_DATA" \
    "localhost:$GRPC_PORT"

# 清理后台服务器进程
if [ ! -z "$SERVER_PID" ]; then
    echo ""
    echo "🧹 清理服务器进程..."
    kill $SERVER_PID 2>/dev/null
fi

echo ""
echo "✅ 压测完成"