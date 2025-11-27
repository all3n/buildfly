# Go Wrapper (gow) for Windows PowerShell

param(
    [Parameter(Position=0)]
    [string]$Command,

    [Parameter(Position=1, ValueFromRemainingArguments=$true)]
    [string[]]$Arguments
)

$GOW_VERSION = "2.0.0"
$GO_VERSION_FILE = ".go-version"
$GOW_GLOBAL_VERSION_FILE = "$env:USERPROFILE\.gow\global-version"
$GOW_DIR = ".gow"
$GOW_HOME = "$env:USERPROFILE\.gow"
$GOW_VERSIONS_DIR = "$GOW_HOME\versions"
$GOW_CACHE_DIR = "$GOW_HOME\cache"

# 颜色输出函数
function Write-Info { Write-Host "[INFO] $($args[0])" -ForegroundColor Blue }
function Write-Warn { Write-Host "[WARN] $($args[0])" -ForegroundColor Yellow }
function Write-Error { Write-Host "[ERROR] $($args[0])" -ForegroundColor Red }
function Write-Success { Write-Host "[SUCCESS] $($args[0])" -ForegroundColor Green }
function Write-Debug { Write-Host "[DEBUG] $($args[0])" -ForegroundColor Cyan }

# 显示帮助
function Show-Help {
    @"
Go Wrapper (gow) v$GOW_VERSION

用法:
  gow <command> [arguments]

核心命令:
  run         运行 Go 程序
  build       编译项目
  test        运行测试
  mod         Go modules 相关操作

版本管理:
  use <version>    切换到指定 Go 版本
  install <version> 安装指定 Go 版本
  uninstall <version> 卸载指定 Go 版本
  list             列出已安装的版本
  list-remote      列出可用的远程版本
  current          显示当前使用的版本

项目管理:
  init             初始化 gow 项目配置
  deps             管理依赖
  clean            清理构建文件和缓存

工具命令:
  version          显示 gow 和 Go 版本信息
  doctor           检查环境和配置
  update           更新 gow 自身
  self-update      更新 gow 到最新版本

示例:
  gow run main.go
  gow use 1.21.0
  gow install 1.20.5
  gow install 1.20.5 amd64
  gow install 1.20.5 386
  gow install 1.20.5 arm64
  gow test ./...
  gow mod tidy
  gow deps update
  gow doctor

配置文件:
  .go-version      项目级 Go 版本配置
  ~/.gow/global-version  全局默认 Go 版本
"@
}

# 检测 Windows 架构，默认返回 windows-amd64
function Get-Platform {
    $arch = $env:PROCESSOR_ARCHITECTURE

    # 检测系统架构，默认为 amd64
    switch -Wildcard ($arch) {
        "AMD64" { return "windows-amd64" }
        "x86_64" { return "windows-amd64" }
        "x64" { return "windows-amd64" }
        "IA64" { return "windows-amd64" }
        "ARM64" { return "windows-arm64" }
        "ARM" {
            # 检查是否是 ARM64 上的 ARM32
            if ($env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
                return "windows-arm64"
            } else {
                return "windows-386"  # ARM32 暂不支持，回退到 386
            }
        }
        default {
            # 如果不是已知的 64 位，检查 WOW64
            if ($env:PROCESSOR_ARCHITEW6432 -eq "AMD64") {
                return "windows-amd64"
            } elseif ($env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
                return "windows-arm64"
            } else {
                return "windows-386"
            }
        }
    }
}

# 检查系统是否有 Go
function Test-SystemGo {
    try {
        $goPath = Get-Command go -ErrorAction SilentlyContinue
        if ($goPath) {
            & $goPath.Source version 2>$null
            return $true
        }
    } catch {
        return $false
    }
    return $false
}

# 标准化 Go 版本号
function Normalize-GoVersion {
    param([string]$Version)

    # 移除 'go' 前缀
    $Version = $Version -replace '^go', ''

    # 验证版本号格式 (如 1.21.0)
    if ($Version -notmatch '^\d+\.\d+(\.\d+)?$') {
        Write-Error "无效的 Go 版本格式: $Version"
        return $null
    }

    return $Version
}

# 获取已安装的 Go 版本列表
function Get-InstalledVersions {
    if (Test-Path $GOW_VERSIONS_DIR) {
        Get-ChildItem -Path $GOW_VERSIONS_DIR -Directory -Name "go*" | Sort-Object
    }
}

