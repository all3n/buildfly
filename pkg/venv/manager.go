package venv

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Manager 虚拟环境管理器
type Manager struct {
	config      *VEnvConfig
	rootDir     string
	factory     *BackendFactory
	backend     VEnvBackend
	backendType BackendType
	platform    *PlatformInfo
}

// NewManager 创建新的虚拟环境管理器
func NewManager(config *VEnvConfig, projectRoot string) (*Manager, error) {
	if config == nil {
		config = &VEnvConfig{}
	}

	// 设置根目录
	rootDir := config.RootDir
	if rootDir == "" {
		rootDir = filepath.Join(projectRoot, ".buildfly", "root")
	}

	// 确保路径是绝对路径
	if !filepath.IsAbs(rootDir) {
		rootDir, _ = filepath.Abs(rootDir)
	}

	factory := NewBackendFactory()

	// 确定使用的 backend
	backendType := config.Backend
	if backendType == "" {
		// 默认使用 Pixi
		backendType = BackendPixi
	}

	// 创建 backend 实例
	backend, err := factory.CreateBackend(backendType)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend %s: %w", backendType, err)
	}

	return &Manager{
		config:      config,
		rootDir:     rootDir,
		factory:     factory,
		backend:     backend,
		backendType: backendType,
		platform:    getPlatformInfo(),
	}, nil
}

// NewManagerWithBackend 使用指定 backend 创建管理器
func NewManagerWithBackend(config *VEnvConfig, projectRoot string, backendType BackendType) (*Manager, error) {
	if config == nil {
		config = &VEnvConfig{}
	}

	// 设置根目录
	rootDir := config.RootDir
	if rootDir == "" {
		rootDir = filepath.Join(projectRoot, ".buildfly", "root")
	}

	// 确保路径是绝对路径
	if !filepath.IsAbs(rootDir) {
		rootDir, _ = filepath.Abs(rootDir)
	}

	factory := NewBackendFactory()

	// 创建 backend 实例
	backend, err := factory.CreateBackend(backendType)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend %s: %w", backendType, err)
	}

	return &Manager{
		config:      config,
		rootDir:     rootDir,
		factory:     factory,
		backend:     backend,
		backendType: backendType,
		platform:    getPlatformInfo(),
	}, nil
}

// GetBackendType 获取当前使用的 backend 类型
func (m *Manager) GetBackendType() BackendType {
	return m.backendType
}

// SwitchBackend 切换 backend
func (m *Manager) SwitchBackend(backendType BackendType) error {
	backend, err := m.factory.CreateBackend(backendType)
	if err != nil {
		return fmt.Errorf("failed to create backend %s: %w", backendType, err)
	}

	m.backend = backend
	m.backendType = backendType

	// 更新配置
	m.config.Backend = backendType

	return nil
}

// Initialize 初始化虚拟环境
func (m *Manager) Initialize() error {
	fmt.Printf("Initializing C++ virtual environment with %s backend at: %s\n", m.backendType, m.rootDir)

	// 使用 backend 初始化环境
	if err := m.backend.Initialize(m.config, m.rootDir); err != nil {
		return fmt.Errorf("failed to initialize environment: %w", err)
	}

	if err := m.createEnvironmentInfo(); err != nil {
		return fmt.Errorf("failed to create environment info: %w", err)
	}

	fmt.Println("Virtual environment initialized successfully")
	return nil
}

// ensureBackendInstalled 确保 backend 已安装
func (m *Manager) ensureBackendInstalled() error {
	backendInfo := m.backend.GetInfo()

	if !backendInfo.Installed {
		fmt.Printf("%s not found, installing...\n", m.backendType)

		// 根据 backend 类型获取版本
		var version string
		switch m.backendType {
		case BackendUV:
			version = m.config.UVVersion
		case BackendPixi:
			version = m.config.Pixi.Version
		}

		if err := m.backend.Install(version); err != nil {
			return fmt.Errorf("failed to install %s: %w", m.backendType, err)
		}
		fmt.Printf("%s installed successfully\n", m.backendType)
	} else {
		fmt.Printf("%s already installed: %s (version: %s)\n", m.backendType, backendInfo.Path, backendInfo.Version)
	}

	return nil
}

// createEnvironmentInfo 创建环境信息文件
func (m *Manager) createEnvironmentInfo() error {
	backendInfo := m.backend.GetInfo()

	info := &EnvironmentInfo{
		RootDir:        m.rootDir,
		Activated:      false,
		UVPath:         backendInfo.Path,
		UVInstalled:    backendInfo.Installed,
		UVVersion:      backendInfo.Version,
		PythonVersion:  m.config.Python.Version,
		InstalledTools: make(map[string]string),
		CreatedAt:      time.Now(),
	}

	infoFile := filepath.Join(m.rootDir, "environment.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment info: %w", err)
	}

	if err := os.WriteFile(infoFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment info: %w", err)
	}

	return nil
}

