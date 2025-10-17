package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buildfly/internal/errors"
	"buildfly/pkg/config"
	"buildfly/pkg/utils"
)

// CacheManager 缓存管理器
type CacheManager struct {
	cacheDir string
	maxSize  int64         // 最大缓存大小（字节）
	maxAge   time.Duration // 最大缓存时间
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(cacheDir string, maxSize int64, maxAge time.Duration) *CacheManager {
	return &CacheManager{
		cacheDir: cacheDir,
		maxSize:  maxSize,
		maxAge:   maxAge,
	}
}

// Init 初始化缓存目录
func (cm *CacheManager) Init() error {
	if err := os.MkdirAll(cm.cacheDir, 0755); err != nil {
		return errors.CacheErrorWithCause(err, fmt.Sprintf("failed to create cache directory %s", cm.cacheDir))
	}

	// 创建子目录
	subDirs := []string{"downloads", "builds", "metadata"}
	for _, dir := range subDirs {
		if err := os.MkdirAll(filepath.Join(cm.cacheDir, dir), 0755); err != nil {
			return errors.CacheErrorWithCause(err, fmt.Sprintf("failed to create cache subdirectory %s", dir))
		}
	}

	return nil
}

// GetCacheKey 生成缓存键
func (cm *CacheManager) GetCacheKey(dep config.Dependency) string {
	h := sha256.New()

	// 使用所有 URL 生成缓存键
	urls := dep.Source.GetURLs()
	for _, url := range urls {
		h.Write([]byte(url))
	}

	h.Write([]byte(dep.Version))
	if dep.Source.Tag != "" {
		h.Write([]byte(dep.Source.Tag))
	}
	if dep.Source.Hash != "" {
		h.Write([]byte(dep.Source.Hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// GetDownloadCachePath 获取下载缓存路径
// 规范路径：{cache_dir}/buildfly/{name}/{version}/{filename}.{ext}
func (cm *CacheManager) GetDownloadCachePath(dep config.Dependency) string {
	// 获取文件名和扩展名
	filename := cm.getCacheFileName(dep)

	// 构建标准化路径：{cache_dir}/buildfly/{name}/{version}/{filename}.{ext}
	return filepath.Join(cm.cacheDir, "buildfly", dep.Name, dep.Version, filename)
}

// GetBuildCachePath 获取构建缓存路径
// 规范路径：{cache_dir}/buildfly/{name}/{version}/{build_tag}
func (cm *CacheManager) GetBuildCachePath(dep config.Dependency, buildTag *config.BuildTag) string {
	// 构建标签目录名
	buildTagDir := "default"
	if buildTag != nil {
		buildTagDir = buildTag.ToDirName()
	}

	// 构建标准化路径：{cache_dir}/buildfly/{name}/{version}/{build_tag}
	return filepath.Join(cm.cacheDir, "buildfly", dep.Name, dep.Version, buildTagDir)
}

// getCacheFileName 获取缓存文件名
func (cm *CacheManager) getCacheFileName(dep config.Dependency) string {
	// 尝试从 URL 获取文件名
	urls := dep.Source.GetURLs()
	if len(urls) > 0 {
		filename := getFileNameFromURL(urls[0])
		if filename != "" && filename != "download" {
			return filename
		}
	}

	// 如果无法从 URL 获取，使用缓存键
	cacheKey := cm.GetCacheKey(dep)

	// 根据源类型确定扩展名
	ext := ""
	switch dep.Source.Type {
	case "archive":
		ext = ".tar.gz"
	case "git":
		ext = ".git"
	default:
		ext = ".bin"
	}

	return cacheKey + ext
}

// getFileNameFromURL 从 URL 获取文件名
func getFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// 移除查询参数
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		// 移除片段标识符
		if idx := strings.Index(filename, "#"); idx != -1 {
			filename = filename[:idx]
		}
		return filename
	}
	return ""
}

// GetMetadataPath 获取元数据路径
func (cm *CacheManager) GetMetadataPath(dep config.Dependency) string {
	cacheKey := cm.GetCacheKey(dep)
	return filepath.Join(cm.cacheDir, "metadata", cacheKey+".json")
}

// IsCachedDownloads 检查是否已缓存
func (cm *CacheManager) IsCachedDownloads(dep config.Dependency) bool {
	cachePath := cm.GetDownloadCachePath(dep)
	if _, err := os.Stat(cachePath); err != nil {
		return false
	}

	// 检查缓存是否过期
	// if cm.isExpired(cachePath) {
	// 	cm.Invalidate(dep)
	// 	return false
	// }
	return true
}

// IsBuildCached 检查构建是否已缓存
func (cm *CacheManager) IsBuildCached(dep config.Dependency, buildTag *config.BuildTag) bool {
	// {{.BuildFlyGlobalDir}}/install/{name}/{version}/{build_tag}
	cachePath := cm.GetBuildCachePath(dep, buildTag)
	if _, err := os.Stat(cachePath); err != nil {
		return false
	}

	// 检查缓存是否过期 TODO
	//if cm.isExpired(cachePath) {
	//	cm.InvalidateBuild(dep)
	//	return false
	//}

	return true
}

// Store 存储到缓存
func (cm *CacheManager) Store(dep config.Dependency, sourcePath string) error {
	cachePath := cm.GetDownloadCachePath(dep)

	// 确保缓存目录存在
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return errors.CacheErrorWithCause(err, "failed to create cache directory")
	}

	// 如果是目录，递归复制
	if info, err := os.Stat(sourcePath); err == nil && info.IsDir() {
		return utils.CopyDir(sourcePath, cachePath)
	}

	// 复制文件
	return utils.CopyFile(sourcePath, cachePath)
}

// StoreBuild 存储构建结果到缓存
func (cm *CacheManager) StoreBuild(dep config.Dependency, buildPath string, buildTag *config.BuildTag) error {
	cachePath := cm.GetBuildCachePath(dep, buildTag)

	// 确保缓存目录存在
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return errors.CacheErrorWithCause(err, "failed to create build cache directory")
	}

	// 如果是目录，递归复制
	if info, err := os.Stat(buildPath); err == nil && info.IsDir() {
		return utils.CopyDir(buildPath, cachePath)
	}

	// 复制文件
	return utils.CopyFile(buildPath, cachePath)
}

// Retrieve 从缓存检索
func (cm *CacheManager) Retrieve(dep config.Dependency, targetPath string) error {
	cachePath := cm.GetDownloadCachePath(dep)

	if _, err := os.Stat(cachePath); err != nil {
		return errors.CacheError(fmt.Sprintf("cache not found for dependency %s", dep.Name))
	}

	// 如果是目录，递归复制
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		fmt.Printf("Copying directory from %s to %s\n", cachePath, targetPath)
		return utils.CopyDir(cachePath, targetPath)
	}

	// 复制文件
	fmt.Printf("Copying file from %s to %s\n", cachePath, targetPath)
	return utils.CopyFile(cachePath, targetPath)
}

