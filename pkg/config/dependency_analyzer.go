package config

import (
	"fmt"
	"strings"
)

// DependencyAnalyzer 依赖分析器
type DependencyAnalyzer struct {
	projectConfig *ProjectConfig
	buildTag      *BuildTag
}

// NewDependencyAnalyzer 创建依赖分析器
func NewDependencyAnalyzer(config *ProjectConfig, buildTag *BuildTag) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		projectConfig: config,
		buildTag:      buildTag,
	}
}

// AnalyzeDependencies 分析所有依赖项
func (da *DependencyAnalyzer) AnalyzeDependencies() ([]DependencyAnalysis, error) {
	var analyses []DependencyAnalysis

	for name, dep := range da.projectConfig.Dependencies {
		analysis, err := da.AnalyzeDependency(name, dep)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze dependency %s: %w", name, err)
		}
		analyses = append(analyses, *analysis)
	}

	return analyses, nil
}

// AnalyzeDependency 分析单个依赖项
func (da *DependencyAnalyzer) AnalyzeDependency(name string, dep Dependency) (*DependencyAnalysis, error) {
	analysis := &DependencyAnalysis{
		Name:    name,
		Version: dep.Version,
		Dep:     dep,
	}

	// 生成构建标签
	analysis.BuildTag = da.generateDependencyBuildTag(dep)

	// 生成标准化的路径
	analysis.BuildDir = da.getStandardBuildDir(dep, analysis.BuildTag)
	analysis.InstallDir = da.getStandardInstallDir(dep, analysis.BuildTag)
	analysis.CacheKey = da.generateCacheKey(dep, analysis.BuildTag)

	// 分析构建要求
	analysis.BuildRequirements = da.analyzeBuildRequirements(dep)

	return analysis, nil
}

// DependencyAnalysis 依赖分析结果
type DependencyAnalysis struct {
	Name              string            `json:"name"`
	Version           string            `json:"version"`
	Dep               Dependency        `json:"dependency"`
	BuildTag          *BuildTag         `json:"build_tag"`
	BuildDir          string            `json:"build_dir"`
	InstallDir        string            `json:"install_dir"`
	CacheKey          string            `json:"cache_key"`
	BuildRequirements BuildRequirements `json:"build_requirements"`
}

// BuildRequirements 构建要求
type BuildRequirements struct {
	RequiredTools    []string `json:"required_tools"`
	RequiredLibs     []string `json:"required_libs"`
	PlatformSpecific bool     `json:"platform_specific"`
	HasTests         bool     `json:"has_tests"`
	HeaderOnly       bool     `json:"header_only"`
}

// generateDependencyBuildTag 为依赖生成构建标签
func (da *DependencyAnalyzer) generateDependencyBuildTag(dep Dependency) *BuildTag {
	// 从项目构建标签开始
	var buildTag *BuildTag
	if da.buildTag != nil {
		buildTag = da.buildTag.Clone()
	} else {
		buildTag = &BuildTag{}
	}

	// 检查依赖是否有特定的构建标签要求
	if dep.EnvVariables != nil {
		// 从环境变量中提取构建标签信息
		if cxxStandard, exists := dep.EnvVariables["CXX_STANDARD"]; exists {
			buildTag.Std = da.mapCPPStandardToTag(cxxStandard)
		}

		if buildType, exists := dep.EnvVariables["BUILD_TYPE"]; exists {
			// 构建类型通常不包含在构建标签中，但某些特殊情况下可能需要
			if buildType == "Debug" {
				// 可以添加 debug 后缀或其他标识
			}
		}
	}

	// 根据构建系统调整构建标签
	switch dep.BuildSystem {
	case "cmake":
		// CMake 项目可能需要特殊的构建标签
		if buildTag.Std == "" {
			buildTag.Std = "cpp17" // CMake 默认使用 C++17
		}
	case "custom":
		// 自定义构建可能有特殊要求
		if strings.Contains(dep.CustomScript, "gcc") || strings.Contains(dep.CustomScript, "g++") {
			if buildTag.Compiler == "" {
				buildTag.Compiler = "gcc"
			}
		}
	}

	return buildTag
}

// mapCPPStandardToTag 将 C++ 标准映射到构建标签格式
func (da *DependencyAnalyzer) mapCPPStandardToTag(standard string) string {
	standard = strings.ToLower(strings.TrimSpace(standard))

	switch {
	case strings.Contains(standard, "11"):
		return "cpp11"
	case strings.Contains(standard, "14"):
		return "cpp14"
	case strings.Contains(standard, "17"):
		return "cpp17"
	case strings.Contains(standard, "20"):
		return "cpp20"
	case strings.Contains(standard, "23"):
		return "cpp23"
	default:
		return "cpp17" // 默认
	}
}

// getStandardBuildDir 获取标准化构建目录
func (da *DependencyAnalyzer) getStandardBuildDir(dep Dependency, buildTag *BuildTag) string {
	baseDir := "~/.buildfly/build"

	// 构建标准化路径：~/.buildfly/build/{name}/{version}/{build_tag}
	if buildTag != nil && buildTag.String() != "" {
		return fmt.Sprintf("%s/%s/%s/%s", baseDir, dep.Name, dep.Version, buildTag.ToDirName())
	}

	return fmt.Sprintf("%s/%s/%s", baseDir, dep.Name, dep.Version)
}

