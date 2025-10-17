#!/bin/bash

# BuildFly C++ 虚拟环境测试脚本

set -e

echo "=== BuildFly C++ 虚拟环境测试 ==="
echo

# 检查 BuildFly 是否可用
if ! command -v buildfly &> /dev/null; then
    echo "错误: buildfly 命令未找到"
    echo "请先编译 buildfly: go build -o bin/buildfly cmd/main.go"
    exit 1
fi

echo "✓ BuildFly 命令已找到"

# 清理之前的环境
echo
echo "=== 清理之前的环境 ==="
if [ -d ".buildfly" ]; then
    rm -rf .buildfly
    echo "✓ 清理了旧的 .buildfly 目录"
fi

if [ -d "build" ]; then
    rm -rf build
    echo "✓ 清理了旧的 build 目录"
fi

# 初始化虚拟环境
echo
echo "=== 初始化虚拟环境 ==="
buildfly venv init --force
echo "✓ 虚拟环境初始化完成"

# 查看环境状态
echo
echo "=== 查看环境状态 ==="
buildfly venv status --verbose
echo

# 激活虚拟环境
echo "=== 激活虚拟环境 ==="
buildfly venv activate
echo "✓ 虚拟环境已激活"

# 检查环境变量
echo
echo "=== 检查环境变量 ==="
echo "BUILDFLY_ENV_ROOT: $BUILDFLY_ENV_ROOT"
echo "PATH 包含环境目录: $(echo $PATH | grep -o '.buildfly/root[^:]*' || echo '未找到')"

# 检查 UV 是否可用
echo
echo "=== 检查 UV ==="
if command -v uv &> /dev/null; then
    echo "✓ UV 已安装: $(uv --version)"
else
    echo "✗ UV 未找到"
fi

# 安装依赖
echo
echo "=== 安装依赖 ==="
buildfly install
echo "✓ 依赖安装完成"

# 构建项目
echo
echo "=== 构建项目 ==="
mkdir -p build
cd build

# 使用环境中的 CMake
if [ -n "$BUILDFLY_ENV_ROOT" ] && [ -f "$BUILDFLY_ENV_ROOT/tools/cmake/bin/cmake" ]; then
    CMAKE_CMD="$BUILDFLY_ENV_ROOT/tools/cmake/bin/cmake"
    echo "使用虚拟环境中的 CMake: $CMAKE_CMD"
else
    CMAKE_CMD="cmake"
    echo "使用系统 CMake: $CMAKE_CMD"
fi

$CMAKE_CMD ..
echo "✓ CMake 配置完成"

# 构建
if command -v ninja &> /dev/null; then
    echo "使用 Ninja 构建"
    ninja
else
    echo "使用 Make 构建"
    make
fi
echo "✓ 构建完成"

# 运行程序
echo
echo "=== 运行程序 ==="
if [ -f "venv-demo" ]; then
    ./venv-demo
    echo "✓ 程序运行成功"
elif [ -f "venv-demo.exe" ]; then
    ./venv-demo.exe
    echo "✓ 程序运行成功"
else
    echo "✗ 找不到可执行文件"
fi

cd ..

# 列出已安装的工具
echo
echo "=== 列出已安装的工具 ==="
buildfly venv list

# 测试 JSON 输出
echo
echo "=== 测试 JSON 状态输出 ==="
buildfly venv status --json

# 停用环境
echo
echo "=== 停用虚拟环境 ==="
buildfly venv deactivate
echo "✓ 虚拟环境已停用"

# 显示最终状态
echo
echo "=== 最终状态 ==="
buildfly venv status

echo
echo "=== 测试完成 ==="
echo "虚拟环境功能正常工作！"

# 显示目录结构
echo
echo "=== 生成的目录结构 ==="
if command -v tree &> /dev/null; then
    tree .buildfly -L 3
else
    find .buildfly -type d | head -20
fi

echo
echo "=== 激活脚本内容 ==="
echo "Unix/Linux/macOS 激活脚本:"
if [ -f ".buildfly/root/activate.sh" ]; then
    echo "----------------------------------------"
    head -20 .buildfly/root/activate.sh
    echo "----------------------------------------"
fi

echo
echo "使用建议:"
echo "1. 在新项目中运行: buildfly venv init"
echo "2. 激活环境: buildfly venv activate"
echo "3. 或手动激活: source .buildfly/root/activate.sh"
echo "4. 安装依赖: buildfly install"
echo "5. 构建项目: 使用环境中的工具"
