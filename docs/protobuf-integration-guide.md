# SLG游戏Protobuf协议集成指南

## 🎯 概述

这个指南说明如何将研发团队的protobuf协议文件集成到测试框架中，支持版本管理和持续迭代。

> **支持平台**: Windows、Linux、macOS  
> **Go版本要求**: 1.24+  
> **依赖工具**: protoc、buf（可选）

## 📁 目录结构设计

### 当前结构

```text
GoSlgBenchmarkTest/
├── proto/                    # 测试框架内置协议（保持不变）
│   └── game/v1/
│       └── game.proto        # 框架基础协议
└── ...
```

### 推荐的SLG游戏协议集成结构

```text
GoSlgBenchmarkTest/
├── proto/                    # 测试框架基础协议
│   └── game/v1/
│       └── game.proto
├── slg-proto/                # SLG游戏协议目录 ✨ 新增
│   ├── v1.0.0/              # 版本1.0.0
│   │   ├── combat/          # 战斗系统
│   │   │   ├── battle.proto
│   │   │   └── skill.proto
│   │   ├── building/        # 建筑系统
│   │   │   ├── city.proto
│   │   │   └── construction.proto
│   │   ├── alliance/        # 联盟系统
│   │   │   └── alliance.proto
│   │   └── common/          # 通用类型
│   │       ├── types.proto
│   │       └── error.proto
│   ├── v1.1.0/              # 版本1.1.0（增量更新）
│   │   ├── combat/
│   │   │   ├── battle.proto # 覆盖v1.0.0版本
│   │   │   └── pvp.proto    # 新增PVP功能
│   │   └── event/           # 新增活动系统
│   │       └── activity.proto
│   └── latest -> v1.1.0     # 软链接指向最新版本
├── generated/                # 生成的Go代码 ✨ 新增
│   ├── proto/game/v1/       # 框架协议生成代码
│   └── slg/                 # SLG协议生成代码
│       ├── v1_0_0/
│       └── v1_1_0/
├── test/slg/                 # SLG专用测试 ✨ 新增
│   ├── combat_test.go
│   ├── building_test.go
│   └── integration_test.go
└── configs/                  # 配置文件 ✨ 新增
    ├── buf-slg.yaml         # SLG协议buf配置
    └── proto-versions.yaml   # 版本管理配置
```

## 🚀 集成步骤

### 步骤1: 设置SLG协议目录

#### Windows (PowerShell/CMD)

```powershell
# PowerShell
md slg-proto\v1.0.0\combat, slg-proto\v1.0.0\building, slg-proto\v1.0.0\alliance, slg-proto\v1.0.0\common
md generated\slg\v1_0_0
md test\slg
md configs
```

```cmd
:: CMD
mkdir slg-proto\v1.0.0\combat
mkdir slg-proto\v1.0.0\building
mkdir slg-proto\v1.0.0\alliance
mkdir slg-proto\v1.0.0\common
mkdir generated\slg\v1_0_0
mkdir test\slg
mkdir configs
```

#### Linux/macOS (Bash)

```bash
mkdir -p slg-proto/v1.0.0/{combat,building,alliance,common}
mkdir -p generated/slg/v1_0_0
mkdir -p test/slg
mkdir -p configs
```

### 步骤2: 配置版本管理

创建 `configs/proto-versions.yaml`:

```yaml
versions:
  current: "v1.1.0"
  supported:
    - "v1.0.0"
    - "v1.1.0"
  deprecated:
    - "v0.9.0"  # 30天后移除
    
compatibility_tests:
  - from: "v1.0.0"
    to: "v1.1.0"
    
build_targets:
  - version: "v1.0.0"
    output: "generated/slg/v1_0_0"
  - version: "v1.1.0" 
    output: "generated/slg/v1_1_0"
```

### 步骤3: 研发协议投递流程

#### Windows 环境

##### 方法1: PowerShell脚本（推荐）

