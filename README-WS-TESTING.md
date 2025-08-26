# WebSocket集成测试指南

## 📋 概述

本文档介绍如何在本地开发环境和GitHub CI环境中运行WebSocket集成测试。

## 🚀 快速开始

### 本地开发环境

#### 方法1: 使用测试脚本 (推荐)
```bash
# 克隆项目
git clone <repository-url>
cd GoSlgBenchmarkTest

# 运行WebSocket测试
./scripts/run-websocket-tests.sh -l

# 或指定端口和超时时间
./scripts/run-websocket-tests.sh -p 18090 -t 15m
```

#### 方法2: 手动启动
```bash
# 终端1: 启动WebSocket服务器
go run tools/grpc-server.go websocket

# 终端2: 运行WebSocket测试
go test ./test/session -v -run TestTimelineAnalysisRealWallClock -timeout=10m
```

### GitHub CI环境

#### 自动触发
- **Push到主分支**: 自动运行完整测试套件
- **PR标记触发**: 为PR添加`websocket-test`标签来触发WebSocket测试

#### 手动触发
```yaml
# 在GitHub Actions中查看 WebSocket Integration Tests 工作流
name: WebSocket Integration Tests
```

## 🧪 测试内容

### 1. 5分钟墙钟时间测试
- **测试名称**: `TestTimelineAnalysisRealWallClock`
- **持续时间**: 5分钟真实时间
- **消息数量**: ~3000个消息
- **验证指标**:
  - 业务成功率 ≥90%
  - 传输丢包率 =0%
  - 应用超时率 =10% (预设值)
  - P95延迟 ≤300ms
  - P99延迟 ≤400ms

### 2. 会话录制和回放测试
- **测试名称**: `TestSessionRecordingAndReplay`
- **功能**: 验证WebSocket会话的录制和回放机制
- **验证**: 事件序列一致性、消息完整性

### 3. 会话断言测试
- **测试名称**: `TestSessionAssertions`
- **功能**: 验证各种会话质量断言
- **包含断言**:
  - 消息顺序断言
  - 延迟断言
  - 重连断言
  - 错误率断言

## 🏗️ 项目结构

```
GoSlgBenchmarkTest/
├── tools/
│   └── grpc-server.go          # 多功能服务器 (gRPC + WebSocket)
├── scripts/
│   └── run-websocket-tests.sh  # WebSocket测试启动脚本
├── test/session/
│   ├── session_integration_test.go  # WebSocket集成测试
│   └── session_test.go              # 单元测试
└── internal/
    ├── testserver/                  # WebSocket测试服务器
    │   └── ws_server.go
    └── session/                     # 会话管理
        ├── recorder.go              # 会话录制
        ├── replayer.go              # 会话回放
        ├── timeline.go              # 时间线分析
        └── assertions.go            # 断言系统
```

## ⚙️ 配置选项

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `SERVER_TYPE` | 服务器类型 (grpc/websocket) | websocket |
| `WS_PORT` | WebSocket服务器端口 | 18090 |
| `GRPC_PORT` | gRPC服务器端口 | 19001 |
| `CI` | 是否在CI环境 | false |

### 命令行参数

#### 服务器启动参数
```bash
go run tools/grpc-server.go [grpc|websocket|help]
```

#### 测试脚本参数
```bash
./scripts/run-websocket-tests.sh [选项]

选项:
  -h, --help          显示帮助信息
  -p, --port PORT     设置端口 (默认: 18090)
  -t, --timeout TIME  设置超时时间 (默认: 10m)
  -c, --ci            CI环境模式
  -l, --local         本地开发模式
```

## 🔧 CI配置详解

### GitHub Actions工作流

#### 触发条件
```yaml
websocket_integration:
  if: github.event_name == 'push' ||
      (github.event_name == 'pull_request' &&
       contains(github.event.pull_request.labels.*.name, 'websocket-test'))
```