// createActivationScripts 创建环境激活脚本
func (m *Manager) createActivationScripts() error {
	// Unix/Linux/macOS 激活脚本
	unixScript := fmt.Sprintf(`#!/bin/bash
# BuildFly C++ Environment Activation Script

export BUILDFLY_ENV_ROOT="%s"
export PATH="%s/bin:$PATH"

# 设置编译器环境变量
if [ -f "%s/bin/gcc" ]; then
    export CC="%s/bin/gcc"
    export CXX="%s/bin/g++"
fi

if [ -f "%s/bin/clang" ]; then
    export CC="%s/bin/clang"
    export CXX="%s/bin/clang++"
fi

# 设置 CMake 环境变量
if [ -d "%s/tools/cmake" ]; then
    export CMAKE_PREFIX_PATH="%s/tools/cmake:$CMAKE_PREFIX_PATH"
    export PATH="%s/tools/cmake/bin:$PATH"
fi

# 设置 Python 环境
if [ -d "%s/env" ]; then
    export PYTHONPATH="%s/env/lib/python*/site-packages:$PYTHONPATH"
fi

echo "BuildFly C++ environment activated"
echo "Environment root: $BUILDFLY_ENV_ROOT"
`, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir)

	unixScriptFile := filepath.Join(m.rootDir, "activate.sh")
	if err := os.WriteFile(unixScriptFile, []byte(unixScript), 0755); err != nil {
		return fmt.Errorf("failed to write Unix activation script: %w", err)
	}

	// Windows 激活脚本
	windowsScript := fmt.Sprintf(`@echo off
REM BuildFly C++ Environment Activation Script

set BUILDFLY_ENV_ROOT=%s
set PATH=%s\bin;%%PATH%%

REM 设置编译器环境变量
if exist "%s\bin\gcc.exe" (
    set CC=%s\bin\gcc.exe
    set CXX=%s\bin\g++.exe
)

if exist "%s\bin\clang.exe" (
    set CC=%s\bin\clang.exe
    set CXX=%s\bin\clang++.exe
)

REM 设置 CMake 环境变量
if exist "%s\tools\cmake" (
    set CMAKE_PREFIX_PATH=%s\tools\cmake;%%CMAKE_PREFIX_PATH%%
    set PATH=%s\tools\cmake\bin;%%PATH%%
)

echo BuildFly C++ environment activated
echo Environment root: %%BUILDFLY_ENV_ROOT%%
`, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir, m.rootDir)

	windowsScriptFile := filepath.Join(m.rootDir, "activate.bat")
	if err := os.WriteFile(windowsScriptFile, []byte(windowsScript), 0755); err != nil {
		return fmt.Errorf("failed to write Windows activation script: %w", err)
	}

	return nil
}

// Activate 激活虚拟环境
func (m *Manager) Activate() error {
	// 检查环境是否已初始化
	if !m.isInitialized() {
		return fmt.Errorf("virtual environment not initialized. Run 'buildfly venv init' first")
	}

	// 更新环境信息
	if err := m.updateActivationStatus(true); err != nil {
		return fmt.Errorf("failed to update activation status: %w", err)
	}

	// 设置环境变量
	if err := m.setEnvironmentVariables(); err != nil {
		return fmt.Errorf("failed to set environment variables: %w", err)
	}

	fmt.Printf("Virtual environment activated: %s\n", m.rootDir)
	return nil
}

// Deactivate 停用虚拟环境
func (m *Manager) Deactivate() error {
	// 更新环境信息
	if err := m.updateActivationStatus(false); err != nil {
		return fmt.Errorf("failed to update activation status: %w", err)
	}

	// 清理环境变量
	m.clearEnvironmentVariables()

	fmt.Println("Virtual environment deactivated")
	return nil
}

// isInitialized 检查环境是否已初始化
func (m *Manager) isInitialized() bool {
	infoFile := filepath.Join(m.rootDir, "environment.json")
	if _, err := os.Stat(infoFile); err != nil {
		return false
	}
	return true
}

