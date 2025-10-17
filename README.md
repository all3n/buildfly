# BuildFly - C++ 依赖管理器

BuildFly 是一个用 Golang 开发的 C++ 依赖管理器，支持 YAML 配置文件，可以管理 C++ 项目的依赖下载、构建和安装。

## 功能特性

- 📝 **YAML 配置文件** - 使用简洁的 YAML 语法定义项目依赖
- 🔧 **多构建系统支持** - 支持 CMake、Make、Configure 和自定义构建脚本
- 🚀 **依赖缓存** - 智能缓存机制，避免重复下载和构建
- 🌍 **跨平台支持** - 支持 Linux、macOS 和 Windows
- 📦 **版本管理** - 支持语义化版本控制和依赖锁定
- 🎯 **构建配置文件** - 支持不同环境（开发、测试、生产）的依赖配置

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/all3n/buildfly.git
cd buildfly

# 构建
go build -o buildfly cmd/main.go

# 安装到系统路径
sudo mv buildfly /usr/local/bin/
```

### 初始化项目

```bash
# 创建新项目
buildfly init --name my-project --template cmake

# 这将创建一个 buildfly.yaml 配置文件
```

### 基本用法

```bash
# 安装所有依赖
buildfly install

# 安装特定依赖
buildfly install boost fmt

# 使用构建配置文件安装
buildfly install --profile linux-release

# 构建依赖（不下载）
buildfly build

# 清理缓存
buildfly clean --cache

# 列出依赖
buildfly list --verbose
```

## 配置文件示例

```yaml
project:
  name: "my-cpp-project"
  version: "1.0.0"
  variables:
    install_dir: "${HOME}/.buildfly/install"
    build_type: "Release"
    cxx_compiler: "g++"

dependencies:
  boost:
    version: "1.75.0"
    source: 
      type: "archive"
      url: "https://boostorg.jfrog.io/artifactory/main/release/1.75.0/source/boost_1_75_0.tar.gz"
    build_system: "custom"
    custom_script: |
      #!/bin/bash
      ./bootstrap.sh --prefix=${INSTALL_DIR}
      ./b2 install --with-system --with-filesystem

  fmt:
    version: "8.0.1"
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"
    cmake_options:
      - "FMT_TEST=OFF"
      - "CMAKE_POSITION_INDEPENDENT_CODE=ON"

build_profiles:
  release:
    variables:
      build_type: "Release"
    dependencies:
      - "boost"
      - "fmt"
  
  debug:
    variables:
      build_type: "Debug"
    dependencies:
      - "fmt"
```

## 支持的依赖源

### Git 仓库

```yaml
dependencies:
  mylib:
    version: "1.0.0"
    source:
      type: "git"
      url: "https://github.com/user/mylib.git"
      tag: "v1.0.0"  # 或 branch, commit
```

### 压缩包

```yaml
dependencies:
  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
      hash: "sha256:..."  # 可选的校验和
```

### 直接下载

```yaml
dependencies:
  header_only:
    version: "1.0.0"
    source:
      type: "direct"
      url: "https://example.com/header.hpp"
```

## 构建系统

### CMake

```yaml
dependencies:
  fmt:
    build_system: "cmake"
    cmake_options:
      - "FMT_TEST=OFF"
      - "CMAKE_BUILD_TYPE=${BUILD_TYPE}"
    build_commands:
      configure: "cmake -B ${BUILD_DIR} -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR}"
      build: "cmake --build ${BUILD_DIR} --parallel ${CPU_CORES}"
      install: "cmake --install ${BUILD_DIR}"
```

### Make

```yaml
dependencies:
  zlib:
    build_system: "make"
    make_options:
      - "-j${CPU_CORES}"
    build_commands:
      build: "make ${MAKE_OPTIONS}"
      install: "make install"
```

### Configure

```yaml
dependencies:
  autoconf:
    build_system: "configure"
    configure_options:
      - "--prefix=${INSTALL_DIR}"
      - "--enable-shared"
```

### 自定义脚本

```yaml
dependencies:
  boost:
    build_system: "custom"
    custom_script: |
      #!/bin/bash
      set -e
      ./bootstrap.sh --prefix=${INSTALL_DIR}
      ./b2 install --with-system --with-filesystem
```

## 变量系统

BuildFly 支持强大的变量替换系统：

### 内置变量

- `${INSTALL_DIR}` - 安装目录
- `${BUILD_DIR}` - 构建目录
- `${SOURCE_DIR}` - 源代码目录
- `${BUILD_TYPE}` - 构建类型 (Debug/Release)
- `${CXX_COMPILER}` - C++ 编译器
- `${CPU_CORES}` - CPU 核心数
- `${OS}` - 操作系统
- `${ARCH}` - 系统架构

### 环境变量

- `${HOME}` - 用户主目录
- `${PATH}` - 系统路径
- 其他系统环境变量

### 自定义变量

```yaml
project:
  variables:
    my_version: "1.0.0"
    custom_path: "${HOME}/my-project"

dependencies:
  mylib:
    build_commands:
      configure: "./configure --version=${my_version} --prefix=${custom_path}"
```

## 命令参考

### install

安装依赖：

```bash
buildfly install [options] [dependencies...]

选项：
  -f, --force         强制重新安装
      --no-cache       不使用缓存
  -p, --profile       使用构建配置文件
  -t, --target        目标安装目录
```

### build

构建依赖：

```bash
buildfly build [options] [dependencies...]

选项：
  -f, --force         强制重新构建
  -p, --profile       使用构建配置文件
```

### clean

清理缓存和文件：

```bash
buildfly clean [options]

选项：
      --all            清理所有文件
      --cache          清理缓存
      --deps           清理已安装的依赖
      --dry-run        显示将要删除的文件，但不实际删除
```

### init

初始化项目：

```bash
buildfly init [options]

选项：
  -n, --name          项目名称
  -t, --template      项目模板 (basic, cmake, make)
      --force          覆盖现有配置文件
```

### list

列出依赖和缓存信息：

```bash
buildfly list [options]

选项：
  -v, --verbose       显示详细信息
      --cache          显示缓存信息
```

### config

配置管理：

```bash
buildfly config show    # 显示当前配置
buildfly config set <key> <value>    # 设置配置
buildfly config reset   # 重置配置
```


### 构建

```bash
./manager build
```

### 测试

```bash
go test ./...
```