// getStandardInstallDir 获取标准化安装目录
func (da *DependencyAnalyzer) getStandardInstallDir(dep Dependency, buildTag *BuildTag) string {
	baseDir := "~/.buildfly/install"

	// 构建标准化路径：~/.buildfly/install/{name}/{version}/{build_tag}
	if buildTag != nil && buildTag.String() != "" {
		return fmt.Sprintf("%s/%s/%s/%s", baseDir, dep.Name, dep.Version, buildTag.ToDirName())
	}

	return fmt.Sprintf("%s/%s/%s", baseDir, dep.Name, dep.Version)
}

// generateCacheKey 生成缓存键
func (da *DependencyAnalyzer) generateCacheKey(dep Dependency, buildTag *BuildTag) string {
	// 缓存键格式：{name}@{version}#{build_tag}
	key := fmt.Sprintf("%s@%s", dep.Name, dep.Version)

	if buildTag != nil && buildTag.String() != "" {
		key += "#" + buildTag.String()
	}

	// 添加源码哈希（如果有）
	if dep.Source.SHA256 != "" {
		key += ":" + dep.Source.SHA256[:8] // 使用前8位
	} else if dep.Source.MD5 != "" {
		key += ":" + dep.Source.MD5[:8]
	}

	return key
}

// analyzeBuildRequirements 分析构建要求
func (da *DependencyAnalyzer) analyzeBuildRequirements(dep Dependency) BuildRequirements {
	req := BuildRequirements{}

	// 根据构建系统确定所需工具
	switch dep.BuildSystem {
	case "cmake":
		req.RequiredTools = []string{"cmake", "make"}
	case "make":
		req.RequiredTools = []string{"make"}
	case "configure":
		req.RequiredTools = []string{"autoconf", "automake", "make", "gcc", "g++"}
	case "custom":
		req.RequiredTools = da.analyzeCustomScriptTools(dep.CustomScript)
	case "none":
		req.RequiredTools = []string{}
		req.HeaderOnly = true
	}

	// 检查是否需要特定库
	if dep.BuildCommands.Configure != "" {
		if strings.Contains(dep.BuildCommands.Configure, "pthread") {
			req.RequiredLibs = append(req.RequiredLibs, "pthread")
		}
		if strings.Contains(dep.BuildCommands.Configure, "ssl") {
			req.RequiredLibs = append(req.RequiredLibs, "ssl", "crypto")
		}
	}

	// 检查平台特定要求
	if da.buildTag != nil {
		switch da.buildTag.Platform {
		case "windows":
			req.PlatformSpecific = true
			// Windows 可能需要 Visual Studio 或 MinGW
		case "linux", "darwin":
			req.PlatformSpecific = true
			// Unix 系统通常需要标准工具
		}
	}

	// 检查是否有测试
	if dep.BuildCommands.Test != "" || strings.Contains(dep.CustomScript, "test") {
		req.HasTests = true
	}

	return req
}

// analyzeCustomScriptTools 分析自定义脚本所需的工具
func (da *DependencyAnalyzer) analyzeCustomScriptTools(script string) []string {
	var tools []string
	toolMap := map[string]bool{
		"gcc":     false,
		"g++":     false,
		"clang":   false,
		"clang++": false,
		"make":    false,
		"cmake":   false,
		"python":  false,
		"bash":    false,
		"sh":      false,
	}

	lines := strings.Split(script, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 查找命令调用
		for tool := range toolMap {
			if strings.Contains(line, tool+" ") || strings.HasSuffix(line, tool) {
				toolMap[tool] = true
			}
		}
	}

	// 收集被标记为 true 的工具
	for tool, found := range toolMap {
		if found {
			tools = append(tools, tool)
		}
	}

	if len(tools) == 0 {
		// 默认工具
		tools = []string{"gcc", "make"}
	}

	return tools
}

// ValidateDependencies 验证所有依赖项
func (da *DependencyAnalyzer) ValidateDependencies() error {
	for name, dep := range da.projectConfig.Dependencies {
		if err := da.validateDependency(name, dep); err != nil {
			return fmt.Errorf("dependency %s validation failed: %w", name, err)
		}
	}
	return nil
}

// validateDependency 验证单个依赖项
func (da *DependencyAnalyzer) validateDependency(name string, dep Dependency) error {
	if dep.Version == "" {
		return fmt.Errorf("version is required")
	}

	// 对于外部index类型，URLs可能不是必需的
	if dep.Source.Type != "external_index" && len(dep.Source.GetURLs()) == 0 {
		return fmt.Errorf("source URLs are required")
	}

	// 验证构建系统
	validBuildSystems := []string{"cmake", "make", "configure", "custom", "none"}
	valid := false
	for _, bs := range validBuildSystems {
		if dep.BuildSystem == bs {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid build system: %s", dep.BuildSystem)
	}

	// 验证自定义构建
	if dep.BuildSystem == "custom" && dep.CustomScript == "" && dep.BuildCommands.Configure == "" {
		return fmt.Errorf("custom build system requires either custom_script or build_commands.configure")
	}

	return nil
}

// GetDependencyOrder 获取依赖构建顺序（用于处理依赖间的依赖关系）
func (da *DependencyAnalyzer) GetDependencyOrder() ([]string, error) {
	// 简单实现：按名称排序
	// 在实际应用中，这里应该实现拓扑排序来处理依赖关系

	var names []string
	for name := range da.projectConfig.Dependencies {
		names = append(names, name)
	}

	// 这里可以添加依赖关系分析逻辑
	// 目前返回所有依赖的名称

	return names, nil
}