# 检查指定版本是否已安装
function Test-VersionInstalled {
    param([string]$Version)
    $normalizedVersion = Normalize-GoVersion $Version
    if ($normalizedVersion) {
        $installDir = "$GOW_VERSIONS_DIR\go$normalizedVersion"
        return (Test-Path $installDir) -and (Test-Path "$installDir\bin\go.exe")
    }
    return $false
}

# 获取 Go 下载 URL - Windows 平台使用 zip 格式
function Get-GoDownloadUrl {
    param([string]$Version, [string]$Architecture = "amd64")

    $filename = "go$Version.windows-$Architecture.zip"
    return "https://go.dev/dl/$filename"
}

# 安装指定版本的 Go - Windows 支持三种架构
function Install-GoVersion {
    param([string]$Version, [string]$Architecture = "amd64")

    $normalizedVersion = Normalize-GoVersion $Version
    if (-not $normalizedVersion) {
        return $false
    }

    if (Test-VersionInstalled $normalizedVersion) {
        Write-Info "Go $normalizedVersion 已经安装"
        return $true
    }

    Write-Info "正在安装 Go $normalizedVersion (Windows $Architecture)..."

    # 创建必要目录
    New-Item -ItemType Directory -Path $GOW_VERSIONS_DIR -Force | Out-Null
    New-Item -ItemType Directory -Path $GOW_CACHE_DIR -Force | Out-Null

    $downloadUrl = Get-GoDownloadUrl $normalizedVersion $Architecture
    $filename = "go$normalizedVersion.windows-$Architecture.zip"
    $cacheFile = "$GOW_CACHE_DIR\$filename"

    # 下载 Go (如果缓存中没有)
    if (-not (Test-Path $cacheFile)) {
        Write-Info "正在下载 Go $normalizedVersion 从 $downloadUrl..."
        try {
            if (Get-Command curl -ErrorAction SilentlyContinue) {
                curl -L -o $cacheFile $downloadUrl
            } elseif (Get-Command wget -ErrorAction SilentlyContinue) {
                wget -O $cacheFile $downloadUrl
            } else {
                Write-Error "需要 curl 或 wget 来下载 Go"
                return $false
            }
        } catch {
            Write-Error "下载失败: $($_.Exception.Message)"
            return $false
        }
    } else {
        Write-Info "使用缓存的文件: $cacheFile"
    }

    # 解压安装
    $installDir = "$GOW_VERSIONS_DIR\go$normalizedVersion"
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    Write-Info "正在解压到 $installDir..."
    try {
        if (Get-Command Expand-Archive -ErrorAction SilentlyContinue) {
            # 使用 PowerShell 的 Expand-Archive
            $tempDir = "$GOW_CACHE_DIR\temp_extract_$([System.Guid]::NewGuid())"
            New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
            Expand-Archive -Path $cacheFile -DestinationPath $tempDir -Force

            # 移动文件到目标目录（Go zip 包包含 go 目录）
            $goDir = Join-Path $tempDir "go"
            if (Test-Path $goDir) {
                Get-ChildItem -Path $goDir | Move-Item -Destination $installDir
            } else {
                # 如果没有 go 子目录，直接移动所有文件
                Get-ChildItem -Path $tempDir | Move-Item -Destination $installDir
            }

            Remove-Item -Path $tempDir -Recurse -Force
        } else {
            Write-Error "需要 PowerShell 5.0+ 来解压文件"
            return $false
        }
    } catch {
        Write-Error "解压失败: $($_.Exception.Message)"
        return $false
    }

    # 验证安装
    $goBinary = "$installDir\bin\go.exe"
    if (Test-Path $goBinary) {
        try {
            $installedVersion = (& $goBinary version).Split(' ')[2].Replace('go', '')
            if ($installedVersion -eq $normalizedVersion) {
                Write-Success "Go $normalizedVersion ($Architecture) 安装成功!"
                Write-Info "安装路径: $installDir"
                return $true
            } else {
                Write-Warn "版本不匹配: 期望 $normalizedVersion, 实际 $installedVersion"
                Write-Info "安装路径: $installDir"
                return $false
            }
        } catch {
            Write-Error "验证安装失败: $($_.Exception.Message)"
            return $false
        }
    } else {
        Write-Error "Go 安装失败: 找不到 go.exe"
        return $false
    }
}

