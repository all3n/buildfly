package venv

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// UVDetector UV 检测器
type UVDetector struct{}

// NewUVDetector 创建新的 UV 检测器
func NewUVDetector() *UVDetector {
	return &UVDetector{}
}

// DetectUV 检测 UV 是否已安装
func (d *UVDetector) DetectUV() (*UVInfo, error) {
	// 检查 UV 是否在 PATH 中
	uvPath, err := exec.LookPath("uv")
	if err != nil {
		return &UVInfo{
			Installed: false,
		}, nil
	}

	// 获取 UV 版本
	version, err := d.getUVVersion(uvPath)
	if err != nil {
		return &UVInfo{
			Installed: true,
			Path:      uvPath,
		}, nil
	}

	return &UVInfo{
		Installed: true,
		Version:   version,
		Path:      uvPath,
	}, nil
}

// getUVVersion 获取 UV 版本
func (d *UVDetector) getUVVersion(uvPath string) (string, error) {
	cmd := exec.Command(uvPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 解析版本信息，通常输出格式为: uv 0.1.0
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("unable to parse uv version")
}

// InstallUV 安装 UV
func (d *UVDetector) InstallUV(version string) error {
	switch runtime.GOOS {
	case "windows":
		return d.installUVWindows(version)
	case "linux", "darwin":
		return d.installUVUnix(version)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// installUVUnix 在 Unix 系统 (Linux/macOS) 上安装 UV
func (d *UVDetector) installUVUnix(version string) error {
	var installScript string
	var cmd *exec.Cmd

	if version == "" || version == "latest" {
		// 安装最新版本
		installScript = "curl -LsSf https://astral.sh/uv/install.sh | sh"
		cmd = exec.Command("sh", "-c", installScript)
	} else {
		// 安装指定版本
		installScript = fmt.Sprintf("curl -LsSf https://astral.sh/uv/%s/install.sh | sh", version)
		cmd = exec.Command("sh", "-c", installScript)
	}

	// 设置环境变量
	cmd.Env = append(os.Environ(), "HOME="+os.Getenv("HOME"))

	// 执行安装
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install uv: %w\nOutput: %s", err, string(output))
	}

	// 检查是否需要更新 PATH
	if err := d.ensureUVInPathUnix(); err != nil {
		return fmt.Errorf("uv installed but PATH update failed: %w", err)
	}

	return nil
}

// installUVWindows 在 Windows 上安装 UV
func (d *UVDetector) installUVWindows(version string) error {
	var installScript string
	var cmd *exec.Cmd

	if version == "" || version == "latest" {
		// 安装最新版本
		installScript = "powershell -ExecutionPolicy ByPass -c \"irm https://astral.sh/uv/install.ps1 | iex\""
		cmd = exec.Command("powershell", "-Command", installScript)
	} else {
		// 安装指定版本
		installScript = fmt.Sprintf("powershell -ExecutionPolicy ByPass -c \"irm https://astral.sh/uv/%s/install.ps1 | iex\"", version)
		cmd = exec.Command("powershell", "-Command", installScript)
	}

	// 执行安装
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install uv: %w\nOutput: %s", err, string(output))
	}

	// 检查是否需要更新 PATH
	if err := d.ensureUVInPathWindows(); err != nil {
		return fmt.Errorf("uv installed but PATH update failed: %w", err)
	}

	return nil
}

// ensureUVInPathUnix 确保 UV 在 Unix 系统 PATH 中
func (d *UVDetector) ensureUVInPathUnix() error {
	// 检查 UV 是否已经在 PATH 中
	_, err := exec.LookPath("uv")
	if err == nil {
		return nil // UV 已经在 PATH 中
	}

	// 常见的 UV 安装路径
	possiblePaths := []string{
		fmt.Sprintf("%s/.cargo/bin", os.Getenv("HOME")),
		fmt.Sprintf("%s/.local/bin", os.Getenv("HOME")),
		"/usr/local/bin",
	}

	for _, path := range possiblePaths {
		uvPath := fmt.Sprintf("%s/uv", path)
		if _, err := os.Stat(uvPath); err == nil {
			// 找到 UV，添加到 PATH
			currentPath := os.Getenv("PATH")
			newPath := fmt.Sprintf("%s:%s", path, currentPath)
			os.Setenv("PATH", newPath)
			return nil
		}
	}

	return fmt.Errorf("uv installed but not found in standard locations")
}

// ensureUVInPathWindows 确保 UV 在 Windows PATH 中
func (d *UVDetector) ensureUVInPathWindows() error {
	// 检查 UV 是否已经在 PATH 中
	_, err := exec.LookPath("uv")
	if err == nil {
		return nil // UV 已经在 PATH 中
	}

	// 常见的 UV 安装路径
	possiblePaths := []string{
		fmt.Sprintf("%s\\.cargo\\bin", os.Getenv("USERPROFILE")),
		fmt.Sprintf("%s\\AppData\\Local\\Programs\\uv\\bin", os.Getenv("USERPROFILE")),
		"C:\\Program Files\\uv\\bin",
	}

	for _, path := range possiblePaths {
		uvPath := fmt.Sprintf("%s\\uv.exe", path)
		if _, err := os.Stat(uvPath); err == nil {
			// 找到 UV，添加到 PATH
			currentPath := os.Getenv("PATH")
			newPath := fmt.Sprintf("%s;%s", path, currentPath)
			os.Setenv("PATH", newPath)
			return nil
		}
	}

	return fmt.Errorf("uv installed but not found in standard locations")
}

// GetPlatformInfo 获取平台信息
func (d *UVDetector) GetPlatformInfo() *PlatformInfo {
	info := &PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	switch runtime.GOOS {
	case "linux":
		info.Platform = "linux"
		info.Compiler = d.detectLinuxCompiler()
		info.PackageManager = d.detectLinuxPackageManager()
	case "darwin":
		info.Platform = "darwin"
		info.Compiler = d.detectDarwinCompiler()
		info.PackageManager = "brew"
	case "windows":
		info.Platform = "windows"
		info.Compiler = "msvc"
		info.PackageManager = "choco"
	}

	return info
}

// detectLinuxCompiler 检测 Linux 编译器
func (d *UVDetector) detectLinuxCompiler() string {
	if _, err := exec.LookPath("gcc"); err == nil {
		return "gcc"
	}
	if _, err := exec.LookPath("clang"); err == nil {
		return "clang"
	}
	return "unknown"
}

// detectDarwinCompiler 检测 macOS 编译器
func (d *UVDetector) detectDarwinCompiler() string {
	if _, err := exec.LookPath("clang"); err == nil {
		return "clang"
	}
	if _, err := exec.LookPath("gcc"); err == nil {
		return "gcc"
	}
	return "unknown"
}

// detectLinuxPackageManager 检测 Linux 包管理器
func (d *UVDetector) detectLinuxPackageManager() string {
	if _, err := exec.LookPath("apt"); err == nil {
		return "apt"
	}
	if _, err := exec.LookPath("yum"); err == nil {
		return "yum"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "dnf"
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		return "pacman"
	}
	if _, err := exec.LookPath("zypper"); err == nil {
		return "zypper"
	}
	return "unknown"
}
