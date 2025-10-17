# BuildFly 使用指南

## 概述

BuildFly 是一个用 Golang 开发的 C++ 依赖管理器，支持 YAML 配置文件，可以管理 C++ 项目的依赖下载、构建和安装。

## 快速开始

### 1. 安装 BuildFly

```bash
# 克隆仓库
git clone https://github.com/your-username/buildfly.git
cd buildfly

# 构建
go build -o buildfly cmd/main.go

# 安装到系统路径
sudo mv buildfly /usr/local/bin/
```

### 2. 初始化项目

```bash
# 创建新项目
buildfly init --name my-project

# 这将创建一个 buildfly.yaml 配置文件
```

### 3. 配置依赖

编辑 `buildfly.yaml` 文件来定义你的项目依赖：

```yaml
project:
  name: "my-cpp-project"
  version: "1.0.0"

dependencies:
  fmt:
    version: "8.0.1"
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"
```

### 4. 安装依赖

```bash
# 安装所有依赖
buildfly install

# 安装特定依赖
buildfly install fmt

# 使用构建配置文件安装
buildfly install --profile release
```

## 配置文件详解

### 项目配置

```yaml
project:
  name: "my-project"           # 项目名称
  version: "1.0.0"             # 项目版本
  variables:                   # 自定义变量
    install_dir: "${HOME}/.buildfly/install"
    build_type: "Release"
    cxx_compiler: "g++"
```

### 依赖配置

#### Git 仓库依赖

```yaml
dependencies:
  mylib:
    version: "1.0.0"
    source:
      type: "git"
      url: "https://github.com/user/mylib.git"
      tag: "v1.0.0"           # 可以是 tag、branch 或 commit
    build_system: "cmake"
```

#### 压缩包依赖

```yaml
dependencies:
  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
      hash: "sha256:..."      # 可选的校验和
    build_system: "configure"
```

#### 直接下载依赖

```yaml
dependencies:
  header_only:
    version: "1.0.0"
    source:
      type: "direct"
      url: "https://example.com/header.hpp"
```

### 构建系统配置

#### CMake

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

#### Make

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

#### Configure

```yaml
dependencies:
  autoconf:
    build_system: "configure"
    configure_options:
      - "--prefix=${INSTALL_DIR}"
      - "--enable-shared"
```

#### 自定义脚本

```yaml
dependencies:
  boost:
    build_system: "custom"
    custom_script: |
      #!/bin/bash
      set -e
      ./bootstrap.sh --prefix=${INSTALL_DIR}
      ./b2 install --with-system --with-filesystem
    env_variables:
      BOOST_VERSION: "1.75.0"
```

### 构建配置文件

```yaml
build_profiles:
  debug:
    variables:
      build_type: "Debug"
      cxx_flags: "-g -O0"
    dependencies:
      - "fmt"
      - "spdlog"

  release:
    variables:
      build_type: "Release"
      cxx_flags: "-O3 -DNDEBUG"
    dependencies:
      - "fmt"
```

## 变量系统

### 内置变量

- `${INSTALL_DIR}` - 安装目录
- `${BUILD_DIR}` - 构建目录
- `${SOURCE_DIR}` - 源代码目录
- `${BUILD_TYPE}` - 构建类型 (Debug/Release)
- `${CXX_COMPILER}` - C++ 编译器
- `${CXX_FLAGS}` - 编译标志
- `${CPU_CORES}` - CPU 核心数
- `${OS}` - 操作系统
- `${ARCH}` - 系统架构

### 环境变量

- `${HOME}` - 用户主目录
- `${PATH}` - 系统路径
- 其他系统环境变量

### 自定义变量

在 `project.variables` 中定义的自定义变量可以在配置文件的其他部分引用：

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

## 使用示例

### 示例 1: 基本的 CMake 项目

```bash
# 创建项目目录
mkdir my-cpp-project && cd my-cpp-project

# 初始化项目
buildfly init --name my-cpp-project --template cmake

# 编辑 buildfly.yaml 添加依赖
cat > buildfly.yaml << EOF
project:
  name: "my-cpp-project"
  version: "1.0.0"

dependencies:
  fmt:
    version: "8.0.1"
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"
    cmake_options:
      - "FMT_TEST=OFF"
EOF

# 安装依赖
buildfly install

# 编写你的 C++ 代码...
```

### 示例 2: 使用 Boost

```bash
# 使用 Boost 示例
cp -r examples/boost-custom my-boost-project
cd my-boost-project

# 安装最小 Boost 配置
buildfly install --profile minimal

# 或者安装完整 Boost
buildfly install --profile full
```

### 示例 3: 自定义构建脚本

```yaml
dependencies:
  special_lib:
    version: "1.0.0"
    source:
      type: "git"
      url: "https://github.com/user/special-lib.git"
      tag: "v1.0.0"
    build_system: "custom"
    custom_script: |
      #!/bin/bash
      set -e
      
      # 自定义构建逻辑
      cd ${SOURCE_DIR}
      
      # 运行自定义配置脚本
      ./configure.sh --prefix=${INSTALL_DIR}
      
      # 编译
      make -j${CPU_CORES}
      
      # 安装
      make install
      
      # 运行测试
      make test || echo "Tests failed but continuing..."
```

## 最佳实践

### 1. 使用构建配置文件

为不同的开发环境创建不同的构建配置文件：

```yaml
build_profiles:
  development:
    variables:
      build_type: "Debug"
    dependencies:
      - "fmt"
      - "spdlog"

  production:
    variables:
      build_type: "Release"
    dependencies:
      - "fmt"

  ci:
    variables:
      build_type: "Release"
    dependencies:
      - "fmt"
```

### 2. 缓存管理

合理使用缓存来提高构建速度：

```bash
# 清理缓存（当依赖更新时）
buildfly clean --cache

# 强制重新安装
buildfly install --force

# 不使用缓存（用于调试）
buildfly install --no-cache
```

### 3. 版本锁定

在 CI/CD 环境中，建议锁定依赖版本：

```yaml
dependencies:
  fmt:
    version: "8.0.1"          # 固定版本
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"            # 固定 tag
```

### 4. 安全性

使用校验和验证下载的文件：

```yaml
dependencies:
  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
      hash: "sha256:1c9f418ee0e4be921b5df78d058b6fcc3acc0a9bf22b692f66a9a8f6b8fa3e0f"
```

## 故障排除

### 常见问题

1. **构建失败**
   - 检查构建工具是否安装（cmake, make, gcc 等）
   - 查看详细错误信息：`buildfly install --verbose`

2. **下载失败**
   - 检查网络连接
   - 验证 URL 是否正确
   - 尝试使用代理：`export https_proxy=your-proxy`

3. **缓存问题**
   - 清理缓存：`buildfly clean --cache`
   - 强制重新安装：`buildfly install --force`

### 调试技巧

```bash
# 查看详细输出
buildfly install --verbose

# 检查配置文件
buildfly list --verbose

# 测试单个依赖
buildfly install fmt

# 使用临时目录
buildfly install --target /tmp/test-build
```

## 贡献

欢迎贡献代码和文档！请参考 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。
