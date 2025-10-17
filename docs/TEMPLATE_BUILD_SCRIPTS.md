# Go Template 构建脚本

## 概述

Buildfly 现在支持使用 Go template 语法来生成动态构建脚本。这使得构建脚本可以根据项目配置、系统环境和依赖信息动态生成，提供了更大的灵活性和可重用性。

## Template 数据结构

构建脚本模板可以访问以下数据：

### 依赖信息
- `{{.Dependency.Name}}` - 依赖名称
- `{{.Dependency.Version}}` - 依赖版本
- `{{.Dependency.BuildSystem}}` - 构建系统类型
- `{{.Dependency.Source.URL}}` - 源码 URL

### 项目信息
- `{{.ProjectName}}` - 项目名称
- `{{.ProjectVersion}}` - 项目版本

### 路径变量
- `{{.SourceDir}}` - 源码目录
- `{{.BuildDir}}` - 构建目录
- `{{.InstallDir}}` - 安装目录

### 构建变量
- `{{.BuildType}}` - 构建类型（Debug/Release）
- `{{.CPUCount}}` - CPU 核心数

### 系统信息
- `{{.OS}}` - 操作系统（linux/darwin/windows）
- `{{.Arch}}` - 系统架构（amd64/arm64等）

### 自定义变量
- `{{.Variables.variable_name}}` - 项目自定义变量

## Template 语法示例

### 基本变量替换
```yaml
custom_script: |
  #!/bin/bash
  echo "Building {{.Dependency.Name}} version {{.Dependency.Version}}"
  echo "Installing to {{.InstallDir}}"
  
  cd {{.SourceDir}}
  make -j{{.CPUCount}}
  make install PREFIX={{.InstallDir}}
```

### 条件判断
```yaml
custom_script: |
  #!/bin/bash
  
  {{if eq .OS "linux"}}
  echo "Linux detected, using gcc"
  export CC=gcc
  export CXX=g++
  {{else if eq .OS "darwin"}}
  echo "macOS detected, using clang"
  export CC=clang
  export CXX=clang++
  {{else}}
  echo "Windows detected, using MSVC"
  {{end}}
  
  cmake {{.SourceDir}} -DCMAKE_INSTALL_PREFIX={{.InstallDir}}
```

### 循环（如果需要）
```yaml
custom_script: |
  #!/bin/bash
  
  {{range $key, $value := .Variables}}
  echo "Variable {{$key}} = {{$value}}"
  {{end}}
  
  # 构建命令
  make -j{{.CPUCount}}
```

### 复杂示例
```yaml
custom_script: |
  #!/bin/bash
  set -e
  
  echo "=== Building {{.Dependency.Name}} ==="
  echo "Project: {{.ProjectName}} v{{.ProjectVersion}}"
  echo "System: {{.OS}} ({{.Arch}})"
  echo "Build Type: {{.BuildType}}"
  
  # 根据构建类型设置不同的编译选项
  {{if eq .BuildType "Debug"}}
  CMAKE_BUILD_TYPE="Debug"
  CFLAGS="-g -O0"
  {{else}}
  CMAKE_BUILD_TYPE="Release"
  CFLAGS="-O3 -DNDEBUG"
  {{end}}
  
  # 创建构建目录
  mkdir -p {{.BuildDir}}
  cd {{.BuildDir}}
  
  # 配置
  cmake {{.SourceDir}} \
    -DCMAKE_INSTALL_PREFIX={{.InstallDir}} \
    -DCMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE} \
    -DCMAKE_C_FLAGS="${CFLAGS}" \
    -DCMAKE_CXX_FLAGS="${CFLAGS}"
  
  # 构建
  cmake --build . --config {{.BuildType}} --parallel {{.CPUCount}}
  
  # 安装
  cmake --install . --config {{.BuildType}}
  
  echo "✓ {{.Dependency.Name}} installed successfully"
```

## 使用方法

1. 在配置文件中设置 `build_system: "custom"`
2. 在 `custom_script` 字段中使用 Go template 语法
3. 运行 `buildfly install`，系统会自动渲染模板并执行生成的脚本

## 生成的脚本文件

模板渲染后的脚本会保存为 `.buildfly_build_script.sh`，位于构建工作目录（`.buildfly/build/{dependency-name}/`）中。这便于调试和检查生成的脚本内容。

## 自定义构建脚本执行位置

- **自定义构建脚本**：在构建目录（`.buildfly/build/{dependency-name}/`）中执行
- **其他构建系统**：
  - CMake：在构建目录中执行
  - Make：在源码目录中执行
  - Configure：在源码目录中执行

这种设计确保了自定义构建脚本与项目的构建目录结构保持一致，便于管理和调试。

## 错误处理

如果模板语法错误或变量不存在，构建会失败并显示详细的错误信息。可以检查生成的 `.buildfly_build_script.sh` 文件来调试问题。

## 最佳实践

1. **使用条件判断**来处理不同操作系统的差异
2. **利用变量**来避免硬编码路径和参数
3. **添加错误检查**确保构建过程的可靠性
4. **使用 echo 命令**输出构建过程信息，便于调试
5. **保持脚本简洁**，复杂的逻辑可以考虑分解为多个步骤

## 示例配置

完整示例请参考 `examples/template-demo/buildfly.yaml` 文件。

## 注意事项

1. Template 语法使用 `{{` 和 `}}` 作为分隔符
2. 在 YAML 中使用多行字符串时，注意缩进
3. 生成的脚本会在 Bash 环境中执行
4. 确保模板中的所有变量都已定义，避免运行时错误
