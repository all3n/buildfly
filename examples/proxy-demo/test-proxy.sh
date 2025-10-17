#!/bin/bash

echo "=== 测试代理配置功能 ==="
echo

# 进入代理演示目录
cd examples/proxy-demo

echo "当前目录: $(pwd)"
echo

echo "=== 配置文件内容 ==="
cat buildfly.yaml
echo

echo "=== 测试配置加载 ==="
# 使用 verbose 模式查看配置加载过程
../../bin/buildfly --verbose list deps 2>&1 | head -20
echo

echo "=== 代理配置测试完成 ==="
