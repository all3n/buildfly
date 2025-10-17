# UV 虚拟环境实现文档

## 概述

本文档描述了 BuildFly 中基于 UV 的 C++ 虚拟环境功能的实现。该功能提供了隔离的 C++ 构建环境，支持自动安装和管理构建工具。

## 架构设计

### 核心组件

1. **pkg/venv/types.go** - 定义虚拟环境相关的数据结构
2. **pkg/venv/uv_detector.go** - UV 工具检测和安装逻辑
3. **pkg/venv/manager.go** - 虚拟环境管理器
4. **cmd/cli/venv.go** - CLI 命令接口
5. **pkg/config/types.go** - 配置类型扩展

### 目录结构

```
.buildfly/root/                 # 虚拟环境根目录
├── bin/                        # 可执行文件
├── lib/                        # 库文件
├── include/                    # 头文件
├── share/                      # 共享资源
├── tools/                      # 构建工具
├── env/                        # Python 环境
├── activate.sh                 # Unix/Linux/macOS 激活脚本
├── activate.bat                # Windows 激活脚本
└── environment.json            # 环境信息文件
```

## 功能特性

### 1. UV 工具管理

- **自动检测**: 检测系统是否已安装 UV
- **自动安装**: 如果未安装则自动安装最新版本
- **版本管理**: 支持指定 UV 版本
- **跨平台**: 支持 Linux、macOS 和 Windows

### 2. 虚拟环境管理

- **初始化**: 创建隔离的构建环境
- **激活/停用**: 管理环境状态
- **状态查看**: 显示环境详细信息
- **重置功能**: 完全重置环境

### 3. 构建工具支持

- **CMake**: 自动安装和管理 CMake
- **Ninja**: 支持 Ninja 构建系统
- **编译器**: 支持 GCC、Clang、MSVC
- **Python**: 集成 Python 环境管理

## 配置系统

### YAML 配置示例

```yaml
project:
  name: "my-cpp-project"
  version: "1.0.0"

venv:
  enabled: true
  uv_version: "latest"
  root_dir: ".buildfly/root"
  auto_activate: false
  cpp_tools:
    cmake:
      version: "3.28.0"
      enabled: true
    ninja:
      version: "1.11.1"
      enabled: true
    gcc:
      version: "13.2.0"
      enabled: false
    clang:
      version: "17.0.6"
      enabled: false
  python:
    version: "3.11"
    packages:
      - "cmake"
      - "ninja"
```

### 配置字段说明

- `enabled`: 是否启用虚拟环境
- `uv_version`: UV 工具版本
- `root_dir`: 环境根目录
- `auto_activate`: 是否自动激活
- `cpp_tools`: C++ 工具配置
- `python`: Python 环境配置

## CLI 命令

### 基本命令

```bash
# 初始化虚拟环境
buildfly venv init

# 激活虚拟环境
buildfly venv activate

# 停用虚拟环境
buildfly venv deactivate

# 查看环境状态
buildfly venv status

# 列出已安装工具
buildfly venv list

# 重置环境
buildfly venv reset
```

### 高级选项

```bash
# 强制重新初始化
buildfly venv init --force

# 指定 UV 版本
buildfly venv init --uv-version 0.8.0

# 指定环境根目录
buildfly venv init --root-dir /path/to/env

# JSON 格式输出
buildfly venv status --json

# 详细信息输出
buildfly venv status --verbose
```

## 环境变量

### 系统环境变量

虚拟环境激活后会设置以下环境变量：

- `BUILDFLY_ENV_ROOT`: 环境根目录
- `PATH`: 添加环境 bin 目录
- `CC`: C 编译器路径（如果已安装）
- `CXX`: C++ 编译器路径（如果已安装）
- `CMAKE_PREFIX_PATH`: CMake 路径（如果已安装）
- `PYTHONPATH`: Python 路径（如果已安装）

### 激活脚本

#### Unix/Linux/macOS (activate.sh)

