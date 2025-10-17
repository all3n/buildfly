package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader_LoadWithHierarchy(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "buildfly-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建全局配置目录
	globalConfigDir := filepath.Join(tempDir, ".config", "buildfly")
	err = os.MkdirAll(globalConfigDir, 0755)
	require.NoError(t, err)

	// 创建项目目录
	projectDir := filepath.Join(tempDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	t.Run("OnlyLocalConfig", func(t *testing.T) {
		// 只创建本地配置
		localConfig := `project:
  name: "test-project"
  version: "1.0.0"
dependencies:
  fmt:
    version: "8.0.1"
    build_system: "cmake"
    source:
      type: "git"
      urls:
        - "https://github.com/fmtlib/fmt.git"
`
		localConfigPath := filepath.Join(projectDir, "buildfly.yaml")
		err = os.WriteFile(localConfigPath, []byte(localConfig), 0644)
		require.NoError(t, err)

		loader := NewConfigLoader(projectDir)
		config, err := loader.LoadWithHierarchy()
		require.NoError(t, err)

		assert.Equal(t, "test-project", config.Project.Name)
		assert.Equal(t, "1.0.0", config.Project.Version)
		assert.Contains(t, config.Dependencies, "fmt")
	})

	t.Run("OnlyGlobalConfig", func(t *testing.T) {
		// 只创建全局配置
		globalConfig := `project:
  name: "global-project"
  version: "2.0.0"
dependencies:
  boost:
    version: "1.75.0"
    build_system: "cmake"
    source:
      type: "git"
      urls:
        - "https://github.com/boostorg/boost.git"
`
		globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
		err = os.WriteFile(globalConfigPath, []byte(globalConfig), 0644)
		require.NoError(t, err)

		// 确保本地配置不存在
		localConfigPath := filepath.Join(projectDir, "buildfly.yaml")
		os.Remove(localConfigPath)

		// 临时修改 HOME 环境变量指向测试目录
		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", tempDir)

		loader := NewConfigLoader(projectDir)
		config, err := loader.LoadWithHierarchy()
		require.NoError(t, err)

		assert.Equal(t, "global-project", config.Project.Name)
		assert.Equal(t, "2.0.0", config.Project.Version)
		assert.Contains(t, config.Dependencies, "boost")
	})

	t.Run("BothConfigsMerge", func(t *testing.T) {
		// 创建全局配置
		globalConfig := `project:
  name: "global-project"
  version: "2.0.0"
  variables:
    build_type: "Release"
    global_var: "from-global"
dependencies:
  boost:
    version: "1.75.0"
    build_system: "cmake"
    source:
      type: "git"
      urls:
        - "https://github.com/boostorg/boost.git"
  zlib:
    version: "1.2.11"
    build_system: "cmake"
    source:
      type: "git"
      urls:
        - "https://github.com/madler/zlib.git"
`
		globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
		err = os.WriteFile(globalConfigPath, []byte(globalConfig), 0644)
		require.NoError(t, err)

		// 创建本地配置
		localConfig := `project:
  name: "local-project"
  version: "1.0.0"
  variables:
    build_type: "Debug"
    local_var: "from-local"
dependencies:
  fmt:
    version: "8.0.1"
    build_system: "cmake"
    source:
      type: "git"
      urls:
        - "https://github.com/fmtlib/fmt.git"
  zlib:
    version: "1.3.0"
    build_system: "cmake"
    source:
      type: "git"
      urls:
        - "https://github.com/madler/zlib.git"
`
		localConfigPath := filepath.Join(projectDir, "buildfly.yaml")
		err = os.WriteFile(localConfigPath, []byte(localConfig), 0644)
		require.NoError(t, err)

		// 临时修改 HOME 环境变量指向测试目录
		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", tempDir)

		loader := NewConfigLoader(projectDir)
		config, err := loader.LoadWithHierarchy()
		require.NoError(t, err)

		// 本地配置应该覆盖全局配置
		assert.Equal(t, "local-project", config.Project.Name)                  // 本地覆盖
		assert.Equal(t, "1.0.0", config.Project.Version)                       // 本地覆盖
		assert.Equal(t, "Debug", config.Project.Variables["build_type"])       // 本地覆盖
		assert.Equal(t, "from-local", config.Project.Variables["local_var"])   // 本地独有
		assert.Equal(t, "from-global", config.Project.Variables["global_var"]) // 全局保留

		// 依赖项应该合并，本地覆盖同名依赖
		assert.Contains(t, config.Dependencies, "boost")              // 全局独有
		assert.Contains(t, config.Dependencies, "fmt")                // 本地独有
		assert.Contains(t, config.Dependencies, "zlib")               // 本地覆盖版本
		assert.Equal(t, "1.3.0", config.Dependencies["zlib"].Version) // 本地版本
	})
}

func TestConfigLoader_getGlobalConfigPath(t *testing.T) {
	t.Run("EnvironmentVariableOverride", func(t *testing.T) {
		// 创建临时配置文件
		tempFile, err := os.CreateTemp("", "custom-config.yaml")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		// 设置环境变量
		oldEnv := os.Getenv("BUILDFLY_CONFIG_FILE")
		defer os.Setenv("BUILDFLY_CONFIG_FILE", oldEnv)

		os.Setenv("BUILDFLY_CONFIG_FILE", tempFile.Name())

		loader := NewConfigLoader(".")
		path := loader.getGlobalConfigPath()
		assert.Equal(t, tempFile.Name(), path)
	})

	t.Run("DefaultGlobalPath", func(t *testing.T) {
		// 确保环境变量未设置
		oldEnv := os.Getenv("BUILDFLY_CONFIG_FILE")
		defer os.Setenv("BUILDFLY_CONFIG_FILE", oldEnv)
		os.Unsetenv("BUILDFLY_CONFIG_FILE")

		// 创建临时目录作为 home
		tempHome, err := os.MkdirTemp("", "test-home")
		require.NoError(t, err)
		defer os.RemoveAll(tempHome)

		// 创建全局配置目录和文件
		globalConfigDir := filepath.Join(tempHome, ".config", "buildfly")
		err = os.MkdirAll(globalConfigDir, 0755)
		require.NoError(t, err)

		globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
		err = os.WriteFile(globalConfigPath, []byte("test: config"), 0644)
		require.NoError(t, err)

		// 临时修改 HOME 环境变量
		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", tempHome)

		loader := NewConfigLoader(".")
		path := loader.getGlobalConfigPath()
		assert.Equal(t, globalConfigPath, path)
	})

	t.Run("NoGlobalConfig", func(t *testing.T) {
		// 确保环境变量未设置
		oldEnv := os.Getenv("BUILDFLY_CONFIG_FILE")
		defer os.Setenv("BUILDFLY_CONFIG_FILE", oldEnv)
		os.Unsetenv("BUILDFLY_CONFIG_FILE")

		// 使用不存在的 home 目录
		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", "/nonexistent")

		loader := NewConfigLoader(".")
		path := loader.getGlobalConfigPath()
		assert.Equal(t, "", path)
	})
}

func TestConfigLoader_mergeConfigs(t *testing.T) {
	loader := NewConfigLoader(".")

	globalConfig := &ProjectConfig{
		Project: Project{
			Name:    "global-project",
			Version: "2.0.0",
			Variables: map[string]string{
				"global_var": "from-global",
				"shared_var": "global-value",
			},
		},
		Dependencies: map[string]Dependency{
			"boost": {
				Name:        "boost",
				Version:     "1.75.0",
				BuildSystem: "cmake",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/boostorg/boost.git"},
				},
			},
			"zlib": {
				Name:        "zlib",
				Version:     "1.2.11",
				BuildSystem: "cmake",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/madler/zlib.git"},
				},
			},
		},
		BuildFlyBaseDir: "/global/.buildfly",
		InstallDir:      "/global/install",
		BuildDir:        "/global/build",
		CacheDir:        "/global/cache",
	}

	localConfig := &ProjectConfig{
		Project: Project{
			Name:    "local-project",
			Version: "1.0.0",
			Variables: map[string]string{
				"local_var":  "from-local",
				"shared_var": "local-value",
			},
		},
		Dependencies: map[string]Dependency{
			"fmt": {
				Name:        "fmt",
				Version:     "8.0.1",
				BuildSystem: "cmake",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/fmtlib/fmt.git"},
				},
			},
			"zlib": {
				Name:        "zlib",
				Version:     "1.3.0",
				BuildSystem: "cmake",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/madler/zlib.git"},
				},
			},
		},
		BuildFlyBaseDir: "/local/.buildfly",
		InstallDir:      "/local/install",
		ProjectRoot:     "/local/project",
	}

	merged := loader.mergeConfigs(globalConfig, localConfig)

	// 验证项目信息合并
	assert.Equal(t, "local-project", merged.Project.Name)                  // 本地覆盖
	assert.Equal(t, "1.0.0", merged.Project.Version)                       // 本地覆盖
	assert.Equal(t, "from-local", merged.Project.Variables["local_var"])   // 本地独有
	assert.Equal(t, "local-value", merged.Project.Variables["shared_var"]) // 本地覆盖
	assert.Equal(t, "from-global", merged.Project.Variables["global_var"]) // 全局保留

	// 验证依赖项合并
	assert.Contains(t, merged.Dependencies, "boost")              // 全局独有
	assert.Contains(t, merged.Dependencies, "fmt")                // 本地独有
	assert.Contains(t, merged.Dependencies, "zlib")               // 本地覆盖
	assert.Equal(t, "1.3.0", merged.Dependencies["zlib"].Version) // 本地版本

	// 验证目录配置合并（本地优先）
	assert.Equal(t, "/local/.buildfly", merged.BuildFlyBaseDir)
	assert.Equal(t, "/local/install", merged.InstallDir)
	assert.Equal(t, "/global/build", merged.BuildDir) // 本地未设置，保留全局
	assert.Equal(t, "/global/cache", merged.CacheDir) // 本地未设置，保留全局

	// 验证项目根目录
	assert.Equal(t, "/local/project", merged.ProjectRoot)
}
