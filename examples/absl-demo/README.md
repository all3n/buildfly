# Abseil 依赖安装优化示例

这个示例展示了如何使用 BuildFly 优化后的依赖安装流程来安装和管理 Google Abseil 库。

## 项目结构

```
absl-demo/
├── buildfly.yaml          # BuildFly 配置文件
├── main.cpp              # 示例 C++ 程序
├── CMakeLists.txt        # CMake 构建配置
└── README.md             # 本文件
```

## 优化特性

### 1. 配置读取和合并
- 支持全局配置 (`~/.config/buildfly/config.yaml`)
- 支持本地项目配置 (`buildfly.yaml`)
- 项目配置优先级高于全局配置

### 2. 构建标签生成
- 自动检测系统架构、平台、编译器
- 支持构建配置文件中的构建标签覆盖
- 确保不同构建环境下的隔离性

### 3. 标准化路径结构
```
# 下载缓存
~/.buildfly/cache/buildfly/{name}/{version}/{filename}

# 构建目录（含构建标签）
~/.buildfly/build/{name}/{version}/{build_tag}

# 安装目录（含构建标签）
~/.buildfly/install/{name}/{version}/{build_tag}

# 项目链接目录
{project_root}/.buildfly/install/{dep_name}
```

### 4. 原子性链接管理
- 使用临时目录确保操作的原子性
- 维护安装清单记录所有链接操作
- 支持链接完整性验证

### 5. 依赖分析
- 自动分析依赖构建要求
- 生成标准化缓存键
- 支持构建工具依赖检测

## 使用方法

### 安装依赖
```bash
# 使用默认配置安装
./buildfly install

# 使用指定构建配置文件
./buildfly install --profile debug

# 指定自定义构建标签
./buildfly install --build-tag "arch=x86_64,platform=linux,runtime=glibc_2.31,compiler=gcc_11,std=cpp17"
```

### 构建项目
```bash
# 创建构建目录
mkdir build && cd build

# 配置 CMake
cmake -DCMAKE_BUILD_TYPE=Release ..

# 编译
make -j$(nproc)

# 运行示例
./absl-demo
```

## 构建配置文件

### debug 配置
- 构建类型：Debug
- C++ 标准：17
- 包含完整的调试信息

### release 配置
- 构建类型：Release
- C++ 标准：17
- 优化级别：O2

### header_only 配置
- 仅使用头文件（如果支持）
- 最小化编译时间

## 输出示例

```
=== Abseil Demo Program ===
StrCat result: Hello Abseil!
StrJoin result: Optimized Dependency Management
String view: Hello
Optional value: 42
Span elements: 1 2 3 4 5
=== Demo completed successfully! ===
```

## 高级特性

### 缓存管理
```bash
# 查看缓存状态
./buildfly list --cache

# 清理缓存
./buildfly clean --cache

# 强制重新安装
./buildfly install --force
```

### 链接验证
```bash
# 验证链接完整性
./buildfly verify

# 修复损坏的链接
./buildfly repair
```

## 技术细节

### 构建标签格式
```
arch=x86_64,platform=linux,runtime=glibc_2.31,compiler=gcc_11,std=cpp17,abi=sysv
```

### 缓存键格式
```
{name}@{version}#{build_tag}:{checksum_prefix}
```

### 原子操作流程
1. 在临时目录中准备所有操作
2. 备份现有文件/链接
3. 执行删除和创建操作
4. 记录到安装清单
5. 清理临时文件

这个优化后的流程确保了依赖安装的可靠性、可重复性和可追溯性。