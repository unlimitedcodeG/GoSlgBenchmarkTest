# 修复 proto 文件中的 go_package 选项，使用完整的模块路径
Write-Host "Fixing go_package options to use full module paths for Go modules..." -ForegroundColor Blue

# 修复 v1.0.0 文件
$files_v1_0_0 = @(
    "slg-proto/v1.0.0/building/city.proto",
    "slg-proto/v1.0.0/combat/battle.proto",
    "slg-proto/v1.0.0/common/types.proto"
)

foreach ($file in $files_v1_0_0) {
    if (Test-Path $file) {
        Write-Host "Fixing $file..." -ForegroundColor Yellow
        $content = Get-Content $file -Raw
        
        # 根据文件类型设置正确的 go_package（包含完整模块路径）
        if ($file -like "*/building/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_0_0/building;building";'
        } elseif ($file -like "*/combat/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_0_0/combat;combat";'
        } elseif ($file -like "*/common/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_0_0/common;common";'
        }
        
        Set-Content $file $content -NoNewline
    }
}

# 修复 v1.1.0 文件
$files_v1_1_0 = @(
    "slg-proto/v1.1.0/building/city.proto",
    "slg-proto/v1.1.0/combat/battle.proto",
    "slg-proto/v1.1.0/combat/pvp.proto",
    "slg-proto/v1.1.0/event/activity.proto",
    "slg-proto/v1.1.0/common/types.proto"
)

foreach ($file in $files_v1_1_0) {
    if (Test-Path $file) {
        Write-Host "Fixing $file..." -ForegroundColor Yellow
        $content = Get-Content $file -Raw
        
        # 根据文件类型设置正确的 go_package（包含完整模块路径）
        if ($file -like "*/building/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_1_0/building;building";'
        } elseif ($file -like "*/combat/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_1_0/combat;combat";'
        } elseif ($file -like "*/event/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_1_0/event;event";'
        } elseif ($file -like "*/common/*") {
            $content = $content -replace 'option go_package = ".*";', 'option go_package = "GoSlgBenchmarkTest/generated/slg/v1_1_0/common;common";'
        }
        
        Set-Content $file $content -NoNewline
    }
}

Write-Host "go_package options fixed with full module paths!" -ForegroundColor Green
Write-Host "Generated imports will be: GoSlgBenchmarkTest/generated/slg/v1_x_x/..." -ForegroundColor Cyan