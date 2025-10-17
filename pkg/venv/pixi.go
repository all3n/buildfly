package venv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PixiBackend Pixi 后端实现
type PixiBackend struct {
	info     *PixiInfo
	rootDir  string
	config   *VEnvConfig
	platform *PlatformInfo
}

// NewPixiBackend 创建 Pixi 后端实例
func NewPixiBackend() *PixiBackend {
	return &PixiBackend{
		platform: getPlatformInfo(),
	}
}

// Detect 检测 Pixi 是否已安装
func (p *PixiBackend) Detect() (*BackendInfo, error) {
	info, err := p.detectPixi()
	if err != nil {
		return &BackendInfo{
			Name:      "pixi",
			Installed: false,
		}, nil
	}

	p.info = info

	return &BackendInfo{
		Name:      "pixi",
		Version:   info.Version,
		Path:      info.Path,
		Installed: info.Installed,
		Environment: map[string]string{
			"PIXI_HOME":    os.Getenv("PIXI_HOME"),
			"PIXI_VERSION": os.Getenv("PIXI_VERSION"),
		},
	}, nil
}

// detectPixi 检测 Pixi 安装状态
func (p *PixiBackend) detectPixi() (*PixiInfo, error) {
	// 检查环境变量中的路径
	pixiPath := os.Getenv("PIXI_HOME")
	if pixiPath != "" {
		pixiExe := filepath.Join(pixiPath, "bin", "pixi")
		if runtime.GOOS == "windows" {
			pixiExe += ".exe"
		}

		if _, err := os.Stat(pixiExe); err == nil {
			version, _ := p.getPixiVersion(pixiExe)
			return &PixiInfo{
				Installed: true,
				Version:   version,
				Path:      pixiExe,
			}, nil
		}
	}

	// 检查系统 PATH 中的 pixi
	path, err := exec.LookPath("pixi")
	if err == nil {
		version, _ := p.getPixiVersion(path)
		return &PixiInfo{
			Installed: true,
			Version:   version,
			Path:      path,
		}, nil
	}

	return &PixiInfo{
		Installed: false,
	}, nil
}

// getPixiVersion 获取 Pixi 版本
func (p *PixiBackend) getPixiVersion(pixiPath string) (string, error) {
	cmd := exec.Command(pixiPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 解析版本信息，输出格式类似: pixi 0.35.1
	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "pixi ") {
		return strings.TrimPrefix(outputStr, "pixi "), nil
	}

	return outputStr, nil
}

// Install 安装 Pixi
func (p *PixiBackend) Install(version string) error {
	fmt.Printf("Installing Pixi %s...\n", version)

	switch runtime.GOOS {
	case "windows":
		return p.installPixiWindows(version)
	default:
		return p.installPixiUnix(version)
	}
}

// installPixiUnix 在 Unix 系统上安装 Pixi
func (p *PixiBackend) installPixiUnix(version string) error {
	// 设置环境变量
	env := os.Environ()
	if version != "" && version != "latest" {
		env = append(env, fmt.Sprintf("PIXI_VERSION=%s", version))
	}

	// 使用官方安装脚本
	cmd := exec.Command("curl", "-fsSL", "https://pixi.sh/install.sh", "|", "sh")
	if version != "" && version != "latest" {
		cmd.Args = []string{"curl", "-fsSL", fmt.Sprintf("https://pixi.sh/install.sh|%s", version), "|", "sh"}
	}

	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Pixi: %w", err)
	}

	fmt.Println("Pixi installed successfully")
	return nil
}

// installPixiWindows 在 Windows 上安装 Pixi
func (p *PixiBackend) installPixiWindows(version string) error {
	// 设置环境变量
	env := os.Environ()
	if version != "" && version != "latest" {
		env = append(env, fmt.Sprintf("PIXI_VERSION=%s", version))
	}

	// 使用 PowerShell 安装脚本
	cmd := exec.Command("powershell", "-ExecutionPolicy", "ByPass", "-c", "irm -useb https://pixi.sh/install.ps1 | iex")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Pixi on Windows: %w", err)
	}

	fmt.Println("Pixi installed successfully")
	return nil
}

// Initialize 初始化 Pixi 环境
func (p *PixiBackend) Initialize(config *VEnvConfig, rootDir string) error {
	p.config = config
	p.rootDir = rootDir

	fmt.Printf("Initializing Pixi environment at: %s\n", rootDir)

	// 确保目录存在
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	// 创建 pixi.toml 配置文件
	if err := p.CreateConfig(config, rootDir); err != nil {
		return fmt.Errorf("failed to create Pixi config: %w", err)
	}

	// 初始化 Pixi 项目
	if err := p.initPixiProject(); err != nil {
		return fmt.Errorf("failed to initialize Pixi project: %w", err)
	}

	// 安装 C++ 工具
	if err := p.installCPPTools(); err != nil {
		return fmt.Errorf("failed to install C++ tools: %w", err)
	}

	fmt.Println("Pixi environment initialized successfully")
	return nil
}