// updateActivationStatus 更新激活状态
func (m *Manager) updateActivationStatus(activated bool) error {
	infoFile := filepath.Join(m.rootDir, "environment.json")
	data, err := os.ReadFile(infoFile)
	if err != nil {
		return fmt.Errorf("failed to read environment info: %w", err)
	}

	var info EnvironmentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("failed to unmarshal environment info: %w", err)
	}

	info.Activated = activated
	if activated {
		info.LastActivated = time.Now()
	}

	data, err = json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment info: %w", err)
	}

	if err := os.WriteFile(infoFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment info: %w", err)
	}

	return nil
}

// setEnvironmentVariables 设置环境变量
func (m *Manager) setEnvironmentVariables() error {
	os.Setenv("BUILDFLY_ENV_ROOT", m.rootDir)

	// 添加 bin 目录到 PATH
	binDir := filepath.Join(m.rootDir, "bin")
	currentPath := os.Getenv("PATH")
	newPath := binDir + string(filepath.ListSeparator) + currentPath
	os.Setenv("PATH", newPath)

	return nil
}

// clearEnvironmentVariables 清理环境变量
func (m *Manager) clearEnvironmentVariables() {
	os.Unsetenv("BUILDFLY_ENV_ROOT")
	// 注意：这里不完全清理 PATH，因为这可能影响其他程序
}

// GetEnvironmentInfo 获取环境信息
func (m *Manager) GetEnvironmentInfo() (*EnvironmentInfo, error) {
	if !m.isInitialized() {
		return &EnvironmentInfo{
			RootDir:   m.rootDir,
			Activated: false,
		}, nil
	}

	infoFile := filepath.Join(m.rootDir, "environment.json")
	data, err := os.ReadFile(infoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment info: %w", err)
	}

	var info EnvironmentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment info: %w", err)
	}

	return &info, nil
}

// Reset 重置虚拟环境
func (m *Manager) Reset() error {
	if !m.isInitialized() {
		return fmt.Errorf("virtual environment not initialized")
	}

	fmt.Printf("Resetting virtual environment: %s\n", m.rootDir)

	// 删除整个环境目录
	if err := os.RemoveAll(m.rootDir); err != nil {
		return fmt.Errorf("failed to remove environment directory: %w", err)
	}

	fmt.Println("Virtual environment reset successfully")
	return nil
}

// GetRootDir 获取环境根目录
func (m *Manager) GetRootDir() string {
	return m.rootDir
}

// GetActivationScript 获取激活脚本路径
func (m *Manager) GetActivationScript() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(m.rootDir, "activate.bat")
	default:
		return filepath.Join(m.rootDir, "activate.sh")
	}
}

// IsActivated 检查环境是否已激活
func (m *Manager) IsActivated() bool {
	envRoot := os.Getenv("BUILDFLY_ENV_ROOT")
	return envRoot == m.rootDir
}

// GetEnvironmentVars 获取虚拟环境的环境变量
func (m *Manager) GetEnvironmentVars() (map[string]string, error) {
	vars := make(map[string]string)

	if m.IsActivated() {
		vars["BUILDFLY_ENV_ROOT"] = m.rootDir
		vars["PATH"] = filepath.Join(m.rootDir, "bin") + string(filepath.ListSeparator) + os.Getenv("PATH")

		// 添加其他可能的环境变量
		binDir := filepath.Join(m.rootDir, "bin")
		if _, err := os.Stat(filepath.Join(binDir, "gcc")); err == nil {
			vars["CC"] = filepath.Join(binDir, "gcc")
			vars["CXX"] = filepath.Join(binDir, "g++")
		}
		if _, err := os.Stat(filepath.Join(binDir, "clang")); err == nil {
			vars["CC"] = filepath.Join(binDir, "clang")
			vars["CXX"] = filepath.Join(binDir, "clang++")
		}
	}

	return vars, nil
}

// IsToolInstalled 检查工具是否安装在虚拟环境中
func (m *Manager) IsToolInstalled(toolName string) bool {
	if !m.IsActivated() {
		return false
	}

	binDir := filepath.Join(m.rootDir, "bin")
	toolPath := filepath.Join(binDir, toolName)

	// 检查工具是否存在
	if _, err := os.Stat(toolPath); err == nil {
		return true
	}

	// 检查带扩展名的工具（Windows）
	if runtime.GOOS == "windows" {
		extToolPath := toolPath + ".exe"
		if _, err := os.Stat(extToolPath); err == nil {
			return true
		}
	}

	return false
}

// Run 在虚拟环境中运行命令
func (m *Manager) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	// 如果环境未初始化，自动初始化
	if !m.isInitialized() {
		fmt.Println("Virtual environment not initialized, creating...")
		if err := m.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize virtual environment: %w", err)
		}
	}

	// 使用 backend 运行命令
	return m.backend.Run(args)
}

