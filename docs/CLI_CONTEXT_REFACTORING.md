# CLI 上下文重构总结

## 重构目标

解决 `cmd/cli/root.go` 中全局选项和命令重复的问题，通过结构体封装和全局上下文对象来统一管理。

## 重构内容

### 1. 创建了新的上下文结构

#### `cmd/cli/context.go`
- **CLIContext**: 全局CLI上下文，包含所有共享状态
- **GlobalOptions**: 全局选项结构体，封装所有命令行参数
- **CommandContext**: 命令上下文，用于传递命令特定的状态

#### 核心结构体设计
```go
type CLIContext struct {
    Options        *GlobalOptions
    ProjectConfig  *config.ProjectConfig
    CacheManager   *cache.CacheManager
    DownloadManager *downloader.DownloadManager
    ConfigLoader   *config.ConfigLoader
    ConfigFile     string
    WorkDir        string
    initialized    bool
}

type GlobalOptions struct {
    ConfigFile     string
    Verbose        bool
    CacheDir       string
    MaxCacheAge    string
}
```

### 2. 重构了 `root.go`

#### 改进前的问题
- 全局变量散乱（`cfgFile`, `verbose`, `cacheDir`, `maxCacheAge`）
- 重复的标志定义
- 缺乏统一的状态管理

#### 改进后的优势
- 统一的上下文管理
- 清晰的选项封装
- 自动初始化机制
- 更好的代码组织

### 3. 重构了 `install.go`

#### 主要改进
- 移除了全局变量依赖
- 使用 `GlobalCLIContext` 访问共享状态
- 统一的配置加载和管理
- 更好的错误处理

## 重构效果

### 1. 解决了重复问题
- ✅ 全局选项不再与命令选项重复
- ✅ 统一的选项管理机制
- ✅ 避免了全局变量污染

### 2. 提高了代码质量
- ✅ 更好的封装性
- ✅ 清晰的职责分离
- ✅ 统一的状态管理
- ✅ 更容易测试和维护

### 3. 增强了可扩展性
- ✅ 新命令可以轻松访问共享状态
- ✅ 配置管理更加灵活
- ✅ 支持更复杂的初始化逻辑

## 使用方式

### 全局访问
```go
// 确保上下文已初始化
if err := GlobalCLIContext.Initialize(); err != nil {
    return fmt.Errorf("failed to initialize context: %w", err)
}

// 访问项目配置
projectConfig := GlobalCLIContext.ProjectConfig

// 访问管理器
cacheManager := GlobalCLIContext.CacheManager
downloadManager := GlobalCLIContext.DownloadManager
```

### 命令创建
```go
func newInstallCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "install [dependency...]",
        Short: "安装依赖",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runInstall(args, force, noCache, profile, targetDir, buildTag)
        },
    }
    
    // 添加标志（不再需要全局变量）
    cmd.Flags().BoolVar(&force, "force", false, "强制重新安装")
    // ...
    
    return cmd
}
```

## 向后兼容性

- ✅ 保持了所有现有的命令行接口
- ✅ 保持了所有现有功能
- ✅ 用户无需修改使用方式

## 测试结果

### 编译测试
```bash
go build -o bin/buildfly ./cmd/main.go
# ✅ 编译成功
```

### 功能测试
```bash
./bin/buildfly --help
# ✅ 显示帮助信息

./bin/buildfly install --help
# ✅ 显示安装命令帮助
```

## 后续改进建议

1. **其他命令重构**: 将其他命令文件也重构为使用新的上下文
2. **配置验证**: 在上下文初始化时添加配置验证
3. **错误处理**: 统一错误处理机制
4. **日志系统**: 集成统一的日志系统
5. **性能优化**: 优化上下文初始化性能

## 文件变更

### 新增文件
- `cmd/cli/context.go` - 上下文管理

### 修改文件
- `cmd/cli/root.go` - 重构为使用上下文
- `cmd/cli/install.go` - 重构为使用上下文

### 保持不变
- 其他命令文件（待后续重构）
- 核心业务逻辑
- 配置系统

## 总结

这次重构成功地解决了全局选项和命令重复的问题，通过引入统一的上下文管理机制，提高了代码的可维护性和可扩展性。重构后的代码结构更清晰，职责分离更明确，为后续的功能开发奠定了良好的基础。
