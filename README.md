# GoSlgBenchmarkTest

Unity长连接+Protobuf自动化测试框架，专为实时游戏场景设计。

## 🎯 项目特性

- ✅ **WebSocket长连接** - 稳定的全双工通信
- ✅ **自动重连机制** - 智能断线检测和指数退避重连
- ✅ **Protobuf序列化** - 高效的二进制协议
- ✅ **心跳机制** - 实时RTT统计和连接保活
- ✅ **消息去重** - 基于序列号的重复消息过滤
- ✅ **完整测试套件** - 端到端、模糊测试、基准测试
- ✅ **压力测试工具** - 多客户端并发测试
- ✅ **CI/CD Pipeline** - 自动化构建和测试

## 🚀 快速开始

### 1. 环境要求

- Go 1.24+
- buf (Protocol Buffers工具)
- make (可选，用于构建脚本)

### 2. 安装依赖

```bash
# 安装项目依赖
make install-deps

# 安装开发工具
make tools-install

# 生成Protobuf代码
make proto
```

### 3. 运行演示

```bash
# 查看项目介绍
go run main.go

# 启动测试服务器
go run main.go -mode=server

# 运行客户端压力测试（另一个终端）
go run main.go -mode=client -clients=10 -duration=60s
```

## 📋 项目结构

```text
GoSlgBenchmarkTest/
├── proto/game/v1/          # Protobuf协议定义
│   └── game.proto
├── internal/
│   ├── protocol/           # 帧协议和操作码
│   ├── wsclient/          # WebSocket客户端
│   └── testserver/        # 测试服务器
├── test/                  # 测试套件
│   ├── e2e_ws_test.go     # 端到端测试
│   ├── compat_pb_test.go  # 兼容性测试
│   ├── fuzz_decode_test.go # 模糊测试
│   └── benchmark_test.go   # 基准测试
├── tools/                 # 工具脚本
├── testdata/             # 测试数据
├── Makefile              # 构建脚本
└── main.go               # 程序入口
```

## 🧪 测试

### 运行所有测试

```bash
make test           # 基本测试
make test-race      # 竞态检测
make test-coverage  # 覆盖率测试
```

### 专项测试

```bash
make fuzz          # 模糊测试
make bench         # 基准测试
make profile       # 性能分析
```

### 验证项目

```bash
make verify        # 完整验证（测试+检查）
make quick         # 快速检查
make release-check # 发布前检查
```

## 🔧 开发

### 代码检查和格式化

```bash
make lint          # 代码检查
make format        # 代码格式化
```

### 构建

```bash
make build-server  # 构建服务器
make build-client  # 构建客户端
make build-all     # 构建所有
```

### 开发环境设置

```bash
make dev-setup     # 一键设置开发环境
make watch         # 监控文件变化（需要inotify-tools）
```

## 📊 性能测试

### 基准测试结果

运行 `make bench` 查看详细的性能指标：

- **单客户端往返延迟** - 测量请求-响应延迟
- **并发客户端吞吐量** - 多客户端并发性能
- **Protobuf序列化性能** - 编码/解码效率
- **帧处理性能** - 协议层开销
- **内存分配分析** - 不同消息大小的内存使用

### 压力测试

```bash
# 10个客户端，运行60秒
go run main.go -mode=client -clients=10 -duration=60s

# 自定义服务器地址
go run main.go -mode=client -url=ws://remote-server:8080/ws -clients=50
```

## 🌐 网络模拟

在Linux环境下测试弱网条件：

```bash
make simulate-network  # 应用网络延迟和丢包
make reset-network     # 重置网络设置
```

## 🏗️ 架构设计

### 协议层

- **帧格式**: `| opcode(2字节) | length(4字节) | body(变长) |`
- **操作码**: 分类管理不同类型消息
- **流式解码**: 支持TCP流式数据解析

### 客户端设计

- **状态管理**: 精确的连接状态跟踪
- **重连策略**: 指数退避算法
- **消息去重**: 基于序列号的单调性检查
- **并发安全**: 全面的锁保护

### 服务器设计

- **连接管理**: 高效的连接池
- **消息广播**: 批量推送优化
- **统计收集**: 实时性能指标
- **优雅关闭**: 安全的资源清理

## 🔌 协议消息

### 基础消息

- `LoginReq/LoginResp` - 认证握手
- `Heartbeat/HeartbeatResp` - 心跳保活
- `ErrorResp` - 错误响应

### 游戏消息

- `BattlePush` - 战斗状态推送
- `PlayerAction` - 玩家操作
- `ChatAction` - 聊天消息

### 扩展性

协议设计支持：
- 向后兼容的字段演进
- 新枚举值的处理
- 大消息的分片（待实现）

## 🧩 Unity集成

### C#客户端示例

```csharp
// 基于该Go项目的协议，Unity端可以使用：
// 1. Unity WebSocket插件
// 2. Google.Protobuf for Unity
// 3. 相同的帧格式和操作码定义
```

### 推荐集成方案

1. 使用相同的.proto文件生成C#代码
2. 实现相同的帧编解码逻辑
3. 复用心跳和重连机制
4. 对接相同的测试服务器

## 🐛 故障排除

### 常见问题

1. **连接失败**: 检查服务器地址和防火墙设置
2. **编译错误**: 确保已运行 `make proto` 生成代码
3. **测试失败**: 检查端口占用，尝试更换测试端口
4. **性能问题**: 使用 `make profile` 进行性能分析

### 调试技巧

```bash
# 查看详细日志
go run main.go -mode=server -v

# 监控服务器状态
curl http://localhost:8080/stats

# 运行单个测试
go test ./test -run TestReconnectAndSequenceMonotonic -v
```

## 📈 监控

### 内置指标

- 连接数统计
- 消息吞吐量
- RTT分布
- 错误率
- 内存使用

### 扩展监控

项目预留了Prometheus指标接口，可以轻松集成：
- Grafana仪表板
- 告警规则
- 链路追踪

## 🤝 贡献

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

### 开发规范

- 遵循Go代码规范
- 添加必要的单元测试
- 更新相关文档
- 通过所有CI检查

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Unity WebSocket](https://docs.unity3d.com/Manual/webgl-networking.html)
- [Go Testing](https://golang.org/pkg/testing/)

---

祝您编码愉快！**Happy Coding! 🎮**