```powershell
# 集成研发协议
.\slg-help.ps1 integrate -DevPath "C:\dev\proto" -Version "v1.1.0"

# 验证协议格式
.\slg-help.ps1 validate -Version "v1.1.0"

# 生成Go代码
.\slg-help.ps1 generate -Version "v1.1.0"

# 运行兼容性测试
.\slg-help.ps1 compatibility -From "v1.0.0" -To "v1.1.0"
```

##### 方法2: 直接使用Go工具

```powershell
# 研发团队投递新协议
Copy-Item -Recurse "C:\path\to\dev\proto\*" "slg-proto\v1.1.0\"

# 验证协议格式
go run tools\slg-proto-manager.go validate v1.1.0

# 生成Go代码
go run tools\slg-proto-manager.go generate v1.1.0

# 运行兼容性测试
go run tools\slg-proto-manager.go compatibility-check v1.0.0 v1.1.0
```

#### Linux/macOS 环境

##### 使用Makefile（推荐）

```bash
# 研发团队投递新协议
cp -r /path/to/dev/proto/* slg-proto/v1.1.0/

# 验证协议格式
make validate-slg-proto VERSION=v1.1.0

# 生成Go代码
make generate-slg-proto VERSION=v1.1.0

# 运行兼容性测试
make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0
```

##### 使用Go工具

```bash
# 使用工具脚本
go run tools/slg-proto-manager.go integrate /path/to/dev/proto v1.1.0
go run tools/slg-proto-manager.go validate v1.1.0
go run tools/slg-proto-manager.go generate v1.1.0
go run tools/slg-proto-manager.go compatibility-check v1.0.0 v1.1.0
```

## 🔧 配置文件

### SLG专用buf配置 (`configs/buf-slg.yaml`)

```yaml
version: v2
modules:
  - path: slg-proto
deps: []
lint:
  use:
    - STANDARD
  except:
    - FIELD_LOWER_SNAKE_CASE
    - ENUM_VALUE_PREFIX
    - ENUM_ZERO_VALUE_SUFFIX
  ignore:
    - slg-proto/v*/deprecated  # 忽略废弃的协议
breaking:
  use:
    - FILE
  ignore:
    - slg-proto/v*/experimental  # 允许实验性协议破坏兼容性
```

### 构建脚本配置

#### Makefile扩展 (Linux/macOS)

```makefile
# SLG协议相关命令
.PHONY: generate-slg-proto validate-slg-proto test-slg-compatibility integrate-dev-proto

# 生成指定版本的SLG协议
generate-slg-proto:
    @if [ -z "$(VERSION)" ]; then echo "Usage: make generate-slg-proto VERSION=v1.0.0"; exit 1; fi
    go run tools/slg-proto-manager.go generate $(VERSION)

# 验证SLG协议格式
validate-slg-proto:
    @if [ -z "$(VERSION)" ]; then echo "Usage: make validate-slg-proto VERSION=v1.0.0"; exit 1; fi
    go run tools/slg-proto-manager.go validate $(VERSION)

# 运行SLG兼容性测试
test-slg-compatibility:
    @if [ -z "$(FROM)" ] || [ -z "$(TO)" ]; then echo "Usage: make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0"; exit 1; fi
    go run tools/slg-proto-manager.go compatibility-check $(FROM) $(TO)

# 快速集成研发协议
integrate-dev-proto:
    @echo "🔄 集成研发协议..."
    @read -p "输入协议版本 (例: v1.2.0): " VERSION; \
    read -p "输入研发协议路径: " DEV_PATH; \
    go run tools/slg-proto-manager.go integrate $$DEV_PATH $$VERSION; \
    echo "✅ 协议集成完成！"
```

#### PowerShell脚本 (Windows)

详见项目根目录的 `slg-help.ps1` 脚本，提供完整的Windows环境支持。

## 🧪 测试集成

### SLG协议测试模板 (`test/slg/protocol_test.go`)

