# Pixi Backend 实现文档

## 概述

Pixi backend 为 buildfly 提供了基于 Pixi 的 C++ 虚拟环境管理功能。Pixi 是一个现代的 Python 包管理器，支持 conda 生态系统，特别适合管理复杂的 C++ 构建环境。

## 架构设计

### 1. Backend 接口

Pixi backend 实现了 `VEnvBackend` 接口，提供统一的后端功能：

```go
type VEnvBackend interface {
    GetType() BackendType
    GetInfo() *BackendInfo
    Install(version string) error
    Initialize(config *VEnvConfig, rootDir string) error
    Run(args []string) error
    InstallTool(toolName, version string) error
}
```

### 2. 核心组件

#### PixiBackend
- 实现所有 VEnvBackend 接口方法
- 管理 Pixi 项目的生命周期
- 处理依赖安装和环境配置

#### PixiDetector
- 检测系统中 Pixi 的安装状态
- 支持自动安装 Pixi
- 跨平台兼容性

#### BackendFactory
- 创建和管理不同类型的 backend
- 自动检测最佳可用 backend
- 提供 backend 切换功能

## 配置结构

### PixiConfig

```go
type PixiConfig struct {
    Version string                 `yaml:"version"`
    Config  map[string]interface{} `yaml:"config"`
}
```

### VEnvConfig 扩展

```go
type VEnvConfig struct {
    Backend    BackendType `yaml:"backend"`
    Pixi       PixiConfig  `yaml:"pixi"`
    // ... 其他字段
}
```

## 功能特性

### 1. 项目初始化

```bash
# 使用 Pixi backend 初始化
buildfly venv init --backend pixi

# 指定 Pixi 版本
buildfly venv init --backend pixi --pixi-version 0.35.1
```

### 2. 自动安装 Pixi

- 检测系统中是否已安装 Pixi
- 未安装时自动下载并安装
- 支持指定版本安装

### 3. 项目配置管理

- 自动生成 `pixi.toml` 配置文件
- 支持多平台依赖配置
- 灵活的任务定义

### 4. 依赖管理

- Python 包依赖
- C++ 构建工具（CMake、Ninja、编译器）
- 平台特定的依赖配置
- 测试框架集成

## 配置示例

### 基本配置

```yaml
venv:
  enabled: true
  backend: "pixi"
  pixi:
    version: "0.35.1"
    config:
      name: "my-project"
      version: "0.1.0"
      channels: ["conda-forge"]
      platforms: ["linux-64", "osx-arm64"]
      dependencies:
        python: "3.11.*"
        cmake: ">=3.20"
        ninja: "*"
```

### 高级配置

```yaml
venv:
  enabled: true
  backend: "pixi"
  pixi:
    version: "0.35.1"
    config:
      # 项目信息
      name: "my-project"
      version: "0.1.0"
      description: "C++ project with Pixi"
      authors: ["developer@example.com"]
      
      # 通道配置
      channels: ["conda-forge", "rapidsai"]
      
      # 平台支持
      platforms: ["linux-64", "osx-64", "osx-arm64", "win-64"]
      
      # 依赖配置
      dependencies:
        python: "3.11.*"
        cmake: ">=3.20"
        ninja: "*"
        pkg-config: "*"
        gtest: ">=1.17.0,<2"
        
      # 构建依赖（平台特定）
      build_dependencies:
        - target: "linux-64"
          dependencies:
            gcc_linux-64: ">=11"
            gxx_linux-64: ">=11"
            
        - target: "osx-arm64"
          dependencies:
            clang_osx-arm64: ">=11"
            clangxx_osx-arm64: ">=11"
      
      # 任务定义
      tasks:
        configure: "cmake -B build ."
        build: "cmake --build build"
        test: "cd build && ctest"
        clean: "rm -rf build"
        dev:
          depends_on: ["configure", "build"]
```

## 安装脚本

### Linux/macOS