// RetrieveBuild 从缓存检索构建结果
func (cm *CacheManager) RetrieveBuild(dep config.Dependency, targetPath string, buildTag *config.BuildTag) error {
	cachePath := cm.GetBuildCachePath(dep, buildTag)

	if _, err := os.Stat(cachePath); err != nil {
		return errors.CacheError(fmt.Sprintf("build cache not found for dependency %s", dep.Name))
	}

	// 如果是目录，递归复制
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		return utils.CopyDir(cachePath, targetPath)
	}

	// 复制文件
	return utils.CopyFile(cachePath, targetPath)
}

// Invalidate 使缓存失效
func (cm *CacheManager) Invalidate(dep config.Dependency) error {
	cachePath := cm.GetDownloadCachePath(dep)
	if err := os.RemoveAll(cachePath); err != nil {
		return errors.CacheErrorWithCause(err, "failed to invalidate cache")
	}

	// 删除元数据
	metadataPath := cm.GetMetadataPath(dep)
	os.Remove(metadataPath)

	return nil
}

// InvalidateBuild 使构建缓存失效
func (cm *CacheManager) InvalidateBuild(dep config.Dependency, buildTag *config.BuildTag) error {
	cachePath := cm.GetBuildCachePath(dep, buildTag)
	if err := os.RemoveAll(cachePath); err != nil {
		return errors.CacheErrorWithCause(err, "failed to invalidate build cache")
	}

	return nil
}

// Cleanup 清理过期缓存
func (cm *CacheManager) Cleanup() error {
	// 清理下载缓存
	if err := cm.cleanupDir(filepath.Join(cm.cacheDir, "downloads")); err != nil {
		return err
	}

	// 清理构建缓存
	if err := cm.cleanupDir(filepath.Join(cm.cacheDir, "builds")); err != nil {
		return err
	}

	// 清理元数据
	if err := cm.cleanupDir(filepath.Join(cm.cacheDir, "metadata")); err != nil {
		return err
	}

	return nil
}

// cleanupDir 清理目录
func (cm *CacheManager) cleanupDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.CacheErrorWithCause(err, fmt.Sprintf("failed to read cache directory %s", dir))
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		if cm.isExpired(entryPath) {
			if err := os.RemoveAll(entryPath); err != nil {
				return errors.CacheErrorWithCause(err, fmt.Sprintf("failed to remove expired cache %s", entryPath))
			}
		}
	}

	return nil
}

// isExpired 检查缓存是否过期
func (cm *CacheManager) isExpired(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	return time.Since(info.ModTime()) > cm.maxAge
}

// GetCacheSize 获取缓存大小
func (cm *CacheManager) GetCacheSize() (int64, error) {
	var size int64

	err := filepath.Walk(cm.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// Clear 清空所有缓存
func (cm *CacheManager) Clear() error {
	if err := os.RemoveAll(cm.cacheDir); err != nil {
		return errors.CacheErrorWithCause(err, "failed to clear cache")
	}
	return cm.Init()
}

// CacheInfo 缓存信息
type CacheInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Expired bool
}

// ListCache 列出缓存内容
func (cm *CacheManager) ListCache() ([]CacheInfo, error) {
	var cacheInfos []CacheInfo

	err := filepath.Walk(cm.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			cacheInfos = append(cacheInfos, CacheInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Expired: cm.isExpired(path),
			})
		}
		return nil
	})

	return cacheInfos, err
}
