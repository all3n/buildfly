package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathManager 路径管理器，处理标准化的路径操作
type PathManager struct {
	homeDir       string
	buildflyBase  string
	buildflyCache string
}

// NewPathManager 创建路径管理器
func NewPathManager() (*PathManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return &PathManager{
		homeDir:       home,
		buildflyBase:  filepath.Join(home, ".buildfly"),
		buildflyCache: filepath.Join(home, ".buildfly", "cache"),
	}, nil
}

// GetDownloadCachePath 获取下载缓存路径
// 格式：{cache_dir}/buildfly/{name}/{version}/{filename}.{ext}
func (pm *PathManager) GetDownloadCachePath(depName, version, filename string) string {
	return filepath.Join(pm.buildflyCache, "buildfly", depName, version, filename)
}

// GetBuildDir 获取构建目录路径
// 格式：~/.buildfly/build/{name}/{version}/{build_tag}
func (pm *PathManager) GetBuildDir(depName, version, buildTag string) string {
	basePath := filepath.Join(pm.buildflyBase, "build", depName, version)
	if buildTag != "" && buildTag != "default" {
		return filepath.Join(basePath, buildTag)
	}
	return basePath
}

// GetInstallDir 获取安装目录路径
// 格式：~/.buildfly/install/{name}/{version}/{build_tag}
func (pm *PathManager) GetInstallDir(depName, version, buildTag string) string {
	basePath := filepath.Join(pm.buildflyBase, "install", depName, version)
	if buildTag != "" && buildTag != "default" {
		return filepath.Join(basePath, buildTag)
	}
	return basePath
}

// GetProjectInstallDir 获取项目安装目录路径
// 格式：{project_root}/.buildfly/install/{dep_name}
func (pm *PathManager) GetProjectInstallDir(projectRoot, depName string) string {
	return filepath.Join(projectRoot, ".buildfly", "install", depName)
}

// EnsureDirectory 确保目录存在，如果不存在则创建
func (pm *PathManager) EnsureDirectory(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// StandardizePath 标准化路径（处理 ~ 展开等）
func (pm *PathManager) StandardizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(pm.homeDir, path[2:])
	}
	return filepath.Clean(path)
}

// ExpandPathVariables 展开路径变量
func (pm *PathManager) ExpandPathVariables(path string, variables map[string]string) string {
	// 替换标准变量
	replacements := map[string]string{
		"${HOME}":          pm.homeDir,
		"${USERHOME}":      pm.homeDir,
		"${BUILDFLY}":      pm.buildflyBase,
		"${BUILDFLY_BASE}": pm.buildflyBase,
		"${CACHE}":         pm.buildflyCache,
	}

	for key, value := range replacements {
		path = strings.ReplaceAll(path, key, value)
	}

	// 替换用户定义的变量
	for key, value := range variables {
		// 支持 ${VAR} 和 $VAR 格式
		path = strings.ReplaceAll(path, fmt.Sprintf("${%s}", key), value)
		path = strings.ReplaceAll(path, fmt.Sprintf("$%s", key), value)
	}

	return path
}

// IsSubPath 检查一个路径是否是另一个路径的子路径
func (pm *PathManager) IsSubPath(child, parent string) (bool, error) {
	absChild, err := filepath.Abs(child)
	if err != nil {
		return false, err
	}

	absParent, err := filepath.Abs(parent)
	if err != nil {
		return false, err
	}

	rel, err := filepath.Rel(absParent, absChild)
	if err != nil {
		return false, err
	}

	return !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel), nil
}

// GetRelativePath 获取相对路径
func (pm *PathManager) GetRelativePath(target, base string) (string, error) {
	return filepath.Rel(base, target)
}

// JoinPath 连接路径组件，处理跨平台
func (pm *PathManager) JoinPath(components ...string) string {
	return filepath.Join(components...)
}

// SplitPath 分离路径的目录和文件名
func (pm *PathManager) SplitPath(path string) (dir, file string) {
	return filepath.Split(path)
}

// GetFileExtension 获取文件扩展名
func (pm *PathManager) GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// GetFileName 获取文件名（不含扩展名）
func (pm *PathManager) GetFileName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// GetFileNameWithExt 获取文件名（含扩展名）
func (pm *PathManager) GetFileNameWithExt(path string) string {
	return filepath.Base(path)
}

// PathExists 检查路径是否存在
func (pm *PathManager) PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile 检查路径是否是文件
func (pm *PathManager) IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDirectory 检查路径是否是目录
func (pm *PathManager) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetUniqueDirPath 获取唯一的目录路径（避免冲突）
func (pm *PathManager) GetUniqueDirPath(basePath string) string {
	if !pm.PathExists(basePath) {
		return basePath
	}

	counter := 1
	for {
		uniquePath := fmt.Sprintf("%s_%d", basePath, counter)
		if !pm.PathExists(uniquePath) {
			return uniquePath
		}
		counter++
	}
}

// CleanPath 清理路径（移除冗余的分隔符等）
func (pm *PathManager) CleanPath(path string) string {
	return filepath.Clean(path)
}

// GetCommonPrefix 获取多个路径的公共前缀
func (pm *PathManager) GetCommonPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	prefix := paths[0]
	for _, path := range paths[1:] {
		prefix = getCommonPrefix(prefix, path)
		if prefix == "" {
			break
		}
	}

	return prefix
}

// getCommonPrefix 获取两个路径的公共前缀
func getCommonPrefix(path1, path2 string) string {
	// 标准化路径
	path1 = filepath.Clean(path1)
	path2 = filepath.Clean(path2)

	// 分割路径组件
	components1 := strings.Split(path1, string(filepath.Separator))
	components2 := strings.Split(path2, string(filepath.Separator))

	var commonComponents []string
	minLen := len(components1)
	if len(components2) < minLen {
		minLen = len(components2)
	}

	for i := 0; i < minLen; i++ {
		if components1[i] == components2[i] {
			commonComponents = append(commonComponents, components1[i])
		} else {
			break
		}
	}

	if len(commonComponents) == 0 {
		return ""
	}

	return filepath.Join(commonComponents...)
}

// ValidatePath 验证路径是否合法
func (pm *PathManager) ValidatePath(path string) error {
	// 检查路径是否包含非法字符
	illegalChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range illegalChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains illegal character: %s", char)
		}
	}

	// 检查路径长度
	if len(path) > 260 { // Windows MAX_PATH 限制
		return fmt.Errorf("path too long (max 260 characters): %s", path)
	}

	return nil
}

// GetTempDir 获取临时目录
func (pm *PathManager) GetTempDir() string {
	return os.TempDir()
}

// GetTempDirWithPrefix 获取带前缀的临时目录
func (pm *PathManager) GetTempDirWithPrefix(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}
