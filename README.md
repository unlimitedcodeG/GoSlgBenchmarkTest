# GoSlgBenchmarkTest

**SLG游戏长连接+Protobuf自动化测试框架** - 专为实时游戏场景和研发团队协议集成设计的完整解决方案。

## 🎯 项目特性

### 🔧 测试框架核心能力

- ✅ **WebSocket长连接** - 稳定的全双工通信和智能重连
- ✅ **Protobuf序列化** - 高效的二进制协议处理
- ✅ **心跳机制** - 实时RTT统计和连接保活
- ✅ **消息去重** - 基于序列号的重复消息过滤
- ✅ **完整测试套件** - 端到端、模糊测试、基准测试、压力测试
- ✅ **并发安全** - 多客户端并发测试和竞态检测
- ✅ **CI/CD Pipeline** - 自动化构建和测试流水线

### 🎮 SLG协议集成能力

- ✅ **版本化协议管理** - 支持多版本协议并存和演进
- ✅ **自动代码生成** - 一键生成Go代码，测试框架立即可用
- ✅ **兼容性验证** - 自动检查协议向后兼容性
- ✅ **研发协议无缝集成** - 支持研发团队协议快速导入
- ✅ **跨平台工具链** - Windows/Linux/macOS完整支持
- ✅ **模块化设计** - 战斗、建筑、联盟等系统独立管理

## 🚀 快速开始

### 1. 环境要求

- **Go 1.24+** - 支持最新Go特性
- **buf** - Protocol Buffers工具（可选）
- **make** - Linux/macOS构建工具（可选）
- **PowerShell** - Windows环境支持（可选）

### 2. 基础使用

```bash
# 查看项目介绍和快速开始
go run main.go

# 安装依赖和生成协议代码
make install-deps && make proto

# 运行完整测试套件
make test

# 启动测试服务器
go run main.go -mode=server

# 运行客户端压力测试（另一个终端）
go run main.go -mode=client -clients=10 -duration=60s
```

### 3. SLG协议集成使用

#### Windows环境（推荐PowerShell）

```powershell
# 查看SLG工具帮助
.\slg-help.ps1

# 集成研发团队协议（交互式）
.\slg-help.ps1 integrate

# 或直接指定参数
.\slg-help.ps1 integrate -DevPath "C:\dev\proto" -Version "v1.1.0"

# 验证和生成代码
.\slg-help.ps1 validate -Version "v1.1.0"
.\slg-help.ps1 generate -Version "v1.1.0"

# 运行SLG专用测试
go test .\test\slg -v
```

#### Linux/macOS环境

```bash
# 查看SLG工具帮助
make slg-help

# 集成研发团队协议（交互式）
make integrate-dev-proto

# 验证和生成特定版本
make validate-slg-proto VERSION=v1.1.0
make generate-slg-proto VERSION=v1.1.0

# 兼容性测试
make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0

# 运行SLG专用测试
make test-slg
```

## 📋 项目结构

```text
GoSlgBenchmarkTest/
├── 🔧 测试框架核心
│   ├── proto/game/v1/          # 框架基础协议（保持不变）
│   │   └── game.proto          # WebSocket通信协议
│   ├── internal/               # 核心组件
│   │   ├── protocol/           # 帧协议和操作码定义
│   │   ├── wsclient/          # WebSocket客户端实现
│   │   ├── testserver/        # 测试服务器实现
│   │   └── session/           # 会话录制和回放
│   ├── test/                   # 框架测试套件
│   │   ├── e2e_ws_test.go     # 端到端WebSocket测试
│   │   ├── fuzz_decode_test.go # 协议模糊测试
│   │   ├── benchmark_test.go   # 性能基准测试
│   │   └── session/           # 会话管理测试
│   ├── testdata/              # 测试数据和样例
│   └── main.go                # 程序入口点
│
├── 🎮 SLG协议集成层
│   ├── slg-proto/             # SLG游戏协议目录 ✨
│   │   ├── v1.0.0/           # 版本1.0.0协议
│   │   │   ├── combat/       # 战斗系统 (battle.proto)
│   │   │   ├── building/     # 建筑系统 (city.proto)
│   │   │   ├── alliance/     # 联盟系统 (待扩展)
│   │   │   └── common/       # 通用类型 (types.proto)
│   │   └── v1.1.0/           # 版本1.1.0协议
│   │       ├── combat/       # 战斗系统（增强版+PVP）
│   │       ├── building/     # 建筑系统（兼容升级）
│   │       ├── common/       # 通用类型（扩展版）
│   │       └── event/        # 活动系统（新增）
│   ├── generated/slg/         # 生成的Go代码 ✨
│   │   ├── v1_0_0/           # v1.0.0版本生成代码
│   │   └── v1_1_0/           # v1.1.0版本生成代码
│   ├── test/slg/             # SLG专用测试 ✨
│   │   ├── slg_protocol_test.go    # 协议兼容性测试
│   │   └── integration_test.go     # 集成测试
│   └── configs/              # 配置文件 ✨
│       ├── proto-versions.yaml     # 版本管理配置
│       └── buf-slg.yaml           # SLG协议buf配置
│
└── 🛠️ 工具和文档
    ├── tools/                # 管理工具
    │   ├── slg-proto-manager/ # 协议管理工具
    │   └── generate_testdata.go # 测试数据生成
    ├── docs/                 # 完整文档
    │   ├── protobuf-integration-guide.md # 集成指南
    │   └── QUICK-REFERENCE.md            # 快速参考
    ├── slg-help.ps1         # Windows PowerShell工具脚本
    ├── Makefile             # Linux/macOS构建脚本
    └── README-SLG-INTEGRATION.md # SLG集成方案说明
```

