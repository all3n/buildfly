# BuildFly Manager 脚本使用指南

## 概述

`manager` 脚本是 BuildFly 项目的根管理脚本，提供了常用的开发命令封装，包括编译、运行、测试、格式化、lint 等功能。

## 基本用法

```bash
./manager <命令> [选项]
```

## 可用命令

### 🔨 构建相关

#### `build`
编译项目，生成二进制文件到 `bin/` 目录。

```bash
# 标准编译
./manager build

# 发布模式编译（优化，去除调试信息）
./manager build --release

# 调试模式编译（包含调试符号）
./manager build --debug

# 详细输出
./manager build --verbose
```

#### `run`
运行编译后的项目。

```bash
# 运行项目
./manager run

# 传递参数给程序
./manager run --help
./manager run install some-package
```

### 🧪 测试相关

#### `test`
运行项目测试。

```bash
# 运行所有测试
./manager test

# 生成覆盖率报告
./manager test --cover

# 竞态检测
./manager test --race

# 详细测试输出
./manager test --verbose
```

### 📝 代码质量

#### `fmt`
格式化代码。

```bash
./manager fmt
```

#### `lint`
代码检查。

```bash
./manager lint
```

### 🧹 清理相关

#### `clean`
清理构建文件和缓存。

```bash
./manager clean
```

#### `install`
安装依赖和开发工具。

```bash
./manager install
```

### 🐳 Docker 相关

#### `docker`
Docker 相关操作。

```bash
# 构建 Docker 镜像
./manager docker build

# 运行 Docker 容器
./manager docker run

# 推送镜像到仓库
./manager docker push v1.0.0
```

### 🚀 开发相关

#### `dev`
开发模式，监听文件变化并自动重新编译。

```bash
./manager dev
```

> 注意：需要安装 `fswatch` 工具

#### `release`
发布版本。

```bash
./manager release v1.0.0
```

### ❓ 帮助

#### `help`
显示帮助信息。

```bash
./manager help
```

## 环境要求

### 必需
- Go 1.19+
- Bash 4.0+

### 可选（用于增强功能）
- `goimports` - Go 代码导入格式化
- `golangci-lint` - Go 代码检查
- `shfmt` - Shell 脚本格式化
- `shellcheck` - Shell 脚本检查
- `yamllint` - YAML 文件检查
- `fswatch` - 文件变化监听（开发模式）
- Docker - Docker 相关功能

## 项目结构

脚本会自动处理以下目录结构：

```
buildfly/
├── bin/                    # 编译输出目录
│   ├── buildfly           # 主程序
├── coverage.out           # 测试覆盖率数据
├── coverage.html          # 测试覆盖率报告
├── releases/              # 发布版本目录
│   └── v1.0.0/
│       ├── buildfly
└── manager                # 管理脚本
```

## 配置

脚本会自动检测项目信息：
- 项目名称从 `go.mod` 文件中读取
- 版本信息从 Git 标签中获取

## 示例工作流

### 日常开发
```bash
# 安装依赖
./manager install

# 编译项目
./manager build

# 运行测试
./manager test --cover

# 格式化代码
./manager fmt

# 代码检查
./manager lint
```

### 发布流程
```bash
# 确保代码质量
./manager test
./manager lint
./manager fmt

# 提交代码
git add .
git commit -m "Prepare for release v1.0.0"
git push

# 发布版本
./manager release v1.0.0
```

### 开发模式
```bash
# 启动开发模式，自动监听文件变化
./manager dev
```

## 故障排除

### 权限问题
如果脚本无法执行，请检查权限：
```bash
chmod +x manager
```

### 工具缺失
如果某些功能不可用，请安装相应的工具：
```bash
# 安装 Go 工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装 fswatch（macOS）
brew install fswatch

# 安装 fswatch（Linux）
sudo apt-get install fswatch
```

### 编译失败
确保在项目根目录运行脚本，并且 `go.mod` 文件存在。

## 贡献

如果需要为脚本添加新功能，请：

1. 在 `main()` 函数中添加新的 case
2. 实现对应的功能函数
3. 更新 `show_help()` 函数
4. 测试新功能
5. 更新此文档

## 许可证

本脚本遵循项目的开源许可证。
