# Go Wrapper (gow) for Windows PowerShell

param(
    [Parameter(Position=0)]
    [string]$Command,
    
    [Parameter(Position=1, ValueFromRemainingArguments=$true)]
    [string[]]$Arguments
)

$GOW_VERSION = "1.0.0"
$GO_VERSION_FILE = ".go-version"
$GOW_DIR = ".gow"
$WRAPPER_DIR = "$GOW_DIR\wrapper"

# 颜色输出函数
function Write-Info { Write-Host "[INFO] $($args[0])" -ForegroundColor Blue }
function Write-Warn { Write-Host "[WARN] $($args[0])" -ForegroundColor Yellow }
function Write-Error { Write-Host "[ERROR] $($args[0])" -ForegroundColor Red }
function Write-Success { Write-Host "[SUCCESS] $($args[0])" -ForegroundColor Green }

# 显示帮助
function Show-Help {
    @"
Go Wrapper (gow) v$GOW_VERSION

用法:
  gow <command> [arguments]

命令:
  build       编译项目
  run         运行 Go 程序
  test        运行测试
  mod         Go modules 相关操作
  deps        管理依赖
  clean       清理构建文件
  version     显示版本信息
  init        初始化 gow 配置
  update      更新 gow 自身

示例:
  gow run main.go
  gow test ./...
  gow mod tidy
  gow deps update
"@
}

# 检测平台
function Get-Platform {
    $os = "windows"
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64" -or $env:PROCESSOR_ARCHITEW6432 -eq "AMD64") { "amd64" } else { "386" }
    "$os-$arch"
}

# 初始化 gow
function Initialize-Gow {
    if (-not (Test-Path $GO_VERSION_FILE)) {
        try {
            $currentGo = (go version).Split(' ')[2].Replace('go', '')
            $currentGo | Out-File -FilePath $GO_VERSION_FILE -Encoding utf8
            Write-Info "创建 $GO_VERSION_FILE，当前 Go 版本: $currentGo"
        } catch {
            Write-Warn "无法检测当前 Go 版本，请手动创建 $GO_VERSION_FILE 文件"
        }
    }
    
    if (-not (Test-Path $GOW_DIR)) {
        New-Item -ItemType Directory -Path $GOW_DIR -Force | Out-Null
        Write-Info "创建 $GOW_DIR 目录"
    }
    
    # 创建 gitignore
    $gitignorePath = "$GOW_DIR\.gitignore"
    if (-not (Test-Path $gitignorePath)) {
        @"
# gow 忽略文件
wrapper/
cache/
temp/
"@ | Out-File -FilePath $gitignorePath -Encoding utf8
    }
    
    Write-Success "gow 初始化完成"
}

# 检查 Go 版本
function Test-GoVersion {
    if (Test-Path $GO_VERSION_FILE) {
        $requiredVersion = Get-Content $GO_VERSION_FILE -Raw
        try {
            $currentVersion = (go version).Split(' ')[2].Replace('go', '')
            if ($currentVersion -ne $requiredVersion.Trim()) {
                Write-Warn "当前 Go 版本 ($currentVersion) 与项目要求 ($($requiredVersion.Trim())) 不匹配"
                Write-Warn "请使用 Go $($requiredVersion.Trim()) 或更新 .go-version 文件"
            }
        } catch {
            Write-Warn "无法检测 Go 版本"
        }
    }
}

# 管理依赖
function Manage-Dependencies {
    param([string]$Action)
    
    switch ($Action) {
        "update" {
            Write-Info "更新依赖..."
            go get -u ./...
            go mod tidy
        }
        "vendor" {
            Write-Info "创建 vendor 目录..."
            go mod vendor
        }
        "verify" {
            Write-Info "验证依赖..."
            go mod verify
        }
        default {
            Write-Host "可用操作: update, vendor, verify"
        }
    }
}

# 主函数
function Main {
    if (-not $Command) {
        Show-Help
        return
    }

    switch ($Command) {
        "init" {
            Initialize-Gow
        }
        "version" {
            Write-Host "Go Wrapper v$GOW_VERSION"
            try {
                go version
            } catch {
                Write-Host "Go 未安装"
            }
        }
        "deps" {
            Manage-Dependencies -Action $Arguments[0]
        }
        "update" {
            Write-Info "更新 gow..."
            Write-Warn "自我更新功能尚未实现"
        }
        "build" { 
            Test-GoVersion
            go @Arguments
        }
        "run" { 
            Test-GoVersion
            go @Arguments
        }
        "test" { 
            Test-GoVersion
            go @Arguments
        }
        "mod" { 
            Test-GoVersion
            go @Arguments
        }
        "clean" { 
            Test-GoVersion
            go @Arguments
        }
        default {
            # 检查是否是文件
            if (Test-Path $Command) {
                Test-GoVersion
                go run @($Command) + $Arguments
            } else {
                Write-Error "未知命令: $Command"
                Show-Help
                exit 1
            }
        }
    }
}

# 运行主函数
Main
