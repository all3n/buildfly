package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	baseDir string
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader(baseDir string) *ConfigLoader {
	return &ConfigLoader{
		baseDir: baseDir,
	}
}

// LoadWithHierarchy 按优先级层次加载配置文件
// 1. 全局配置 (~/.config/buildfly/config.yaml 或 BUILDFLY_CONFIG_FILE 环境变量指定)
// 2. 本地配置 (当前目录 buildfly.yaml)
// 本地配置会覆盖全局配置中的同名配置
func (cl *ConfigLoader) LoadWithHierarchy() (*ProjectConfig, error) {
	var globalConfig *ProjectConfig
	var err error

	// 1. 加载全局配置
	globalConfigPath := cl.getGlobalConfigPath()
	if globalConfigPath != "" {
		globalConfig, err = cl.Load(globalConfigPath)
		if err != nil {
			// 全局配置加载失败时记录警告但不中断
			fmt.Printf("Warning: failed to load global config from %s: %v\n", globalConfigPath, err)
			globalConfig = nil
		}
	}

	// 2. 尝试加载本地配置
	localConfigPath := filepath.Join(cl.baseDir, "buildfly.yaml")
	if _, err := os.Stat(localConfigPath); err == nil {
		// 本地配置存在
		localConfig, err := cl.Load(localConfigPath)
		if err != nil {
			if globalConfig != nil {
				// 如果本地配置加载失败但有全局配置，使用全局配置
				fmt.Printf("Using global config only (local config load failed: %v)\n", err)
				return globalConfig, nil
			}
			return nil, fmt.Errorf("failed to load local config from %s: %w", localConfigPath, err)
		}

		// 3. 合并配置
		if globalConfig != nil {
			return cl.mergeConfigs(globalConfig, localConfig), nil
		}

		// 只有本地配置
		return localConfig, nil
	}

	// 本地配置不存在
	if globalConfig != nil {
		fmt.Printf("Using global config only (no local config found)\n")
		return globalConfig, nil
	}

	return nil, fmt.Errorf("no config file found: neither global nor local config available")
}

// getGlobalConfigPath 获取全局配置文件路径
func (cl *ConfigLoader) getGlobalConfigPath() string {
	// 优先使用环境变量指定的路径
	if envPath := os.Getenv("BUILDFLY_CONFIG_FILE"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		fmt.Printf("Warning: BUILDFLY_CONFIG_FILE %s not accessible\n", envPath)
	}

	// 默认全局配置路径
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	globalPath := filepath.Join(home, ".config", "buildfly", "config.yaml")
	if _, err := os.Stat(globalPath); err == nil {
		return globalPath
	}

	return ""
}

// mergeConfigs 合并两个配置，localConfig 会覆盖 globalConfig 中的同名配置
func (cl *ConfigLoader) mergeConfigs(globalConfig, localConfig *ProjectConfig) *ProjectConfig {
	merged := &ProjectConfig{}

	// 深拷贝全局配置作为基础
	*merged = *globalConfig

	// 合并项目信息
	if localConfig.Project.Name != "" {
		merged.Project.Name = localConfig.Project.Name
	}
	if localConfig.Project.Version != "" {
		merged.Project.Version = localConfig.Project.Version
	}

	// 合并项目变量
	if merged.Project.Variables == nil {
		merged.Project.Variables = make(map[string]string)
	}
	if localConfig.Project.Variables != nil {
		for k, v := range localConfig.Project.Variables {
			merged.Project.Variables[k] = v
		}
	}

	// 合并依赖项
	if merged.Dependencies == nil {
		merged.Dependencies = make(map[string]Dependency)
	}
	if localConfig.Dependencies != nil {
		for k, v := range localConfig.Dependencies {
			// 设置依赖项名称
			v.Name = k
			merged.Dependencies[k] = v
		}
	}

	// 合并构建配置文件
	if merged.BuildProfiles == nil {
		merged.BuildProfiles = make(map[string]BuildProfile)
	}
	if localConfig.BuildProfiles != nil {
		for k, v := range localConfig.BuildProfiles {
			merged.BuildProfiles[k] = v
		}
	}

	// 合并目录配置（本地配置优先）
	if localConfig.BuildFlyBaseDir != "" {
		merged.BuildFlyBaseDir = localConfig.BuildFlyBaseDir
	}
	if localConfig.InstallDir != "" {
		merged.InstallDir = localConfig.InstallDir
	}
	if localConfig.BuildDir != "" {
		merged.BuildDir = localConfig.BuildDir
	}
	if localConfig.CacheDir != "" {
		merged.CacheDir = localConfig.CacheDir
	}

	// 合并代理配置（本地配置优先）
	if localConfig.Proxy != nil {
		merged.Proxy = localConfig.Proxy
	}

	// 项目根目录设置为本地配置的目录
	merged.ProjectRoot = localConfig.ProjectRoot

	return merged
}

// Load 加载配置文件
func (cl *ConfigLoader) Load(configFile string) (*ProjectConfig, error) {
	// 如果是相对路径，则基于 baseDir
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(cl.baseDir, configFile)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config %s: %w", configFile, err)
	}

	// 设置依赖项名称
	for name, dep := range config.Dependencies {
		dep.Name = name
		config.Dependencies[name] = dep
	}
	config.ProjectRoot = filepath.Dir(configFile)

	// 验证配置
	if err := cl.Validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	cl.fillConfigDefaults(&config)

	return &config, nil
}
func (cl *ConfigLoader) fillConfigDefaults(cfg *ProjectConfig) {
	if cfg.BuildFlyBaseDir == "" {
		cfg.BuildFlyBaseDir = filepath.Join(cfg.ProjectRoot, ".buildfly")
	}
	if cfg.InstallDir == "" {
		// install use home share dir by default
		home, _ := os.UserHomeDir()
		cfg.InstallDir = filepath.Join(home, ".buildfly", "install")
	}
	if cfg.BuildDir == "" {
		cfg.BuildDir = filepath.Join(cfg.BuildFlyBaseDir, "build")
	}
	// cache use home share dir by default
	if cfg.CacheDir == "" {
		home, _ := os.UserHomeDir()
		cfg.CacheDir = filepath.Join(home, ".buildfly", "cache")
	}
}

