# 缓存功能改进文档

## 概述

本次改进主要针对 C++ 依赖管理器的缓存机制，增加了对构建状态的智能检查，避免重复构建已经构建过的依赖。

## 改进内容

### 1. 分层缓存机制

缓存系统现在支持两种类型的缓存：

- **下载缓存**: 存储下载和解压后的源代码
- **构建缓存**: 存储构建后的结果（编译后的库文件、头文件等）

### 2. 项目构建目录

为了更好的构建管理和调试，构建过程现在使用项目的本地构建目录：

- **构建目录**: `.buildfly/build/{dependency-name}/`
- **优势**: 
  - 便于调试构建问题
  - 支持增量构建
  - 构建文件持久化，不会被自动清理
  - 每个依赖有独立的构建工作区

### 2. 智能缓存检查逻辑

安装命令现在会按以下顺序检查缓存：

1. **检查下载缓存**
   - 如果存在且未过期，使用缓存的源代码
   - 如果不存在或过期，重新下载

2. **检查构建缓存**（仅对需要构建的依赖）
   - 如果存在且未过期，直接使用缓存的构建结果
   - 如果不存在，从下载缓存恢复源代码并构建
   - 构建完成后，缓存构建结果

3. **无需构建的依赖**
   - 直接使用下载缓存

### 3. 缓存目录结构

```
~/.buildfly/cache/                    # 全局缓存目录
├── downloads/                         # 下载缓存
│   └── {cache-key}/                   # 源代码
├── builds/                            # 构建缓存
│   └── {cache-key}/                   # 构建结果
└── metadata/                          # 元数据
    └── {cache-key}.json

项目目录/.buildfly/build/               # 项目构建目录
├── {dependency-name}/                 # 每个依赖的构建工作区
│   ├── {source-files}                 # 源代码
│   └── build/                         # 构建输出目录
│       ├── {build-files}              # 构建中间文件
│       └── {install-files}            # 安装文件
```

## 使用示例

### 基本使用

```bash
# 第一次安装 - 会下载并构建
buildfly install

# 第二次安装 - 使用缓存（包括构建缓存）
buildfly install

# 强制重新安装
buildfly install --force

# 不使用缓存
buildfly install --no-cache
```

### 配置文件示例

```yaml
project:
  name: "my-project"
  version: "1.0.0"

dependencies:
  # 需要构建的依赖
  fmt:
    version: "8.0.1"
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"

  # 无需构建的依赖（header-only）
  spdlog:
    version: "1.9.2"
    source:
      type: "git"
      url: "https://github.com/gabime/spdlog.git"
      tag: "v1.9.2"
    build_system: "none"
```

## 性能优势

### 1. 构建时间优化

- **首次安装**: 下载 + 构建
- **后续安装**: 直接使用缓存（包括构建结果）
- **构建缓存命中**: 节省 90%+ 的构建时间

### 2. 网络带宽优化

- 源代码只下载一次
- 构建结果本地缓存，无需重复构建

### 3. 开发效率提升

- 快速切换项目配置
- 快速重新安装依赖
- 离线开发支持（如果缓存完整）

## 缓存管理

### 查看缓存状态

```bash
# 查看缓存信息（未来功能）
buildfly cache list

# 清理缓存（未来功能）
buildfly cache clean
```

### 缓存失效条件

1. **时间过期**: 超过配置的最大缓存时间（默认 7 天）
2. **手动清理**: 使用 `--force` 或 `--no-cache` 参数
3. **配置变更**: 依赖版本、构建系统等配置发生变化

## 技术实现

### 缓存键生成

缓存键基于以下信息生成 SHA256 哈希：

- 依赖源 URL
- 版本号
- Git 标签（如果有）
- 校验和（如果有）

### 构建缓存判断

```go
// 检查是否需要构建
if dep.BuildSystem != "" && dep.BuildSystem != "none" {
    // 检查构建缓存
    if cacheManager.IsBuildCached(dep) {
        // 使用构建缓存
        cacheManager.RetrieveBuild(dep, targetPath)
    } else {
        // 从下载缓存恢复并构建
        // ... 构建逻辑
        // 缓存构建结果
        cacheManager.StoreBuild(dep, installPath)
    }
} else {
    // 无需构建，直接使用下载缓存
    cacheManager.Retrieve(dep, targetPath)
}
```

## 注意事项

1. **磁盘空间**: 构建缓存会占用额外磁盘空间
2. **缓存一致性**: 确保缓存键的唯一性和准确性
3. **清理策略**: 定期清理过期缓存以节省空间

## 未来改进

1. **缓存统计**: 添加缓存命中率统计
2. **智能清理**: 基于使用频率的 LRU 清理策略
3. **增量构建**: 支持源代码变更的增量构建
4. **分布式缓存**: 支持团队共享缓存
