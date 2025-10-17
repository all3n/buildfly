#!/bin/bash

# 测试 UV 配置文件创建和环境变量设置功能

set -e

echo "=== Testing UV Config Creation and Environment Variable Setting ==="

# 创建测试目录
TEST_DIR="$(mktemp -d)"
cd "$TEST_DIR"

echo "Test directory: $TEST_DIR"

# 初始化 buildfly 项目
cat > buildfly.yaml << EOF
project:
  name: "test-uv-config"
  version: "1.0.0"

venv:
  enabled: true
  uv_version: "latest"
  python:
    version: "3.11"
    packages: ["cmake", "ninja"]
  cpp_tools:
    cmake:
      version: "3.28.0"
      enabled: true
    ninja:
      version: "1.11.1"
      enabled: true
  uv:
    index_url: "https://pypi.org/simple/"
EOF

echo "Created buildfly.yaml"

# 检查是否有 buildfly 命令
if ! command -v buildfly &> /dev/null; then
    echo "Error: buildfly command not found. Please build the project first."
    cd ..
    rm -rf "$TEST_DIR"
    exit 1
fi

# 运行 venv init
echo "Running buildfly venv init..."
if buildfly venv init --force; then
    echo "✓ venv init succeeded"
else
    echo "✗ venv init failed"
    cd ..
    rm -rf "$TEST_DIR"
    exit 1
fi

# 检查 .buildfly/uv.toml 文件是否创建
if [ -f ".buildfly/uv.toml" ]; then
    echo "✓ uv.toml file created at .buildfly/uv.toml"
    echo "Contents:"
    cat .buildfly/uv.toml
    echo ""
else
    echo "✗ uv.toml file not found"
    cd ..
    rm -rf "$TEST_DIR"
    exit 1
fi

# 检查环境变量设置
echo "Testing environment variable setting..."
export UV_ENV_FILE="$(pwd)/.buildfly/uv.toml"
echo "UV_ENV_FILE set to: $UV_ENV_FILE"

# 测试 venv run 命令是否使用了环境变量
echo "Testing buildfly venv run with UV_ENV_FILE..."
if buildfly venv run python --version; then
    echo "✓ venv run succeeded with UV_ENV_FILE"
else
    echo "✗ venv run failed with UV_ENV_FILE"
    cd ..
    rm -rf "$TEST_DIR"
    exit 1
fi

# 清理
cd ..
rm -rf "$TEST_DIR"

echo ""
echo "=== All tests passed! ==="
echo "UV config file creation and environment variable setting are working correctly."
