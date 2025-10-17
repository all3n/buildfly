package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

// VariableContext 变量上下文
type VariableContext struct {
	// 项目变量
	ProjectName    string
	ProjectVersion string
	ProjectRoot    string // 项目根目录

	// 路径变量
	InstallDir   string
	BuildDir     string
	DownloadPath string
	ExtractDir   string
	SourceDir    string

	// 构建变量
	BuildType   string
	CXXCompiler string
	CXXFlags    string

	// 系统变量
	OS       string
	Arch     string
	CPUCount int

	// 自定义变量
	CustomVars map[string]string

	// 依赖特定变量
	DepName    string
	DepVersion string

	// BuildTag 构建标签
	BuildTag *BuildTag

	// ProjectConfig 项目配置引用
	ProjectConfig *ProjectConfig
}

// NewVariableContext 创建新的变量上下文
func NewVariableContext(project Project, depName string) *VariableContext {
	return &VariableContext{
		ProjectName:    project.Name,
		ProjectVersion: project.Version,
		InstallDir:     "${INSTALL_DIR}",
		BuildDir:       "${BUILD_DIR}",
		DownloadPath:   "${DOWNLOAD_PATH}",
		ExtractDir:     "${EXTRACT_DIR}",
		SourceDir:      "${SOURCE_DIR}",
		BuildType:      "Release",
		CXXCompiler:    "g++",
		CXXFlags:       "-O2 -std=c++17",
		OS:             getOS(),
		Arch:           getArch(),
		CPUCount:       getCPUCount(),
		CustomVars:     project.Variables,
		DepName:        depName,
	}
}

// SetPaths 设置路径变量
func (vc *VariableContext) SetPaths(installDir, buildDir, downloadPath, extractDir, sourceDir string) {
	vc.InstallDir = installDir
	vc.BuildDir = buildDir
	vc.DownloadPath = downloadPath
	vc.ExtractDir = extractDir
	vc.SourceDir = sourceDir
}

// SetBuildVars 设置构建变量
func (vc *VariableContext) SetBuildVars(buildType, compiler, flags string) {
	vc.BuildType = buildType
	vc.CXXCompiler = compiler
	vc.CXXFlags = flags
}

// SetDepInfo 设置依赖信息
func (vc *VariableContext) SetDepInfo(name, version string) {
	vc.DepName = name
	vc.DepVersion = version
}

// SetBuildTag 设置构建标签
func (vc *VariableContext) SetBuildTag(buildTag *BuildTag) {
	vc.BuildTag = buildTag
}

// GetBuildTagDir 获取构建标签目录名
func (vc *VariableContext) GetBuildTagDir() string {
	if vc.BuildTag == nil {
		return "default"
	}
	return vc.BuildTag.ToDirName()
}

// GetBuildPathWithVersion 获取包含版本和构建标签的构建路径
func (vc *VariableContext) GetBuildPathWithVersion(depName, depVersion string) string {
	buildTagDir := vc.GetBuildTagDir()
	return filepath.Join(vc.BuildDir, depName, depVersion, buildTagDir)
}

// GetInstallPathWithVersion 获取包含版本和构建标签的安装路径
func (vc *VariableContext) GetInstallPathWithVersion(depName, depVersion string) string {
	buildTagDir := vc.GetBuildTagDir()
	return filepath.Join(vc.InstallDir, depName, depVersion, buildTagDir)
}

// GetCacheKey 获取包含构建标签的缓存键
func (vc *VariableContext) GetCacheKey(depName, depVersion string) string {
	if vc.BuildTag == nil {
		return fmt.Sprintf("%s:%s", depName, depVersion)
	}
	return fmt.Sprintf("%s:%s:%s", depName, depVersion, vc.BuildTag.String())
}

// SetProjectRoot 设置项目根目录
func (vc *VariableContext) SetProjectRoot(root string) {
	vc.ProjectRoot = root
}