# 卸载指定版本的 Go
function Uninstall-GoVersion {
    param([string]$Version)

    $normalizedVersion = Normalize-GoVersion $Version
    if (-not $normalizedVersion) {
        return $false
    }

    if (-not (Test-VersionInstalled $normalizedVersion)) {
        Write-Warn "Go $normalizedVersion 未安装"
        return $false
    }

    Write-Info "正在卸载 Go $normalizedVersion..."
    Remove-Item -Path "$GOW_VERSIONS_DIR\go$normalizedVersion" -Recurse -Force
    Write-Success "Go $normalizedVersion 已卸载"
    return $true
}

# 切换 Go 版本
function Use-GoVersion {
    param([string]$Version)

    $normalizedVersion = Normalize-GoVersion $Version
    if (-not $normalizedVersion) {
        return $false
    }

    if (-not (Test-VersionInstalled $normalizedVersion)) {
        Write-Warn "Go $normalizedVersion 未安装"
        Write-Info "使用 'gow install $normalizedVersion' 来安装"
        return $false
    }

    # 设置项目级版本
    $normalizedVersion | Out-File -FilePath $GO_VERSION_FILE -Encoding UTF8
    Write-Success "项目 Go 版本已设置为 $normalizedVersion"
    Write-Info "运行 'gow doctor' 来验证配置"
    return $true
}

# 获取当前应该使用的 Go 版本
function Get-CurrentGoVersion {
    # 优先级：项目版本 > 全局版本 > 系统版本
    if (Test-Path $GO_VERSION_FILE) {
        $content = Get-Content $GO_VERSION_FILE -Raw
        return $content.Trim()
    } elseif (Test-Path $GOW_GLOBAL_VERSION_FILE) {
        $content = Get-Content $GOW_GLOBAL_VERSION_FILE -Raw
        return $content.Trim()
    } else {
        # 返回系统版本（如果有）
        if (Test-SystemGo) {
            try {
                $systemVersion = (go version).Split(' ')[2].Replace('go', '')
                return $systemVersion
            } catch {
                return $null
            }
        }
    }
    return $null
}

# 获取当前 Go 二进制路径
function Get-GoBinary {
    $version = Get-CurrentGoVersion
    if ($version) {
        $normalizedVersion = Normalize-GoVersion $version
        if (Test-VersionInstalled $normalizedVersion) {
            return "$GOW_VERSIONS_DIR\go$normalizedVersion\bin\go.exe"
        }
    }

    # 回退到系统 Go
    if (Test-SystemGo) {
        return (Get-Command go).Source
    }

    return $null
}

# 初始化 gow 配置
function Initialize-Gow {
    Write-Info "初始化 gow 项目配置..."

    # 创建全局目录
    New-Item -ItemType Directory -Path $GOW_HOME -Force | Out-Null
    New-Item -ItemType Directory -Path $GOW_VERSIONS_DIR -Force | Out-Null
    New-Item -ItemType Directory -Path $GOW_CACHE_DIR -Force | Out-Null

    # 创建项目级目录
    if (-not (Test-Path $GOW_DIR)) {
        New-Item -ItemType Directory -Path $GOW_DIR -Force | Out-Null
        Write-Info "创建 $GOW_DIR 目录"
    }

    # 检测当前 Go 版本并设置
    $currentGo = ""
    if (Test-SystemGo) {
        try {
            $currentGo = (go version).Split(' ')[2].Replace('go', '')
        } catch {
            # 忽略错误
        }
    }

    if (-not (Test-Path $GO_VERSION_FILE)) {
        if ($currentGo) {
            $currentGo | Out-File -FilePath $GO_VERSION_FILE -Encoding UTF8
            Write-Info "创建 $GO_VERSION_FILE，当前 Go 版本: $currentGo"
        } else {
            Write-Warn "未检测到系统 Go，请手动设置版本: gow use <version>"
        }
    }

    # 创建项目 gitignore
    $projectGitignore = "$GOW_DIR\.gitignore"
    if (-not (Test-Path $projectGitignore)) {
        @"
# gow 忽略文件
.cache/
build/
dist/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work.sum
"@ | Out-File -FilePath $projectGitignore -Encoding UTF8
    }

    # 创建全局 gitignore
    $globalGitignore = "$GOW_HOME\.gitignore"
    if (-not (Test-Path $globalGitignore)) {
        @"
# gow 全局忽略文件
cache/
versions/
*.log
"@ | Out-File -FilePath $globalGitignore -Encoding UTF8
    }

    Write-Success "gow 初始化完成"
    Write-Info "运行 'gow doctor' 检查环境配置"
}