```bash
#!/bin/bash
export BUILDFLY_ENV_ROOT="/path/to/.buildfly/root"
export PATH="/path/to/.buildfly/root/bin:$PATH"

# 设置编译器环境变量
if [ -f "/path/to/.buildfly/root/bin/gcc" ]; then
    export CC="/path/to/.buildfly/root/bin/gcc"
    export CXX="/path/to/.buildfly/root/bin/g++"
fi

# 设置 CMake 环境变量
if [ -d "/path/to/.buildfly/root/tools/cmake" ]; then
    export CMAKE_PREFIX_PATH="/path/to/.buildfly/root/tools/cmake:$CMAKE_PREFIX_PATH"
    export PATH="/path/to/.buildfly/root/tools/cmake/bin:$PATH"
fi
```

#### Windows (activate.bat)

```batch
@echo off
set BUILDFLY_ENV_ROOT=C:\path\to\.buildfly\root
set PATH=C:\path\to\.buildfly\root\bin;%PATH%

REM 设置编译器环境变量
if exist "C:\path\to\.buildfly\root\bin\gcc.exe" (
    set CC=C:\path\to\.buildfly\root\bin\gcc.exe
    set CXX=C:\path\to\.buildfly\root\bin\g++.exe
)
```

## 集成方式

### 1. 依赖管理集成

虚拟环境与依赖管理系统集成，确保依赖在隔离环境中构建：

```go
// 在构建执行器中使用虚拟环境
if venvManager.IsActivated() {
    envVars, _ := venvManager.GetEnvironmentVars()
    for key, value := range envVars {
        os.Setenv(key, value)
    }
}
```

### 2. 构建系统集成

构建系统自动检测虚拟环境状态：

```go
// 检查是否在虚拟环境中
if venvManager != nil && venvManager.IsActivated() {
    // 使用虚拟环境中的工具
    cmakePath := filepath.Join(venvManager.GetRootDir(), "tools", "cmake", "bin", "cmake")
}
```

## 错误处理

### 常见错误

1. **UV 未安装**: 自动安装 UV
2. **权限不足**: 提示用户检查权限
3. **网络问题**: 提供离线安装选项
4. **路径冲突**: 检查路径可用性

### 错误恢复

- 自动重试机制
- 降级处理方案
- 详细的错误信息
- 建议的解决方案

## 性能优化

### 缓存策略

- UV 工具缓存
- 构建工具缓存
- 依赖包缓存
- 环境状态缓存

### 并发处理

- 并发下载工具
- 并行安装依赖
- 异步状态检查

## 安全考虑

### 路径安全

- 验证路径合法性
- 防止路径遍历
- 检查权限设置

### 工具验证

- 校验和验证
- 签名检查
- 版本兼容性

## 扩展性

### 插件系统

- 自定义工具安装器
- 第三方工具集成
- 平台特定扩展

### 配置扩展

- 自定义配置字段
- 环境特定配置
- 用户偏好设置

## 测试

### 单元测试

- UV 检测器测试
- 环境管理器测试
- 配置加载测试

### 集成测试

- 完整工作流测试
- 跨平台兼容性测试
- 性能基准测试

## 未来计划

### 短期目标

- 完善工具安装逻辑
- 增加更多构建工具支持
- 优化性能和稳定性

### 长期目标

- 容器化支持
- 云端环境同步
- GUI 管理界面

## 总结

UV 虚拟环境功能为 BuildFly 提供了强大的环境隔离能力，确保 C++ 项目构建的一致性和可重现性。通过自动化工具管理和环境配置，大大简化了 C++ 项目的设置和维护工作。

该实现遵循了以下设计原则：

1. **简洁性**: 易于使用和配置
2. **灵活性**: 支持多种工具和平台
3. **可靠性**: 错误处理和恢复机制
4. **扩展性**: 支持未来功能扩展
5. **性能**: 优化的缓存和并发处理

通过这些特性，BuildFly 的虚拟环境功能能够满足从简单项目到复杂企业级应用的各种需求。