// resolvePath 解析路径，如果是相对路径则转换为绝对路径
func (vc *VariableContext) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if vc.ProjectRoot != "" {
		return filepath.Join(vc.ProjectRoot, path)
	}
	return path
}

// ExpandVariables 展开模板变量
func (vc *VariableContext) ExpandVariables(input string) (string, error) {
	tmpl, err := template.New("vars").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
	}).Parse(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vc); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExpandCommand 展开命令中的变量
func (vc *VariableContext) ExpandCommand(script string) (string, error) {
	// 首先替换环境变量格式 ${VAR} 或 $VAR
	expanded := os.Expand(script, func(key string) string {
		// 优先使用自定义变量
		if val, exists := vc.CustomVars[key]; exists {
			return val
		}

		// 获取结构体字段值
		switch strings.ToUpper(key) {
		case "INSTALL_DIR":
			return vc.InstallDir
		case "BUILD_DIR":
			return vc.BuildDir
		case "DOWNLOAD_PATH":
			return vc.DownloadPath
		case "EXTRACT_DIR":
			return vc.ExtractDir
		case "SOURCE_DIR":
			return vc.SourceDir
		case "BUILD_TYPE":
			return vc.BuildType
		case "CXX_COMPILER":
			return vc.CXXCompiler
		case "CXX_FLAGS":
			return vc.CXXFlags
		case "CPU_CORES":
			return strconv.Itoa(vc.CPUCount)
		case "OS":
			return vc.OS
		case "ARCH":
			return vc.Arch
		case "PROJECT_NAME":
			return vc.ProjectName
		case "PROJECT_VERSION":
			return vc.ProjectVersion
		case "DEP_NAME":
			return vc.DepName
		case "DEP_VERSION":
			return vc.DepVersion
		case "BUILD_TAG":
			if vc.BuildTag != nil {
				return vc.BuildTag.String()
			}
			return ""
		case "BUILD_TAG_DIR":
			return vc.GetBuildTagDir()
		default:
			// 尝试从环境变量获取
			if envVal := os.Getenv(key); envVal != "" {
				return envVal
			}
			return "${" + key + "}" // 保持原样
		}
	})

	// 然后展开模板变量
	result, err := vc.ExpandVariables(expanded)
	if err != nil {
		return "", err
	}

	// 对于路径相关的变量，如果是相对路径则转换为绝对路径
	if strings.Contains(result, "${INSTALL_DIR}") {
		if installDir, exists := vc.CustomVars["install_dir"]; exists {
			resolvedPath := vc.resolvePath(installDir)
			result = strings.ReplaceAll(result, "${INSTALL_DIR}", resolvedPath)
		}
	}

	return result, nil
}

// ExpandList 展开字符串列表中的变量
func (vc *VariableContext) ExpandList(list []string) ([]string, error) {
	result := make([]string, len(list))
	for i, item := range list {
		expanded, err := vc.ExpandCommand(item)
		if err != nil {
			return nil, err
		}
		result[i] = expanded
	}
	return result, nil
}

// getCPUCount 获取 CPU 核心数
func getCPUCount() int {
	if runtime.NumCPU() > 0 {
		return runtime.NumCPU()
	}
	return 4 // 默认值
}

// getOS 获取操作系统
func getOS() string {
	return runtime.GOOS
}

// getArch 获取架构
func getArch() string {
	return runtime.GOARCH
}

// MergeVariables 合并变量
func (vc *VariableContext) MergeVariables(additional map[string]string) {
	if vc.CustomVars == nil {
		vc.CustomVars = make(map[string]string)
	}
	for k, v := range additional {
		vc.CustomVars[k] = v
	}
}

// Clone 克隆变量上下文
func (vc *VariableContext) Clone() *VariableContext {
	clone := *vc
	if vc.CustomVars != nil {
		clone.CustomVars = make(map[string]string)
		for k, v := range vc.CustomVars {
			clone.CustomVars[k] = v
		}
	}
	return &clone
}