# 环境检查
function Invoke-Doctor {
    Write-Info "正在检查 gow 环境..."

    Write-Host ""
    Write-Host "=== 系统信息 ==="
    Write-Host "平台: $(Get-Platform)"
    Write-Host "Shell: $env:COMSPEC"

    Write-Host ""
    Write-Host "=== Go 版本信息 ==="

    # 检查系统 Go
    if (Test-SystemGo) {
        try {
            $systemGo = go version
            Write-Host "系统 Go: $systemGo"
        } catch {
            Write-Host "系统 Go: 版本检测失败"
        }
    } else {
        Write-Host "系统 Go: 未安装"
    }

    # 检查当前项目版本
    $currentVersion = Get-CurrentGoVersion
    if ($currentVersion) {
        Write-Host "项目配置版本: go$currentVersion"
    } else {
        Write-Host "项目配置版本: 未设置"
    }

    Write-Host ""
    Write-Host "=== gow 管理的版本 ==="
    $installedVersions = Get-InstalledVersions
    if ($installedVersions) {
        Write-Host "已安装的版本:"
        foreach ($version in $installedVersions) {
            $versionName = $version -replace '^go', ''
            if ($versionName -eq $currentVersion) {
                Write-Host "  ✓ $version (当前)"
            } else {
                Write-Host "    $version"
            }
        }
    } else {
        Write-Host "已安装的版本: 无"
    }

    Write-Host ""
    Write-Host "=== 配置文件 ==="
    if (Test-Path $GO_VERSION_FILE) {
        $content = Get-Content $GO_VERSION_FILE -Raw
        Write-Host "项目版本文件: $GO_VERSION_FILE ($($content.Trim()))"
    } else {
        Write-Host "项目版本文件: 不存在"
    }

    if (Test-Path $GOW_GLOBAL_VERSION_FILE) {
        $content = Get-Content $GOW_GLOBAL_VERSION_FILE -Raw
        Write-Host "全局版本文件: $GOW_GLOBAL_VERSION_FILE ($($content.Trim()))"
    } else {
        Write-Host "全局版本文件: 不存在"
    }

    Write-Host ""
    Write-Host "=== 目录结构 ==="
    Write-Host "gow 主目录: $GOW_HOME"
    Write-Host "版本目录: $GOW_VERSIONS_DIR"
    Write-Host "缓存目录: $GOW_CACHE_DIR"

    Write-Host ""
    Write-Host "=== 路径检查 ==="
    $goBinary = Get-GoBinary
    if ($goBinary) {
        Write-Host "当前 Go 二进制: $goBinary"
        if (Test-Path $goBinary) {
            try {
                $actualVersion = (& $goBinary version)
                Write-Host "版本验证: $actualVersion"
                Write-Success "环境配置正常"
            } catch {
                Write-Error "版本验证失败: $($_.Exception.Message)"
            }
        } else {
            Write-Error "Go 二进制不可执行: $goBinary"
        }
    } else {
        Write-Error "未找到可用的 Go 二进制"
    }
}

# 列出远程可用版本
function Get-RemoteVersions {
    Write-Info "正在获取可用版本列表..."

    try {
        if (Get-Command curl -ErrorAction SilentlyContinue) {
            $response = curl -L -s "https://golang.org/VERSION?m=text" 2>$null
            if ($response) {
                $lines = $response -split "`n"
                for ($i = 0; $i -lt [Math]::Min(10, $lines.Count); $i++) {
                    $line = $lines[$i].Trim()
                    if ($line -match '^go(\d+\.\d+(\.\d+)?)') {
                        $matches[1]
                    }
                }
            } else {
                Write-Warn "无法获取版本列表"
            }
        } else {
            Write-Warn "需要 curl 来获取远程版本列表"
            Write-Info "访问 https://golang.org/dl/ 查看可用版本"
        }
    } catch {
        Write-Error "获取远程版本失败: $($_.Exception.Message)"
    }
}

