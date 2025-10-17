# Build Tag 功能实现文档

## 概述

Build Tag 是 buildfly 项目中的一个核心功能，用于区分不同的构建配置。它允许用户根据架构、平台、编译器、运行时等参数来构建和安装不同版本的依赖库。

## 功能特性

### 1. 构建标签格式
Build Tag 使用键值对格式，支持多种构建参数：

```
arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17,abi=sysv
```

### 2. 支持的参数

#### 基础参数
- **arch**: 架构 (x86_64, arm64, x64, i386, aarch64, universal)
- **platform**: 平台 (linux, darwin, windows)
- **runtime**: 运行时 (glibc, libcxx, msvcrt, musl)
- **compiler**: 编译器 (gcc, clang, msvc, nvcc, apple-clang)
- **std**: C++ 标准 (cpp11, cpp14, cpp17, cpp20, cpp23)
- **abi**: ABI (sysv, macho, msabi)
- **target**: 平台特定目标 (如 macos14.0)

#### GPU 相关参数
- **cuda**: CUDA 版本 (如 12.0)
- **cuda_arch**: CUDA 架构 (如 compute_80|compute_90)
- **rocm**: ROCm 版本 (如 5.0)
- **rocm_arch**: ROCm 架构 (如 gfx900|gfx1030)
- **opencl**: OpenCL 版本 (如 2.0)
- **gpu_backend**: GPU 后端 (cuda, rocm, opencl, none)
- **gpu_enabled**: 是否启用 GPU (true, false)

### 3. 版本范围支持
- **+**: 版本或更高 (如 gcc_11+)
- **|**: 多个选择 (如 compute_80|compute_90)
- **,**: 分隔多个键值对

## 实现架构

### 1. 核心数据结构

```go
type BuildTag struct {
    Arch     string   `json:"arch,omitempty"`
    Platform string   `json:"platform,omitempty"`
    Runtime  string   `json:"runtime,omitempty"`
    Compiler string   `json:"compiler,omitempty"`
    Std      string   `json:"std,omitempty"`
    ABI      string   `json:"abi,omitempty"`
    Target   string   `json:"target,omitempty"`
    GPU      *GPUInfo `json:"gpu,omitempty"`
}

type GPUInfo struct {
    Backend string      `json:"backend"`
    CUDA    *CUDABackend  `json:"cuda,omitempty"`
    ROCm    *ROCmBackend  `json:"rocm,omitempty"`
    OpenCL  *OpenCLBackend `json:"opencl,omitempty"`
}
```

### 2. 核心功能

#### 解析功能
- `ParseBuildTag()`: 将字符串解析为 BuildTag 结构体
- 支持复杂的嵌套结构（如 GPU 配置）
- 自动验证格式和参数有效性

#### 目录生成
- `ToDirName()`: 将 BuildTag 转换为目录名
- 自动处理特殊字符（+ → plus, | → or）
- 确保目录名在不同操作系统上的兼容性

#### 验证功能
- `Validate()`: 验证 BuildTag 的有效性
- 检查参数值是否在允许范围内
- 提供详细的错误信息

#### 字符串表示
- `String()`: 将 BuildTag 转换回字符串格式
- 保持与输入格式的一致性

### 3. 集成点

#### 配置系统集成
- 在 `ProjectConfig` 中添加 `BuildTag` 字段
- 支持 YAML 配置文件中的 build tag 定义
- 支持命令行参数覆盖配置

#### 变量系统集成
- 在 `VariableContext` 中集成 BuildTag
- 自动设置相关的环境变量
- 支持构建脚本中的变量替换

#### 安装系统集成
- 在 `install.go` 中添加 build tag 处理逻辑
- 根据构建标签调整构建和安装目录
- 支持缓存系统的 build tag 感知

## 使用方式

### 1. 配置文件方式

```yaml
project:
  name: "my-project"
  build_tag:
    arch: "x86_64"
    platform: "linux"
    runtime: "glibc_2.35"
    compiler: "gcc_11"
    std: "cpp17"
    abi: "sysv"
```

### 2. 命令行方式

