package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoader_Load(t *testing.T) {
	// 创建临时测试配置文件
	tempDir, err := os.MkdirTemp("", "buildfly-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
project:
  name: "test-project"
  version: "1.0.0"
  variables:
    install_dir: "/tmp/test"
    build_type: "Release"

dependencies:
  boost:
    version: "1.75.0"
    source:
      type: "archive"
      urls:
        - "https://boostorg.jfrog.io/artifactory/main/release/1.75.0/source/boost_1_75_0.tar.gz"
    build_system: "custom"
    custom_script: |
      #!/bin/bash
      ./bootstrap.sh --prefix=${INSTALL_DIR}
      ./b2 install

  fmt:
    version: "8.0.1"
    source:
      type: "git"
      urls:
        - "https://github.com/fmtlib/fmt.git"
      tag: "8.0.1"
    build_system: "cmake"
    cmake_options:
      - "FMT_TEST=OFF"
      - "CMAKE_POSITION_INDEPENDENT_CODE=ON"

build_profiles:
  release:
    variables:
      build_type: "Release"
    dependencies:
      - "boost"
      - "fmt"
  
  debug:
    variables:
      build_type: "Debug"
    dependencies:
      - "fmt"
`

	configFile := filepath.Join(tempDir, "buildfly.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 测试配置加载
	loader := NewConfigLoader(tempDir)
	config, err := loader.Load("buildfly.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证项目配置
	if config.Project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", config.Project.Name)
	}
	if config.Project.Version != "1.0.0" {
		t.Errorf("Expected project version '1.0.0', got '%s'", config.Project.Version)
	}
	if config.Project.Variables["install_dir"] != "/tmp/test" {
		t.Errorf("Expected install_dir '/tmp/test', got '%s'", config.Project.Variables["install_dir"])
	}

	// 验证依赖配置
	if len(config.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(config.Dependencies))
	}

	boost, exists := config.Dependencies["boost"]
	if !exists {
		t.Fatal("Boost dependency not found")
	}
	if boost.Version != "1.75.0" {
		t.Errorf("Expected boost version '1.75.0', got '%s'", boost.Version)
	}
	if boost.Source.Type != "archive" {
		t.Errorf("Expected boost source type 'archive', got '%s'", boost.Source.Type)
	}
	if boost.BuildSystem != "custom" {
		t.Errorf("Expected boost build system 'custom', got '%s'", boost.BuildSystem)
	}

	fmt, exists := config.Dependencies["fmt"]
	if !exists {
		t.Fatal("Fmt dependency not found")
	}
	if fmt.Version != "8.0.1" {
		t.Errorf("Expected fmt version '8.0.1', got '%s'", fmt.Version)
	}
	if fmt.Source.Type != "git" {
		t.Errorf("Expected fmt source type 'git', got '%s'", fmt.Source.Type)
	}
	if fmt.BuildSystem != "cmake" {
		t.Errorf("Expected fmt build system 'cmake', got '%s'", fmt.BuildSystem)
	}

	// 验证构建配置文件
	if len(config.BuildProfiles) != 2 {
		t.Errorf("Expected 2 build profiles, got %d", len(config.BuildProfiles))
	}

	releaseProfile, exists := config.BuildProfiles["release"]
	if !exists {
		t.Fatal("Release profile not found")
	}
	if releaseProfile.Variables["build_type"] != "Release" {
		t.Errorf("Expected release build_type 'Release', got '%s'", releaseProfile.Variables["build_type"])
	}
	if len(releaseProfile.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies in release profile, got %d", len(releaseProfile.Dependencies))
	}
}

func TestConfigLoader_Validate(t *testing.T) {
	loader := NewConfigLoader(".")

	// 测试有效配置
	validConfig := &ProjectConfig{
		Project: Project{
			Name:    "test",
			Version: "1.0.0",
		},
		Dependencies: map[string]Dependency{
			"test": {
				Name:    "test",
				Version: "1.0.0",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/test/test.git"},
				},
				BuildSystem: "cmake",
			},
		},
	}

	if err := loader.Validate(validConfig); err != nil {
		t.Errorf("Valid config should not fail validation: %v", err)
	}

	// 测试无效配置 - 缺少项目名称
	invalidConfig := &ProjectConfig{
		Project: Project{
			Version: "1.0.0",
		},
		Dependencies: map[string]Dependency{},
	}

	if err := loader.Validate(invalidConfig); err == nil {
		t.Error("Invalid config should fail validation")
	}

	// 测试无效配置 - 缺少依赖版本
	invalidConfig2 := &ProjectConfig{
		Project: Project{
			Name:    "test",
			Version: "1.0.0",
		},
		Dependencies: map[string]Dependency{
			"test": {
				Name: "test",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/test/test.git"},
				},
				BuildSystem: "cmake",
			},
		},
	}

	if err := loader.Validate(invalidConfig2); err == nil {
		t.Error("Invalid config should fail validation")
	}

	// 测试无效配置 - 不支持的构建系统
	invalidConfig3 := &ProjectConfig{
		Project: Project{
			Name:    "test",
			Version: "1.0.0",
		},
		Dependencies: map[string]Dependency{
			"test": {
				Name:    "test",
				Version: "1.0.0",
				Source: SourceInfo{
					Type: "git",
					URLS: []string{"https://github.com/test/test.git"},
				},
				BuildSystem: "unsupported",
			},
		},
	}

	if err := loader.Validate(invalidConfig3); err == nil {
		t.Error("Invalid config should fail validation")
	}
}
