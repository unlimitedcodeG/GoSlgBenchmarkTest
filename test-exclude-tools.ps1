# GoSlgBenchmarkTest - 排除tools目录的测试脚本
# Windows PowerShell专用

param(
    [switch]$Race,      # 运行竞态检测
    [switch]$Bench,     # 运行性能基准测试
    [switch]$Verbose    # 详细输出
)

Write-Host "🧪 GoSlgBenchmarkTest - 智能测试脚本" -ForegroundColor Magenta
Write-Host "=====================================" -ForegroundColor Magenta

# 设置详细输出
if ($Verbose) {
    $VerbosePreference = "Continue"
}

# 测试目录列表（排除tools）
$testDirs = @(
    ".",
    "./api/handlers",
    "./cmd/test-platform",
    "./game/v1",
    "./generated/slg/v1_0_0/building",
    "./generated/slg/v1_0_0/combat",
    "./generated/slg/v1_0_0/common",
    "./generated/slg/v1_1_0/building",
    "./generated/slg/v1_1_0/combat",
    "./generated/slg/v1_1_0/common",
    "./generated/slg/v1_1_0/event",
    "./internal/config",
    "./internal/database",
    "./internal/db",
    "./internal/grpcserver",
    "./internal/httpserver",
    "./internal/loadtest",
    "./internal/logger",
    "./internal/protocol",
    "./internal/session",
    "./internal/testrunner",
    "./internal/testserver",
    "./internal/testutil",
    "./internal/wsclient",
    "./pkg/analyzer",
    "./pkg/dashboard",
    "./proto/game/v1",
    "./test",
    "./test/session",
    "./test/slg"
)

function Run-UnitTests {
    Write-Host "📋 运行单元测试..." -ForegroundColor Cyan

    $testCommand = "go test"
    if ($Race) {
        $testCommand += " -race"
        Write-Host "  🔄 启用竞态检测" -ForegroundColor Yellow
    }

    $testCommand += " -v -count=1 -timeout=10m"

    foreach ($dir in $testDirs) {
        if (Test-Path $dir) {
            Write-Host "  🧪 测试目录: $dir" -ForegroundColor White
            try {
                Push-Location $dir
                $result = Invoke-Expression "$testCommand 2>&1"
                $result | Tee-Object -Variable testOutput | Out-Null

                # 检查Go测试的退出状态和最终结果
                $lastExitCode = $LASTEXITCODE
                $finalResult = $result | Select-String "^(PASS|FAIL)$" | Select-Object -Last 1

                if ($lastExitCode -ne 0 -or $finalResult -match "FAIL") {
                    Write-Host "  ❌ 测试失败: $dir" -ForegroundColor Red
                    $allPassed = $false
                } else {
                    Write-Host "  ✅ 测试通过: $dir" -ForegroundColor Green
                }
                Pop-Location
            }
            catch {
                Write-Host "  ❌ 错误: $($_.Exception.Message)" -ForegroundColor Red
                Pop-Location
                return $false
            }
        }
    }

    Write-Host "  ✅ 单元测试通过" -ForegroundColor Green
    return $true
}

function Run-BenchmarkTests {
    Write-Host "📊 运行性能基准测试..." -ForegroundColor Cyan

    $benchCommand = "go test -bench=. -benchmem -count=1 -timeout=10m ./test"

    try {
        Invoke-Expression "$benchCommand 2>&1" | Tee-Object -Variable benchOutput
        Write-Host "  ✅ 性能基准测试完成" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "  ❌ 性能测试失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 主测试流程
$allPassed = $true

if ($Bench) {
    $allPassed = Run-BenchmarkTests
} else {
    $allPassed = Run-UnitTests
}

# 结果总结
Write-Host ""
Write-Host "📊 测试结果汇总:" -ForegroundColor Magenta
if ($allPassed) {
    Write-Host "  ✅ 所有测试通过!" -ForegroundColor Green
    Write-Host "  🎯 测试覆盖范围: $(($testDirs | Where-Object { Test-Path $_ }).Count) 个目录" -ForegroundColor Green
} else {
    Write-Host "  ❌ 部分测试失败" -ForegroundColor Red
}

Write-Host ""
Write-Host "💡 提示:" -ForegroundColor Yellow
Write-Host "  • tools目录被排除是因为包含独立的main程序" -ForegroundColor White
Write-Host "  • 如需测试tools，请单独运行: go run tools/<tool>.go" -ForegroundColor White