# SLG协议管理快速参考

## 🎯 核心命令对照表

| 操作 | Windows (PowerShell) | Windows (CMD) | Linux/macOS (Make) | Linux/macOS (Go) |
|------|----------------------|---------------|-------------------|------------------|
| **查看帮助** | `.\slg-help.ps1` | `go run tools\slg-proto-manager.go` | `make slg-help` | `go run tools/slg-proto-manager.go` |
| **集成协议** | `.\slg-help.ps1 integrate -DevPath "C:\path" -Version "v1.1.0"` | `go run tools\slg-proto-manager.go integrate "C:\path" "v1.1.0"` | `make integrate-dev-proto` | `go run tools/slg-proto-manager.go integrate /path v1.1.0` |
| **列出版本** | `.\slg-help.ps1 list-versions` | `go run tools\slg-proto-manager.go list-versions` | `make list-slg-versions` | `go run tools/slg-proto-manager.go list-versions` |
| **验证协议** | `.\slg-help.ps1 validate -Version "v1.1.0"` | `go run tools\slg-proto-manager.go validate "v1.1.0"` | `make validate-slg-proto VERSION=v1.1.0` | `go run tools/slg-proto-manager.go validate v1.1.0` |
| **生成代码** | `.\slg-help.ps1 generate -Version "v1.1.0"` | `go run tools\slg-proto-manager.go generate "v1.1.0"` | `make generate-slg-proto VERSION=v1.1.0` | `go run tools/slg-proto-manager.go generate v1.1.0` |
| **兼容性测试** | `.\slg-help.ps1 compatibility -From "v1.0.0" -To "v1.1.0"` | `go run tools\slg-proto-manager.go compatibility-check "v1.0.0" "v1.1.0"` | `make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0` | `go run tools/slg-proto-manager.go compatibility-check v1.0.0 v1.1.0` |
| **运行测试** | `go test .\test\slg -v` | `go test .\test\slg -v` | `make test-slg` | `go test ./test/slg -v` |

## 🚀 常见工作流程

### 1. 首次设置项目

#### Windows

```powershell
# 检查Go版本
go version

# 查看帮助
.\slg-help.ps1

# 列出现有版本
.\slg-help.ps1 list-versions
```

#### Linux/macOS

```bash
# 检查Go版本
go version

# 查看帮助
make slg-help

# 列出现有版本
make list-slg-versions
```

### 2. 集成新协议版本

#### Windows集成流程

```powershell
# 方法1: 使用交互式脚本
.\slg-help.ps1 integrate

# 方法2: 直接指定参数
.\slg-help.ps1 integrate -DevPath "C:\dev\proto" -Version "v1.2.0"
```

#### Linux/macOS集成流程

```bash
# 方法1: 使用交互式Makefile
make integrate-dev-proto

# 方法2: 直接使用Go工具
go run tools/slg-proto-manager.go integrate /path/to/dev/proto v1.2.0
```

### 3. 验证和测试新版本

#### Windows验证流程

```powershell
# 验证协议格式
.\slg-help.ps1 validate -Version "v1.2.0"

# 生成Go代码
.\slg-help.ps1 generate -Version "v1.2.0"

# 检查兼容性
.\slg-help.ps1 compatibility -From "v1.1.0" -To "v1.2.0"

# 运行测试
go test .\test\slg -v
```

#### Linux/macOS验证流程

```bash
# 验证协议格式
make validate-slg-proto VERSION=v1.2.0

# 生成Go代码
make generate-slg-proto VERSION=v1.2.0

# 检查兼容性
make test-slg-compatibility FROM=v1.1.0 TO=v1.2.0

# 运行测试
make test-slg
```

## 📁 关键目录和文件

```text
项目根目录/
├── slg-proto/                    # 📁 协议定义目录
│   ├── v1.0.0/                  # 版本化协议
│   ├── v1.1.0/
│   └── v1.2.0/
├── generated/slg/               # 📁 生成的Go代码
│   ├── v1_0_0/
│   ├── v1_1_0/
│   └── v1_2_0/
├── test/slg/                    # 📁 SLG专用测试
├── configs/                     # 📁 配置文件
│   ├── proto-versions.yaml     # 版本配置
│   └── buf-slg.yaml            # Buf配置
├── tools/                       # 📁 工具脚本
│   └── slg-proto-manager.go    # 核心管理工具
├── slg-help.ps1                # 🖥️ Windows PowerShell脚本
└── Makefile                     # 🐧 Linux/macOS构建脚本
```

## ⚡ 快速故障排除

### 常见问题

#### 1. PowerShell脚本执行被阻止 (Windows)

```powershell
# 问题：execution policy不允许脚本运行
# 解决：设置执行策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 2. make命令不存在 (Linux/macOS)

```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install
# 或者
brew install make
```

#### 3. 协议生成失败

```bash
# 检查Go模块
go mod tidy

# 检查protoc安装
protoc --version

# 手动验证协议文件语法
protoc --proto_path=slg-proto/v1.0.0 --go_out=. slg-proto/v1.0.0/combat/battle.proto
```

#### 4. 版本不存在错误

```bash
# 检查可用版本
# Windows: .\slg-help.ps1 list-versions
# Linux/macOS: make list-slg-versions

# 检查目录结构
tree slg-proto/    # Linux/macOS
dir slg-proto /s   # Windows
```

### 调试模式

#### Windows调试

```powershell
# 启用详细输出
$DebugPreference = "Continue"
.\slg-help.ps1 validate -Version "v1.0.0"
```

#### Linux/macOS调试

```bash
# 使用详细模式
make validate-slg-proto VERSION=v1.0.0 VERBOSE=1

# 直接调试Go工具
go run -x tools/slg-proto-manager.go validate v1.0.0
```

## 🔗 相关文档链接

- [完整集成指南](./protobuf-integration-guide.md)
- [SLG集成方案](../README-SLG-INTEGRATION.md)
- [项目主文档](../README.md)

## 🆘 获取帮助

1. **查看内置帮助**
   - Windows: `.\slg-help.ps1`
   - Linux/macOS: `make slg-help`

2. **检查项目状态**
   - 运行 `go test ./test/slg -v` 验证环境
   - 检查 `configs/proto-versions.yaml` 查看配置

3. **常用诊断命令**

   ```bash
   go version                    # 检查Go版本
   go mod tidy                   # 整理依赖
   go build ./...               # 检查编译
   ```

---

> 💡 **提示**: 建议将此文档添加到书签，作为日常操作的快速参考！
