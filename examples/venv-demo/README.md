# BuildFly C++ 虚拟环境演示

这个示例演示了如何使用 BuildFly 的 UV 虚拟环境功能来管理隔离的 C++ 构建环境。

## 功能特性

- **环境隔离**: 使用 UV 创建隔离的 C++ 构建环境
- **工具管理**: 自动安装和管理 CMake、Ninja 等构建工具
- **跨平台支持**: 支持 Linux、macOS 和 Windows
- **依赖管理**: 集成 BuildFly 的依赖管理系统

## 快速开始

### 1. 初始化虚拟环境

```bash
cd examples/venv-demo
buildfly venv init
```

这将：
- 检测并安装 UV（如果未安装）
- 创建 `.buildfly/root` 目录
- 设置环境结构
- 生成激活脚本

### 2. 激活虚拟环境

```bash
# 方式 1: 使用 BuildFly 命令
buildfly venv activate

# 方式 2: 手动激活脚本
source .buildfly/root/activate.sh
```

### 3. 查看环境状态

```bash
buildfly venv status
```

输出示例：
```
Virtual Environment Status:
  Root Directory: /path/to/project/.buildfly/root
  Activated: true
  UV Installed: true
  UV Version: 0.1.0
  Python Version: 3.11
  Created: 2024-01-01 12:00:00
  Last Activated: 2024-01-01 12:30:00
```

### 4. 安装依赖

```bash
buildfly install
```

这将下载并构建 fmt 库，使用虚拟环境中的工具。

### 5. 构建项目

```bash
# 在虚拟环境中构建
mkdir build
cd build
cmake ..
make

# 运行程序
./venv-demo
```

## 配置说明

### buildfly.yaml 配置

```yaml
venv:
  enabled: true                    # 启用虚拟环境
  uv_version: "latest"            # UV 版本
  root_dir: ".buildfly/root"      # 环境根目录
  auto_activate: true             # 自动激活
  
  # C++ 构建工具配置
  cpp_tools:
    cmake:
      version: "3.28.0"
      enabled: true
    ninja:
      version: "1.11.1"
      enabled: true
    gcc:
      version: "11.4.0"
      enabled: false              # 在 macOS 上禁用
    clang:
      version: "16.0.0"
      enabled: true               # 在 macOS 上启用
  
  # Python 工具配置
  python:
    version: "3.11"
    packages:
      - "cmake"
      - "ninja"
      - "conan"
      - "meson"
```

## 命令参考

### buildfly venv init

初始化虚拟环境

```bash
buildfly venv init [选项]

选项:
  --force              强制重新初始化
  --uv-version string  指定 UV 版本 (默认: "latest")
  --root-dir string    指定环境根目录
```

### buildfly venv activate

激活虚拟环境

```bash
buildfly venv activate [选项]

选项:
  --shell              输出激活脚本供 shell 执行
```

### buildfly venv deactivate

停用虚拟环境

```bash
buildfly venv deactivate
```

### buildfly venv status

查看环境状态

```bash
buildfly venv status [选项]

选项:
  -v, --verbose        显示详细信息
  --json               以 JSON 格式输出
```

### buildfly venv reset

重置虚拟环境

```bash
buildfly venv reset [选项]

选项:
  --force              强制重置不询问确认
```

### buildfly venv list

列出已安装的工具

```bash
buildfly venv list
```

## 环境变量

激活虚拟环境后，会设置以下环境变量：

- `BUILDFLY_ENV_ROOT`: 环境根目录
- `PATH`: 添加环境 bin 目录
- `CC`/`CXX`: 设置编译器（如果可用）
- `CMAKE_PREFIX_PATH`: CMake 搜索路径
- `PYTHONPATH`: Python 包路径

## 目录结构

```
.buildfly/root/
├── activate.sh          # Unix/Linux/macOS 激活脚本
├── activate.bat         # Windows 激活脚本
├── environment.json     # 环境信息文件
├── bin/                 # 可执行文件目录
├── lib/                 # 库文件目录
├── include/             # 头文件目录
├── share/               # 共享文件目录
├── tools/               # 构建工具目录
│   ├── cmake/           # CMake 安装
│   └── ninja/           # Ninja 安装
└── env/                 # Python 环境
```

## 故障排除

### UV 安装失败

如果 UV 安装失败，可以手动安装：

**Linux/macOS:**
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows:**
```bash
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
```

### 环境激活失败

确保激活脚本有执行权限：

```bash
chmod +x .buildfly/root/activate.sh
```

### 工具找不到

检查环境变量是否正确设置：

```bash
echo $BUILDFLY_ENV_ROOT
echo $PATH
which cmake
```

## 平台特定说明

### Linux
- 使用系统包管理器安装编译器
- 支持 GCC 和 Clang
- 兼容主流发行版

### macOS
- 使用 Xcode Clang 或安装新版 Clang
- 支持 Homebrew 集成
- ARM64 和 Intel 兼容

### Windows
- 集成 Visual Studio Build Tools
- 支持 MSVC 编译器
- PowerShell 和 CMD 兼容

## 高级用法

### 自定义工具版本

在 `buildfly.yaml` 中指定特定版本：

```yaml
cpp_tools:
  cmake:
    version: "3.25.0"
    enabled: true
```

### 环境导入/导出

```bash
# 导出环境配置
buildfly venv status --json > environment-config.json

# 在其他项目中使用相同配置
# 复制 environment-config.json 并调整路径
```

### 集成 CI/CD

在 CI/CD 管道中使用：

```yaml
# .github/workflows/build.yml
- name: Setup BuildFly Environment
  run: |
    buildfly venv init
    buildfly venv activate
    buildfly install

- name: Build Project
  run: |
    mkdir build && cd build
    cmake ..
    make
```

## 相关文档

- [BuildFly 主文档](../../README.md)
- [配置参考](../../docs/CONFIGURATION.md)
- [UV 官方文档](https://docs.astral.sh/uv/)