#### 资源配置
```yaml
runs-on: ubuntu-latest
timeout-minutes: 15  # 15分钟超时
```

#### 测试步骤
1. **环境准备**: Go安装、依赖下载、代码生成
2. **服务器启动**: 后台启动WebSocket服务器
3. **健康检查**: 验证服务器启动成功
4. **测试执行**: 运行5分钟集成测试
5. **清理**: 停止服务器进程
6. **报告上传**: 保存测试结果

### CI优化策略

#### 1. 并行执行
- **基础测试**: 快速单元测试和代码检查
- **WebSocket测试**: 独立工作流，处理长时间测试

#### 2. 超时保护
- **服务器启动**: 60秒超时
- **测试执行**: 10分钟超时
- **整体工作流**: 15分钟超时

#### 3. 资源管理
- **进程清理**: 自动清理后台进程
- **端口检查**: 确保端口可用性
- **错误处理**: 优雅的失败处理

## 📊 性能基准

### 延迟指标 (内网/同机房)
- **P50**: ≤100ms
- **P90**: ≤300ms
- **P95**: ≤300ms
- **P99**: ≤400ms

### 容量指标
- **并发连接**: 8.97 msg/s
- **消息吞吐**: 3000消息/5分钟
- **业务成功率**: ≥90%

### 稳定性指标
- **连接成功率**: 100%
- **重连恢复时间**: <1秒
- **内存使用**: <100MB

## 🐛 故障排除

### 常见问题

#### 1. 端口被占用
```bash
# 检查端口占用
lsof -i :18090

# 停止占用进程
kill -9 <PID>
```

#### 2. 服务器启动失败
```bash
# 检查防火墙
sudo ufw status

# 检查网络连接
nc -z localhost 18090
```

#### 3. 测试超时
```bash
# 增加超时时间
go test -timeout=15m

# 检查系统资源
top -p <PID>
```

#### 4. CI环境问题
```yaml
# 在PR中添加标签触发测试
labels: [websocket-test]
```

### 调试模式

#### 启用详细日志
```bash
export CI=true
go run tools/grpc-server.go
```

#### 查看测试输出
```bash
go test -v -run TestTimelineAnalysisRealWallClock 2>&1 | tee test.log
```

## 📈 监控和报告

### 测试报告
- **覆盖率报告**: 通过Codecov上传
- **测试结果**: 自动保存到GitHub Actions artifacts
- **性能指标**: 实时输出到控制台

### 监控指标
- **测试执行时间**: 5分钟基准
- **成功率统计**: 业务/传输层分别统计
- **延迟分布**: P50/P90/P95/P99百分位数
- **连接稳定性**: 重连次数和恢复时间

## 🎯 最佳实践

### 本地开发
1. **使用测试脚本**: 自动化服务器管理和测试执行
2. **资源监控**: 注意系统资源使用情况
3. **日志分析**: 查看详细的测试输出和错误信息

### CI环境
1. **标签触发**: 使用PR标签控制测试执行
2. **并行优化**: 避免长时间测试阻塞其他工作流
3. **资源配置**: 合理设置超时时间和计算资源

### 性能调优
1. **基准测试**: 定期运行性能基准测试
2. **阈值调整**: 根据实际业务需求调整性能阈值
3. **容量规划**: 基于测试结果进行容量规划

## 🤝 贡献指南

### 添加新的WebSocket测试
1. 在`test/session/`目录下添加测试文件
2. 使用`Test*`命名规范
3. 添加适当的超时设置
4. 更新CI配置包含新测试

### 修改测试配置
1. 更新`scripts/run-websocket-tests.sh`
2. 修改`.github/workflows/ci.yml`
3. 测试在本地和CI环境中的表现
4. 更新本文档

---

## 📞 技术支持

如果遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查GitHub Issues是否有类似问题
3. 提交新的Issue描述问题
4. 提供详细的错误日志和环境信息

---

**Happy WebSocket Testing! 🚀**