# 清理缓存
function Clear-Cache {
    Write-Info "正在清理 gow 缓存..."

    if (Test-Path $GOW_CACHE_DIR) {
        try {
            $cacheSize = (Get-ChildItem -Path $GOW_CACHE_DIR -Recurse -File | Measure-Object -Property Length -Sum).Sum
            Remove-Item -Path "$GOW_CACHE_DIR\*" -Recurse -Force
            if ($cacheSize -gt 0) {
                $sizeString = [math]::Round($cacheSize / 1MB, 2)
                Write-Success "缓存已清理 (释放空间: ${sizeString}MB)"
            } else {
                Write-Info "缓存已清理"
            }
        } catch {
            Write-Warn "清理缓存时出错: $($_.Exception.Message)"
        }
    } else {
        Write-Info "缓存目录不存在"
    }

    # 清理 Go 构建缓存
    $goBinary = Get-GoBinary
    if ($goBinary -and (Test-Path $goBinary)) {
        try {
            Write-Info "清理 Go 构建缓存..."
            & $goBinary clean -cache 2>$null
            & $goBinary clean -modcache 2>$null
        } catch {
            # 忽略清理错误
        }
    }
}

# 确保使用正确的 Go 版本
function Invoke-EnsureGoVersion {
    $goBinary = Get-GoBinary
    if (-not $goBinary) {
        Write-Error "未找到 Go，请安装 Go 或使用 'gow install <version>'"
        exit 1
    }

    if (-not (Test-Path $goBinary)) {
        Write-Error "Go 二进制不可执行: $goBinary"
        exit 1
    }

    # 设置环境变量
    $versionDir = Split-Path (Split-Path $goBinary -Parent) -Parent
    $env:GOROOT = $versionDir

    # 确保 gow 的 Go bin 目录在 PATH 最前面
    $env:PATH = "$versionDir\bin;$env:PATH"

    Write-Debug "使用 Go: $goBinary"
    Write-Debug "GOROOT: $env:GOROOT"
    Write-Debug "PATH: $env:PATH"
}

# 管理依赖
function Manage-Dependencies {
    param([string]$Action, [string[]]$RemainingArgs)

    Invoke-EnsureGoVersion
    $goBinary = Get-GoBinary

    switch ($Action) {
        "update" {
            Write-Info "更新依赖..."
            & $goBinary get -u ./...
            & $goBinary mod tidy
        }
        "vendor" {
            Write-Info "创建 vendor 目录..."
            & $goBinary mod vendor
        }
        "verify" {
            Write-Info "验证依赖..."
            & $goBinary mod verify
        }
        "download" {
            Write-Info "下载依赖..."
            & $goBinary mod download
        }
        "why" {
            if (-not $RemainingArgs -or $RemainingArgs.Count -eq 0) {
                Write-Error "请指定要查询的包"
                return
            }
            Write-Info "查询依赖关系: $($RemainingArgs -join ' ')"
            & $goBinary mod why $RemainingArgs
        }
        "graph" {
            Write-Info "显示依赖图..."
            & $goBinary mod graph
        }
        default {
            Write-Host "可用操作:"
            Write-Host "  update    - 更新所有依赖"
            Write-Host "  vendor    - 创建 vendor 目录"
            Write-Host "  verify    - 验证依赖"
            Write-Host "  download  - 下载依赖"
            Write-Host "  why <pkg> - 查询为什么需要某个包"
            Write-Host "  graph     - 显示依赖图"
        }
    }
}

# 自我更新
function Invoke-SelfUpdate {
    Write-Info "正在更新 gow..."

    $scriptUrl = "https://raw.githubusercontent.com/all3n/buildfly/main/gow.ps1"
    $currentScript = $PSCommandPath

    try {
        if (Get-Command curl -ErrorAction SilentlyContinue) {
            $tempFile = [System.IO.Path]::GetTempFileName()
            if (curl -s -o $tempFile $scriptUrl) {
                if ((Get-Item $tempFile).Length -gt 0) {
                    Copy-Item $tempFile $currentScript -Force
                    Write-Success "gow 更新完成"
                } else {
                    Write-Error "下载的文件为空"
                }
                Remove-Item $tempFile -ErrorAction SilentlyContinue
            } else {
                Write-Error "下载失败"
                Remove-Item $tempFile -ErrorAction SilentlyContinue
            }
        } else {
            Write-Error "需要 curl 来更新 gow"
        }
    } catch {
        Write-Error "更新失败: $($_.Exception.Message)"
    }
}

