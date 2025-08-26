# GoSlg 负载测试平台启动脚本 (PowerShell)
# 支持 WebSocket、gRPC、HTTP API 压力测试

param(
    [switch]$GenProto,      # 生成 protobuf 代码
    [switch]$BuildOnly,     # 仅构建，不启动
    [switch]$CheckDeps,     # 仅检查依赖
    [switch]$Help           # 显示帮助信息
)

# 错误处理
$ErrorActionPreference = "Stop"

# 项目根目录
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $ProjectRoot

# 颜色函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $colorMap = @{
        "Red" = "Red"
        "Green" = "Green"
        "Yellow" = "Yellow"
        "Blue" = "Blue"
        "Magenta" = "Magenta"
        "Cyan" = "Cyan"
        "White" = "White"
    }
    
    Write-Host $Message -ForegroundColor $colorMap[$Color]
}

function Log-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Log-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Log-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Log-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 检查依赖
function Check-Dependencies {
    Log-Info "检查依赖..."
    
    # 检查 Go
    try {
        $null = Get-Command go -ErrorAction Stop
        Log-Success "Go 已安装: $(go version)"
    }
    catch {
        Log-Error "Go 未安装或不在 PATH 中"
        Log-Error "请从 https://golang.org/dl/ 下载并安装 Go"
        exit 1
    }
    
    # 检查 protoc
    try {
        $null = Get-Command protoc -ErrorAction Stop
        Log-Success "protoc 已安装: $(protoc --version)"
    }
    catch {
        Log-Warning "protoc 未安装，gRPC 功能可能无法正常工作"
        Log-Warning "请从 https://github.com/protocolbuffers/protobuf/releases 下载安装"
    }
    
    # 检查 buf (可选)
    try {
        $null = Get-Command buf -ErrorAction Stop
        Log-Success "buf 已安装: $(buf --version)"
    }
    catch {
        Log-Warning "buf 未安装，将使用 protoc 生成代码"
    }
    
    Log-Success "依赖检查完成"
}

# 构建项目
function Build-Project {
    Log-Info "构建项目..."
    
    # 清理旧的构建文件
    if (Test-Path "test-platform-extended.exe") {
        Remove-Item "test-platform-extended.exe" -Force
    }
    
    try {
        # 构建扩展测试平台
        & go build -o test-platform-extended.exe cmd/test-platform/main_extended.go
        
        if ($LASTEXITCODE -eq 0) {
            Log-Success "项目构建成功"
        } else {
            throw "构建失败，退出码: $LASTEXITCODE"
        }
    }
    catch {
        Log-Error "项目构建失败: $($_.Exception.Message)"
        exit 1
    }
}

# 生成 protobuf 代码
function Generate-Proto {
    if ($GenProto) {
        Log-Info "生成 protobuf 代码..."
        
        try {
            # 检查 buf
            $null = Get-Command buf -ErrorAction Stop
            & buf generate
            if ($LASTEXITCODE -eq 0) {
                Log-Success "使用 buf 生成 protobuf 代码"
            } else {
                throw "buf generate 失败"
            }
        }
        catch {
            try {
                # 使用 protoc 手动生成
                $null = Get-Command protoc -ErrorAction Stop
                
                # 确保目录存在
                if (!(Test-Path "proto/game/v1")) {
                    New-Item -ItemType Directory -Path "proto/game/v1" -Force | Out-Null
                }
                
                # 生成 Go 代码
                & protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/game/v1/game.proto proto/game/v1/game_service.proto
                
                if ($LASTEXITCODE -eq 0) {
                    Log-Success "使用 protoc 生成 protobuf 代码"
                } else {
                    throw "protoc 生成失败"
                }
            }
            catch {
                Log-Warning "无法生成 protobuf 代码，protoc 和 buf 都未找到或失败"
            }
        }
    }
}