```bash
# 基本使用
buildfly install --build-tag "arch=x86_64,platform=linux"

# 复杂配置
buildfly install --build-tag "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17,abi=sysv"

# CUDA 支持
buildfly install --build-tag "arch=x86_64,platform=linux,cuda=12.0,cuda_arch=compute_80|compute_90,gpu_enabled=true,gpu_backend=cuda"

# macOS 支持
buildfly install --build-tag "arch=arm64,platform=darwin,runtime=libcxx_15,compiler=apple-clang_15.0,std=cpp20,abi=macho,target=macos14.0"
```

### 3. 目录结构

使用 build tag 后，目录结构会自动包含构建标签信息：

```
build/
├── fmt/
│   └── 8.0.1/
│       └── arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv/
└── zlib/
    └── 1.2.11/
        └── arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv/

install/
└── arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv/
    ├── fmt/
    └── zlib/
```

## 最佳实践

### 1. 构建标签命名规范
- 使用一致的命名约定
- 包含足够的信息以区分不同的构建配置
- 避免过于复杂的组合

### 2. 配置管理
- 在配置文件中定义常用的构建标签
- 使用命令行参数进行临时覆盖
- 为不同的部署环境创建不同的配置文件

### 3. 缓存策略
- 利用 build tag 进行缓存隔离
- 定期清理不需要的构建缓存
- 监控缓存使用情况

### 4. 构建脚本
- 在构建脚本中使用构建标签变量
- 根据不同的构建标签调整构建参数
- 确保构建脚本的跨平台兼容性

## 示例场景

### 1. 多平台支持
```bash
# Linux 构建
buildfly install --build-tag "arch=x86_64,platform=linux"

# macOS 构建
buildfly install --build-tag "arch=arm64,platform=darwin"

# Windows 构建
buildfly install --build-tag "arch=x64,platform=windows"
```

### 2. 不同编译器支持
```bash
# GCC 构建
buildfly install --build-tag "arch=x86_64,platform=linux,compiler=gcc_11"

# Clang 构建
buildfly install --build-tag "arch=x86_64,platform=linux,compiler=clang_15"

# MSVC 构建
buildfly install --build-tag "arch=x64,platform=windows,compiler=msvc_19.38"
```

### 3. GPU 加速支持
```bash
# CUDA 支持
buildfly install --build-tag "arch=x86_64,platform=linux,cuda=12.0,cuda_arch=compute_80,gpu_backend=cuda"

# ROCm 支持
buildfly install --build-tag "arch=x86_64,platform=linux,rocm=5.0,rocm_arch=gfx900,gpu_backend=rocm"
```

## 测试覆盖

### 1. 单元测试
- BuildTag 解析测试
- 验证功能测试
- 目录名生成测试
- 字符串表示测试

### 2. 集成测试
- 配置文件加载测试
- 命令行参数测试
- 安装流程测试
- 缓存功能测试

### 3. 示例和演示
- 完整的配置文件示例
- 交互式演示脚本
- 各种使用场景的测试用例

## 性能考虑

### 1. 解析性能
- BuildTag 解析是 O(n) 复杂度
- 使用字符串操作优化性能
- 缓存解析结果以避免重复解析

### 2. 内存使用
- BuildTag 结构体大小合理
- 避免不必要的字符串复制
- 使用指针优化内存布局

### 3. 磁盘空间
- 根据构建标签隔离目录
- 避免重复下载和构建
- 支持缓存清理功能

## 扩展性

### 1. 新参数支持
- 可以轻松添加新的构建参数
- 保持向后兼容性
- 支持自定义验证规则

### 2. 新平台支持
- 可以支持新的操作系统和架构
- 添加新的运行时和编译器支持
- 扩展 GPU 后端支持

### 3. 插件系统
- 支持自定义构建标签解析器
- 允许第三方扩展
- 提供插件 API 接口

## 总结

Build Tag 功能为 buildfly 提供了强大的构建配置管理能力，支持复杂的跨平台构建需求。通过灵活的配置系统和完善的验证机制，确保了构建的可靠性和一致性。

该功能的设计考虑了易用性、性能和扩展性，为 C++ 项目的依赖管理提供了坚实的基础。
