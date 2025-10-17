package venv

import "os/exec"

// VEnvBackend 虚拟环境后端接口
type VEnvBackend interface {
	// 检测工具是否已安装
	Detect() (*BackendInfo, error)

	// 安装工具
	Install(version string) error

	// 初始化环境
	Initialize(config *VEnvConfig, rootDir string) error

	// 激活环境
	Activate() error

	// 停用环境
	Deactivate() error

	// 运行命令
	Run(args []string) error

	// 获取环境信息
	GetInfo() *BackendInfo

	// 创建配置文件
	CreateConfig(config *VEnvConfig, rootDir string) error

	// 安装工具
	InstallTool(toolName, version string) error
}

// BackendInfo 后端信息
type BackendInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Path        string            `json:"path"`
	Installed   bool              `json:"installed"`
	Environment map[string]string `json:"environment"`
}

// BackendFactory 后端工厂
type BackendFactory struct{}

// NewBackendFactory 创建后端工厂
func NewBackendFactory() *BackendFactory {
	return &BackendFactory{}
}

// CreateBackend 创建后端实例
func (f *BackendFactory) CreateBackend(backendType BackendType) (VEnvBackend, error) {
	switch backendType {
	case BackendUV:
		return NewUVBackend(), nil
	case BackendPixi:
		return NewPixiBackend(), nil
	default:
		// 默认使用 Pixi
		return NewPixiBackend(), nil
	}
}

// DetectAvailableBackends 检测可用的后端
func (f *BackendFactory) DetectAvailableBackends() []BackendType {
	var available []BackendType

	// 检测 UV
	uvBackend := NewUVBackend()
	if uvInfo, err := uvBackend.Detect(); err == nil && uvInfo.Installed {
		available = append(available, BackendUV)
	}

	// 检测 Pixi
	pixiBackend := NewPixiBackend()
	if pixiInfo, err := pixiBackend.Detect(); err == nil && pixiInfo.Installed {
		available = append(available, BackendPixi)
	}

	return available
}

// GetBestBackend 获取最佳后端（优先 Pixi）
func (f *BackendFactory) GetBestBackend() BackendType {
	available := f.DetectAvailableBackends()

	// 优先选择 Pixi
	for _, backend := range available {
		if backend == BackendPixi {
			return BackendPixi
		}
	}

	// 如果 Pixi 不可用，选择 UV
	for _, backend := range available {
		if backend == BackendUV {
			return BackendUV
		}
	}

	// 如果都不可用，默认返回 Pixi（会触发自动安装）
	return BackendPixi
}

// runCommand 执行命令的辅助函数
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// runCommandWithOutput 执行命令并获取输出
func runCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
