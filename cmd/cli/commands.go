package cli

import (
	"buildfly/internal/constants"
	"buildfly/pkg/cache"
	"buildfly/pkg/config"
	"buildfly/pkg/venv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newBuildCmd 创建 build 命令
func newBuildCmd() *cobra.Command {
	var (
		force   bool
		profile string
	)

	cmd := &cobra.Command{
		Use:   "build [dependency...]",
		Short: "构建依赖",
		Long: `构建指定的依赖或配置文件中的所有依赖。

只执行构建过程，不下载依赖。适用于需要重新构建的场景。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(args, force, profile)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "强制重新构建")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "构建配置文件")

	return cmd
}

// newCleanCmd 创建 clean 命令
func newCleanCmd() *cobra.Command {
	var (
		all    bool
		cache  bool
		deps   bool
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "清理缓存和构建文件",
		Long: `清理缓存、构建文件和已安装的依赖。

支持选择性清理不同类型的文件。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(all, cache, deps, dryRun)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "清理所有文件")
	cmd.Flags().BoolVar(&cache, "cache", false, "清理缓存")
	cmd.Flags().BoolVar(&deps, "deps", false, "清理已安装的依赖")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "显示将要删除的文件，但不实际删除")

	return cmd
}

// newInitCmd 创建 init 命令
func newInitCmd() *cobra.Command {
	var (
		name     string
		template string
		force    bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化项目",
		Long: `在当前目录创建一个新的 BuildFly 项目配置文件。

可以指定项目名称和模板类型。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(name, template, force)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "项目名称")
	cmd.Flags().StringVarP(&template, "template", "t", "basic", "项目模板 (basic, cmake, make)")
	cmd.Flags().BoolVar(&force, "force", false, "覆盖现有配置文件")

	return cmd
}

// newListCmd 创建 list 命令
func newListCmd() *cobra.Command {
	var (
		verbose bool
		cache   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出依赖和缓存信息",
		Long: `显示项目中配置的依赖和缓存状态。

可以显示详细的依赖信息和缓存统计。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(verbose, cache)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细信息")
	cmd.Flags().BoolVar(&cache, "cache", false, "显示缓存信息")

	return cmd
}

// newConfigCmd 创建 config 命令
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
		Long: `管理 BuildFly 的配置选项。

支持查看、设置和重置配置选项。`,
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigResetCmd())

	return cmd
}

// newConfigShowCmd 创建 config show 命令
func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "显示当前配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow()
		},
	}
}

// newConfigSetCmd 创建 config set 命令
func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "设置配置选项",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSet(args[0], args[1])
		},
	}
}

// newConfigResetCmd 创建 config reset 命令
func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "重置配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigReset()
		},
	}
}

// runBuild 执行构建
func runBuild(deps []string, force bool, profile string) error {
	fmt.Println("Build command not implemented yet")
	return nil
}

