基于您的需求，我为您创建增强版的 YAML 配置规则：
## 📁 `.cline/rules/yaml-build-config.md`

```markdown
# YAML 构建配置系统规则

## 配置系统概述
支持多构建系统（Makefile、Configure、CMake）和自定义脚本构建，特别处理像 Boost 这样的特殊项目。

## YAML 配置结构

### 完整配置示例
```yaml
# cppdep.yaml
project:
  name: "my-cpp-project"
  version: "1.0.0"
  variables:
    install_dir: "${HOME}/.cppdeps"
    build_type: "Release"
    cxx_compiler: "g++"
    cxx_flags: "-O2 -std=c++17"

dependencies:
  boost:
    version: "1.89.0"
    source: 
      type: "archive"
      urls:
        -  "https://archives.boost.io/release/1.89.0/source/boost_1_89_0.tar.gz"
    build_system: "custom"
    custom_script: |
      #!/bin/bash
      set -e
      echo "Building Boost with custom configuration..."
      
      # 解压并进入目录
      tar xzf ${DOWNLOAD_PATH} -C ${EXTRACT_DIR}
      cd ${EXTRACT_DIR}/boost_1_89_0
      
      # Bootstrap
      ./bootstrap.sh \
        --prefix=${INSTALL_DIR} \
        --with-toolset=${CXX_COMPILER}
      
      # 构建特定库
      ./b2 install \
        --prefix=${INSTALL_DIR} \
        --build-dir=${BUILD_DIR} \
        toolset=${CXX_COMPILER} \
        variant=${BUILD_TYPE} \
        link=static,shared \
        --with-system \
        --with-filesystem \
        --with-program_options \
        -j${CPU_CORES}
      
      echo "Boost installation completed"

  fmt:
    version: "8.0.1"
    source:
      type: "git"
      urls: 
        - "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"
    cmake_options:
      - "FMT_TEST=OFF"
      - "FMT_DOC=OFF"
      - "CMAKE_POSITION_INDEPENDENT_CODE=ON"
    build_commands:
      configure: |
        cmake -B ${BUILD_DIR} \
          -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR} \
          -DCMAKE_BUILD_TYPE=${BUILD_TYPE} \
          ${CMAKE_OPTIONS}
      build: "cmake --build ${BUILD_DIR} --config ${BUILD_TYPE} -j${CPU_CORES}"
      install: "cmake --install ${BUILD_DIR} --config ${BUILD_TYPE}"

  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
    build_system: "configure"
    configure_options:
      - "--prefix=${INSTALL_DIR}"
      - "--shared"
    build_commands:
      configure: "./configure ${CONFIGURE_OPTIONS}"
      build: "make -j${CPU_CORES}"
      install: "make install"

build_profiles:
  linux-release:
    variables:
      build_type: "Release"
      cxx_flags: "-O3 -DNDEBUG -std=c++17"
    dependencies:
      - "boost"
      - "fmt"
      - "zlib"

  windows-debug:
    variables:
      build_type: "Debug"
      cxx_flags: "-g -O0 -std=c++17"
    dependencies:
      - "fmt"
```

### 内置变量
- `${INSTALL_DIR}` - 安装目录
- `${BUILD_DIR}` - 构建目录  
- `${DOWNLOAD_PATH}` - 下载文件路径
- `${EXTRACT_DIR}` - 解压目录
- `${SOURCE_DIR}` - 源代码目录
- `${BUILD_TYPE}` - 构建类型 (Debug/Release)
- `${CXX_COMPILER}` - C++ 编译器
- `${CXX_FLAGS}` - 编译标志
- `${CPU_CORES}` - CPU 核心数
- `${OS}` - 操作系统
- `${ARCH}` - 架构

### 环境变量
- 自动从系统环境变量继承
- 支持 `${HOME}`, `${PATH}` 等标准变量

### 自定义变量
- 在 YAML 的 `project.variables` 中定义
- 支持嵌套引用其他变量

这套配置系统提供了强大的灵活性和可扩展性，能够处理从简单的 Makefile 项目到复杂的自定义构建脚本如 Boost。
```

这个增强版的配置系统支持：

1. **多构建系统**: Makefile、CMake、Configure、自定义脚本
2. **Boost 特殊处理**: 完整的多行自定义脚本支持
3. **变量替换**: 强大的变量系统，支持内置变量、环境变量和自定义变量
4. **灵活的构建命令**: 支持多行命令和复杂脚本
5. **配置验证**: 自动验证配置文件的正确性
6. **跨平台支持**: 自动处理不同操作系统的差异
