# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building
```bash
# Build the main binary
go build -o buildfly cmd/main.go

# Alternative: build to bin directory
mkdir -p bin && go build -o bin/buildfly cmd/main.go
```
### 目录结构 示例
```
cmd/                 # 命令行入口
  main.go            # 主程序入口
pkg/
  dependency/        # 依赖管理核心逻辑
  resolver/          # 依赖解析器
  downloader/        # 下载管理器
  cache/             # 缓存管理
  config/            # 配置管理
  utils/             # 工具函数
internal/            # 内部包（不对外暴露）
  types/             # 数据类型定义
  constants/         # 常量定义
configs/             # 配置文件模板
  global.yaml
  templates/
    cmake.yaml
    configure.yaml
    makefile.yaml
examples/            # 使用示例
  example-1/
  example-2/
data/                 # 数据文件
tests/                # 测试脚本 或者 shell 测试文件
manager               # 项目根管理脚本 包含编译运行测试 fmt lint 等命令封装
bin/                  # 脚本shell目录
```
不要在根目录创建测试shell文件

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/config
go test ./pkg/downloader
go test ./pkg/cache

# Run a single test file
go test -run TestSpecificFunction ./pkg/config/loader_test.go
```

### Development Workflow
```bash
# Initialize a new project
./buildfly init --name my-project --template cmake

# Install dependencies
./buildfly install

# List dependencies with cache info
./buildfly list --verbose --cache

# Clean cache and dependencies
./buildfly clean --cache --deps

# Build specific dependencies
./buildfly build boost fmt
```

## Project Architecture

### High-Level Structure
BuildFly is a C++ dependency manager written in Go that uses YAML configuration files. The architecture follows a clean separation of concerns:

- **CLI Layer** (`cmd/cli/`): Command-line interface using Cobra
- **Core Packages** (`pkg/`): Main business logic
- **Configuration** (`pkg/config/`): YAML loading, validation, and variable substitution
- **Download Management** (`pkg/downloader/`): Git, archive, and direct file downloads
- **Cache Management** (`pkg/cache/`): Intelligent caching with expiration
- **Build Execution** (`pkg/builder/`): Build system integration (CMake, Make, Configure, custom)
- **Utilities** (`pkg/utils/`): File system helpers

### Key Components

#### Configuration System (`pkg/config/`)
- **loader.go**: Main configuration loading with hierarchy support
- **types.go**: Core data structures for projects and dependencies
- **variables.go**: Variable substitution system (built-in, environment, custom variables)
- **buildtag.go**: Build tag caching and optimization
- **detect.go**: Automatic configuration detection

#### Download Management (`pkg/downloader/`)
- **manager.go**: Download orchestration and parallel processing
- **git.go**: Git repository cloning and checkout
- **archive.go**: Archive extraction (tar, zip, etc.)
- **direct.go**: Direct file downloads

#### CLI Context (`cmd/cli/context.go`)
Global CLI context that manages:
- Project configuration loading
- Cache manager initialization
- Global options and flags
- Command execution state

### Variable System
BuildFly supports powerful variable substitution:
- **Built-in variables**: `${INSTALL_DIR}`, `${BUILD_DIR}`, `${SOURCE_DIR}`, `${BUILD_TYPE}`, `${CPU_CORES}`, `${OS}`, `${ARCH}`
- **Environment variables**: `${HOME}`, `${PATH}`, etc.
- **Custom variables**: Defined in `project.variables` section

### Build System Support
- **CMake**: Full integration with custom options and build commands
- **Make**: Traditional makefile-based builds
- **Configure**: Autoconf-style builds
- **Custom Scripts**: Arbitrary shell scripts (used for complex builds like Boost)

### Dependency Sources
- **Git repositories**: GitHub, GitLab, etc. with tag/branch/commit support
- **Archives**: tar.gz, zip, etc. with optional hash verification
- **Direct downloads**: Single files (header-only libraries)

### Caching Strategy
- **Download cache**: Avoids re-downloading the same version
- **Build cache**: Caches compiled artifacts per build tag
- **Expiration**: Configurable cache lifetime (default 7 days)
- **Size management**: Automatic cleanup of expired cache items

### CLI Command Structure
- **install**: Download and build dependencies
- **build**: Build dependencies only (no download)
- **clean**: Clean cache and installed files
- **init**: Initialize new project with templates
- **list**: Show dependencies and cache status
- **config**: Configuration management
- **detect**: Auto-detect project configuration

### Configuration File Format
The project uses `buildfly.yaml` with this structure:
```yaml
project:
  name: "project-name"
  version: "1.0.0"

dependencies:
  dependency_name:
    version: "1.0.0"
    source:
      type: "git|archive|direct"
      url: "..."
      tag: "..."  # for git
    build_system: "cmake|make|configure|custom"
    # build-system specific options

build_profiles:
  profile_name:
    variables:
      build_type: "Release"
    dependencies:
      - "dependency_name"
```

## Development Guidelines

### Code Organization
- Follow Go package naming conventions
- Use interfaces for abstraction (Downloader, CacheManager, etc.)
- Separate concerns clearly between packages
- Use dependency injection where appropriate

### Error Handling
- Use `fmt.Errorf("operation failed: %w", err)` for error wrapping
- Define clear error types in each package
- Handle context cancellation appropriately
- Provide meaningful error messages to users

### Testing Strategy
- Unit tests for all core packages
- Integration tests for CLI commands
- Use table-driven tests for multiple scenarios
- Mock external dependencies (HTTP, file system)

### Performance Considerations
- Parallel downloads with configurable concurrency
- Intelligent caching to avoid redundant operations
- Efficient file operations using proper buffering
- Memory-conscious processing of large files

### Platform Compatibility
- Handle path separators using `filepath.Join()`
- Consider Windows-specific requirements (long paths, PowerShell)
- Test on multiple platforms during development
- Use `runtime.GOOS` for platform-specific logic