// joinArgs 将字符串数组连接成适合在 bash 中执行的命令字符串
func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}

	var escapedArgs []string
	for _, arg := range args {
		// 对包含空格、引号等特殊字符的参数进行转义
		if strings.ContainsAny(arg, " \t\n\r\"'`$&*()[]{}|;<>?") {
			escapedArgs = append(escapedArgs, fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\"'\"'")))
		} else {
			escapedArgs = append(escapedArgs, arg)
		}
	}

	return strings.Join(escapedArgs, " ")
}

// GetStatus 获取虚拟环境状态（为了向后兼容）
func (m *Manager) GetStatus() *VEnvStatus {
	info, err := m.GetEnvironmentInfo()
	if err != nil {
		return &VEnvStatus{
			RootDir:        m.rootDir,
			Activated:      false,
			UVInstalled:    false,
			UVVersion:      "",
			PixiInstalled:  false,
			PixiVersion:    "",
			PythonVersion:  "",
			InstalledTools: map[string]string{},
		}
	}

	status := &VEnvStatus{
		RootDir:        info.RootDir,
		Activated:      info.Activated,
		PythonVersion:  info.PythonVersion,
		InstalledTools: info.InstalledTools,
		CreatedAt:      info.CreatedAt,
		LastActivated:  info.LastActivated,
	}

	// 根据 backend 类型设置相应的字段
	if m.backendType == BackendUV {
		status.UVInstalled = info.UVInstalled
		status.UVVersion = info.UVVersion
		status.PixiInstalled = false
		status.PixiVersion = ""
	} else if m.backendType == BackendPixi {
		// 对于 Pixi backend，使用 backend 的 info
		backendInfo := m.backend.GetInfo()
		status.PixiInstalled = backendInfo.Installed
		status.PixiVersion = backendInfo.Version
		status.UVInstalled = false
		status.UVVersion = ""
	}

	return status
}

// installTool 安装工具的通用方法
func (m *Manager) installTool(toolName, version string) error {
	return m.backend.InstallTool(toolName, version)
}

// installGCC 安装GCC（仅在Linux/macOS上）
func (m *Manager) installGCC() error {
	fmt.Printf("Installing GCC %s...\n", m.config.CPPTools.GCC.Version)

	// 这里简化处理，实际实现可能需要根据平台使用不同的安装方法
	// 例如：在Ubuntu上使用apt，在macOS上使用brew，或者从源码编译

	switch m.platform.OS {
	case "linux":
		if m.platform.PackageManager == "apt" {
			cmd := exec.Command("sudo", "apt", "update")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update package list: %w", err)
			}

			cmd = exec.Command("sudo", "apt", "install", "build-essential")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install build-essential: %w", err)
			}
		}
	case "darwin":
		if m.platform.PackageManager == "brew" {
			cmd := exec.Command("brew", "install", "gcc")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install gcc via brew: %w", err)
			}
		}
	}

	// 更新工具信息
	m.updateToolInfo("gcc", m.config.CPPTools.GCC.Version)

	fmt.Println("GCC installed successfully")
	return nil
}

// installClang 安装Clang
func (m *Manager) installClang() error {
	fmt.Printf("Installing Clang %s...\n", m.config.CPPTools.Clang.Version)

	switch m.platform.OS {
	case "linux":
		if m.platform.PackageManager == "apt" {
			cmd := exec.Command("sudo", "apt", "install", "-y", "clang")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install clang: %w", err)
			}
		}
	case "darwin":
		// macOS通常已经预装了Clang
		fmt.Println("Clang is typically pre-installed on macOS")
	}

	// 更新工具信息
	m.updateToolInfo("clang", m.config.CPPTools.Clang.Version)

	fmt.Println("Clang installed successfully")
	return nil
}

// DetectAvailableBackends 检测可用的后端
func (m *Manager) DetectAvailableBackends() []BackendType {
	return m.factory.DetectAvailableBackends()
}

// GetBestBackend 获取最佳后端
func (m *Manager) GetBestBackend() BackendType {
	return m.factory.GetBestBackend()
}

// updateToolInfo 更新工具信息
func (m *Manager) updateToolInfo(toolName, version string) {
	infoFile := filepath.Join(m.rootDir, "environment.json")
	data, err := os.ReadFile(infoFile)
	if err != nil {
		return
	}

	var info EnvironmentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return
	}

	if info.InstalledTools == nil {
		info.InstalledTools = make(map[string]string)
	}
	info.InstalledTools[toolName] = version

	data, _ = json.MarshalIndent(info, "", "  ")
	os.WriteFile(infoFile, data, 0644)
}
