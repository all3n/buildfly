#!/bin/bash

# Build Tag 功能演示脚本

set -e

echo "=== Build Tag 功能演示 ==="
echo

# 检查 buildfly 是否存在
if ! command -v buildfly &> /dev/null; then
    echo "错误: buildfly 命令未找到，请先编译项目"
    echo "运行: go build -o buildfly ./cmd/main.go"
    exit 1
fi

# 创建演示目录
DEMO_DIR="build-tag-demo"
if [ -d "$DEMO_DIR" ]; then
    rm -rf "$DEMO_DIR"
fi
mkdir -p "$DEMO_DIR"
cd "$DEMO_DIR"

echo "1. 复制配置文件..."
cp ../examples/build-tag-demo/buildfly.yaml .

echo "2. 显示配置文件内容:"
echo "----------------------------------------"
cat buildfly.yaml
echo "----------------------------------------"
echo

echo "3. 测试不同的构建标签..."

# 测试基本 Linux 构建
echo
echo "3.1 测试基本 Linux 构建 (x86_64):"
echo "buildfly install --build-tag \"arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17\""
echo

# 测试 macOS 构建
echo "3.2 测试 macOS 构建 (arm64):"
echo "buildfly install --build-tag \"arch=arm64,platform=darwin,runtime=libcxx_15,compiler=apple-clang_15.0,std=cpp20,abi=macho,target=macos14.0\""
echo

# 测试 CUDA 构建
echo "3.3 测试 CUDA 构建 (x86_64 + GPU):"
echo "buildfly install --build-tag \"arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=nvcc_12.0+gcc_11.4,std=cpp17,abi=sysv,cuda=12.0,cuda_arch=compute_80|compute_90,gpu_enabled=true,gpu_backend=cuda\""
echo

# 测试使用配置文件中的 build tag
echo "3.4 使用配置文件中的 build tag:"
echo "buildfly install"
echo

echo "4. 预期的目录结构:"
echo "----------------------------------------"
echo "build/"
echo "├── fmt/"
echo "│   └── 8.0.1/"
echo "│       └── arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv/"
echo "└── zlib/"
echo "    └── 1.2.11/"
echo "        └── arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv/"
echo
echo "install/"
echo "└── arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv/"
echo "    ├── fmt/"
echo "    └── zlib/"
echo "----------------------------------------"
echo

echo "5. Build Tag 格式说明:"
echo "----------------------------------------"
echo "基本格式: key=value,key=value,..."
echo
echo "支持的键:"
echo "  arch      - 架构 (x86_64, arm64, x64, i386, aarch64)"
echo "  platform  - 平台 (linux, darwin, windows)"
echo "  runtime   - 运行时 (glibc, libcxx, msvcrt, musl)"
echo "  compiler  - 编译器 (gcc, clang, msvc, nvcc, apple-clang)"
echo "  std       - C++ 标准 (cpp11, cpp14, cpp17, cpp20, cpp23)"
echo "  abi       - ABI (sysv, macho, msabi)"
echo "  target    - 平台特定目标 (如 macos14.0)"
echo
echo "GPU 相关键:"
echo "  cuda      - CUDA 版本 (如 12.0)"
echo "  cuda_arch - CUDA 架构 (如 compute_80|compute_90)"
echo "  rocm      - ROCm 版本 (如 5.0)"
echo "  rocm_arch - ROCm 架构 (如 gfx900|gfx1030)"
echo "  opencl    - OpenCL 版本 (如 2.0)"
echo "  gpu_backend - GPU 后端 (cuda, rocm, opencl, none)"
echo "  gpu_enabled - 是否启用 GPU (true, false)"
echo
echo "版本范围支持:"
echo "  +         - 版本或更高 (如 gcc_11+)"
echo "  |         - 多个选择 (如 compute_80|compute_90)"
echo "  ,         - 分隔多个键值对"
echo "----------------------------------------"
echo

echo "6. 实际使用示例:"
echo "----------------------------------------"
echo "# 安装 Linux x86_64 版本"
echo "buildfly install --build-tag \"arch=x86_64,platform=linux\""
echo
echo "# 安装 macOS ARM64 版本"
echo "buildfly install --build-tag \"arch=arm64,platform=darwin\""
echo
echo "# 安装 CUDA 支持版本"
echo "buildfly install --build-tag \"arch=x86_64,platform=linux,cuda=12.0,cuda_arch=compute_80\""
echo
echo "# 安装多架构支持版本"
echo "buildfly install --build-tag \"arch=x86_64,platform=linux,cuda_arch=compute_80|compute_90\""
echo
echo "# 安装使用配置文件中的 build tag"
echo "buildfly install"
echo "----------------------------------------"
echo

echo "演示完成！"
echo
echo "注意: 这是一个演示脚本，实际运行需要网络连接和足够的磁盘空间"
echo "要真正执行安装，请取消注释下面的命令或手动运行"
echo

# 取消注释以下命令来实际运行安装
# buildfly install --build-tag "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17"

cd ..