```go
package slg_test

import (
    "testing"
    "path/filepath"
    "os"
    
    "google.golang.org/protobuf/proto"
    "github.com/stretchr/testify/require"
    
    // 根据版本导入生成的包
    v1_0_0_combat "GoSlgBenchmarkTest/generated/slg/v1_0_0/combat"
    v1_1_0_combat "GoSlgBenchmarkTest/generated/slg/v1_1_0/combat"
)

// TestSLGProtocolVersionCompatibility 测试版本兼容性
func TestSLGProtocolVersionCompatibility(t *testing.T) {
    // 测试v1.0.0的战斗消息能否被v1.1.0解析
    oldBattle := &v1_0_0_combat.BattleRequest{
        BattleId: "battle_001",
        PlayerId: "player_123", 
        // 旧版本字段...
    }
    
    data, err := proto.Marshal(oldBattle)
    require.NoError(t, err)
    
    // 用新版本解析
    newBattle := &v1_1_0_combat.BattleRequest{}
    err = proto.Unmarshal(data, newBattle)
    require.NoError(t, err)
    
    // 验证关键字段保持兼容
    require.Equal(t, oldBattle.BattleId, newBattle.BattleId)
    require.Equal(t, oldBattle.PlayerId, newBattle.PlayerId)
}

// TestSLGProtocolWithFramework 测试SLG协议与框架集成
func TestSLGProtocolWithFramework(t *testing.T) {
    // 测试SLG协议能否通过测试框架的WebSocket传输
    // 这里可以复用框架的测试服务器和客户端
}
```

## 📦 版本管理最佳实践

### 1. 语义化版本控制

- **主版本号**: 不兼容的API更改
- **次版本号**: 向后兼容的功能性新增
- **修订号**: 向后兼容的问题修正

### 2. 分支策略

```text
slg-proto/
├── v1.0.0/          # 稳定版本，只修复critical bug
├── v1.1.0/          # 当前开发版本
├── v1.2.0-beta/     # 下一版本预览
└── experimental/    # 实验性功能，不保证兼容性
```

### 3. 协议演进规则

- **添加字段**: ✅ 兼容，在消息末尾添加可选字段
- **删除字段**: ❌ 不兼容，使用reserved标记
- **修改字段类型**: ❌ 不兼容，创建新字段
- **重命名字段**: ❌ 不兼容，保留旧字段并标记deprecated

### 4. 自动化工具

#### 协议版本检查工具 (`tools/proto-version-check.go`)

```go
// 检查协议版本兼容性的工具
func main() {
    // 比较两个版本的协议差异
    // 生成兼容性报告
    // 自动标记破坏性更改
}
```

#### CI/CD集成

```yaml
# .github/workflows/slg-proto-ci.yml
name: SLG Protocol CI

on:
  push:
    paths:
      - 'slg-proto/**'
      
jobs:
  validate-protocol:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate Protocol Changes
        run: |
          make validate-slg-proto VERSION=${{ github.ref_name }}
          make test-slg-compatibility FROM=v1.0.0 TO=${{ github.ref_name }}
```

## 🔄 日常工作流程

### 研发团队协议提交工作流

#### Windows工作流

1. 研发完成新版本协议文件
2. 调用 `.\slg-help.ps1 integrate -DevPath <path> -Version <version>`
3. 自动验证协议格式和兼容性
4. 生成Go代码并运行测试
5. 提交到版本控制系统

#### Linux/macOS工作流

1. 研发完成新版本协议文件
2. 调用 `make integrate-dev-proto`
3. 自动验证协议格式和兼容性
4. 生成Go代码并运行测试
5. 提交到版本控制系统

### 测试团队日常操作

#### Windows操作流程

1. 列出可用版本: `.\slg-help.ps1 list-versions`
2. 验证特定版本: `.\slg-help.ps1 validate -Version v1.1.0`
3. 运行测试: `go test .\test\slg -v`
4. 生成兼容性报告: `.\slg-help.ps1 compatibility -From v1.0.0 -To v1.1.0`

#### Linux/macOS操作流程

1. 列出可用版本: `make list-slg-versions`
2. 验证特定版本: `make validate-slg-proto VERSION=v1.1.0`
3. 运行测试: `make test-slg`
4. 生成兼容性报告: `make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0`

这个方案既保持了测试框架的独立性，又能很好地集成SLG游戏的实际协议，支持版本演进和团队协作。