## 🧪 测试体系

### 框架核心测试

```bash
# 基础测试套件
make test              # 基本功能测试
make test-race         # 竞态条件检测
make test-coverage     # 测试覆盖率分析

# 专项测试
make fuzz             # 协议模糊测试（6个模糊测试函数）
make bench            # 性能基准测试
make profile          # 性能分析和优化
make verify           # 完整项目验证
```

### SLG协议专项测试

```bash
# Windows环境
.\slg-help.ps1 compatibility -From "v1.0.0" -To "v1.1.0"
go test .\test\slg -v

# Linux/macOS环境  
make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0
make test-slg

# 测试覆盖内容
# ✅ 协议版本兼容性验证
# ✅ SLG协议序列化/反序列化
# ✅ WebSocket集成测试
# ✅ 性能基准测试
# ✅ 并发安全性测试
```

### CI/CD自动化测试

项目包含完整的CI/CD流水线，支持：

- ✅ 多Go版本兼容性测试
- ✅ 跨平台构建验证  
- ✅ 模糊测试持续运行
- ✅ 协议兼容性回归测试
- ✅ 性能基准对比
- ✅ 代码质量检查

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

### 双重架构体系

**🔧 测试框架层** - 提供稳定的通信基础设施

- **帧协议**: `| opcode(2字节) | length(4字节) | body(变长) |`
- **WebSocket长连接**: 全双工通信 + 智能重连
- **会话管理**: 录制/回放/分析测试会话
- **并发安全**: 全面的锁保护和竞态检测

**🎮 SLG协议层** - 支持游戏协议版本管理

- **版本隔离**: 多版本协议并存和独立管理
- **向后兼容**: 自动兼容性验证和渐进式升级  
- **模块化设计**: 战斗/建筑/联盟等系统解耦
- **工具链集成**: 跨平台协议管理工具

### 核心组件设计

#### 客户端设计 (`internal/wsclient/`)

- **状态管理**: 精确的连接状态跟踪和转换
- **重连策略**: 指数退避算法和断线检测
- **消息去重**: 基于序列号的单调性检查
- **性能监控**: 实时RTT统计和吞吐量分析

#### 服务器设计 (`internal/testserver/`)

- **连接管理**: 高效的连接池和生命周期管理
- **消息广播**: 批量推送优化和选择性广播
- **统计收集**: 实时性能指标和连接状态监控
- **优雅关闭**: 安全的资源清理和goroutine同步

#### 协议管理 (`tools/slg-proto-manager/`)

- **版本控制**: 语义化版本管理和演进策略
- **代码生成**: 自动化Go代码生成和包管理
- **兼容性检查**: 破坏性更改检测和兼容性报告
- **工具集成**: 跨平台工具链和CI/CD集成

## 🔌 协议体系

### 框架基础协议 (`proto/game/v1/`)

**通信协议** - 稳定的WebSocket通信基础

- `LoginReq/LoginResp` - 认证握手和会话建立
- `Heartbeat/HeartbeatResp` - 心跳保活和RTT测量
- `ErrorResp` - 统一错误响应
- `BattlePush` - 战斗状态推送（演示）
- `PlayerAction` - 玩家操作消息（演示）

### SLG游戏协议 (`slg-proto/`)

#### v1.0.0 基础版本

- `combat/battle.proto` - 基础战斗系统（PVE支持）
- `building/city.proto` - 城市建设系统  
- `common/types.proto` - 通用数据类型（位置、物品、资源）

#### v1.1.0 增强版本

- `combat/battle.proto` - 增强战斗系统（向后兼容 + 新字段）
- `combat/pvp.proto` - 新增PVP战斗系统
- `building/city.proto` - 兼容升级的建筑系统
- `common/types.proto` - 扩展的通用类型（稀有度、附魔）
- `event/activity.proto` - 全新活动系统

### 协议演进策略

#### 兼容性规则

- ✅ **添加字段**: 在消息末尾添加可选字段（兼容）
- ✅ **添加枚举值**: 新增枚举值（兼容）
- ❌ **删除字段**: 使用`reserved`标记（不兼容）
- ❌ **修改字段类型**: 创建新字段（不兼容）

#### 版本管理

- **主版本号**: 不兼容的API更改
- **次版本号**: 向后兼容的功能性新增  
- **修订号**: 向后兼容的问题修正

## 🧩 Unity集成方案

### C#客户端集成

**基础集成** - 复用测试框架协议

