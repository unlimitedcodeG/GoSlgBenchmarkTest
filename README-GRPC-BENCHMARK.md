# gRPC 压测指南

## 🚀 快速开始

### 1. 启动gRPC服务器
```bash
# Windows
go run tools/grpc-server.go grpc

# Linux/macOS
./run-grpc-benchmark.sh
```

### 2. 运行压测
```bash
# 使用配置文件 (推荐)
ghz --config ghz-config.json

# 或直接命令行
ghz --insecure \
    --proto proto/game/v1/game_service.proto \
    --import-paths proto/game/v1 \
    --call game.v1.GameService.Login \
    --concurrency 1000 \
    --total 10000 \
    --duration 60s \
    --timeout 10s \
    --keepalive 30s \
    --data '{"token": "test-token-12345", "client_version": "1.0.0", "device_id": "device-test-001"}' \
    localhost:19001
```

## 📊 最新压测结果

- **QPS**: 80,975.72
- **平均延迟**: 12.21ms
- **成功率**: 99.98%
- **并发数**: 1000

## 🛠️ 配置文件说明

### ghz-config.json
```json
{
  "name": "grpc-benchmark-test",
  "proto": "proto/game/v1/game_service.proto",
  "import-paths": ["proto/game/v1"],
  "call": "game.v1.GameService.Login",
  "host": "localhost:19001",
  "concurrency": 1000,
  "total": 10000,
  "duration": "60s",
  "timeout": "10s",
  "keepalive": "30s",
  "data": {
    "token": "test-token-12345",
    "client_version": "1.0.0",
    "device_id": "device-test-001"
  },
  "insecure": true,
  "connections": 100,
  "connect-timeout": "5s"
}
```

## 🎯 测试的其他接口

### GetPlayerStatus
```bash
ghz --insecure \
    --proto proto/game/v1/game_service.proto \
    --import-paths proto/game/v1 \
    --call game.v1.GameService.GetPlayerStatus \
    --concurrency 1000 \
    --duration 60s \
    --data '{"player_id": "player_test"}' \
    localhost:19001
```

### SendPlayerAction
```bash
ghz --insecure \
    --proto proto/game/v1/game_service.proto \
    --import-paths proto/game/v1 \
    --call game.v1.GameService.SendPlayerAction \
    --concurrency 1000 \
    --duration 60s \
    --data '{
      "action_seq": 1,
      "player_id": "player_test",
      "action_type": "ACTION_TYPE_MOVE",
      "action_data": {
        "move": {
          "target_position": {"x": 10, "y": 20, "z": 0},
          "move_speed": 5.0
        }
      },
      "client_timestamp": 1234567890
    }' \
    localhost:19001
```

## 📈 性能调优建议

### 服务器端优化
1. **KeepAlive配置**: 已优化为30s空闲时间
2. **并发流限制**: 设置为10,000个
3. **消息大小**: 最大4MB
4. **连接超时**: 10s

### 客户端优化
1. **连接池**: 使用100个连接处理1000并发
2. **KeepAlive**: 30s保活时间
3. **超时设置**: 10s请求超时

## 📋 故障排除

### 常见问题
1. **"connection closed" 错误**: 正常现象，发生在测试结束时
2. **"Unimplemented" 错误**: 检查方法名是否正确
3. **压缩相关错误**: 已移除压缩配置解决

### 调试技巧
```bash
# 启用详细日志
ghz --config ghz-config.json --verbose

# 测试单个请求
ghz --config ghz-config.json --concurrency 1 --total 1
```

## 📊 监控指标

- **QPS**: 每秒查询数
- **Latency**: 响应延迟分布
- **Error Rate**: 错误率
- **CPU使用率**: 服务器CPU占用
- **内存使用**: 服务器内存占用

## 🔧 自定义配置

### 修改并发数
```json
{
  "concurrency": 500,  // 降低到500并发
  "total": 5000        // 相应减少总请求数
}
```

### 修改测试时长
```json
{
  "duration": "30s",   // 30秒测试
  "concurrency": 1000  // 保持1000并发
}
```

### 修改测试数据
```json
{
  "data": {
    "token": "your-custom-token",
    "client_version": "2.0.0",
    "device_id": "custom-device"
  }
}
```

## 📝 完整报告

详细的压测报告请查看 `benchmark-report.md`。