package downloader

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"buildfly/internal/errors"
	"buildfly/pkg/config"
	"buildfly/pkg/utils"
)

// DirectDownloader 直接文件下载器
type DirectDownloader struct {
	client *http.Client
}

// Download 直接下载文件
func (dd *DirectDownloader) Download(ctx context.Context, dep config.Dependency, targetDir string, callback ProgressCallback) error {
	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to create target directory %s", targetDir))
	}

	// 获取所有可用的 URL 并尝试下载
	urls := dep.Source.GetAllAvailableURLs()
	var lastErr error

	for i, url := range urls {
		fmt.Printf("Attempting to download from URL %d/%d: %s\n", i+1, len(urls), url)

		// 从 URL 获取文件名
		filename := getFileNameFromURL(url)
		if filename == "" || filename == "download" {
			filename = dep.Name
		}

		targetPath := filepath.Join(targetDir, filename)

		// 对于本地文件，直接复制
		if dep.Source.IsLocalURL(url) {
			if err := dd.copyLocalFile(url, targetPath); err != nil {
				lastErr = err
				fmt.Printf("Failed to copy local file %s: %v\n", url, err)
				continue
			}
			fmt.Printf("Successfully copied local file: %s\n", url)
			return nil
		}

		// 对于网络文件，使用 HTTP 下载
		if err := dd.downloadFile(ctx, url, targetPath, callback); err != nil {
			lastErr = err
			fmt.Printf("Failed to download from %s: %v\n", url, err)
			continue
		}

		fmt.Printf("Successfully downloaded from: %s\n", url)
		return nil
	}

	// 所有 URL 都失败了
	if lastErr != nil {
		return lastErr
	}
	return errors.DownloadError("no valid URLs found for download")
}

// copyLocalFile 复制本地文件
func (dd *DirectDownloader) copyLocalFile(srcPath, dstPath string) error {
	// 检查源文件是否存在
	if _, err := os.Stat(srcPath); err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("source file not found: %s", srcPath))
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to create destination directory")
	}

	// 复制文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to open source file: %s", srcPath))
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to create destination file: %s", dstPath))
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to copy file")
	}

	return nil
}

// downloadFile 下载文件
func (dd *DirectDownloader) downloadFile(ctx context.Context, url, targetPath string, callback ProgressCallback) error {
	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("invalid URL: %s", url))
	}

	// 发送请求
	resp, err := dd.client.Do(req)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to download %s", url))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.DownloadError(fmt.Sprintf("HTTP error: %d", resp.StatusCode))
	}

	// 创建目标文件
	file, err := os.Create(targetPath)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to create file %s", targetPath))
	}
	defer file.Close()

	// 创建进度写入器
	var writer io.Writer = file
	if callback != nil && resp.ContentLength > 0 {
		progressWriter := NewProgressWriter(resp.ContentLength, callback)
		writer = io.MultiWriter(file, progressWriter)
	}

	// 复制数据
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to write file")
	}

	return nil
}

// Verify 验证下载的文件
func (dd *DirectDownloader) Verify(dep config.Dependency, filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return errors.DownloadError(fmt.Sprintf("file not found: %s", filePath))
	}

	// 如果指定了哈希，验证文件完整性
	if dep.Source.Hash != "" {
		if err := dd.verifyChecksum(filePath, dep.Source.Hash); err != nil {
			return err
		}
	}

	return nil
}

// verifyChecksum 验证文件校验和
func (dd *DirectDownloader) verifyChecksum(filePath, expectedHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to open file for checksum verification")
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to calculate file checksum")
	}

	actualHash := fmt.Sprintf("%x", hasher.Sum(nil))
	if actualHash != expectedHash {
		return errors.DownloadError(fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedHash, actualHash))
	}

	return nil
}

// isDirectURL 检查是否为直接下载链接
func (dd *DirectDownloader) isDirectURL(url string) bool {
	// 检查是否为常见的直接下载链接
	directPatterns := []string{
		".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar.xz", ".txz",
		".tar", ".zip", ".gz", ".bz2", ".xz",
		".h", ".hpp", ".c", ".cpp", ".cc", ".cxx",
		".o", ".a", ".so", ".dll", ".dylib",
		".txt", ".md", ".readme",
	}

	url = strings.ToLower(url)
	for _, pattern := range directPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}

	// 检查是否为已知的文件下载服务
	downloadHosts := []string{
		"github.com", "raw.githubusercontent.com",
		"gitlab.com", "bitbucket.org",
		"sourceforge.net", "releases.hashicorp.com",
	}

	for _, host := range downloadHosts {
		if strings.Contains(url, host) {
			return true
		}
	}

	return false
}

// getFileType 获取文件类型
func (dd *DirectDownloader) getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".h", ".hpp", ".hxx":
		return "header"
	case ".c":
		return "c-source"
	case ".cpp", ".cc", ".cxx":
		return "cpp-source"
	case ".o":
		return "object"
	case ".a":
		return "static-lib"
	case ".so":
		return "shared-lib"
	case ".dll":
		return "dll"
	case ".dylib":
		return "dylib"
	case ".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar.xz", ".txz", ".tar", ".zip":
		return "archive"
	case ".gz", ".bz2", ".xz":
		return "compressed"
	case ".txt", ".md":
		return "text"
	default:
		return "unknown"
	}
}

// shouldExtract 检查是否需要解压
func (dd *DirectDownloader) shouldExtract(filename string) bool {
	fileType := dd.getFileType(filename)
	return fileType == "archive" || fileType == "compressed"
}

// extractIfNeeded 如果需要则解压文件
func (dd *DirectDownloader) extractIfNeeded(filePath, targetDir string) error {
	if !dd.shouldExtract(filePath) {
		return nil
	}

	// 创建临时解压目录
	tempDir, err := os.MkdirTemp("", "buildfly-extract-*")
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to create temp extract directory")
	}
	defer os.RemoveAll(tempDir)

	// 使用 archive 下载器解压
	archiveDownloader := &ArchiveDownloader{client: dd.client}
	if err := archiveDownloader.extractArchive(filePath, tempDir); err != nil {
		return err
	}

	// 移动解压后的文件到目标目录
	return dd.moveExtractedFiles(tempDir, targetDir)
}

// moveExtractedFiles 移动解压后的文件
func (dd *DirectDownloader) moveExtractedFiles(sourceDir, targetDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to read extract directory")
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(targetDir, entry.Name())

		if entry.IsDir() {
			// 递归复制目录
			if err := utils.CopyDir(sourcePath, targetPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := utils.CopyFile(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}
