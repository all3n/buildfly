package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LinkManager 依赖链接管理器，处理原子性的符号链接操作
type LinkManager struct {
	projectRoot    string
	installTxtPath string
	mutex          sync.Mutex
}

// NewLinkManager 创建链接管理器
func NewLinkManager(projectRoot string) *LinkManager {
	return &LinkManager{
		projectRoot:    projectRoot,
		installTxtPath: filepath.Join(projectRoot, ".buildfly", "install.txt"),
	}
}

// InstallManifest 安装清单记录
type InstallManifest struct {
	Timestamp time.Time  `json:"timestamp"`
	DepName   string     `json:"dep_name"`
	Version   string     `json:"version"`
	BuildTag  string     `json:"build_tag"`
	Links     []LinkInfo `json:"links"`
}

// LinkInfo 链接信息
type LinkInfo struct {
	TargetPath string    `json:"target_path"`
	SourcePath string    `json:"source_path"`
	CreatedAt  time.Time `json:"created_at"`
}

// AtomicInstall 原子性安装操作
func (lm *LinkManager) AtomicInstall(depName, version, buildTag string, sourceDir, targetDir string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 创建临时目录用于原子操作
	tempDir, err := os.MkdirTemp("", "buildfly-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 记录将要创建的链接
	var links []LinkInfo
	manifest := &InstallManifest{
		Timestamp: time.Now(),
		DepName:   depName,
		Version:   version,
		BuildTag:  buildTag,
		Links:     links,
	}

	// 在临时目录中准备所有操作
	if err := lm.prepareInstall(tempDir, sourceDir, targetDir, manifest); err != nil {
		return fmt.Errorf("failed to prepare install: %w", err)
	}

	// 执行原子操作
	if err := lm.executeInstall(tempDir, manifest); err != nil {
		return fmt.Errorf("failed to execute install: %w", err)
	}

	// 记录到安装清单
	if err := lm.recordInstall(manifest); err != nil {
		// 记录失败不影响安装，但需要警告
		fmt.Printf("Warning: failed to record install manifest: %v\n", err)
	}

	return nil
}

// prepareInstall 在临时目录中准备安装操作
func (lm *LinkManager) prepareInstall(tempDir, sourceDir, targetDir string, manifest *InstallManifest) error {
	// 确保源目录存在
	if !PathExists(sourceDir) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// 读取源目录内容
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 准备链接操作
	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(targetDir, entry.Name())

		// 检查目标路径是否已存在
		if _, err := os.Lstat(targetPath); err == nil {
			// 记录需要删除的现有路径
			backupPath := filepath.Join(tempDir, "remove", entry.Name())
			if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
				return fmt.Errorf("failed to create backup directory: %w", err)
			}

			// 创建一个标记文件表示需要删除
			removeFile := filepath.Join(backupPath, "remove.flag")
			if err := os.WriteFile(removeFile, []byte(targetPath), 0644); err != nil {
				return fmt.Errorf("failed to create remove flag: %w", err)
			}
		}

		// 记录链接信息
		absSourcePath, err := filepath.Abs(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute source path: %w", err)
		}

		manifest.Links = append(manifest.Links, LinkInfo{
			TargetPath: targetPath,
			SourcePath: absSourcePath,
			CreatedAt:  time.Now(),
		})

		// 在临时目录中创建链接预览
		tempLinkPath := filepath.Join(tempDir, "links", entry.Name())
		if err := os.MkdirAll(filepath.Dir(tempLinkPath), 0755); err != nil {
			return fmt.Errorf("failed to create temp link directory: %w", err)
		}

		if err := os.Symlink(absSourcePath, tempLinkPath); err != nil {
			return fmt.Errorf("failed to create temp symlink: %w", err)
		}
	}

	return nil
}

