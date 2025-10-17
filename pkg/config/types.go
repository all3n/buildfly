package config

import (
	"buildfly/pkg/venv"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectConfig 项目配置结构
type ProjectConfig struct {
	Project      Project               `yaml:"project"`
	Dependencies map[string]Dependency `yaml:"dependencies"`

	ProjectRoot     string `yaml:"-"`
	BuildFlyBaseDir string `yaml:"buildfly_base_dir,omitempty"`
	InstallDir      string `yaml:"install_dir,omitempty"`
	BuildDir        string `yaml:"build_dir,omitempty"`
	CacheDir        string `yaml:"cache_dir,omitempty"`

	BuildProfiles map[string]BuildProfile `yaml:"build_profiles,omitempty"`

	// BuildTag 当前项目的构建标签
	BuildTag *BuildTag `yaml:"build_tag,omitempty"`

	// Proxy 代理配置
	Proxy *ProxyConfig `yaml:"proxy,omitempty"`

	// VEnv 虚拟环境配置
	VEnv *venv.VEnvConfig `yaml:"venv,omitempty"`
}

// 项目信息
type Project struct {
	Name      string            `yaml:"name"`
	Version   string            `yaml:"version"`
	Variables map[string]string `yaml:"variables,omitempty"`
}

// 依赖项配置
type Dependency struct {
	Name             string            `yaml:"-"`
	Version          string            `yaml:"version"`
	Source           SourceInfo        `yaml:"source"`
	BuildSystem      string            `yaml:"build_system"` // make, cmake, configure, custom
	CMakeOptions     []string          `yaml:"cmake_options,omitempty"`
	MakeOptions      []string          `yaml:"make_options,omitempty"`
	ConfigureOptions []string          `yaml:"configure_options,omitempty"`
	CustomScript     string            `yaml:"custom_script,omitempty"`
	BuildCommands    BuildCommands     `yaml:"build_commands,omitempty"`
	EnvVariables     map[string]string `yaml:"env_variables,omitempty"`
	CacheKey         string            `yaml:"-"` // 缓存键，自动生成
	LastUpdated      time.Time         `yaml:"-"` // 最后更新时间
}

// 源码信息
type SourceInfo struct {
	Type      string            `yaml:"type"` // git, archive, direct
	URLS      []string          `yaml:"urls"`
	Tag       string            `yaml:"tag,omitempty"`
	Hash      string            `yaml:"hash,omitempty"`      // SHA256 哈希值（向后兼容）
	MD5       string            `yaml:"md5,omitempty"`       // MD5 哈希值
	SHA1      string            `yaml:"sha1,omitempty"`      // SHA1 哈希值
	SHA256    string            `yaml:"sha256,omitempty"`    // SHA256 哈希值
	SHA512    string            `yaml:"sha512,omitempty"`    // SHA512 哈希值
	Checksums map[string]string `yaml:"checksums,omitempty"` // 通用校验和映射
}

// 构建命令
type BuildCommands struct {
	Configure string `yaml:"configure,omitempty"`
	Build     string `yaml:"build,omitempty"`
	Install   string `yaml:"install,omitempty"`
	Test      string `yaml:"test,omitempty"`
}

// 构建配置文件
type BuildProfile struct {
	Variables    map[string]string `yaml:"variables,omitempty"`
	Dependencies []string          `yaml:"dependencies"`
	BuildTag     *BuildTag         `yaml:"build_tag,omitempty"`
}

// 解析后的依赖项
type ResolvedDependency struct {
	Dependency
	InstallPath string   `yaml:"-"`
	IncludeDirs []string `yaml:"-"`
	LibDirs     []string `yaml:"-"`
	Libs        []string `yaml:"-"`
}

// 依赖冲突信息
type Conflict struct {
	Dependency string   `yaml:"dependency"`
	Versions   []string `yaml:"versions"`
	Reason     string   `yaml:"reason"`
}

// 解析结果
type ResolutionResult struct {
	Dependencies []ResolvedDependency `yaml:"dependencies"`
	Conflicts    []Conflict           `yaml:"conflicts"`
	DownloadSize int64                `yaml:"download_size"`
	BuildTime    time.Duration        `yaml:"build_time"`
}

// GetURLs 获取所有 URL，支持向后兼容
func (s *SourceInfo) GetURLs() []string {
	if len(s.URLS) > 0 {
		return s.URLS
	}

	// 向后兼容：如果 URLS 为空，尝试从其他字段获取
	// 这里可以添加从旧字段转换的逻辑
	return []string{}
}

// GetFirstAvailableURL 获取第一个可用的 URL
// 优先级：本地文件 > 网络文件
func (s *SourceInfo) GetFirstAvailableURL() (string, error) {
	urls := s.GetURLs()
	if len(urls) == 0 {
		return "", &SourceError{Message: "no URLs available"}
	}

	// 分离本地文件和网络 URL
	var localURLs []string
	var networkURLs []string

	for _, url := range urls {
		if s.IsLocalURL(url) {
			localURLs = append(localURLs, url)
		} else {
			networkURLs = append(networkURLs, url)
		}
	}

	// 优先尝试本地文件
	for _, url := range localURLs {
		if s.fileExists(url) {
			return url, nil
		}
	}

	// 如果本地文件不存在，尝试网络 URL
	if len(networkURLs) > 0 {
		return networkURLs[0], nil
	}

	// 如果都没有可用的，返回第一个本地 URL（即使不存在）
	if len(localURLs) > 0 {
		return localURLs[0], nil
	}

	return urls[0], nil
}

// IsLocalURL 检查是否为本地 URL
func (s *SourceInfo) IsLocalURL(url string) bool {
	return !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "git@") &&
		!strings.Contains(url, "://")
}

// fileExists 检查文件是否存在
func (s *SourceInfo) fileExists(path string) bool {
	// 处理相对路径
	if !filepath.IsAbs(path) {
		// 这里可以添加当前工作目录的逻辑
		// 暂时使用相对路径检查
	}

	_, err := os.Stat(path)
	return err == nil
}

// GetAllAvailableURLs 按优先级获取所有可用的 URL
// 本地文件在前，网络 URL 在后
func (s *SourceInfo) GetAllAvailableURLs() []string {
	urls := s.GetURLs()
	if len(urls) == 0 {
		return []string{}
	}

	var localURLs []string
	var networkURLs []string

	for _, url := range urls {
		if s.IsLocalURL(url) {
			localURLs = append(localURLs, url)
		} else {
			networkURLs = append(networkURLs, url)
		}
	}

	// 本地文件优先，然后是网络 URL
	result := append(localURLs, networkURLs...)
	return result
}

// HasURLs 检查是否有可用的 URL
func (s *SourceInfo) HasURLs() bool {
	urls := s.GetURLs()
	return len(urls) > 0
}

// GetPrimaryURL 获取第一个 URL
func (s *SourceInfo) GetPrimaryURL() string {
	urls := s.GetURLs()
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	HTTP    string   `yaml:"http" json:"http"`
	HTTPS   string   `yaml:"https" json:"https"`
	NoProxy []string `yaml:"no_proxy" json:"no_proxy"`
}

// SourceError 源码相关错误
type SourceError struct {
	Message string
}

func (e *SourceError) Error() string {
	return e.Message
}