// LoadWithProfile 加载配置并应用构建配置文件
func (cl *ConfigLoader) LoadWithProfile(configFile, profileName string) (*ProjectConfig, error) {
	config, err := cl.Load(configFile)
	if err != nil {
		return nil, err
	}

	if profileName != "" {
		profile, exists := config.BuildProfiles[profileName]
		if !exists {
			return nil, fmt.Errorf("build profile '%s' not found", profileName)
		}

		// 应用配置文件的变量
		if config.Project.Variables == nil {
			config.Project.Variables = make(map[string]string)
		}
		for k, v := range profile.Variables {
			config.Project.Variables[k] = v
		}

		// 应用配置文件的构建标签（如果有）
		if profile.BuildTag != nil {
			// 如果配置文件有 build_tag，则覆盖全局 build_tag
			config.BuildTag = profile.BuildTag.Clone()
		}

		// 过滤依赖项，只保留配置文件中指定的
		if len(profile.Dependencies) > 0 {
			filteredDeps := make(map[string]Dependency)
			for _, depName := range profile.Dependencies {
				if dep, exists := config.Dependencies[depName]; exists {
					filteredDeps[depName] = dep
				}
			}
			config.Dependencies = filteredDeps
		}
	}

	return config, nil
}

// Save 保存配置到文件
func (cl *ConfigLoader) Save(config *ProjectConfig, configFile string) error {
	// 如果是相对路径，则基于 baseDir
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(cl.baseDir, configFile)
	}

	// 确保目录存在
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configFile, err)
	}

	return nil
}

// Validate 验证配置
func (cl *ConfigLoader) Validate(config *ProjectConfig) error {
	if config.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if config.Project.Version == "" {
		return fmt.Errorf("project version is required")
	}

	// 验证依赖项
	for name, dep := range config.Dependencies {
		if err := cl.validateDependency(name, dep); err != nil {
			return err
		}
	}

	// 验证构建配置文件
	for profileName, profile := range config.BuildProfiles {
		if err := cl.validateBuildProfile(profileName, profile); err != nil {
			return err
		}
	}

	return nil
}

// validateDependency 验证单个依赖项
func (cl *ConfigLoader) validateDependency(name string, dep Dependency) error {
	if dep.Version == "" {
		return fmt.Errorf("version is required for dependency %s", name)
	}

	// 对于外部index类型，URLs可能不是必需的
	if dep.Source.Type != "external_index" && len(dep.Source.URLS) == 0 {
		return fmt.Errorf("source URLs are required for dependency %s", name)
	}

	// 验证源码类型
	supportedSourceTypes := map[string]bool{
		"git":            true,
		"archive":        true,
		"direct":         true,
		"external_index": true,
	}
	if !supportedSourceTypes[dep.Source.Type] {
		return fmt.Errorf("unsupported source type: %s for dependency %s", dep.Source.Type, name)
	}

	// 验证构建系统
	supportedBuildSystems := map[string]bool{
		"make":      true,
		"cmake":     true,
		"configure": true,
		"custom":    true,
		"none":      true, // 表示不需要构建，直接使用头文件
	}
	if !supportedBuildSystems[dep.BuildSystem] {
		return fmt.Errorf("unsupported build system: %s for dependency %s", dep.BuildSystem, name)
	}

	// 验证自定义脚本
	if dep.BuildSystem == "custom" && dep.CustomScript == "" && dep.BuildCommands.Configure == "" {
		return fmt.Errorf("custom_script or build_commands.configure is required for custom build system in dependency %s", name)
	}

	return nil
}

// validateBuildProfile 验证构建配置文件
func (cl *ConfigLoader) validateBuildProfile(name string, profile BuildProfile) error {
	// 检查依赖项是否存在
	for _, depName := range profile.Dependencies {
		if depName == "" {
			return fmt.Errorf("empty dependency name in build profile %s", name)
		}
	}

	return nil
}

// FindConfigFile 查找配置文件
func (cl *ConfigLoader) FindConfigFile() (string, error) {
	// 常见的配置文件名
	configNames := []string{
		"cppdep.yaml",
		"buildfly.yaml",
		".cppdep.yaml",
		".buildfly.yaml",
	}

	// 在当前目录和父目录中查找
	dir := cl.baseDir
	for {
		for _, name := range configNames {
			configPath := filepath.Join(dir, name)
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil
			}
		}

		// 移动到父目录
		parent := filepath.Dir(dir)
		if parent == dir {
			break // 已经到达根目录
		}
		dir = parent
	}

	return "", fmt.Errorf("no config file found")
}

// GetDefaultConfig 获取默认配置
func (cl *ConfigLoader) GetDefaultConfig() *ProjectConfig {
	return &ProjectConfig{
		Project: Project{
			Name:    "buildfly-demo",
			Version: "1.0.0",
			Variables: map[string]string{
				"build_type":   "Release",
				"cxx_compiler": "g++",
				"cxx_flags":    "-O2 -std=c++17",
			},
		},
		Dependencies: make(map[string]Dependency),
		BuildProfiles: map[string]BuildProfile{
			"default": {
				Dependencies: []string{},
			},
		},
	}
}
