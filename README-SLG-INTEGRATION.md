# SLG游戏协议集成完整方案

## 🎯 项目概述

本项目已完成**与SLG游戏研发团队协议集成**的完整方案，支持：

- ✅ 版本化协议管理
- ✅ 自动代码生成
- ✅ 兼容性测试
- ✅ 研发协议无缝集成
- ✅ 完整工具链支持

## 📁 项目结构

```text
GoSlgBenchmarkTest/
├── 🔧 测试框架核心
│   ├── proto/game/v1/          # 框架基础协议（保持不变）
│   ├── internal/               # 核心组件
│   └── test/                   # 框架测试
│
├── 🎮 SLG协议集成
│   ├── slg-proto/              # SLG游戏协议目录 ✨
│   │   └── v1.0.0/            # 版本化协议
│   │       ├── combat/        # 战斗系统 (battle.proto)
│   │       ├── building/      # 建筑系统 (city.proto)
│   │       ├── alliance/      # 联盟系统 (待添加)
│   │       └── common/        # 通用类型 (types.proto)
│   │
│   ├── generated/slg/          # 生成的Go代码 ✨
│   ├── test/slg/              # SLG专用测试 ✨
│   └── configs/               # 配置文件 ✨
│
└── 🛠️ 工具和文档
    ├── tools/slg-proto-manager.go  # 协议管理工具
    ├── slg-help.ps1              # PowerShell工具脚本
    ├── docs/                     # 集成指南
    └── Makefile.slg             # 扩展构建脚本
```

## 🚀 研发团队协议投递流程

### Windows环境

> **前置要求**:
>
> - 安装Go 1.24+
> - 确保PowerShell执行策略允许脚本运行: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`

#### 方法1: PowerShell工具（推荐）

```powershell
# 1. 集成研发协议
.\slg-help.ps1 integrate -DevPath "C:\path\to\dev\proto" -Version "v1.1.0"

# 2. 验证协议格式
.\slg-help.ps1 validate -Version "v1.1.0"

# 3. 生成Go代码
.\slg-help.ps1 generate -Version "v1.1.0"

# 4. 兼容性测试
.\slg-help.ps1 compatibility -From "v1.0.0" -To "v1.1.0"

# 5. 运行测试
go test .\test\slg -v
```

#### 方法2: 使用CMD

```cmd
:: 集成协议
go run tools\slg-proto-manager.go integrate "C:\dev\proto" "v1.1.0"

:: 列出版本
go run tools\slg-proto-manager.go list-versions

:: 验证协议
go run tools\slg-proto-manager.go validate "v1.1.0"

:: 生成代码
go run tools\slg-proto-manager.go generate "v1.1.0"

:: 运行测试
go test .\test\slg -v
```

### Linux/macOS环境

> **前置要求**:
>
> - 安装Go 1.24+
> - 安装make工具
> - 确保有protoc工具（可选，用于高级功能）

#### 方法1: Makefile工具（推荐）

```bash
# 集成协议
make integrate-dev-proto

# 生成特定版本
make generate-slg-proto VERSION=v1.1.0

# 验证协议
make validate-slg-proto VERSION=v1.1.0

# 兼容性测试
make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0

# 运行测试
make test-slg
```

#### 方法2: 直接使用Go工具

```bash
# 集成协议
go run tools/slg-proto-manager.go integrate /path/to/dev/proto v1.1.0

# 列出版本
go run tools/slg-proto-manager.go list-versions

# 验证协议
go run tools/slg-proto-manager.go validate v1.1.0

# 生成代码
go run tools/slg-proto-manager.go generate v1.1.0

# 运行测试
go test ./test/slg -v
```

## 📋 版本管理策略

### 1. 研发投递新协议版本

```text
研发团队 → 测试团队
1. 提供协议文件夹
2. 指定版本号（如v1.1.0）
3. 说明主要变更
```

### 2. 测试团队集成流程

```text
1. .\slg-help.ps1 integrate -DevPath <path> -Version <version>
   ↓ 自动复制协议文件到 slg-proto/v1.1.0/

2. .\slg-help.ps1 validate -Version v1.1.0
   ↓ 验证协议格式和语法

3. .\slg-help.ps1 generate -Version v1.1.0
   ↓ 生成Go代码到 generated/slg/v1_1_0/

4. .\slg-help.ps1 compatibility -From v1.0.0 -To v1.1.0
   ↓ 检查向后兼容性

5. go test .\test\slg -v
   ↓ 运行SLG专用测试
```

### 3. 版本演进示例

```text
slg-proto/
├── v1.0.0/                 # 初始版本
│   ├── combat/battle.proto
│   ├── building/city.proto
│   └── common/types.proto
│
├── v1.1.0/                 # 新版本（增加功能）
│   ├── combat/
│   │   ├── battle.proto    # 兼容更新
│   │   └── pvp.proto       # 新增PVP
│   ├── building/city.proto # 保持兼容
│   ├── common/types.proto  # 添加新类型
│   └── event/              # 新增活动系统
│       └── activity.proto
│
└── v2.0.0/                 # 主版本（破坏性更改）
    └── ...
