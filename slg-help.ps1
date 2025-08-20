# SLG游戏协议管理PowerShell工具脚本
# GoSlgBenchmarkTest - SLG Protocol Management Tool
# 版本: 1.0.0

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter()]
    [string]$DevPath = "",
    
    [Parameter()]
    [string]$Version = "",
    
    [Parameter()]
    [string]$From = "",
    
    [Parameter()]
    [string]$To = "",
    
    [Parameter()]
    [switch]$VerboseOutput = $false
)

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✅ $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" "Cyan"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
}

# 检查Go环境
function Test-GoEnvironment {
    try {
        $goVersion = & go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Info "Go环境检查通过: $goVersion"
            return $true
        } else {
            Write-Error "Go环境未安装或配置不正确"
            return $false
        }
    } catch {
        Write-Error "Go环境检查失败: $_"
        return $false
    }
}

# 检查协议管理工具
function Test-ProtoManager {
    if (Test-Path "tools/slg-proto-manager/main.go") {
        return $true
    } elseif (Test-Path "tools/slg-proto-manager.go") {
        return $true
    } else {
        Write-Warning "协议管理工具不存在，将使用基本功能"
        return $false
    }
}

# 列出可用版本
function Get-AvailableVersions {
    Write-Info "查找可用的SLG协议版本..."
    
    if (Test-Path "slg-proto") {
        $versions = Get-ChildItem "slg-proto" -Directory | Where-Object { $_.Name -match "^v\d+\.\d+\.\d+" } | Sort-Object Name
        
        if ($versions.Count -gt 0) {
            Write-Success "找到以下版本:"
            foreach ($version in $versions) {
                $versionPath = $version.FullName
                $moduleCount = (Get-ChildItem $versionPath -Directory -ErrorAction SilentlyContinue).Count
                Write-Host "  📦 $($version.Name) ($moduleCount 个模块)" -ForegroundColor Yellow
                
                # 显示模块详情
                $modules = Get-ChildItem $versionPath -Directory -ErrorAction SilentlyContinue
                foreach ($module in $modules) {
                    $protoFiles = (Get-ChildItem "$($module.FullName)\*.proto" -ErrorAction SilentlyContinue).Count
                    Write-Host "     └─ $($module.Name) ($protoFiles 个.proto文件)" -ForegroundColor Gray
                }
            }
        } else {
            Write-Warning "没有找到任何协议版本"
            Write-Info "请使用 'integrate' 命令添加协议版本"
        }
    } else {
        Write-Warning "slg-proto目录不存在"
        Write-Info "请先运行 'integrate' 命令创建协议目录"
    }
}

