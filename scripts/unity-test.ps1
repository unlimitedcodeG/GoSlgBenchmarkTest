# Unity游戏测试录制脚本
# 用于配合Unity客户端进行自动化测试录制

param(
    [Parameter(HelpMessage="环境类型 (development|testing|staging|local)")]
    [ValidateSet("development", "testing", "staging", "local")]
    [string]$Environment = "development",
    
    [Parameter(HelpMessage="测试账号用户名")]
    [string]$Player = "",
    
    [Parameter(HelpMessage="录制时长 (例如: 30m, 1h)")]
    [string]$Duration = "30m",
    
    [Parameter(HelpMessage="配置文件路径")]
    [string]$Config = "configs/test-environments.yaml",
    
    [Parameter(HelpMessage="输出目录")]
    [string]$Output = "",
    
    [Parameter(HelpMessage="启用详细日志")]
    [switch]$VerboseLogging,
    
    [Parameter(HelpMessage="干运行模式，只检查配置")]
    [switch]$DryRun,
    
    [Parameter(HelpMessage="显示帮助信息")]
    [switch]$Help
)

# 显示帮助信息
if ($Help) {
    Write-Host "🎮 Unity游戏测试录制脚本" -ForegroundColor Green
    Write-Host "================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "用法:" -ForegroundColor Yellow
    Write-Host "  .\scripts\unity-test.ps1 [参数]" -ForegroundColor White
    Write-Host ""
    Write-Host "参数:" -ForegroundColor Yellow
    Write-Host "  -Environment    环境类型 (development|testing|staging|local)" -ForegroundColor White
    Write-Host "  -Player         测试账号用户名" -ForegroundColor White
    Write-Host "  -Duration       录制时长 (例如: 30m, 1h)" -ForegroundColor White
    Write-Host "  -Config         配置文件路径" -ForegroundColor White
    Write-Host "  -Output         输出目录" -ForegroundColor White
    Write-Host "  -VerboseLogging 启用详细日志" -ForegroundColor White
    Write-Host "  -DryRun         干运行模式，只检查配置" -ForegroundColor White
    Write-Host "  -Help           显示此帮助信息" -ForegroundColor White
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Yellow
    Write-Host "  # 使用开发环境录制30分钟" -ForegroundColor Green
    Write-Host "  .\scripts\unity-test.ps1 -Environment development -Duration 30m" -ForegroundColor White
    Write-Host ""
    Write-Host "  # 指定测试账号和输出目录" -ForegroundColor Green
    Write-Host "  .\scripts\unity-test.ps1 -Player qa_tester_001 -Output ./recordings/unity -VerboseLogging" -ForegroundColor White
    Write-Host ""
    Write-Host "  # 干运行模式检查配置" -ForegroundColor Green
    Write-Host "  .\scripts\unity-test.ps1 -DryRun" -ForegroundColor White
    return
}

Write-Host "🎮 Unity游戏测试录制工具" -ForegroundColor Green
Write-Host "=========================" -ForegroundColor Green
Write-Host ""

# 检查Go环境
try {
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Go未安装或不在PATH中"
    }
    Write-Host "✅ Go环境检查通过: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go环境检查失败: $_" -ForegroundColor Red
    Write-Host "请确保已安装Go 1.25+并添加到PATH中" -ForegroundColor Yellow
    exit 1
}

# 检查配置文件
if (-not (Test-Path $Config)) {
    Write-Host "❌ 配置文件不存在: $Config" -ForegroundColor Red
    Write-Host "请确保配置文件存在或指定正确的路径" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ 配置文件存在: $Config" -ForegroundColor Green

# 构建录制工具
Write-Host "🔧 构建Unity录制工具..." -ForegroundColor Cyan
try {
    $buildOutput = go build -o "./bin/unity-recorder.exe" "./cmd/unity-recorder" 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "构建失败: $buildOutput"
    }
    Write-Host "✅ 构建完成" -ForegroundColor Green
} catch {
    Write-Host "❌ 构建失败: $_" -ForegroundColor Red
    exit 1
}

# 准备命令行参数
$recorderArgs = @(
    "-env", $Environment,
    "-config", $Config,
    "-duration", $Duration
)

if ($Player) {
    $recorderArgs += @("-player", $Player)
}

if ($Output) {
    $recorderArgs += @("-output", $Output)
}

if ($VerboseLogging) {
    $recorderArgs += @("-verbose")
}

if ($DryRun) {
    $recorderArgs += @("-dry-run")
    Write-Host "🔍 干运行模式 - 只检查配置" -ForegroundColor Cyan
}

# 显示即将执行的命令
Write-Host "🚀 即将执行的命令:" -ForegroundColor Cyan
Write-Host "  .\bin\unity-recorder.exe $($recorderArgs -join ' ')" -ForegroundColor White
Write-Host ""

# 如果不是干运行模式，询问是否继续
if (-not $DryRun) {
    Write-Host "⚠️  录制即将开始，请确保:" -ForegroundColor Yellow
    Write-Host "   1. Unity客户端已准备就绪" -ForegroundColor White
    Write-Host "   2. 游戏服务器正常运行" -ForegroundColor White
    Write-Host "   3. 网络连接稳定" -ForegroundColor White
    Write-Host ""
    
    $confirm = Read-Host "是否开始录制? (y/N)"
    if ($confirm -notmatch "^[Yy]") {
        Write-Host "❌ 用户取消操作" -ForegroundColor Yellow
        exit 0
    }
}

# 执行录制工具
Write-Host "🎬 启动录制工具..." -ForegroundColor Green
Write-Host ""

try {
    # 启动录制工具
    $process = Start-Process -FilePath ".\bin\unity-recorder.exe" -ArgumentList $recorderArgs -NoNewWindow -PassThru -Wait
    
    if ($process.ExitCode -eq 0) {
        Write-Host ""
        Write-Host "🎉 录制完成!" -ForegroundColor Green
        
        # 显示输出目录中的文件
        $outputDir = if ($Output) { $Output } else { "recordings" }
        if (Test-Path $outputDir) {
            Write-Host ""
            Write-Host "📁 生成的文件:" -ForegroundColor Cyan
            Get-ChildItem $outputDir -Filter "session_*" | Sort-Object LastWriteTime -Descending | Select-Object -First 5 | ForEach-Object {
                $size = [math]::Round($_.Length / 1KB, 2)
                Write-Host "   $($_.Name) (${size} KB)" -ForegroundColor White
            }
        }
        
        Write-Host ""
        Write-Host "💡 提示:" -ForegroundColor Yellow
        Write-Host "   - 可以使用生成的JSON文件进行回放分析" -ForegroundColor White
        Write-Host "   - 断言测试结果已包含在输出中" -ForegroundColor White
        Write-Host "   - 如需详细分析，可运行: go run cmd/session-analyzer/main.go" -ForegroundColor White
        
    } else {
        Write-Host "❌ 录制失败，退出代码: $($process.ExitCode)" -ForegroundColor Red
        exit $process.ExitCode
    }
    
} catch {
    Write-Host "❌ 执行录制工具时出错: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🚀 所有操作完成!" -ForegroundColor Green