Pixi 提供了便捷的安装脚本：

```bash
curl -fsSL https://pixi.sh/install.sh | sh
```

支持环境变量：
- `PIXI_VERSION`: 指定版本
- `PIXI_HOME`: 安装目录
- `PIXI_ARCH`: 架构

### Windows

```powershell
powershell -ExecutionPolicy ByPass -c "irm https://pixi.sh/install.ps1 | iex"
```

## 使用流程

### 1. 初始化项目

```bash
# 创建新项目
mkdir my-cpp-project
cd my-cpp-project

# 使用 Pixi backend 初始化
buildfly venv init --backend pixi
```

### 2. 配置依赖

编辑 `buildfly.yaml`：

```yaml
venv:
  enabled: true
  backend: "pixi"
  pixi:
    config:
      dependencies:
        python: "3.11.*"
        cmake: ">=3.20"
        ninja: "*"
```

### 3. 激活环境

```bash
# 激活虚拟环境
buildfly venv activate

# 或者直接运行命令
buildfly venv run cmake --version
```

### 4. 构建项目

```bash
# 使用 Pixi 任务构建
buildfly venv run pixi run build

# 或者直接使用 CMake
buildfly venv run cmake -B build .
buildfly venv run cmake --build build
```

## 优势特点

### 1. 跨平台一致性

- 统一的配置格式
- 平台特定的依赖处理
- 一致的构建体验

### 2. 依赖隔离

- 完全隔离的构建环境
- 避免系统依赖冲突
- 可重现的构建

### 3. 现代工具链

- 基于 conda-forge 生态
- 快速的依赖解析
- 高效的包管理

### 4. 开发友好

- 简单的配置语法
- 丰富的预构建包
- 活跃的社区支持

## 与 UV Backend 对比

| 特性 | Pixi Backend | UV Backend |
|------|--------------|------------|
| 包生态 | conda-forge | PyPI |
| 依赖解析 | SAT 求解器 | 类 pip |
| 启动速度 | 中等 | 快 |
| 跨平台支持 | 优秀 | 良好 |
| C++ 工具支持 | 丰富 | 中等 |
| 配置复杂度 | 中等 | 简单 |

## 最佳实践

### 1. 项目结构

```
my-project/
├── buildfly.yaml
├── CMakeLists.txt
├── src/
├── tests/
└── .buildfly/
    └── root/          # Pixi 环境
        └── pixi.toml  # Pixi 配置
```

### 2. 版本管理

- 固定 Pixi 版本确保一致性
- 使用语义化版本号
- 定期更新依赖

### 3. 依赖配置

- 明确指定版本范围
- 分离构建和运行时依赖
- 利用平台特定依赖

### 4. 任务定义

- 定义常用的构建任务
- 使用任务依赖关系
- 保持任务简洁

## 故障排除

### 1. 安装问题

```bash
# 检查 Pixi 安装
buildfly venv status

# 重新安装
buildfly venv init --backend pixi --force
```

### 2. 依赖冲突

```bash
# 查看依赖树
buildfly venv run pixi tree

# 更新依赖
buildfly venv run pixi update
```

### 3. 平台问题

- 检查平台配置
- 验证依赖可用性
- 使用正确的包名

## 未来规划

### 1. 功能增强

- 更多的预配置模板
- 集成 CI/CD 支持
- 性能优化

### 2. 生态集成

- 与更多构建系统集成
- 支持更多包源
- 改进错误处理

### 3. 用户体验

- 更好的进度显示
- 智能依赖推荐
- 简化配置语法

## 总结

Pixi backend 为 buildfly 提供了强大而灵活的 C++ 虚拟环境管理能力。通过集成 conda-forge 生态系统，它能够轻松处理复杂的依赖关系，提供跨平台一致的构建体验。无论是简单的 C++ 项目还是复杂的跨平台应用，Pixi backend 都能提供可靠的解决方案。