# 启动测试平台
function Start-Platform {
    Log-Info "启动负载测试平台..."
    
    # 设置环境变量
    $env:GOMAXPROCS = "4"
    $env:TEST_LOG_LEVEL = "info"
    
    # 显示启动信息
    Write-ColorOutput "🚀 GoSlg 负载测试平台" "Cyan"
    Write-ColorOutput "===========================================" "Cyan"
    Write-Host "📊 Web界面: " -NoNewline
    Write-ColorOutput "http://localhost:8080" "Green"
    Write-Host "🧪 负载测试: " -NoNewline  
    Write-ColorOutput "http://localhost:8080/loadtest" "Green"
    Write-Host "❤️  健康检查: " -NoNewline
    Write-ColorOutput "http://localhost:8080/api/v1/health" "Green"
    Write-Host "🌐 HTTP测试服务器: " -NoNewline
    Write-ColorOutput "http://localhost:19000" "Green"
    Write-Host "⚡ gRPC测试服务器: " -NoNewline
    Write-ColorOutput "localhost:19001" "Green"
    Write-Host ""
    Write-ColorOutput "支持的测试类型:" "Yellow"
    Write-Host "  • WebSocket 长连接压测"
    Write-Host "  • gRPC 服务接口压测"  
    Write-Host "  • HTTP REST API 压测"
    Write-Host ""
    Write-ColorOutput "按 Ctrl+C 停止服务" "Magenta"
    Write-Host ""
    
    try {
        # 启动平台
        & ./test-platform-extended.exe
    }
    catch {
        Log-Error "启动失败: $($_.Exception.Message)"
        exit 1
    }
}

# 显示帮助信息
function Show-Help {
    Write-ColorOutput "GoSlg 负载测试平台 - 启动脚本" "Cyan"
    Write-Host ""
    Write-ColorOutput "用法:" "Yellow"
    Write-Host "  .\run-loadtest-platform.ps1 [参数]"
    Write-Host ""
    Write-ColorOutput "参数:" "Yellow"
    Write-Host "  -GenProto      生成 protobuf 代码"
    Write-Host "  -BuildOnly     仅构建，不启动"
    Write-Host "  -CheckDeps     仅检查依赖"
    Write-Host "  -Help          显示此帮助信息"
    Write-Host ""
    Write-ColorOutput "示例:" "Yellow"
    Write-Host "  .\run-loadtest-platform.ps1                 # 正常启动"
    Write-Host "  .\run-loadtest-platform.ps1 -GenProto       # 生成 proto 代码并启动"
    Write-Host "  .\run-loadtest-platform.ps1 -BuildOnly      # 仅构建项目"
    Write-Host ""
    Write-ColorOutput "功能特性:" "Yellow"
    Write-Host "  🌐 HTTP API 压力测试"
    Write-Host "  ⚡ gRPC 服务压力测试"
    Write-Host "  📡 WebSocket 长连接测试"
    Write-Host "  📊 实时性能指标监控"
    Write-Host "  🎯 多协议负载均衡测试"
    Write-Host "  📈 详细的测试报告生成"
}

# 清理函数
function Cleanup {
    Log-Info "正在清理..."
    
    # 停止可能运行的进程
    Get-Process -Name "test-platform-extended" -ErrorAction SilentlyContinue | Stop-Process -Force
    
    Log-Success "清理完成"
}

# 主逻辑
function Main {
    try {
        if ($Help) {
            Show-Help
            return
        }
        
        if ($CheckDeps) {
            Check-Dependencies
            return
        }
        
        Check-Dependencies
        Generate-Proto
        Build-Project
        
        if ($BuildOnly) {
            Log-Success "构建完成，可执行文件: test-platform-extended.exe"
            return
        }
        
        Start-Platform
    }
    catch {
        Log-Error "执行失败: $($_.Exception.Message)"
        exit 1
    }
    finally {
        Cleanup
    }
}

# 设置控制台编码为 UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 运行主函数
Main