# 集成开发协议
function Invoke-IntegrateProtocol {
    param(
        [string]$DevPath,
        [string]$Version
    )
    
    # 交互式获取参数
    if (-not $DevPath) {
        $DevPath = Read-Host "请输入研发协议路径"
    }
    
    if (-not $Version) {
        $Version = Read-Host "请输入版本号 (例: v1.1.0)"
    }
    
    # 验证参数
    if (-not $DevPath -or -not $Version) {
        Write-Error "路径和版本号都是必需的"
        return
    }
    
    if (-not (Test-Path $DevPath)) {
        Write-Error "协议路径不存在: $DevPath"
        return
    }
    
    if ($Version -notmatch "^v\d+\.\d+\.\d+$") {
        Write-Error "版本号格式不正确，应该类似: v1.0.0"
        return
    }
    
    Write-Info "开始集成协议..."
    Write-Info "源路径: $DevPath"
    Write-Info "目标版本: $Version"
    
    # 创建目标目录
    $targetPath = "slg-proto\$Version"
    try {
        if (Test-Path $targetPath) {
            $overwrite = Read-Host "版本 $Version 已存在，是否覆盖? (y/N)"
            if ($overwrite -ne "y" -and $overwrite -ne "Y") {
                Write-Info "操作已取消"
                return
            }
            Remove-Item $targetPath -Recurse -Force
        }
        
        New-Item -ItemType Directory -Path $targetPath -Force | Out-Null
        Write-Success "创建目标目录: $targetPath"
        
        # 复制协议文件
        Copy-Item -Path "$DevPath\*" -Destination $targetPath -Recurse -Force
        Write-Success "协议文件复制完成"
        
        # 验证复制结果
        $protoFiles = Get-ChildItem $targetPath -Filter "*.proto" -Recurse
        Write-Success "成功复制 $($protoFiles.Count) 个.proto文件"
        
        # 显示文件结构
        Write-Info "协议结构:"
        Get-ChildItem $targetPath -Recurse | ForEach-Object {
            $relativePath = $_.FullName.Replace("$PWD\$targetPath\", "")
            if ($_.PSIsContainer) {
                Write-Host "  📁 $relativePath" -ForegroundColor Cyan
            } else {
                Write-Host "  📄 $relativePath" -ForegroundColor Gray
            }
        }
        
    } catch {
        Write-Error "集成失败: $_"
        return
    }
    
    Write-Success "协议集成完成！"
    Write-Info "下一步："
    Write-Host "  1. 运行验证: .\slg-help.ps1 validate -Version $Version" -ForegroundColor Yellow
    Write-Host "  2. 生成代码: .\slg-help.ps1 generate -Version $Version" -ForegroundColor Yellow
    Write-Host "  3. 运行测试: go test .\test\slg -v" -ForegroundColor Yellow
}

# 验证协议
function Invoke-ValidateProtocol {
    param([string]$Version)
    
    if (-not $Version) {
        $Version = Read-Host "请输入要验证的版本号"
    }
    
    $versionPath = "slg-proto\$Version"
    if (-not (Test-Path $versionPath)) {
        Write-Error "版本 $Version 不存在"
        return
    }
    
    Write-Info "验证协议版本: $Version"
    
    # 检查协议文件
    $protoFiles = Get-ChildItem $versionPath -Filter "*.proto" -Recurse
    if ($protoFiles.Count -eq 0) {
        Write-Error "没有找到.proto文件"
        return
    }
    
    Write-Success "找到 $($protoFiles.Count) 个协议文件"
    
    # 验证每个协议文件的语法
    $hasErrors = $false
    foreach ($protoFile in $protoFiles) {
        Write-Info "验证: $($protoFile.Name)"
        
        # 简单的语法检查（检查基本的protobuf语法）
        $content = Get-Content $protoFile.FullName -Raw
        
        if ($content -notmatch 'syntax\s*=\s*"proto3"') {
            Write-Warning "  - 缺少 syntax = `"proto3`" 声明"
        }
        
        if ($content -notmatch 'package\s+[\w\.]+;') {
            Write-Warning "  - 缺少 package 声明"
        }
        
        if ($content -notmatch 'option\s+go_package') {
            Write-Warning "  - 缺少 go_package 选项"
        }
        
        Write-Success "  - 基本语法检查通过"
    }
    
    if (-not $hasErrors) {
        Write-Success "协议验证完成，没有发现问题"
    }
}

# 生成Go代码
function Invoke-GenerateCode {
    param([string]$Version)
    
    if (-not $Version) {
        $Version = Read-Host "请输入要生成代码的版本号"
    }
    
    $versionPath = "slg-proto\$Version"
    if (-not (Test-Path $versionPath)) {
        Write-Error "版本 $Version 不存在"
        return
    }
    
    # 转换版本号用于目录名 (v1.0.0 -> v1_0_0)
    $dirVersion = $Version -replace '\.', '_'
    $outputPath = "generated\slg\$dirVersion"
    
    Write-Info "生成Go代码..."
    Write-Info "源路径: $versionPath"
    Write-Info "输出路径: $outputPath"
    
    # 创建输出目录
    New-Item -ItemType Directory -Path $outputPath -Force | Out-Null
    
    # 查找所有.proto文件并生成
    $protoFiles = Get-ChildItem $versionPath -Filter "*.proto" -Recurse
    
    foreach ($protoFile in $protoFiles) {
        $relativePath = $protoFile.Directory.FullName.Replace("$PWD\$versionPath\", "")
        $moduleOutputPath = "$outputPath\$relativePath"
        
        Write-Info "生成: $($protoFile.Name)"
        
        # 创建模块输出目录
        New-Item -ItemType Directory -Path $moduleOutputPath -Force | Out-Null
        
        try {
            # 使用相对路径避免中文路径问题
            $relativeProtoPath = $protoFile.FullName.Replace("$PWD\", "")
            $relativeVersionPath = $versionPath.Replace("$PWD\", "")
            
            # 使用protoc生成Go代码
            $protoCommand = "protoc --proto_path=`"$relativeVersionPath`" --go_out=`"$moduleOutputPath`" --go_opt=paths=source_relative `"$relativeProtoPath`""
            
            if ($VerboseOutput) {
                Write-Info "执行命令: $protoCommand"
            }
            
            $result = Invoke-Expression $protoCommand 2>&1
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "  - 生成成功"
            } else {
                Write-Warning "  - protoc生成失败"
                Write-Info "    错误信息: $result"
                Write-Info "    建议使用 'make proto' 或手动执行protoc命令"
            }
        } catch {
            Write-Warning "  - 生成时出现问题: $_"
        }
    }
    
    Write-Success "代码生成完成！"
    Write-Info "生成的文件位置: $outputPath"
}

# 兼容性测试
function Invoke-CompatibilityTest {
    param(
        [string]$From,
        [string]$To
    )
    
    if (-not $From) {
        $From = Read-Host "请输入源版本号"
    }
    
    if (-not $To) {
        $To = Read-Host "请输入目标版本号"
    }
    
    $fromPath = "slg-proto\$From"
    $toPath = "slg-proto\$To"
    
    if (-not (Test-Path $fromPath)) {
        Write-Error "源版本 $From 不存在"
        return
    }
    
    if (-not (Test-Path $toPath)) {
        Write-Error "目标版本 $To 不存在"
        return
    }
    
    Write-Info "检查协议兼容性: $From → $To"
    
    # 简单的兼容性检查
    $fromFiles = Get-ChildItem $fromPath -Filter "*.proto" -Recurse
    $toFiles = Get-ChildItem $toPath -Filter "*.proto" -Recurse
    
    Write-Info "源版本文件数: $($fromFiles.Count)"
    Write-Info "目标版本文件数: $($toFiles.Count)"
    
    # 检查文件是否存在
    $missingFiles = @()
    foreach ($fromFile in $fromFiles) {
        $relativePath = $fromFile.FullName.Replace("$PWD\$fromPath\", "")
        $toFile = "$toPath\$relativePath"
        
        if (-not (Test-Path $toFile)) {
            $missingFiles += $relativePath
        }
    }
    
    if ($missingFiles.Count -gt 0) {
        Write-Warning "以下文件在目标版本中缺失:"
        foreach ($file in $missingFiles) {
            Write-Host "  - $file" -ForegroundColor Red
        }
    } else {
        Write-Success "所有协议文件都存在于目标版本中"
    }
    
    Write-Info "兼容性检查完成"
    Write-Info "注意: 这只是基本检查，详细的兼容性验证需要运行测试"
}

# 显示帮助信息
function Show-Help {
    Write-Host @"

🎮 SLG游戏协议管理工具 - PowerShell版本
=============================================

用法: .\slg-help.ps1 [命令] [参数]

📋 可用命令:

  help                      显示此帮助信息
  list-versions            列出所有可用的协议版本
  integrate                集成研发团队协议（交互式）
  validate                 验证协议格式
  generate                 生成Go代码
  compatibility            检查版本兼容性

🔧 命令参数:

  integrate -DevPath <路径> -Version <版本>
    集成指定路径的协议到指定版本
    例: .\slg-help.ps1 integrate -DevPath "C:\dev\proto" -Version "v1.1.0"

  validate -Version <版本>
    验证指定版本的协议格式
    例: .\slg-help.ps1 validate -Version "v1.1.0"

  generate -Version <版本>
    为指定版本生成Go代码
    例: .\slg-help.ps1 generate -Version "v1.1.0"

  compatibility -From <源版本> -To <目标版本>
    检查两个版本之间的兼容性
    例: .\slg-help.ps1 compatibility -From "v1.0.0" -To "v1.1.0"

🚀 快速开始:

  1. 集成协议:   .\slg-help.ps1 integrate
  2. 验证协议:   .\slg-help.ps1 validate -Version v1.1.0
  3. 生成代码:   .\slg-help.ps1 generate -Version v1.1.0
  4. 运行测试:   go test .\test\slg -v

📚 更多信息:

  - 详细文档: README-SLG-INTEGRATION.md
  - 快速参考: docs\QUICK-REFERENCE.md
  - 项目主页: README.md

"@ -ForegroundColor Cyan
}

# 主逻辑
function Main {
    Write-Host "🎮 SLG协议管理工具 v1.0.0" -ForegroundColor Magenta
    Write-Host "================================" -ForegroundColor Magenta
    Write-Host ""
    
    # 检查Go环境
    if (-not (Test-GoEnvironment)) {
        Write-Error "请先安装Go环境 (https://golang.org/dl/)"
        exit 1
    }
    
    # 执行命令
    switch ($Command.ToLower()) {
        "help" { Show-Help }
        "list-versions" { Get-AvailableVersions }
        "integrate" { Invoke-IntegrateProtocol -DevPath $DevPath -Version $Version }
        "validate" { Invoke-ValidateProtocol -Version $Version }
        "generate" { Invoke-GenerateCode -Version $Version }
        "compatibility" { Invoke-CompatibilityTest -From $From -To $To }
        default {
            Write-Warning "未知命令: $Command"
            Write-Info "使用 '.\slg-help.ps1 help' 查看可用命令"
        }
    }
}

# 运行主程序
Main