// CreateConfig 创建 Pixi 配置文件
func (p *PixiBackend) CreateConfig(config *VEnvConfig, rootDir string) error {
	pixiTomlPath := filepath.Join(rootDir, "pixi.toml")

	// 检查文件是否已存在
	if _, err := os.Stat(pixiTomlPath); err == nil {
		fmt.Printf("Pixi config file already exists: %s\n", pixiTomlPath)
		return nil
	}

	// 获取默认配置
	pixiConfig := config.Pixi
	if pixiConfig.Version == "" {
		pixiConfig.Version = "0.35.1"
	}
	if len(pixiConfig.Channels) == 0 {
		pixiConfig.Channels = []string{"conda-forge"}
	}
	if len(pixiConfig.Platforms) == 0 {
		pixiConfig.Platforms = p.getDefaultPlatforms()
	}

	// 生成 pixi.toml 内容
	content := fmt.Sprintf(`[project]
name = "buildfly-env"
version = "0.1.0"
description = "BuildFly C++ Development Environment"
authors = ["BuildFly <buildfly@example.com>"]
channels = [%s]
platforms = [%s]

[dependencies]
python = "%s"
pip = "*"

`, p.formatChannels(pixiConfig.Channels), p.formatPlatforms(pixiConfig.Platforms), config.Python.Version)

	// 添加构建工具依赖
	content += p.generateBuildDependencies(config)

	// 添加平台特定依赖
	content += p.generatePlatformDependencies(config)

	// 添加任务配置
	content += p.generateTasks(config)

	// 写入文件
	if err := os.WriteFile(pixiTomlPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write pixi.toml: %w", err)
	}

	fmt.Printf("Created Pixi config file: %s\n", pixiTomlPath)
	return nil
}

// formatChannels 格式化 channels 列表
func (p *PixiBackend) formatChannels(channels []string) string {
	var formatted []string
	for _, channel := range channels {
		formatted = append(formatted, fmt.Sprintf(`"%s"`, channel))
	}
	return strings.Join(formatted, ", ")
}

// formatPlatforms 格式化 platforms 列表
func (p *PixiBackend) formatPlatforms(platforms []string) string {
	var formatted []string
	for _, platform := range platforms {
		formatted = append(formatted, fmt.Sprintf(`"%s"`, platform))
	}
	return strings.Join(formatted, ", ")
}

// getDefaultPlatforms 获取默认平台列表
func (p *PixiBackend) getDefaultPlatforms() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"win-64"}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return []string{"osx-arm64"}
		}
		return []string{"osx-64"}
	default:
		return []string{"linux-64"}
	}
}

// generateBuildDependencies 生成构建工具依赖
func (p *PixiBackend) generateBuildDependencies(config *VEnvConfig) string {
	var content string

	// 构建工具依赖
	content += "[build-dependencies]\n"

	if config.CPPTools.CMake.Enabled {
		content += fmt.Sprintf(`cmake = "%s"
`, config.CPPTools.CMake.Version)
	}

	if config.CPPTools.Ninja.Enabled {
		content += fmt.Sprintf(`ninja = "%s"
`, config.CPPTools.Ninja.Version)
	}

	// 编译器依赖
	if p.platform.OS == "linux" {
		if config.CPPTools.GCC.Enabled {
			content += `gcc_linux-64 = ">=11"
gxx_linux-64 = ">=11"
`
		}
		if config.CPPTools.Clang.Enabled {
			content += `clang_linux-64 = ">=11"
clangxx_linux-64 = ">=11"
`
		}
	} else if p.platform.OS == "darwin" {
		if config.CPPTools.Clang.Enabled {
			content += `clang_osx-64 = ">=11"
clangxx_osx-64 = ">=11"
`
		}
	}

	return content
}