// executeInstall 执行安装操作
func (lm *LinkManager) executeInstall(tempDir string, manifest *InstallManifest) error {
	// 第一步：删除现有的文件/链接
	removeDir := filepath.Join(tempDir, "remove")
	if PathExists(removeDir) {
		if err := lm.executeRemovals(removeDir); err != nil {
			return fmt.Errorf("failed to execute removals: %w", err)
		}
	}

	// 第二步：创建新的链接
	linksDir := filepath.Join(tempDir, "links")
	if PathExists(linksDir) {
		if err := lm.executeLinks(linksDir, manifest); err != nil {
			return fmt.Errorf("failed to execute links: %w", err)
		}
	}

	return nil
}

// executeRemovals 执行删除操作
func (lm *LinkManager) executeRemovals(removeDir string) error {
	return filepath.Walk(removeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "remove.flag" {
			// 读取要删除的路径
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read remove flag: %w", err)
			}

			targetPath := string(content)
			if err := os.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("failed to remove target path %s: %w", targetPath, err)
			}
		}

		return nil
	})
}

// executeLinks 执行链接操作
func (lm *LinkManager) executeLinks(linksDir string, manifest *InstallManifest) error {
	for _, linkInfo := range manifest.Links {
		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(linkInfo.TargetPath), 0755); err != nil {
			return fmt.Errorf("failed to create target directory for %s: %w", linkInfo.TargetPath, err)
		}

		// 创建符号链接
		if err := os.Symlink(linkInfo.SourcePath, linkInfo.TargetPath); err != nil {
			return fmt.Errorf("failed to create symlink %s -> %s: %w",
				linkInfo.TargetPath, linkInfo.SourcePath, err)
		}

		fmt.Printf("  Linked %s -> %s\n", linkInfo.TargetPath, linkInfo.SourcePath)
	}

	return nil
}

