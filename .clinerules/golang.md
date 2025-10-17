## 📁 `.cline/rules/cpp-dependency-manager.md`

```markdown
# C++ 依赖管理器开发规则

## 项目概述
这是一个用 Golang 开发的 C++ 依赖管理器，主要功能包括：
- C++ 包依赖解析和版本管理
- 跨平台构建支持
- 依赖下载和缓存
- 项目配置管理

## 代码结构规范

### 命名约定
- **包名**: 使用小写字母，简洁明了
- **接口名**: 使用 `er` 结尾，如 `Downloader`, `Resolver`
- **配置文件**: 使用 `.toml` 格式
- 错误变量: 使用 `Err` 前缀，如 `ErrDependencyNotFound`

## 开发规范

### 1. 错误处理
- 使用 `errors.Wrap()` 包装错误，保留堆栈信息
- 定义清晰的错误类型和错误码
- 提供可恢复的错误处理机制

```go
func (r *Resolver) Resolve(dep Dependency) (*ResolvedDependency, error) {
    if dep.Name == "" {
        return nil, errors.New("dependency name cannot be empty")
    }
    // ... 解析逻辑
}
```

### 2. 并发处理
- 使用 `sync.WaitGroup` 管理并发下载
- 实现连接池控制并发数量
- 使用 `context.Context` 实现超时控制

```go
func (d *Downloader) DownloadAll(deps []Dependency) error {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 5) // 限制并发数
    
    for _, dep := range deps {
        wg.Add(1)
        go func(dep Dependency) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            if err := d.downloadSingle(dep); err != nil {
                log.Printf("Failed to download %s: %v", dep.Name, err)
            }
        }(dep)
    }
    
    wg.Wait()
    return nil
}
```

### 3. 缓存机制
- 实现基于文件的缓存系统
- 支持缓存清理和验证
- 缓存键包含版本信息和平台标识

### 4. 配置管理
- 支持多格式配置文件（TOML 优先）
- 环境变量覆盖配置
- 配置验证和默认值设置

## API 设计原则

### 接口设计
```go
type DependencyResolver interface {
    Resolve(dep Dependency) (*ResolvedDependency, error)
    ResolveTransitive(deps []Dependency) ([]ResolvedDependency, error)
}

type DownloadManager interface {
    Download(dep ResolvedDependency, targetDir string) error
    VerifyIntegrity(dep ResolvedDependency, filePath string) bool
}

type CacheManager interface {
    Get(key string) ([]byte, bool)
    Set(key string, data []byte) error
    Cleanup() error
}
```

## 测试规范

### 单元测试
- 每个导出函数都需要测试
- 使用表格驱动测试
- 模拟外部依赖

```go
func TestDependencyResolver_Resolve(t *testing.T) {
    tests := []struct {
        name        string
        dep         Dependency
        wantErr     bool
        expectedVer string
    }{
        {
            name:        "valid dependency",
            dep:         Dependency{Name: "boost", Version: "1.75.0"},
            wantErr:     false,
            expectedVer: "1.75.0",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resolver := NewResolver()
            got, err := resolver.Resolve(tt.dep)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got.Version != tt.expectedVer {
                t.Errorf("Resolve() version = %v, want %v", got.Version, tt.expectedVer)
            }
        })
    }
}
```

## 平台兼容性考虑

### Windows 支持
- 处理路径分隔符差异
- 考虑长路径问题
- 支持 PowerShell 和 CMD

### Linux/macOS 支持
- 处理文件权限
- 支持符号链接
- 考虑不同的包管理器集成

## 性能优化要点

1. **并发下载**: 合理控制并发数量
2. **缓存策略**: LRU 缓存淘汰算法
3. **增量更新**: 只下载变更的依赖
4. **内存管理**: 使用对象池减少 GC 压力

## 常用命令模式

```go
// 命令行接口设计
var rootCmd = &cobra.Command{
    Use:   "cppdep",
    Short: "C++ Dependency Manager",
}

var installCmd = &cobra.Command{
    Use:   "install [package]",
    Short: "Install dependencies",
    Run: func(cmd *cobra.Command, args []string) {
        // 安装逻辑
    },
}
```

## 提示词模板

当需要实现新功能时，请按以下结构思考：

```
需求分析: [明确功能需求]
设计思路: [技术实现方案]
代码结构: [包和文件组织]
接口设计: [公开API定义]
错误处理: [异常情况处理]
测试方案: [验证方法]
```

请遵循这些规则来开发和维护 C++ 依赖管理器的 Golang 代码。
```

## 📁 `.cline/rules/golang-best-practices.md`

```markdown
# Golang 最佳实践规则

## 代码风格
- 使用 `gofmt` 格式化代码
- 遵循 Effective Go 指南
- 使用 `interface` 进行抽象
- 避免全局变量

## 错误处理模式
```go
// 好的做法
if err := doSomething(); err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}

// 使用 errors.Is 和 errors.As 进行错误判断
if errors.Is(err, os.ErrNotExist) {
    // 处理文件不存在
}
```

## 性能考虑
- 使用 `sync.Pool` 减少内存分配
- 避免在循环中创建大量临时对象
- 使用 `strings.Builder` 进行字符串拼接
```

## 📁 `.cline/rules/cpp-ecosystem.md`

```markdown
# C++ 生态系统集成规则

## 支持的依赖源
1. **Git 仓库**: GitHub, GitLab, Bitbucket
2. **包管理器**: vcpkg, Conan (可选)
3. **直接下载**: 压缩包、头文件
4. **系统包管理器**: apt, yum, brew

## 版本管理策略
- 语义化版本控制 (SemVer)
- 版本冲突解决
- 版本锁定文件支持
- 依赖树可视化

## 构建系统集成
- CMake 集成
- Makefile 生成
- 编译标志管理
- 跨平台构建配置
```

## 📁 `.cline/rules/ai-prompt-templates.md`

```markdown
# AI 提示词模板

## 实现新功能
```
我需要实现 [功能描述]，请按照以下要求：

功能需求:
- [具体需求1]
- [具体需求2]

技术约束:
- 必须兼容 [平台/版本]
- 性能要求: [性能指标]
- 内存使用: [内存限制]

请提供:
1. 接口设计
2. 核心实现代码
3. 错误处理方案
4. 单元测试示例
```

## 代码审查
```
请审查以下 Go 代码，关注：

代码质量方面:
- 是否符合 Go 最佳实践
- 错误处理是否完善
- 并发安全性
- 性能优化空间

安全方面:
- 是否有潜在的安全风险
- 输入验证是否充分
- 资源管理是否正确

请给出具体的改进建议。
```

## 调试帮助
```
我遇到了这个问题: [问题描述]

错误信息: [错误日志]
相关代码: [代码片段]

我已经尝试过:
- [尝试的解决方案1]
- [尝试的解决方案2]

请帮我分析可能的原因和解决方案。
```
