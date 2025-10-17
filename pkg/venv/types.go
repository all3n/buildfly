package venv

import "time"

// BackendType 虚拟环境后端类型
type BackendType string

const (
	BackendUV   BackendType = "uv"
	BackendPixi BackendType = "pixi"
)

// VEnvConfig 虚拟环境配置
type VEnvConfig struct {
	Enabled      bool           `yaml:"enabled"`
	Backend      BackendType    `yaml:"backend"`              // 默认为 pixi
	UVVersion    string         `yaml:"uv_version,omitempty"` // 向后兼容
	RootDir      string         `yaml:"root_dir"`
	AutoActivate bool           `yaml:"auto_activate"`
	CPPTools     CPPToolsConfig `yaml:"cpp_tools"`
	Python       PythonConfig   `yaml:"python"`
	UV           UVConfig       `yaml:"uv,omitempty"`
	Pixi         PixiConfig     `yaml:"pixi,omitempty"`
}

// CPPToolsConfig C++ 工具配置
type CPPToolsConfig struct {
	CMake ToolConfig `yaml:"cmake"`
	Ninja ToolConfig `yaml:"ninja"`
	GCC   ToolConfig `yaml:"gcc,omitempty"`
	Clang ToolConfig `yaml:"clang,omitempty"`
	MSVC  ToolConfig `yaml:"msvc,omitempty"`
}

// ToolConfig 工具配置
type ToolConfig struct {
	Version string `yaml:"version"`
	Enabled bool   `yaml:"enabled"`
}

// PythonConfig Python 配置
type PythonConfig struct {
	Version  string   `yaml:"version"`
	Packages []string `yaml:"packages"`
}

// EnvironmentInfo 环境信息
type EnvironmentInfo struct {
	RootDir        string            `json:"root_dir"`
	Activated      bool              `json:"activated"`
	UVInstalled    bool              `json:"uv_installed"`
	UVVersion      string            `json:"uv_version"`
	UVPath         string            `json:"uv_path"`
	PythonVersion  string            `json:"python_version"`
	InstalledTools map[string]string `json:"installed_tools"`
	CreatedAt      time.Time         `json:"created_at"`
	LastActivated  time.Time         `json:"last_activated"`
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Installed   bool      `json:"installed"`
	Path        string    `json:"path"`
	InstalledAt time.Time `json:"installed_at"`
}

// UVInfo UV 信息
type UVInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Path      string `json:"path"`
}

// PlatformInfo 平台信息
type PlatformInfo struct {
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	Platform       string `json:"platform"`
	Compiler       string `json:"compiler"`
	PackageManager string `json:"package_manager"`
}

// UVConfig UV 配置
type UVConfig struct {
	IndexURL      string            `yaml:"index_url,omitempty"`
	ExtraIndexURL []string          `yaml:"extra_index_url,omitempty"`
	TrustedHosts  []string          `yaml:"trusted_hosts,omitempty"`
	Timeout       int               `yaml:"timeout,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
}

// PixiConfig Pixi 配置
type PixiConfig struct {
	Version     string            `yaml:"version,omitempty"`
	Channels    []string          `yaml:"channels,omitempty"`
	Platforms   []string          `yaml:"platforms,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// PixiInfo Pixi 信息
type PixiInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Path      string `json:"path"`
}

// VEnvStatus 虚拟环境状态（向后兼容）
type VEnvStatus struct {
	RootDir        string            `json:"root_dir"`
	Activated      bool              `json:"activated"`
	UVInstalled    bool              `json:"uv_installed"`
	UVVersion      string            `json:"uv_version"`
	PixiInstalled  bool              `json:"pixi_installed"`
	PixiVersion    string            `json:"pixi_version"`
	PythonVersion  string            `json:"python_version"`
	InstalledTools map[string]string `json:"installed_tools"`
	CreatedAt      time.Time         `json:"created_at"`
	LastActivated  time.Time         `json:"last_activated"`
}