# 主函数
function Main {
    if (-not $Command) {
        Show-Help
        return
    }

    switch ($Command) {
        "help" {
            Show-Help
        }
        "-h" {
            Show-Help
        }
        "--help" {
            Show-Help
        }
        "init" {
            Initialize-Gow
        }
        "version" {
            Write-Host "Go Wrapper v$GOW_VERSION"
            $goBinary = Get-GoBinary
            if ($goBinary -and (Test-Path $goBinary)) {
                & $goBinary version
            } else {
                Write-Host "Go 未安装或不可用"
            }
        }
        "-v" {
            Write-Host "Go Wrapper v$GOW_VERSION"
            $goBinary = Get-GoBinary
            if ($goBinary -and (Test-Path $goBinary)) {
                & $goBinary version
            } else {
                Write-Host "Go 未安装或不可用"
            }
        }
        "--version" {
            Write-Host "Go Wrapper v$GOW_VERSION"
            $goBinary = Get-GoBinary
            if ($goBinary -and (Test-Path $goBinary)) {
                & $goBinary version
            } else {
                Write-Host "Go 未安装或不可用"
            }
        }
        "current" {
            $version = Get-CurrentGoVersion
            if ($version) {
                Write-Host "go$version"
            } else {
                Write-Host "未设置 Go 版本"
            }
        }
        "use" {
            if (-not $Arguments -or $Arguments.Count -eq 0) {
                Write-Error "用法: gow use <version>"
                exit 1
            }
            Use-GoVersion $Arguments[0]
        }
        "install" {
            if (-not $Arguments -or $Arguments.Count -eq 0) {
                Write-Error "用法: gow install <version> [amd64|386|arm64]"
                Write-Host "示例: gow install 1.21.0"
                Write-Host "      gow install 1.21.0 amd64"
                Write-Host "      gow install 1.21.0 386"
                Write-Host "      gow install 1.21.0 arm64"
                exit 1
            }
            $version = $Arguments[0]
            $architecture = if ($Arguments.Count -gt 1) { $Arguments[1] } else { "amd64" }
            if ($architecture -notin @("amd64", "386", "arm64")) {
                Write-Error "不支持的架构: $architecture (支持: amd64, 386, arm64)"
                exit 1
            }
            Install-GoVersion $version $architecture
        }
        "uninstall" {
            if (-not $Arguments -or $Arguments.Count -eq 0) {
                Write-Error "用法: gow uninstall <version>"
                exit 1
            }
            Uninstall-GoVersion $Arguments[0]
        }
        "list" {
            Write-Host "已安装的 Go 版本:"
            $installedVersions = Get-InstalledVersions
            $currentVersion = Get-CurrentGoVersion
            if ($installedVersions) {
                foreach ($version in $installedVersions) {
                    $versionName = $version -replace '^go', ''
                    if ($versionName -eq $currentVersion) {
                        Write-Host "  ✓ $version (当前)"
                    } else {
                        Write-Host "    $version"
                    }
                }
            } else {
                Write-Host "已安装的版本: 无"
            }
        }
        "list-remote" {
            Get-RemoteVersions
        }
        "doctor" {
            Invoke-Doctor
        }
        "deps" {
            if ($Arguments -and $Arguments.Count -gt 0) {
                $action = $Arguments[0]
                $remainingArgs = $Arguments | Select-Object -Skip 1
                Manage-Dependencies -Action $action -RemainingArgs $remainingArgs
            } else {
                Manage-Dependencies -Action ""
            }
        }
        "clean" {
            Clear-Cache
        }
        "self-update" {
            Invoke-SelfUpdate
        }
        "update" {
            Write-Warn "'update' 命令已弃用，请使用 'self-update'"
            Invoke-SelfUpdate
        }
        "build" {
            Invoke-EnsureGoVersion
            $goBinary = Get-GoBinary
            & $goBinary $Command $Arguments
        }
        "run" {
            Invoke-EnsureGoVersion
            $goBinary = Get-GoBinary
            & $goBinary $Command $Arguments
        }
        "test" {
            Invoke-EnsureGoVersion
            $goBinary = Get-GoBinary
            & $goBinary $Command $Arguments
        }
        "mod" {
            Invoke-EnsureGoVersion
            $goBinary = Get-GoBinary
            & $goBinary $Command $Arguments
        }
        default {
            # 如果是文件，直接运行
            if (Test-Path $Command) {
                Invoke-EnsureGoVersion
                $goBinary = Get-GoBinary
                & $goBinary run @($Command) + $Arguments
            } else {
                Write-Error "未知命令: $Command"
                Write-Host ""
                Show-Help
                exit 1
            }
        }
    }
}

# 运行主函数
Main