```csharp
// 1. 使用Unity WebSocket插件
// 2. 导入Google.Protobuf for Unity
// 3. 复用相同的帧格式: | opcode(2) | length(4) | body(变长) |
// 4. 实现相同的操作码定义和心跳机制
```

**SLG协议集成** - 支持游戏协议版本

```csharp
// 1. 从slg-proto/生成C#代码
// 2. 支持多版本协议切换
// 3. 实现向后兼容的协议解析
// 4. 集成版本管理和兼容性检查
```

### 推荐集成步骤

#### 第一步：基础协议集成

1. 使用`proto/game/v1/game.proto`生成C#代码
2. 实现WebSocket帧编解码逻辑
3. 复用心跳和重连机制
4. 对接测试服务器进行验证

#### 第二步：SLG协议集成

1. 选择目标SLG协议版本（如v1.1.0）
2. 使用`slg-proto/v1.1.0/`生成C#代码
3. 实现版本兼容的消息处理
4. 集成协议版本检查机制

#### 第三步：生产环境优化

1. 启用消息压缩和批量处理
2. 实现客户端性能监控
3. 添加协议降级机制
4. 集成错误报告和诊断

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

## 📈 监控和分析

### 内置性能指标

#### 连接层监控

- 当前连接数和总连接数统计
- 连接建立/断开成功率
- 重连次数和重连成功率
- WebSocket握手延迟

#### 消息层监控

- 消息发送/接收吞吐量（TPS）
- 消息序列号单调性检查
- 消息去重统计
- 协议解析错误率

#### 性能监控

- 实时RTT分布和P99延迟
- 内存使用和GC压力
- goroutine数量和泄漏检测
- CPU使用率和热点分析

### SLG协议专项监控

#### 版本兼容性监控

- 协议版本使用分布
- 兼容性检查通过率
- 向后兼容性违规检测

#### 业务指标监控

- SLG模块调用频次（战斗/建筑/联盟）
- 协议消息大小分布
- 序列化/反序列化性能

### 扩展监控集成

项目预留了完整的监控接口：

```go
// Prometheus指标集成
func (s *Server) GetMetrics() map[string]interface{} {
    return map[string]interface{}{
        "connections_current": s.connCount.Load(),
        "messages_sent_total": s.messagesSent.Load(),
        "rtt_percentiles": s.getRTTPercentiles(),
        "protocol_versions": s.getProtocolVersionStats(),
    }
}
```

#### 推荐监控栈

- **Prometheus** - 指标收集和存储
- **Grafana** - 可视化仪表板和告警
- **Jaeger** - 分布式链路追踪
- **PProf** - Go性能剖析

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

## 📚 完整文档体系

### 快速上手文档

- 📖 [SLG协议集成方案](./README-SLG-INTEGRATION.md) - 完整的集成流程和实战案例
- 📋 [快速参考手册](./docs/QUICK-REFERENCE.md) - 常用命令速查表和故障排除
- 📘 [详细集成指南](./docs/protobuf-integration-guide.md) - 深入的技术集成指导

### 工具使用文档

```bash
# Windows用户
.\slg-help.ps1                    # 查看PowerShell工具帮助

# Linux/macOS用户  
make slg-help                     # 查看Makefile工具帮助

# 通用Go工具
go run tools/slg-proto-manager.go # 查看协议管理工具帮助
```

### 技术实现文档

- 🏗️ **架构设计** - 双重架构体系和核心组件设计
- 🔌 **协议体系** - 框架协议 + SLG协议版本管理
- 🧪 **测试体系** - 完整的测试方案和CI/CD流水线
- 📈 **监控分析** - 性能指标和业务监控

## 🔗 相关技术链接

### 核心依赖

- [Protocol Buffers](https://developers.google.com/protocol-buffers) - 高效的序列化框架
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - Go WebSocket实现
- [Testify](https://github.com/stretchr/testify) - Go测试断言库

### Unity集成

- [Unity WebSocket](https://docs.unity3d.com/Manual/webgl-networking.html) - Unity WebSocket支持
- [Google.Protobuf for Unity](https://github.com/protocolbuffers/protobuf/releases) - Unity Protobuf包

### 开发工具

- [Go Testing](https://golang.org/pkg/testing/) - Go测试框架
- [Go Fuzzing](https://go.dev/security/fuzz/) - Go模糊测试
- [Buf](https://buf.build/) - 现代Protobuf工具链

---

## 🎯 项目总结

**GoSlgBenchmarkTest** 提供了一个完整的解决方案：

- **🔧 稳定的测试框架** - 支持WebSocket长连接、Protobuf协议、完整测试套件
- **🎮 灵活的协议集成** - 支持SLG游戏协议版本管理、兼容性验证、自动代码生成
- **🛠️ 完整的工具链** - 跨平台工具支持、CI/CD集成、监控分析
- **📚 详尽的文档** - 从快速上手到深入集成的完整指导

**立即开始**: `go run main.go` 查看演示，或选择适合您平台的集成方案！

祝您编码愉快！**Happy Coding! 🎮**