// newVersionCmd 创建 version 命令
func newVersionCmd() *cobra.Command {
	var (
		short bool
	)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long: `显示 BuildFly 的版本信息。

包括版本号、构建信息等详细信息。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(short)
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "只显示版本号")

	return cmd
}

// runVersion 执行版本显示
func runVersion(short bool) error {
	buildInfo := constants.GetBuildInfo()

	if short {
		fmt.Println(buildInfo.Version)
		return nil
	}

	fmt.Printf("BuildFly %s\n", buildInfo.Version)
	fmt.Println("A C++ dependency manager written in Go")
	fmt.Println()

	if buildInfo.GitCommit != "unknown" {
		fmt.Printf("Git Commit: %s\n", buildInfo.GitCommit)
	}
	if buildInfo.BuildDate != "unknown" {
		fmt.Printf("Build Date: %s\n", buildInfo.BuildDate)
	}
	if buildInfo.GoVersion != "unknown" {
		fmt.Printf("Go Version: %s\n", buildInfo.GoVersion)
	}

	return nil
}

// runClean 执行清理
func runClean(all, cache, deps, dryRun bool) error {
	if all || cache {
		// 确保上下文已初始化
		if err := GlobalCLIContext.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize context: %w", err)
		}
		cacheDir := GlobalCLIContext.ProjectConfig.CacheDir
		if dryRun {
			fmt.Printf("Would clean cache directory: %s\n", cacheDir)
		} else {
			fmt.Printf("Cleaning cache directory: %s\n", cacheDir)
			if err := os.RemoveAll(cacheDir); err != nil {
				return fmt.Errorf("failed to clean cache: %w", err)
			}
		}
	}

	if all || deps {
		depsDir := "deps"
		if _, err := os.Stat(depsDir); err == nil {
			if dryRun {
				fmt.Printf("Would clean dependencies directory: %s\n", depsDir)
			} else {
				fmt.Printf("Cleaning dependencies directory: %s\n", depsDir)
				if err := os.RemoveAll(depsDir); err != nil {
					return fmt.Errorf("failed to clean dependencies: %w", err)
				}
			}
		}
	}

	fmt.Println("Clean completed")
	return nil
}

// runInit 执行初始化
func runInit(name, template string, force bool) error {
	configFile := "buildfly.yaml"
	if _, err := os.Stat(configFile); err == nil && !force {
		return fmt.Errorf("config file already exists, use --force to overwrite")
	}

	// 确定项目名称
	if name == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		name = filepath.Base(dir)
	}

	// 生成配置内容
	configContent := generateConfig(name, template)

	// 写入配置文件
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Initialized BuildFly project '%s' with template '%s'\n", name, template)
	return nil
}

// generateConfig 生成配置内容
func generateConfig(name, template string) string {
	switch template {
	case "cmake":
		return fmt.Sprintf(`project:
  name: "%s"
  version: "1.0.0"
  variables:
    install_dir: ".buildfly"
    build_type: "Release"

dependencies:
  fmt:
    version: "8.0.1"
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"
    cmake_options:
      - "FMT_TEST=OFF"
      - "CMAKE_POSITION_INDEPENDENT_CODE=ON"

build_profiles:
  debug:
    variables:
      build_type: "Debug"
    dependencies:
      - "fmt"
  
  release:
    variables:
      build_type: "Release"
    dependencies:
      - "fmt"
`, name)
	case "make":
		return fmt.Sprintf(`project:
  name: "%s"
  version: "1.0.0"
  variables:
    install_dir: ".buildfly"

dependencies:
  zlib:
    version: "1.2.11"
    source:
      type: "archive"
      url: "https://zlib.net/zlib-1.2.11.tar.gz"
    build_system: "configure"
    configure_options:
      - "--prefix=${INSTALL_DIR}"

build_profiles:
  default:
    dependencies:
      - "zlib"
`, name)
	default: // basic
		return fmt.Sprintf(`project:
  name: "%s"
  version: "1.0.0"
  variables:
    install_dir: ".buildfly"

dependencies:
  # 在这里添加你的依赖
  # example:
  # boost:
  #   version: "1.89.0"
  #   source:
  #     type: "archive"
  #     url: "https://archives.boost.io/release/1.89.0/source/boost_1_89_0.tar.gz"
  #   build_system: "custom"
  #   custom_script: |
  #     #!/bin/bash
  #     ./bootstrap.sh --prefix=${INSTALL_DIR} --with-libraries=iostreams,random,filesystem
  #     ./b2 install

build_profiles:
  default:
    dependencies: []
`, name)
	}
}

// runList 执行列表
func runList(verbose, showCache bool) error {
	// 确保上下文已初始化
	if err := GlobalCLIContext.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize context: %w", err)
	}

	// 使用上下文中的配置和缓存管理器
	projectConfig := GlobalCLIContext.ProjectConfig
	cacheMgr := GlobalCLIContext.CacheManager

	if showCache {
		return showCacheInfo(cacheMgr, verbose)
	}

	return showDependencyList(projectConfig, cacheMgr, verbose)
}

// showDependencyList 显示依赖列表
func showDependencyList(projectConfig *config.ProjectConfig, cacheMgr *cache.CacheManager, verbose bool) error {
	fmt.Printf("Dependencies in project '%s':\n", projectConfig.Project.Name)

	if len(projectConfig.Dependencies) == 0 {
		fmt.Println("  No dependencies configured")
		return nil
	}

	if verbose {
		// 详细模式
		fmt.Printf("%-15s %-10s %-9s %-7s %-10s %-12s %s\n",
			"NAME", "VERSION", "DOWNLOAD", "BUILD", "SOURCE", "BUILD SYS", "INSTALL PATH")
		fmt.Println(strings.Repeat("-", 80))
	} else {
		// 简单模式
		fmt.Printf("%-15s %-10s %-9s %-7s %-10s %-12s\n",
			"NAME", "VERSION", "DOWNLOAD", "BUILD", "SOURCE", "BUILD SYS")
		fmt.Println(strings.Repeat("-", 70))
	}

	for name, dep := range projectConfig.Dependencies {
		// 检查缓存状态
		downloadCached := cacheMgr.IsCachedDownloads(dep)
		buildCached := cacheMgr.IsBuildCached(dep, projectConfig.BuildTag)

		downloadStatus := "✗"
		if downloadCached {
			downloadStatus = "✓"
		}

		buildStatus := "✗"
		if buildCached {
			buildStatus = "✓"
		}

		if verbose {
			installPath := filepath.Join(getInstallDir(projectConfig), name)
			fmt.Printf("%-15s %-10s %-9s %-7s %-10s %-12s %s\n",
				name, dep.Version, downloadStatus, buildStatus, dep.Source.Type, dep.BuildSystem, installPath)
		} else {
			fmt.Printf("%-15s %-10s %-9s %-7s %-10s %-12s\n",
				name, dep.Version, downloadStatus, buildStatus, dep.Source.Type, dep.BuildSystem)
		}
	}

	// 显示图例
	fmt.Println("\nLegend:")
	fmt.Println("  ✓ = cached/downloaded")
	fmt.Println("  ✗ = not cached/downloaded")

	return nil
}

// showCacheInfo 显示缓存信息
func showCacheInfo(cacheMgr *cache.CacheManager, verbose bool) error {
	fmt.Println("Cache Statistics:")

	// 获取缓存大小
	cacheSize, err := cacheMgr.GetCacheSize()
	if err != nil {
		return fmt.Errorf("failed to get cache size: %w", err)
	}

	// 获取缓存列表
	cacheInfos, err := cacheMgr.ListCache()
	if err != nil {
		return fmt.Errorf("failed to list cache: %w", err)
	}

	expiredCount := 0
	for _, info := range cacheInfos {
		if info.Expired {
			expiredCount++
		}
	}

	fmt.Printf("  Total size: %s\n", formatBytes(cacheSize))
	fmt.Printf("  Items: %d\n", len(cacheInfos))
	fmt.Printf("  Expired: %d\n", expiredCount)

	if verbose && len(cacheInfos) > 0 {
		fmt.Println("\nCache items:")
		fmt.Printf("%-40s %-12s %-5s %s\n", "NAME", "SIZE", "STATUS", "MODIFIED")
		fmt.Println(strings.Repeat("-", 75))

		for _, info := range cacheInfos {
			name := filepath.Base(info.Path)
			sizeStr := formatBytes(info.Size)

			status := "✓"
			if info.Expired {
				status = "✗"
			}

			modTime := info.ModTime.Format("2006-01-02 15:04")
			fmt.Printf("%-40s %-12s %-5s %s\n", name, sizeStr, status, modTime)
		}
	}

	return nil
}

// formatBytes 格式化字节数为人类可读格式
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getInstallDir 获取安装目录路径
func getInstallDir(projectConfig *config.ProjectConfig) string {
	if installDir, exists := projectConfig.Project.Variables["install_dir"]; exists {
		// 展开环境变量
		return os.ExpandEnv(installDir)
	}
	return filepath.Join(os.Getenv("HOME"), ".buildfly", "install")
}

// runConfigShow 执行配置显示
func runConfigShow() error {
	fmt.Printf("Config file: %s\n", GlobalCLIContext.getConfigFile())
	fmt.Printf("Verbose mode: %v\n", GlobalCLIContext.GlobalOptions.Verbose)
	return nil
}

// runConfigSet 执行配置设置
func runConfigSet(key, value string) error {
	fmt.Printf("Set config %s = %s\n", key, value)
	return nil
}

// runConfigReset 执行配置重置
func runConfigReset() error {
	fmt.Println("Reset config")
	return nil
}

// newRunCmd 创建 run 命令
func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command] [args...]",
		Short: "在虚拟环境中运行命令",
		Long: `在虚拟环境中运行指定的命令。

如果虚拟环境未初始化，会自动初始化环境。
这是 'buildfly venv run' 的快捷方式。

示例:
  buildfly run cmake --version
  buildfly run ninja --version
  buildfly run gcc --version`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(args)
		},
		// 禁用标志解析，将所有参数传递给目标命令
		DisableFlagParsing: true,
	}

	return cmd
}

// runRun 执行命令运行
func runRun(args []string) error {
	// 确保上下文已初始化
	if err := GlobalCLIContext.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize CLI context: %w", err)
	}

	// 获取项目根目录
	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	// 创建或获取 VENV 配置
	venvConfig := GlobalCLIContext.ProjectConfig.VEnv
	fmt.Printf("Using venv config: %+v\n", venvConfig)
	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}
	// 运行命令
	if err := manager.Run(args); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}
