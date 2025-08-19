# SLG协议管理PowerShell脚本
param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter()]
    [string]$Version = "",
    
    [Parameter()]
    [string]$DevPath = "",
    
    [Parameter()]
    [string]$From = "",
    
    [Parameter()]
    [string]$To = ""
)

function Show-Help {
    Write-Host "🎮 SLG协议管理工具" -ForegroundColor Blue
    Write-Host "========================================" -ForegroundColor Blue
    Write-Host ""
    Write-Host "📋 命令列表:" -ForegroundColor Green
    Write-Host "  .\slg-help.ps1 integrate -DevPath <path> -Version <version>" -ForegroundColor Yellow
    Write-Host "    集成研发团队提供的协议文件"
    Write-Host ""
    Write-Host "  .\slg-help.ps1 generate -Version <version>" -ForegroundColor Yellow
    Write-Host "    生成指定版本的SLG协议Go代码"
    Write-Host ""
    Write-Host "  .\slg-help.ps1 validate -Version <version>" -ForegroundColor Yellow
    Write-Host "    验证指定版本的协议格式和兼容性"
    Write-Host ""
    Write-Host "  .\slg-help.ps1 compatibility -From <v1> -To <v2>" -ForegroundColor Yellow
    Write-Host "    测试两个版本间的协议兼容性"
    Write-Host ""
    Write-Host "  .\slg-help.ps1 list-versions" -ForegroundColor Yellow
    Write-Host "    列出所有可用的SLG协议版本"
    Write-Host ""
    Write-Host "📁 目录结构:" -ForegroundColor Green
    Write-Host "  slg-proto\v1.0.0\    - 协议定义文件"
    Write-Host "  generated\slg\       - 生成的Go代码"
    Write-Host "  test\slg\           - SLG专用测试"
    Write-Host "  configs\            - 配置文件"
    Write-Host ""
    Write-Host "🔄 典型工作流:" -ForegroundColor Green
    Write-Host "  1. .\slg-help.ps1 integrate -DevPath .\dev-proto -Version v1.1.0"
    Write-Host "  2. .\slg-help.ps1 validate -Version v1.1.0"
    Write-Host "  3. .\slg-help.ps1 generate -Version v1.1.0"
    Write-Host "  4. .\slg-help.ps1 compatibility -From v1.0.0 -To v1.1.0"
    Write-Host "  5. go test .\test\slg -v"
}

function Integrate-Proto {
    if ([string]::IsNullOrEmpty($DevPath) -or [string]::IsNullOrEmpty($Version)) {
        Write-Host "错误: 请指定 -DevPath 和 -Version 参数" -ForegroundColor Red
        Write-Host "示例: .\slg-help.ps1 integrate -DevPath .\dev-proto -Version v1.1.0"
        return
    }
    
    Write-Host "🔄 集成研发协议..." -ForegroundColor Blue
    Write-Host "   源路径: $DevPath"
    Write-Host "   目标版本: $Version"
    
    go run tools\slg-proto-manager.go integrate $DevPath $Version
}

function Generate-Proto {
    if ([string]::IsNullOrEmpty($Version)) {
        Write-Host "错误: 请指定 -Version 参数" -ForegroundColor Red
        Write-Host "示例: .\slg-help.ps1 generate -Version v1.0.0"
        return
    }
    
    Write-Host "🔧 生成SLG协议代码 $Version..." -ForegroundColor Blue
    go run tools\slg-proto-manager.go generate $Version
}

function Validate-Proto {
    if ([string]::IsNullOrEmpty($Version)) {
        Write-Host "错误: 请指定 -Version 参数" -ForegroundColor Red
        Write-Host "示例: .\slg-help.ps1 validate -Version v1.0.0"
        return
    }
    
    Write-Host "🔍 验证SLG协议 $Version..." -ForegroundColor Blue
    go run tools\slg-proto-manager.go validate $Version
}

function Test-Compatibility {
    if ([string]::IsNullOrEmpty($From) -or [string]::IsNullOrEmpty($To)) {
        Write-Host "错误: 请指定 -From 和 -To 参数" -ForegroundColor Red
        Write-Host "示例: .\slg-help.ps1 compatibility -From v1.0.0 -To v1.1.0"
        return
    }
    
    Write-Host "🔍 兼容性测试: $From -> $To" -ForegroundColor Blue
    go run tools\slg-proto-manager.go compatibility-check $From $To
}

function List-Versions {
    Write-Host "📋 SLG协议版本列表" -ForegroundColor Blue
    go run tools\slg-proto-manager.go list-versions
}

# 主逻辑
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "integrate" { Integrate-Proto }
    "generate" { Generate-Proto }
    "validate" { Validate-Proto }
    "compatibility" { Test-Compatibility }
    "list-versions" { List-Versions }
    default {
        Write-Host "未知命令: $Command" -ForegroundColor Red
        Show-Help
    }
}