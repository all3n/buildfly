package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"buildfly/pkg/builder"
	"buildfly/pkg/cache"
	"buildfly/pkg/config"
	"buildfly/pkg/downloader"
	"buildfly/pkg/utils"

	"github.com/spf13/cobra"
)

// newInstallCmd 创建 install 命令
func newInstallCmd() *cobra.Command {
	var (
		force    bool
		noCache  bool
		profile  string
		buildTag string
	)

	cmd := &cobra.Command{
		Use:   "install [dependency...]",
		Short: "安装依赖",
		Long: `安装指定的依赖或配置文件中的所有依赖。

如果不指定依赖名称，则安装配置文件中的所有依赖。
支持指定构建配置文件来安装特定的依赖集合。

支持构建标签来区分不同的构建配置，例如：
--build-tag "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(args, force, noCache, profile, buildTag)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "强制重新安装")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "不使用缓存")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "构建配置文件")
	cmd.Flags().StringVar(&buildTag, "build-tag", "", "构建标签 (例如: arch=x86_64,platform=linux,runtime=glibc_2.35)")

	return cmd
}

// runInstall 执行安装
func runInstall(deps []string, force, noCache bool, profile, buildTag string) error {
	// 确保上下文已初始化
	if err := GlobalCLIContext.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize context: %w", err)
	}

	projectConfig := GlobalCLIContext.ProjectConfig

	// 解析构建标签
	var parsedBuildTag *config.BuildTag
	var err error
	if buildTag != "" {
		// 使用命令行指定的构建标签
		parsedBuildTag, err = config.ParseBuildTag(buildTag)
		if err != nil {
			return fmt.Errorf("invalid build tag: %w", err)
		}
		if err := parsedBuildTag.Validate(); err != nil {
			return fmt.Errorf("invalid build tag: %w", err)
		}
		fmt.Printf("Using build tag from command line: %s\n", parsedBuildTag.String())
	} else {
		// 尝试从环境变量获取
		if envBuildTag, err := config.GetBuildTagFromEnv(); err == nil && envBuildTag != nil {
			parsedBuildTag = envBuildTag
			fmt.Printf("Using build tag from environment: %s\n", parsedBuildTag.String())
		} else if projectConfig.BuildTag != nil {
			// 使用配置文件中的构建标签作为基础
			parsedBuildTag = projectConfig.BuildTag
			fmt.Printf("Using build tag from config: %s\n", parsedBuildTag.String())
		} else {
			// 自动检测构建标签
			fmt.Printf("No build tag specified, auto-detecting...\n")
			parsedBuildTag, err = config.GetDefaultBuildTag(projectConfig.BuildTag)
			if err != nil {
				fmt.Printf("Warning: Failed to auto-detect build tag: %v\n", err)
				fmt.Printf("Using minimal build tag...\n")
				parsedBuildTag = &config.BuildTag{}
			} else {
				fmt.Printf("Auto-detected build tag: %s\n", parsedBuildTag.String())
			}
		}
	}

	// 验证最终的构建标签
	if err := parsedBuildTag.Validate(); err != nil {
		return fmt.Errorf("invalid build tag: %w", err)
	}

	// 更新项目配置的构建标签（用于后续的构建过程）
	// project build tag
	projectConfig.BuildTag = parsedBuildTag

	// 确定要安装的依赖
	dependenciesToInstall, err := resolveDependencies(deps, profile)
	if err != nil {
		return err
	}

	if len(dependenciesToInstall) == 0 {
		fmt.Println("No dependencies to install")
		return nil
	}

	// 使用上下文中的管理器
	cacheManager := GlobalCLIContext.CacheManager
	downloadManager := GlobalCLIContext.DownloadManager

	// 安装每个依赖
	for _, dep := range dependenciesToInstall {
		if err := installDependency(dep, cacheManager, downloadManager, force, noCache); err != nil {
			return err
		}
	}

	fmt.Printf("\nSuccessfully installed %d dependencies\n", len(dependenciesToInstall))
	return nil
}

// resolveDependencies 解析要安装的依赖
func resolveDependencies(deps []string, profile string) ([]config.Dependency, error) {
	projectConfig := GlobalCLIContext.ProjectConfig
	var dependenciesToInstall []config.Dependency

	if len(deps) > 0 {
		// 安装指定的依赖
		for _, depName := range deps {
			if dep, exists := projectConfig.Dependencies[depName]; exists {
				dependenciesToInstall = append(dependenciesToInstall, dep)
			} else {
				return nil, fmt.Errorf("dependency not found: %s", depName)
			}
		}
	} else if profile != "" {
		// 安装指定配置文件的依赖
		if buildProfile, exists := projectConfig.BuildProfiles[profile]; exists {
			for _, depName := range buildProfile.Dependencies {
				if dep, exists := projectConfig.Dependencies[depName]; exists {
					dependenciesToInstall = append(dependenciesToInstall, dep)
				}
			}
		} else {
			return nil, fmt.Errorf("build profile not found: %s", profile)
		}
	} else {
		// 安装所有依赖
		for _, dep := range projectConfig.Dependencies {
			dependenciesToInstall = append(dependenciesToInstall, dep)
		}
	}

	return dependenciesToInstall, nil
}

// installDependency 安装单个依赖
func installDependency(dep config.Dependency, cacheManager *cache.CacheManager, downloadManager *downloader.DownloadManager, force, noCache bool) error {
	fmt.Printf("Installing %s (%s)...\n", dep.Name, dep.Version)

	// 尝试从缓存安装
	if !force && !noCache {
		if installed := tryInstallFromCache(dep, cacheManager, noCache); installed {
			return nil
		}
	}

	// 下载并安装
	return downloadAndInstall(dep, cacheManager, downloadManager, noCache)
}

// tryInstallFromCache 尝试从缓存安装依赖
func tryInstallFromCache(dep config.Dependency, cacheManager *cache.CacheManager, noCache bool) bool {
	if !cacheManager.IsCachedDownloads(dep) {
		return false
	}

	fmt.Printf("  Using cached download\n")

	// 检查是否需要构建
	if dep.BuildSystem != "" && dep.BuildSystem != "none" {
		return tryBuildFromCache(dep, cacheManager, noCache)
	} else {
		// 不需要构建的依赖，直接使用下载缓存并链接到项目
		if err := linkToProjectDir(dep); err != nil {
			fmt.Printf("  Failed to link from cache: %v\n", err)
			return false
		}
		fmt.Printf("  ✓ Installed from cache %s\n", dep.Name)
		return true
	}
}

// tryBuildFromCache 尝试从缓存构建依赖
func tryBuildFromCache(dep config.Dependency, cacheManager *cache.CacheManager, noCache bool) bool {
	projectConfig := GlobalCLIContext.ProjectConfig
	currentBuildTag := projectConfig.BuildTag

	// 生成标准化的构建和安装目录路径
	depBuildDir := getDepBuildDir(dep, currentBuildTag)
	depInstallDir := getDepInstallDir(dep, currentBuildTag)

	// 检查构建缓存
	if cacheManager.IsBuildCached(dep, currentBuildTag) {
		fmt.Printf("  Using cached build\n")

		// 确保安装目录存在
		if err := os.MkdirAll(depInstallDir, 0755); err != nil {
			fmt.Printf("  Failed to create install dir: %v\n", err)
			return false
		}

		if err := cacheManager.RetrieveBuild(dep, depInstallDir, currentBuildTag); err != nil {
			fmt.Printf("  Failed to retrieve build from cache: %v\n", err)
		} else {
			// 链接到项目目录
			if err := linkToProjectDir(dep); err != nil {
				fmt.Printf("  Failed to link to project: %v\n", err)
				return false
			}
			return true
		}
	}

	fmt.Printf("  Build cache not found, will build from download cache\n")

	// 从下载缓存恢复源码，然后构建
	if err := os.MkdirAll(depBuildDir, 0755); err != nil {
		fmt.Printf("  Failed to create build dir: %v\n", err)
		return false
	}

	if err := cacheManager.Retrieve(dep, depBuildDir); err != nil {
		fmt.Printf("  Failed to retrieve from cache: %v\n", err)
		return false
	}

	// 执行构建
	if err := compileInBuildDir(dep, depBuildDir, depInstallDir, noCache); err != nil {
		fmt.Printf("  Failed to build from cache: %v\n", err)
		return false
	}

	// 链接到项目目录
	if err := linkToProjectDir(dep); err != nil {
		fmt.Printf("  Failed to link to project: %v\n", err)
		return false
	}

	return true
}

// downloadAndInstall 下载并安装依赖
func downloadAndInstall(dep config.Dependency, cacheManager *cache.CacheManager, downloadManager *downloader.DownloadManager, noCache bool) error {
	// 步骤1: 下载压缩包（如果需要）
	sourceDir, err := downloadArchivesIfNeeded(dep, cacheManager, downloadManager, noCache)
	if err != nil {
		return fmt.Errorf("failed to download archives: %w", err)
	}

	// 获取当前构建标签
	currentBuildTag := GlobalCLIContext.ProjectConfig.BuildTag

	// 步骤2: 创建构建目录（如果需要）
	depBuildDir, err := createBuildDirIfNeeded(dep, currentBuildTag, sourceDir)
	if err != nil {
		return fmt.Errorf("failed to create build dir: %w", err)
	}

	// 检查是否需要构建
	if dep.BuildSystem != "" && dep.BuildSystem != "none" {
		// 步骤3: 在构建目录中编译
		depInstallDir := getDepInstallDir(dep, currentBuildTag)
		if err := compileInBuildDir(dep, depBuildDir, depInstallDir, noCache); err != nil {
			return fmt.Errorf("failed to compile: %w", err)
		}
	}

	// 步骤4: 链接到项目目录
	if err := linkToProjectDir(dep); err != nil {
		return fmt.Errorf("failed to link to project: %w", err)
	}

	return nil
}

// buildFromDownloadedSource 从下载的源码构建依赖
func buildFromDownloadedSource(dep config.Dependency, tempDir string, cacheManager *cache.CacheManager, noCache bool) error {
	fmt.Printf("  buildFromDownloadedSource with %s...\n", dep.BuildSystem)

	projectConfig := GlobalCLIContext.ProjectConfig
	currentBuildTag := projectConfig.BuildTag

	// 使用标准化的构建目录
	namedBuildDir := getDepBuildDir(dep, currentBuildTag)
	if err := os.MkdirAll(namedBuildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build dir: %w", err)
	}

	// 将下载的源码复制到构建目录
	if err := utils.CopyDir(tempDir, namedBuildDir); err != nil {
		return fmt.Errorf("failed to copy source to build dir: %w", err)
	}

	// 获取标准化的安装目录
	depInstallDir := getDepInstallDir(dep, currentBuildTag)

	// 执行构建
	if err := compileInBuildDir(dep, namedBuildDir, depInstallDir, noCache); err != nil {
		return fmt.Errorf("failed to build %s: %w", dep.Name, err)
	}

	return nil
}

// buildDependency 构建依赖的公共逻辑
func buildDependency(dep config.Dependency, namedBuildDir string, noCache bool) error {
	projectConfig := GlobalCLIContext.ProjectConfig
	// 创建变量上下文
	varCtx := config.NewVariableContext(projectConfig.Project, dep.Name)

	// 获取当前构建标签（从全局变量或项目配置）
	var currentBuildTag *config.BuildTag
	// 这里我们需要传递当前的 build tag，暂时从项目配置获取
	// TODO: 重构这个函数来接收 build tag 参数
	if projectConfig.BuildTag != nil {
		currentBuildTag = projectConfig.BuildTag
	}

	// 设置构建标签
	if currentBuildTag != nil {
		varCtx.SetBuildTag(currentBuildTag)
	}

	// 设置基础路径
	varCtx.InstallDir = getDepInstallDir(dep, currentBuildTag)
	varCtx.BuildDir = getDepBuildDir(dep, currentBuildTag)

	// 确保构建目录存在
	if err := os.MkdirAll(varCtx.BuildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build dir: %w", err)
	}

	// 初始化构建执行器
	executor := builder.NewBuildExecutor(varCtx)

	// 执行构建
	if err := executor.Execute(dep, namedBuildDir, varCtx.BuildDir, varCtx.InstallDir); err != nil {
		return fmt.Errorf("failed to build %s: %w", dep.Name, err)
	}

	// 缓存构建结果
	if !noCache {
		fmt.Printf("  Caching build result...\n")
		maxAge := GlobalCLIContext.parseMaxCacheAge()
		cacheManager := cache.NewCacheManager(projectConfig.CacheDir, 1024*1024*1024, maxAge)
		if err := cacheManager.StoreBuild(dep, varCtx.InstallDir, currentBuildTag); err != nil {
			fmt.Printf("  Warning: failed to cache build %s: %v\n", dep.Name, err)
		} else {
			fmt.Printf("  ✓ Cached build result in %s\n", cacheManager.GetBuildCachePath(dep, currentBuildTag))
		}
	}

	fmt.Printf("  ✓ Built and installed\n")
	return nil
}

// installDirectly 直接安装依赖（无需构建）
func installDirectly(dep config.Dependency, sourceDir, targetDir string) error {
	// 获取项目配置
	projectConfig := GlobalCLIContext.ProjectConfig

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 直接复制到目标目录
	// installPath := filepath.Join(targetDir, dep.Name)
	// fmt.Printf("  Installing directly... %s -> %s \n", sourceDir, installPath)
	// if err := utils.CopyDir(sourceDir, installPath); err != nil {
	// 	return fmt.Errorf("failed to copy %s: %w", dep.Name, err)
	// }

	// 软连接sourceDir到目标目录targetDir
	// list sourceDir下的所有文件和目录
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	linkDirs := make([]string, 0)
	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(targetDir, entry.Name())

		// 删除目标路径已存在的文件或目录
		if _, err := os.Lstat(targetPath); err == nil {
			if err := os.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("failed to remove existing target path %s: %w", targetPath, err)
			}
		}

		// 创建软连接
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", entry.Name(), err)
		}
		linkDirs = append(linkDirs, targetPath)
		fmt.Printf("  Linked %s -> %s\n", sourcePath, targetPath)
	}
	// 记录安装信息 用于卸载
	install_txt := filepath.Join(projectConfig.BuildFlyBaseDir, "install.txt")
	f, err := os.OpenFile(install_txt, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open install record file: %w", err)
	}
	defer f.Close()
	for _, dir := range linkDirs {
		if _, err := f.WriteString(dir + "\n"); err != nil {
			return fmt.Errorf("failed to write install record: %w", err)
		}
	}

	fmt.Printf("  ✓ Installed\n")
	return nil
}

// getDepBuildDir 获取依赖的标准化构建目录路径
func getDepBuildDir(dep config.Dependency, buildTag *config.BuildTag) string {
	projectConfig := GlobalCLIContext.ProjectConfig
	baseDir := filepath.Join(projectConfig.BuildDir, dep.Name, dep.Version)

	if buildTag != nil {
		buildTagDir := buildTag.ToDirName()
		return filepath.Join(baseDir, buildTagDir)
	}

	return baseDir
}

// getDepInstallDir 获取依赖的标准化安装目录路径
func getDepInstallDir(dep config.Dependency, buildTag *config.BuildTag) string {
	projectConfig := GlobalCLIContext.ProjectConfig
	baseDir := filepath.Join(projectConfig.InstallDir, dep.Name, dep.Version)

	if buildTag != nil {
		buildTagDir := buildTag.ToDirName()
		return filepath.Join(baseDir, buildTagDir)
	}

	return baseDir
}

// downloadArchivesIfNeeded 下载压缩包（如果需要）
func downloadArchivesIfNeeded(dep config.Dependency, cacheManager *cache.CacheManager, downloadManager *downloader.DownloadManager, noCache bool) (string, error) {
	// 检查是否已有缓存
	if !noCache && cacheManager.IsCachedDownloads(dep) {
		fmt.Printf("  Using cached download\n")
		// 创建临时目录来恢复缓存内容
		tempDir, err := os.MkdirTemp("", "buildfly-cache-*")
		if err != nil {
			return "", fmt.Errorf("failed to create temp dir: %w", err)
		}

		if err := cacheManager.Retrieve(dep, tempDir); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to retrieve from cache: %w", err)
		}

		return tempDir, nil
	}

	// 创建临时下载目录
	tempDir, err := os.MkdirTemp("", "buildfly-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 下载依赖
	urls := dep.Source.GetURLs()
	if len(urls) > 0 {
		fmt.Printf("  Downloading from %s...\n", urls[0])
	} else {
		fmt.Printf("  Downloading...\n")
	}

	if err := downloadManager.Download(context.Background(), dep, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download %s: %w", dep.Name, err)
	}

	// 缓存下载的源码（保留压缩包在本地 cache 目录）
	if !noCache {
		if err := cacheManager.Store(dep, tempDir); err != nil {
			fmt.Printf("  Warning: failed to cache download %s: %v\n", dep.Name, err)
		} else {
			fmt.Printf("  ✓ Cached download source in %s\n", cacheManager.GetDownloadCachePath(dep))
		}
	}

	return tempDir, nil
}

// createBuildDirIfNeeded 创建构建目录（如果需要）
func createBuildDirIfNeeded(dep config.Dependency, buildTag *config.BuildTag, sourceDir string) (string, error) {
	// 生成标准化的构建目录路径
	depBuildDir := getDepBuildDir(dep, buildTag)

	// 确保构建目录存在
	if err := os.MkdirAll(depBuildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build dir: %w", err)
	}

	// 将源码复制到构建目录
	if err := utils.CopyDir(sourceDir, depBuildDir); err != nil {
		return "", fmt.Errorf("failed to copy source to build dir: %w", err)
	}

	fmt.Printf("  Prepared build directory: %s\n", depBuildDir)
	return depBuildDir, nil
}

// compileInBuildDir 在构建目录中编译
func compileInBuildDir(dep config.Dependency, buildDir, installDir string, noCache bool) error {
	fmt.Printf("  Compiling %s with %s...\n", dep.Name, dep.BuildSystem)

	projectConfig := GlobalCLIContext.ProjectConfig
	currentBuildTag := projectConfig.BuildTag

	// 创建变量上下文
	varCtx := config.NewVariableContext(projectConfig.Project, dep.Name)

	// 设置构建标签
	if currentBuildTag != nil {
		varCtx.SetBuildTag(currentBuildTag)
	}

	// 设置基础路径
	varCtx.BuildDir = buildDir
	varCtx.InstallDir = installDir

	// 确保安装目录存在
	if err := os.MkdirAll(varCtx.InstallDir, 0755); err != nil {
		return fmt.Errorf("failed to create install dir: %w", err)
	}

	// 初始化构建执行器
	executor := builder.NewBuildExecutor(varCtx)

	// 执行构建
	if err := executor.Execute(dep, buildDir, varCtx.BuildDir, varCtx.InstallDir); err != nil {
		return fmt.Errorf("failed to build %s: %w", dep.Name, err)
	}

	// 缓存构建结果
	if !noCache {
		fmt.Printf("  Caching build result...\n")
		maxAge := GlobalCLIContext.parseMaxCacheAge()
		cacheManager := cache.NewCacheManager(projectConfig.CacheDir, 1024*1024*1024, maxAge)
		if err := cacheManager.StoreBuild(dep, varCtx.InstallDir, currentBuildTag); err != nil {
			fmt.Printf("  Warning: failed to cache build %s: %v\n", dep.Name, err)
		} else {
			fmt.Printf("  ✓ Cached build result in %s\n", cacheManager.GetBuildCachePath(dep, currentBuildTag))
		}
	}

	fmt.Printf("  ✓ Compiled %s\n", dep.Name)
	return nil
}

// linkToProjectDir 链接到项目目录
func linkToProjectDir(dep config.Dependency) error {
	projectConfig := GlobalCLIContext.ProjectConfig
	currentBuildTag := projectConfig.BuildTag

	// 获取源目录（安装目录）
	sourceDir := getDepInstallDir(dep, currentBuildTag)

	// 如果不需要构建，源目录是项目安装目录
	if dep.BuildSystem == "" || dep.BuildSystem == "none" {
		sourceDir = projectConfig.InstallDir
	}

	// 目标目录：项目的 .buildfly/install/{depName}/
	targetDepDir := filepath.Join(projectConfig.BuildFlyBaseDir, "install", dep.Name)

	// 确保目标目录存在
	if err := os.MkdirAll(targetDepDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 读取源目录下的所有文件和目录
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	linkDirs := make([]string, 0)
	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(targetDepDir, entry.Name())

		// 删除目标路径已存在的文件或目录
		if _, err := os.Lstat(targetPath); err == nil {
			if err := os.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("failed to remove existing target path %s: %w", targetPath, err)
			}
		}

		// 创建绝对路径的软连接
		absSourcePath, err := filepath.Abs(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for source %s: %w", sourcePath, err)
		}

		if err := os.Symlink(absSourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", entry.Name(), err)
		}
		linkDirs = append(linkDirs, targetPath)
		fmt.Printf("  Linked %s -> %s\n", targetPath, absSourcePath)
	}

	// 记录安装信息用于卸载
	installTxt := filepath.Join(projectConfig.BuildFlyBaseDir, "install.txt")
	f, err := os.OpenFile(installTxt, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open install record file: %w", err)
	}
	defer f.Close()

	for _, dir := range linkDirs {
		if _, err := f.WriteString(dir + "\n"); err != nil {
			return fmt.Errorf("failed to write install record: %w", err)
		}
	}

	fmt.Printf("  ✓ Linked %s to project\n", dep.Name)
	return nil
}

// parseDuration 解析时间字符串（简化版本）
func parseDuration(s string) interface{} {
	// 这里应该实现完整的时间解析逻辑
	// 为了简化，返回一个固定的时间
	return 7 * 24 * time.Hour
}