// generatePlatformDependencies 生成平台特定依赖
func (p *PixiBackend) generatePlatformDependencies(config *VEnvConfig) string {
	var content string

	// Linux 特定依赖
	if p.platform.OS == "linux" {
		content += `
[target.linux-64.dependencies]
`
		if config.CPPTools.GCC.Enabled {
			content += `gcc_linux-64 = ">=11"
gxx_linux-64 = ">=11"
`
		}
	}

	// macOS 特定依赖
	if p.platform.OS == "darwin" {
		content += `
[target.osx-64.dependencies]
clang_osx-64 = ">=11"
clangxx_osx-64 = ">=11"
`
		if runtime.GOARCH == "arm64" {
			content += `
[target.osx-arm64.dependencies]
clang_osx-arm64 = ">=11"
clangxx_osx-arm64 = ">=11"
`
		}
	}

	// Windows 特定依赖
	if p.platform.OS == "windows" {
		content += `
[target.win-64.dependencies]
vs2019_win-64 = "*"
`
	}

	return content
}

// generateTasks 生成任务配置
func (p *PixiBackend) generateTasks(config *VEnvConfig) string {
	return `
[tasks]
configure = "cmake -B build ."
build = "cmake --build build"
install = "cmake --install build"
test = "cd build && ctest"
clean = "rm -rf build"

[activation]
scripts = ["setup_env.sh"]
`
}

// initPixiProject 初始化 Pixi 项目
func (p *PixiBackend) initPixiProject() error {
	if p.info == nil || !p.info.Installed {
		return fmt.Errorf("Pixi is not installed")
	}

	// 运行 pixi install
	cmd := exec.Command(p.info.Path, "install")
	cmd.Dir = p.rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pixi install: %w", err)
	}

	return nil
}

// installCPPTools 安装 C++ 工具
func (p *PixiBackend) installCPPTools() error {
	if p.info == nil || !p.info.Installed {
		return fmt.Errorf("Pixi is not installed")
	}

	fmt.Println("Installing C++ build tools with Pixi...")

	// Pixi 会根据 pixi.toml 自动安装所有依赖
	cmd := exec.Command(p.info.Path, "install")
	cmd.Dir = p.rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install C++ tools: %w", err)
	}

	fmt.Println("C++ tools installed successfully")
	return nil
}

// Activate 激活环境
func (p *PixiBackend) Activate() error {
	if p.info == nil || !p.info.Installed {
		return fmt.Errorf("Pixi is not installed")
	}

	// 设置环境变量
	os.Setenv("BUILDFLY_ENV_ROOT", p.rootDir)
	os.Setenv("BUILDFLY_BACKEND", "pixi")

	// 添加 Pixi 环境到 PATH
	pixiEnvPath := filepath.Join(p.rootDir, ".pixi", "env")
	if _, err := os.Stat(pixiEnvPath); err == nil {
		currentPath := os.Getenv("PATH")
		newPath := filepath.Join(pixiEnvPath, "bin") + string(filepath.ListSeparator) + currentPath
		os.Setenv("PATH", newPath)
	}

	fmt.Printf("Pixi environment activated: %s\n", p.rootDir)
	return nil
}

// Deactivate 停用环境
func (p *PixiBackend) Deactivate() error {
	os.Unsetenv("BUILDFLY_ENV_ROOT")
	os.Unsetenv("BUILDFLY_BACKEND")

	fmt.Println("Pixi environment deactivated")
	return nil
}

// Run 在环境中运行命令
func (p *PixiBackend) Run(args []string) error {
	if p.info == nil || !p.info.Installed {
		return fmt.Errorf("Pixi is not installed")
	}

	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	// 使用 pixi run 运行命令
	cmdArgs := []string{"run"}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(p.info.Path, cmdArgs...)
	cmd.Dir = p.rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// GetInfo 获取后端信息
func (p *PixiBackend) GetInfo() *BackendInfo {
	if p.info == nil {
		info, _ := p.Detect()
		return info
	}

	return &BackendInfo{
		Name:      "pixi",
		Version:   p.info.Version,
		Path:      p.info.Path,
		Installed: p.info.Installed,
		Environment: map[string]string{
			"PIXI_HOME":    os.Getenv("PIXI_HOME"),
			"PIXI_VERSION": os.Getenv("PIXI_VERSION"),
		},
	}
}

// InstallTool 安装指定工具
func (p *PixiBackend) InstallTool(toolName, version string) error {
	if p.info == nil || !p.info.Installed {
		return fmt.Errorf("Pixi is not installed")
	}

	// 添加到 pixi.toml 并重新安装
	// 这里简化处理，实际应该解析和修改配置文件
	pkgSpec := toolName
	if version != "" {
		pkgSpec = fmt.Sprintf("%s=%s", toolName, version)
	}

	cmd := exec.Command(p.info.Path, "add", pkgSpec)
	cmd.Dir = p.rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install tool %s: %w", toolName, err)
	}

	return nil
}

// getPlatformInfo 获取平台信息
func getPlatformInfo() *PlatformInfo {
	return &PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}
