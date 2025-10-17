# 校验和验证功能

Buildfly 现在支持对 archive 类型的依赖进行校验和验证，确保下载的文件完整性和安全性。

## 功能特性

- **多种哈希算法支持**: MD5、SHA1、SHA256、SHA512
- **灵活的配置方式**: 支持多种配置格式
- **向后兼容**: 保持与现有配置的兼容性
- **自动验证**: 下载后自动进行校验和验证

## 配置方式

### 1. 使用专用字段（推荐）

```yaml
dependencies:
  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
      md5: "1c9f418cd4baa9c3d7a0ea8fda56af76"
    build_system: "configure"

  fmt:
    version: "8.0.1"
    source:
      type: "archive"
      url: "https://github.com/fmtlib/fmt/archive/8.0.1.tar.gz"
      sha256: "3d5d4144db202c8929a21983c710bf7b9f83e6bb8e1a1f53d3be9c1c9d5d78c4"
    build_system: "cmake"
```

### 2. 使用通用校验和映射

```yaml
dependencies:
  nlohmann-json:
    version: "3.9.1"
    source:
      type: "archive"
      url: "https://github.com/nlohmann/json/archive/v3.9.1.tar.gz"
      checksums:
        sha256: "8b8e8e1c2e4c5f6e7a8b9c0d1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0"
        md5: "bb8e8e8e1c2e4c5f6e7a8b9c0d1f2a3b"
    build_system: "none"
```

### 3. 向后兼容的 hash 字段

```yaml
dependencies:
  spdlog:
    version: "1.9.2"
    source:
      type: "archive"
      url: "https://github.com/gabime/spdlog/archive/v1.9.2.tar.gz"
      hash: "8d8c8a8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8"
    build_system: "none"
```

## 支持的哈希算法

| 算法 | 配置字段 | 示例 |
|------|----------|------|
| MD5 | `md5` | `md5: "1c9f418cd4baa9c3d7a0ea8fda56af76"` |
| SHA1 | `sha1` | `sha1: "351b0f2b3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a"` |
| SHA256 | `sha256` | `sha256: "3d5d4144db202c8929a21983c710bf7b9f83e6bb8e1a1f53d3be9c1c9d5d78c4"` |
| SHA512 | `sha512` | `sha512: "8b8e8e1c2e4c5f6e7a8b9c0d1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6"` |

## 优先级规则

当存在多个校验和配置时，优先级如下：

1. **专用字段** (`md5`, `sha1`, `sha256`, `sha512`) - 最高优先级
2. **通用映射** (`checksums`) - 中等优先级
3. **向后兼容字段** (`hash`) - 最低优先级，默认为 SHA256

## 使用示例

### 基本使用

```bash
# 安装带有校验和的依赖
buildfly install zlib

# 强制重新下载并验证
buildfly install --force zlib

# 使用详细输出查看验证过程
buildfly install --verbose zlib
```

### 构建配置文件

```yaml
project:
  name: "my-project"
  version: "1.0.0"

dependencies:
  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
      md5: "1c9f418cd4baa9c3d7a0ea8fda56af76"
    build_system: "configure"

  fmt:
    version: "8.0.1"
    source:
      type: "archive"
      url: "https://github.com/fmtlib/fmt/archive/8.0.1.tar.gz"
      sha256: "3d5d4144db202c8929a21983c710bf7b9f83e6bb8e1a1f53d3be9c1c9d5d78c4"
    build_system: "cmake"

build_profiles:
  basic:
    dependencies:
      - "zlib"
  
  full:
    dependencies:
      - "zlib"
      - "fmt"
```

## 验证过程

1. **下载文件**: 从指定 URL 下载压缩包
2. **计算哈希**: 使用配置的算法计算文件哈希值
3. **比较验证**: 将计算结果与预期值比较
4. **结果处理**: 
   - 验证成功：继续解压和安装
   - 验证失败：停止安装并报告错误

## 错误处理

### 校验和不匹配

```
Error: archive verification failed: checksum mismatch: expected 1c9f418cd4baa9c3d7a0ea8fda56af76, got 2d0e529fd4baa9c3d7a0ea8fda56af77
```

### 不支持的算法

```
Error: archive verification failed: unsupported checksum algorithm: sha3
```

### 文件访问错误

```
Error: archive verification failed: failed to open file for checksum verification: no such file or directory
```

## 最佳实践

1. **使用强哈希算法**: 优先使用 SHA256 或 SHA512
2. **验证校验和来源**: 从官方渠道获取校验和值
3. **多种算法**: 对重要依赖可以使用多种算法验证
4. **定期更新**: 及时更新校验和值以匹配新版本
5. **测试配置**: 在生产环境使用前测试配置的正确性

## 示例项目

查看 `examples/checksum-demo/` 目录中的完整示例：

```bash
cd examples/checksum-demo
../../buildfly install --profile basic
```

这将演示如何使用不同的校验和配置方式安装依赖。