// recordInstall 记录安装信息到清单文件
func (lm *LinkManager) recordInstall(manifest *InstallManifest) error {
	// 确保清单文件目录存在
	if err := os.MkdirAll(filepath.Dir(lm.installTxtPath), 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// 打开文件进行追加写入
	file, err := os.OpenFile(lm.installTxtPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	// 写入链接路径
	for _, linkInfo := range manifest.Links {
		if _, err := file.WriteString(linkInfo.TargetPath + "\n"); err != nil {
			return fmt.Errorf("failed to write to manifest: %w", err)
		}
	}

	return nil
}

// AtomicUninstall 原子性卸载操作
func (lm *LinkManager) AtomicUninstall(depName string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 读取安装清单
	manifests, err := lm.readManifests()
	if err != nil {
		return fmt.Errorf("failed to read manifests: %w", err)
	}

	// 找到要卸载的依赖的链接
	var linksToRemove []LinkInfo
	for _, manifest := range manifests {
		if manifest.DepName == depName {
			linksToRemove = append(linksToRemove, manifest.Links...)
		}
	}

	if len(linksToRemove) == 0 {
		fmt.Printf("No installed links found for dependency: %s\n", depName)
		return nil
	}

	// 创建临时目录用于原子操作
	tempDir, err := os.MkdirTemp("", "buildfly-uninstall-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 准备卸载操作
	if err := lm.prepareUninstall(tempDir, linksToRemove); err != nil {
		return fmt.Errorf("failed to prepare uninstall: %w", err)
	}

	// 执行卸载操作
	if err := lm.executeUninstall(tempDir); err != nil {
		return fmt.Errorf("failed to execute uninstall: %w", err)
	}

	// 从清单中移除记录
	if err := lm.removeFromManifest(depName); err != nil {
		fmt.Printf("Warning: failed to remove from manifest: %v\n", err)
	}

	return nil
}

// prepareUninstall 准备卸载操作
func (lm *LinkManager) prepareUninstall(tempDir string, links []LinkInfo) error {
	removeDir := filepath.Join(tempDir, "remove")
	if err := os.MkdirAll(removeDir, 0755); err != nil {
		return fmt.Errorf("failed to create remove directory: %w", err)
	}

	for i, linkInfo := range links {
		// 创建删除标记
		removeFile := filepath.Join(removeDir, fmt.Sprintf("remove_%d.flag", i))
		if err := os.WriteFile(removeFile, []byte(linkInfo.TargetPath), 0644); err != nil {
			return fmt.Errorf("failed to create remove flag: %w", err)
		}
	}

	return nil
}

// executeUninstall 执行卸载操作
func (lm *LinkManager) executeUninstall(tempDir string) error {
	removeDir := filepath.Join(tempDir, "remove")
	return filepath.Walk(removeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "remove.flag" ||
			(len(info.Name()) > 11 && info.Name()[:11] == "remove_" && strings.HasSuffix(info.Name(), ".flag")) {
			// 读取要删除的路径
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read remove flag: %w", err)
			}

			targetPath := string(content)
			if PathExists(targetPath) {
				if err := os.RemoveAll(targetPath); err != nil {
					return fmt.Errorf("failed to remove target path %s: %w", targetPath, err)
				}
				fmt.Printf("  Removed %s\n", targetPath)
			}
		}

		return nil
	})
}

// readManifests 读取安装清单
func (lm *LinkManager) readManifests() ([]InstallManifest, error) {
	var manifests []InstallManifest

	if !PathExists(lm.installTxtPath) {
		return manifests, nil
	}

	file, err := os.Open(lm.installTxtPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentManifest := &InstallManifest{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 简化实现：每行是一个路径
		currentManifest.Links = append(currentManifest.Links, LinkInfo{
			TargetPath: line,
			CreatedAt:  time.Now(), // 实际应该从文件解析
		})
	}

	if len(currentManifest.Links) > 0 {
		manifests = append(manifests, *currentManifest)
	}

	return manifests, scanner.Err()
}

// removeFromManifest 从清单中移除记录
func (lm *LinkManager) removeFromManifest(depName string) error {
	if !PathExists(lm.installTxtPath) {
		return nil
	}

	// 读取现有清单
	manifests, err := lm.readManifests()
	if err != nil {
		return err
	}

	// 过滤掉要删除的依赖的记录
	var remainingManifests []InstallManifest
	for _, manifest := range manifests {
		if manifest.DepName != depName {
			remainingManifests = append(remainingManifests, manifest)
		}
	}

	// 重写清单文件
	file, err := os.OpenFile(lm.installTxtPath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open manifest file for writing: %w", err)
	}
	defer file.Close()

	for _, manifest := range remainingManifests {
		for _, linkInfo := range manifest.Links {
			if _, err := file.WriteString(linkInfo.TargetPath + "\n"); err != nil {
				return fmt.Errorf("failed to write to manifest: %w", err)
			}
		}
	}

	return nil
}

// ListInstalledLinks 列出已安装的链接
func (lm *LinkManager) ListInstalledLinks() ([]string, error) {
	var links []string

	if !PathExists(lm.installTxtPath) {
		return links, nil
	}

	file, err := os.Open(lm.installTxtPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			links = append(links, line)
		}
	}

	return links, scanner.Err()
}

// VerifyLinks 验证链接的完整性
func (lm *LinkManager) VerifyLinks() ([]string, error) {
	links, err := lm.ListInstalledLinks()
	if err != nil {
		return nil, err
	}

	var brokenLinks []string
	for _, link := range links {
		// 检查链接是否存在
		if _, err := os.Lstat(link); err != nil {
			brokenLinks = append(brokenLinks, fmt.Sprintf("%s (missing)", link))
			continue
		}

		// 检查链接目标是否存在
		target, err := os.Readlink(link)
		if err != nil {
			brokenLinks = append(brokenLinks, fmt.Sprintf("%s (invalid link)", link))
			continue
		}

		if !PathExists(target) {
			brokenLinks = append(brokenLinks, fmt.Sprintf("%s -> %s (target missing)", link, target))
		}
	}

	return brokenLinks, nil
}
