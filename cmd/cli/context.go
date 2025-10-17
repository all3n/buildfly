package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"buildfly/pkg/cache"
	"buildfly/pkg/config"
	"buildfly/pkg/downloader"
)

// CLIContext 全局CLI上下文
type CLIContext struct {
	// 全局选项
	GlobalOptions *GlobalOptions

	// 项目配置
	ProjectConfig *config.ProjectConfig

	// 管理器实例
	CacheManager    *cache.CacheManager
	DownloadManager *downloader.DownloadManager

	// 运行时状态
	Initialized bool
}

// GlobalOptions 全局选项结构
type GlobalOptions struct {
	ConfigFile  string
	Verbose     bool
	CacheDir    string
	MaxCacheAge string
}

// NewCLIContext 创建新的CLI上下文
func NewCLIContext() *CLIContext {
	return &CLIContext{
		GlobalOptions: &GlobalOptions{},
		Initialized:   false,
	}
}

// Initialize 初始化上下文
func (ctx *CLIContext) Initialize() error {
	if ctx.Initialized {
		return nil
	}

	// 加载项目配置
	if err := ctx.loadProjectConfig(); err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// 初始化缓存管理器
	if err := ctx.initCacheManager(); err != nil {
		return fmt.Errorf("failed to init cache manager: %w", err)
	}

	// 初始化下载管理器（使用配置中的代理设置）
	ctx.DownloadManager = downloader.NewDownloadManagerFromConfig(5, ctx.ProjectConfig) // 最大并发数

	ctx.Initialized = true

	if ctx.GlobalOptions.Verbose {
		fmt.Println("CLI context initialized successfully")
	}

	return nil
}

// loadProjectConfig 加载项目配置
func (ctx *CLIContext) loadProjectConfig() error {
	configFile := ctx.getConfigFile()
	if configFile == "" {
		return fmt.Errorf("no config file found, run 'buildfly init' to create one")
	}

	loader := config.NewConfigLoader(filepath.Dir(configFile))
	projectConfig, err := loader.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx.ProjectConfig = projectConfig
	return nil
}

// initCacheManager 初始化缓存管理器
func (ctx *CLIContext) initCacheManager() error {
	cacheDir := ctx.ProjectConfig.CacheDir
	maxAge := ctx.parseMaxCacheAge()

	cacheManager := cache.NewCacheManager(cacheDir, 1024*1024*1024, maxAge)

	if err := cacheManager.Init(); err != nil {
		return fmt.Errorf("failed to init cache: %w", err)
	}

	ctx.CacheManager = cacheManager
	return nil
}

// getConfigFile 获取配置文件路径
func (ctx *CLIContext) getConfigFile() string {
	if ctx.GlobalOptions.ConfigFile != "" {
		return ctx.GlobalOptions.ConfigFile
	}

	// 查找默认配置文件
	for _, name := range []string{"buildfly.yaml", "buildfly.yml", ".buildfly.yaml", "BUILDFLY"} {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	return ""
}

// parseMaxCacheAge 解析最大缓存时间
func (ctx *CLIContext) parseMaxCacheAge() time.Duration {
	maxAgeStr := ctx.GlobalOptions.MaxCacheAge
	if maxAgeStr == "" {
		maxAgeStr = "7d" // 默认7天
	}

	// 简化的时间解析，实际应该实现完整的解析逻辑
	switch maxAgeStr {
	case "1d":
		return 24 * time.Hour
	case "7d":
		return 7 * 24 * time.Hour
	case "30d":
		return 30 * 24 * time.Hour
	default:
		return 7 * 24 * time.Hour
	}
}

// Reset 重置上下文
func (ctx *CLIContext) Reset() {
	ctx.ProjectConfig = nil
	ctx.CacheManager = nil
	ctx.DownloadManager = nil
	ctx.Initialized = false
}

// 全局上下文实例
var GlobalCLIContext = NewCLIContext()