```

## 🧪 测试验证

### SLG协议测试覆盖

```go
// test/slg/integration_test.go
func TestSLGProtocolIntegration(t *testing.T) {
    // 测试SLG协议与框架的WebSocket传输
}

func TestSLGMessageSerialization(t *testing.T) {
    // 测试SLG消息的protobuf序列化
}

func TestSLGFrameEncoding(t *testing.T) {
    // 测试SLG消息的帧编码/解码
}

func TestSLGLoadTest(t *testing.T) {
    // 测试SLG场景下的多客户端负载
}
```

### 兼容性测试自动化

```powershell
# 测试所有支持版本的兼容性
.\slg-help.ps1 compatibility -From v1.0.0 -To v1.1.0
.\slg-help.ps1 compatibility -From v1.1.0 -To v1.2.0
```

## 📊 实际使用场景

### 场景1: 研发提交新战斗系统协议

#### Windows操作步骤

```powershell
# 1. 研发：完成新的战斗协议 v1.2.0
# 2. 测试团队集成
.\slg-help.ps1 integrate -DevPath ".\combat-v1.2.0" -Version "v1.2.0"

# 3. 验证：自动检查协议格式和兼容性
.\slg-help.ps1 validate -Version "v1.2.0"

# 4. 生成：自动生成Go代码，测试框架立即可用
.\slg-help.ps1 generate -Version "v1.2.0"

# 5. 测试：运行端到端测试验证新协议
go test .\test\slg -v
```

#### Linux/macOS操作步骤

```bash
# 1. 研发：完成新的战斗协议 v1.2.0
# 2. 测试团队集成
make integrate-dev-proto

# 3. 验证：自动检查协议格式和兼容性
make validate-slg-proto VERSION=v1.2.0

# 4. 生成：自动生成Go代码，测试框架立即可用
make generate-slg-proto VERSION=v1.2.0

# 5. 测试：运行端到端测试验证新协议
make test-slg
```

### 场景2: 游戏版本迭代

```text
v1.0.0 (基础版本)
├── 基础战斗
├── 城市建设
└── 联盟系统

v1.1.0 (功能增强)
├── 战斗系统 ✅ 兼容升级
├── 城市建设 ✅ 兼容升级  
├── 联盟系统 ✅ 保持不变
└── 活动系统 ⭐ 新增

v2.0.0 (大版本重构)
├── 战斗系统 ⚠️ 破坏性更改
└── 全新架构
```

### 场景3: 多分支并行开发

```text
main分支: v1.0.0 (生产版本)
feature/combat: v1.1.0-combat (战斗系统新功能)
feature/social: v1.1.0-social (社交系统新功能)

测试：每个分支独立验证，最后合并测试
```

## 🔧 配置说明

### configs/proto-versions.yaml

```yaml
versions:
  current: "v1.1.0"          # 当前使用版本
  supported: ["v1.0.0", "v1.1.0"]  # 支持的版本
  deprecated: []             # 废弃版本
  
modules:
  combat:
    description: "战斗系统协议"
    owner: "combat-team"
    critical: true           # 核心模块，破坏性更改需特别注意
```

### configs/buf-slg.yaml

```yaml
version: v2
modules:
  - path: slg-proto         # SLG协议目录
lint:
  use: [STANDARD]           # 协议规范检查
breaking:
  use: [FILE]               # 兼容性检查
```

## 📈 扩展建议

### 1. 自动化CI/CD

```yaml
# .github/workflows/slg-proto-ci.yml
on:
  push:
    paths: ['slg-proto/**']
jobs:
  validate-slg-protocol:
    - name: 验证协议更改
    - name: 兼容性测试
    - name: 自动部署测试环境
```

### 2. 协议文档生成

```bash
# 自动从.proto文件生成API文档
buf generate --template buf.gen.doc.yaml
```

### 3. 性能基准

```go
// 针对每个SLG协议版本的性能基准
func BenchmarkSLGProtocol_v1_0_0(b *testing.B)
func BenchmarkSLGProtocol_v1_1_0(b *testing.B)
```

## 🎯 总结

这个集成方案为SLG游戏开发提供了：

✅ **无缝集成**: 研发协议一键导入  
✅ **版本管理**: 支持多版本并存和演进  
✅ **自动化**: 代码生成和测试全自动  
✅ **兼容性**: 自动检查向后兼容性  
✅ **生产就绪**: 完整的测试和验证流程  

现在您可以：

1. 让研发团队按照这个流程提交协议
2. 使用提供的工具快速集成和验证
3. 基于生成的代码进行全面测试
4. 支持游戏的快速迭代和版本管理

**立即开始**:

- **Windows**: `.\slg-help.ps1` 查看所有可用命令
- **Linux/macOS**: `make slg-help` 查看所有可用命令
