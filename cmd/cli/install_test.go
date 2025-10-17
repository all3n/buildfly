package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"buildfly/pkg/cache"
	"buildfly/pkg/config"
)

func TestInstallCacheLogic(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "buildfly-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试配置
	projectConfig := &config.ProjectConfig{
		Project: config.Project{
			Name: "test-project",
		},
		Dependencies: map[string]config.Dependency{
			"test-lib": {
				Name:    "test-lib",
				Version: "1.0.0",
				Source: config.SourceInfo{
					Type: "archive",
					URLS: []string{"https://example.com/test-lib.tar.gz"},
				},
				BuildSystem: "cmake",
			},
		},
	}

	// 创建缓存管理器
	cacheDir := filepath.Join(tempDir, "cache")
	cacheManager := cache.NewCacheManager(cacheDir, 1024*1024*1024, 24*time.Hour)
	if err := cacheManager.Init(); err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}

	// 测试依赖
	dep := projectConfig.Dependencies["test-lib"]

	// 测试初始状态 - 应该没有缓存
	if cacheManager.IsCachedDownloads(dep) {
		t.Error("Expected no cache for new dependency")
	}
	if cacheManager.IsBuildCached(dep, nil) {
		t.Error("Expected no build cache for new dependency")
	}

	// 模拟下载缓存
	downloadCachePath := cacheManager.GetDownloadCachePath(dep)
	if err := os.MkdirAll(downloadCachePath, 0755); err != nil {
		t.Fatalf("Failed to create download cache dir: %v", err)
	}

	// 创建一些测试文件
	testFile := filepath.Join(downloadCachePath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 现在应该有下载缓存
	if !cacheManager.IsCachedDownloads(dep) {
		t.Error("Expected download cache to exist")
	}
	if cacheManager.IsBuildCached(dep, nil) {
		t.Error("Expected no build cache yet")
	}

	// 模拟构建缓存
	buildCachePath := cacheManager.GetBuildCachePath(dep, nil)
	if err := os.MkdirAll(buildCachePath, 0755); err != nil {
		t.Fatalf("Failed to create build cache dir: %v", err)
	}

	// 创建构建结果文件
	buildFile := filepath.Join(buildCachePath, "libtest.a")
	if err := os.WriteFile(buildFile, []byte("fake library content"), 0644); err != nil {
		t.Fatalf("Failed to create build file: %v", err)
	}

	// 现在应该有两种缓存
	if !cacheManager.IsCachedDownloads(dep) {
		t.Error("Expected download cache to exist")
	}
	if !cacheManager.IsBuildCached(dep, nil) {
		t.Error("Expected build cache to exist")
	}

	// 测试缓存路径生成
	cacheKey := cacheManager.GetCacheKey(dep)
	if cacheKey == "" {
		t.Error("Expected non-empty cache key")
	}

	// 实际的下载缓存路径应该是：{cache_dir}/buildfly/{name}/{version}/{filename}
	expectedDownloadPath := filepath.Join(cacheDir, "buildfly", dep.Name, dep.Version, "test-lib.tar.gz")
	if downloadCachePath != expectedDownloadPath {
		t.Errorf("Expected download path %s, got %s", expectedDownloadPath, downloadCachePath)
	}

	// 实际的构建缓存路径应该是：{cache_dir}/buildfly/{name}/{version}/{build_tag}
	expectedBuildPath := filepath.Join(cacheDir, "buildfly", dep.Name, dep.Version, "default")
	if buildCachePath != expectedBuildPath {
		t.Errorf("Expected build path %s, got %s", expectedBuildPath, buildCachePath)
	}
}

func TestInstallCacheRetrieval(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "buildfly-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建缓存管理器
	cacheDir := filepath.Join(tempDir, "cache")
	cacheManager := cache.NewCacheManager(cacheDir, 1024*1024*1024, 24*time.Hour)
	if err := cacheManager.Init(); err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}

	// 创建测试依赖
	dep := config.Dependency{
		Name:    "test-lib",
		Version: "1.0.0",
		Source: config.SourceInfo{
			Type: "archive",
			URLS: []string{"https://example.com/test-lib.tar.gz"},
		},
		BuildSystem: "cmake",
	}

	// 创建源文件缓存
	downloadCachePath := cacheManager.GetDownloadCachePath(dep)
	if err := os.MkdirAll(downloadCachePath, 0755); err != nil {
		t.Fatalf("Failed to create download cache dir: %v", err)
	}

	// 创建测试源文件
	sourceFile := filepath.Join(downloadCachePath, "CMakeLists.txt")
	if err := os.WriteFile(sourceFile, []byte("cmake_minimum_required(VERSION 3.10)\nproject(test-lib)"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// 创建构建缓存
	buildCachePath := cacheManager.GetBuildCachePath(dep, nil)
	if err := os.MkdirAll(buildCachePath, 0755); err != nil {
		t.Fatalf("Failed to create build cache dir: %v", err)
	}

	// 创建构建结果文件
	libFile := filepath.Join(buildCachePath, "libtest.a")
	if err := os.WriteFile(libFile, []byte("fake library content"), 0644); err != nil {
		t.Fatalf("Failed to create library file: %v", err)
	}

	// 测试检索下载缓存
	retrieveDir := filepath.Join(tempDir, "retrieve")
	if err := cacheManager.Retrieve(dep, retrieveDir); err != nil {
		t.Fatalf("Failed to retrieve download cache: %v", err)
	}

	// 验证文件是否存在
	retrievedFile := filepath.Join(retrieveDir, "CMakeLists.txt")
	if _, err := os.Stat(retrievedFile); err != nil {
		t.Errorf("Expected retrieved file to exist: %v", err)
	}

	// 测试检索构建缓存
	retrieveBuildDir := filepath.Join(tempDir, "retrieve-build")
	if err := cacheManager.RetrieveBuild(dep, retrieveBuildDir, nil); err != nil {
		t.Fatalf("Failed to retrieve build cache: %v", err)
	}

	// 验证构建文件是否存在
	retrievedLibFile := filepath.Join(retrieveBuildDir, "libtest.a")
	if _, err := os.Stat(retrievedLibFile); err != nil {
		t.Errorf("Expected retrieved library file to exist: %v", err)
	}
}
