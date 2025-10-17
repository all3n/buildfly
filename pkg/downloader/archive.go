package downloader

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"buildfly/internal/errors"
	"buildfly/pkg/config"

	"github.com/schollz/progressbar/v3"
)

// ArchiveDownloader 压缩包下载器
type ArchiveDownloader struct {
	client *http.Client
}

// Download 下载并解压压缩包
func (ad *ArchiveDownloader) Download(ctx context.Context, dep config.Dependency, targetDir string, callback ProgressCallback) error {
	// 下载压缩包
	archivePath, err := ad.downloadArchive(ctx, dep, callback)
	if err != nil {
		return err
	}
	// defer os.Remove(archivePath)

	// 验证文件完整性
	if err := ad.Verify(dep, archivePath); err != nil {
		return fmt.Errorf("archive verification failed: %w", err)
	}

	// 解压压缩包
	if err := ad.extractArchive(archivePath, targetDir); err != nil {
		return err
	}

	return nil
}

// downloadArchive 下载压缩包
func (ad *ArchiveDownloader) downloadArchive(ctx context.Context, dep config.Dependency, callback ProgressCallback) (string, error) {
	// 获取第一个可用的 URL
	url, err := dep.Source.GetFirstAvailableURL()
	if err != nil {
		return "", errors.DownloadErrorWithCause(err, "failed to get available URL")
	}

	// 从 URL 中提取完整的文件扩展名
	ext := ad.getFullExtension(url)
	if ext == "" {
		ext = ".tar.gz" // 默认扩展名
	}

	// ~/.cache/buildfly/archives/
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", errors.DownloadErrorWithCause(err, "failed to get user cache dir")
	}
	cacheDir := filepath.Join(userCacheDir, ".buildfly", "archives", dep.Version)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", errors.DownloadErrorWithCause(err, "failed to create cache dir")
	}
	fileBaseName := filepath.Base(url)
	cacheFile := filepath.Join(cacheDir, fileBaseName)

	// 检查缓存文件
	if _, err := os.Stat(cacheFile); err == nil {
		if err := ad.Verify(dep, cacheFile); err != nil {
			fmt.Printf("Cached archive verification failed, re-downloading: %s\n", cacheFile)
		} else {
			// 文件存在且验证通过，直接返回缓存文件路径
			fmt.Printf("Using cached archive: %s\n", cacheFile)
			return cacheFile, nil
		}
	}

	// 尝试从多个 URL 下载
	urls := dep.Source.GetAllAvailableURLs()
	var lastErr error

	for i, downloadURL := range urls {
		fmt.Printf("Attempting to download from URL %d/%d: %s\n", i+1, len(urls), downloadURL)

		// 对于本地文件，直接复制到缓存位置
		if dep.Source.IsLocalURL(downloadURL) {
			if err := ad.copyLocalFile(downloadURL, cacheFile); err != nil {
				lastErr = err
				fmt.Printf("Failed to copy local file %s: %v\n", downloadURL, err)
				continue
			}
			fmt.Printf("Successfully copied local file: %s\n", downloadURL)
			return cacheFile, nil
		}

		// 对于网络文件，使用 HTTP 下载
		if err := ad.downloadFromHTTP(ctx, downloadURL, cacheFile, dep.Name, callback); err != nil {
			lastErr = err
			fmt.Printf("Failed to download from %s: %v\n", downloadURL, err)
			continue
		}

		fmt.Printf("Successfully downloaded from: %s\n", downloadURL)
		return cacheFile, nil
	}

	// 所有 URL 都失败了
	if lastErr != nil {
		return "", errors.DownloadErrorWithCause(lastErr, "all URLs failed to download")
	}
	return "", errors.DownloadError("no valid URLs found for download")
}

// copyLocalFile 复制本地文件
func (ad *ArchiveDownloader) copyLocalFile(srcPath, dstPath string) error {
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

// downloadFromHTTP 从 HTTP URL 下载文件
func (ad *ArchiveDownloader) downloadFromHTTP(ctx context.Context, url, cacheFile, depName string, callback ProgressCallback) error {
	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("invalid URL: %s", url))
	}

	// 发送请求
	resp, err := ad.client.Do(req)
	if err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to download %s", url))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.DownloadError(fmt.Sprintf("HTTP error: %d for URL %s", resp.StatusCode, url))
	}

	// 创建临时文件以确保下载完整性
	tempFile := cacheFile + ".tmp"
	if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
		return errors.DownloadErrorWithCause(err, "failed to remove existing temp file")
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to create temp file")
	}
	defer file.Close()

	// 创建进度写入器
	var writer io.Writer = file
	if callback != nil && resp.ContentLength > 0 {
		progressWriter := NewProgressWriter(resp.ContentLength, callback)
		writer = io.MultiWriter(file, progressWriter)
	} else {
		// 如果没有提供回调函数，使用默认进度条
		description := fmt.Sprintf("Downloading %s", depName)
		bar := progressbar.NewOptions64(
			resp.ContentLength,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
		)
		writer = io.MultiWriter(file, bar)
	}

	// 下载文件
	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		os.Remove(tempFile) // 清理失败的下载
		return errors.DownloadErrorWithCause(err, "failed to write archive file")
	}

	// 验证下载完整性
	if resp.ContentLength > 0 && written != resp.ContentLength {
		os.Remove(tempFile) // 清理不完整的下载
		return errors.DownloadError(fmt.Sprintf("download incomplete: expected %d bytes, got %d bytes", resp.ContentLength, written))
	}

	// 重命名临时文件到最终位置
	if err := os.Rename(tempFile, cacheFile); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return errors.DownloadErrorWithCause(err, "failed to move downloaded file to final location")
	}

	return nil
}

// getFullExtension 获取完整的文件扩展名（包括多部分扩展名）
func (ad *ArchiveDownloader) getFullExtension(url string) string {
	// 从 URL 中提取文件名
	filename := filepath.Base(url)
	lowerFilename := strings.ToLower(filename)

	// 检查多部分扩展名
	suffixes := []string{".tar.gz", ".tar.bz2", ".tar.xz", ".tbz2", ".txz", ".tgz"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(lowerFilename, suffix) {
			return suffix
		}
	}

	// 如果不是多部分扩展名，返回单个扩展名
	return filepath.Ext(filename)
}

// extractArchive 解压压缩包
func (ad *ArchiveDownloader) extractArchive(archivePath, targetDir string) error {
	// 根据文件扩展名选择解压方法
	lowerPath := strings.ToLower(archivePath)

	switch {
	case strings.HasSuffix(lowerPath, ".tar.gz") || strings.HasSuffix(lowerPath, ".tgz"):
		return ad.extractTarGz(archivePath, targetDir)
	case strings.HasSuffix(lowerPath, ".tar.bz2") || strings.HasSuffix(lowerPath, ".tbz2"):
		return ad.extractTarBz2(archivePath, targetDir)
	case strings.HasSuffix(lowerPath, ".tar.xz") || strings.HasSuffix(lowerPath, ".txz"):
		return ad.extractTarXz(archivePath, targetDir)
	case strings.HasSuffix(lowerPath, ".tar"):
		return ad.extractTar(archivePath, targetDir)
	case strings.HasSuffix(lowerPath, ".zip"):
		return ad.extractZip(archivePath, targetDir)
	case strings.HasSuffix(lowerPath, ".gz"):
		return errors.DownloadError("standalone .gz files are not supported, please use .tar.gz")
	case strings.HasSuffix(lowerPath, ".bz2"):
		return errors.DownloadError("standalone .bz2 files are not supported, please use .tar.bz2")
	case strings.HasSuffix(lowerPath, ".xz"):
		return errors.DownloadError("standalone .xz files are not supported, please use .tar.xz")
	default:
		return errors.DownloadError(fmt.Sprintf("unsupported archive format: %s", filepath.Ext(archivePath)))
	}
}

// extractTarGz 解压 tar.gz 文件
func (ad *ArchiveDownloader) extractTarGz(archivePath, targetDir string) error {
	// 先检查文件类型
	fileCmd := execCommand("file", archivePath)
	output, err := fileCmd.CombinedOutput()
	if err == nil {
		fmt.Printf("Archive file type: %s\n", string(output))
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to create target directory")
	}

	// 使用 --preserve-permissions 参数保留文件权限
	cmd := execCommand("tar", "xzf", archivePath, "-C", targetDir, "--strip-components=1", "--preserve-permissions")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to extract tar.gz archive")
	}
	return nil
}

// extractTar 解压 tar 文件
func (ad *ArchiveDownloader) extractTar(archivePath, targetDir string) error {
	// 使用 --preserve-permissions 参数保留文件权限
	cmd := execCommand("tar", "xf", archivePath, "-C", targetDir, "--strip-components=1", "--preserve-permissions")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to extract tar archive")
	}
	return nil
}

// extractZip 解压 zip 文件
func (ad *ArchiveDownloader) extractZip(archivePath, targetDir string) error {
	// 使用 -o 参数保留文件权限，-q 静默模式
	cmd := execCommand("unzip", "-q", "-o", archivePath, "-d", targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to extract zip archive")
	}
	return nil
}

// extractTarBz2 解压 tar.bz2 文件
func (ad *ArchiveDownloader) extractTarBz2(archivePath, targetDir string) error {
	// 使用 --preserve-permissions 参数保留文件权限
	cmd := execCommand("tar", "xjf", archivePath, "-C", targetDir, "--strip-components=1", "--preserve-permissions")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to extract tar.bz2 archive")
	}
	return nil
}

// extractTarXz 解压 tar.xz 文件
func (ad *ArchiveDownloader) extractTarXz(archivePath, targetDir string) error {
	// 使用 --preserve-permissions 参数保留文件权限
	cmd := execCommand("tar", "xJf", archivePath, "-C", targetDir, "--strip-components=1", "--preserve-permissions")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to extract tar.xz archive")
	}
	return nil
}

// Verify 验证压缩包
func (ad *ArchiveDownloader) Verify(dep config.Dependency, archivePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(archivePath); err != nil {
		return errors.DownloadError(fmt.Sprintf("archive file not found: %s", archivePath))
	}

	// 获取所有可用的校验和
	checksums := ad.getChecksums(dep.Source)

	if len(checksums) == 0 {
		fmt.Printf("  No checksum specified for verification\n")
		return nil
	}

	// 验证每个指定的校验和
	for algorithm, expectedHash := range checksums {
		if err := ad.verifyChecksumWithAlgorithm(archivePath, expectedHash, algorithm); err != nil {
			return fmt.Errorf("checksum verification failed (%s): %w", algorithm, err)
		}
		fmt.Printf("  ✓ Verified with %s checksum\n", algorithm)
	}

	return nil
}

// getChecksums 获取所有可用的校验和
func (ad *ArchiveDownloader) getChecksums(source config.SourceInfo) map[string]string {
	checksums := make(map[string]string)

	// 优先级：专用字段 > 通用映射 > 向后兼容的 hash 字段

	// 检查专用字段
	if source.MD5 != "" {
		checksums["md5"] = strings.ToLower(source.MD5)
	}
	if source.SHA1 != "" {
		checksums["sha1"] = strings.ToLower(source.SHA1)
	}
	if source.SHA256 != "" {
		checksums["sha256"] = strings.ToLower(source.SHA256)
	}
	if source.SHA512 != "" {
		checksums["sha512"] = strings.ToLower(source.SHA512)
	}

	// 检查通用映射
	if source.Checksums != nil {
		for alg, hash := range source.Checksums {
			checksums[strings.ToLower(alg)] = strings.ToLower(hash)
		}
	}

	// 向后兼容：hash 字段默认为 SHA256
	if source.Hash != "" && checksums["sha256"] == "" {
		checksums["sha256"] = strings.ToLower(source.Hash)
	}

	return checksums
}

// verifyChecksumWithAlgorithm 使用指定算法验证文件校验和
func (ad *ArchiveDownloader) verifyChecksumWithAlgorithm(filePath, expectedHash, algorithm string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to open file for checksum verification")
	}
	defer file.Close()

	var hasher hash.Hash

	// 根据算法选择对应的哈希函数
	switch strings.ToLower(algorithm) {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return errors.DownloadError(fmt.Sprintf("unsupported checksum algorithm: %s", algorithm))
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return errors.DownloadErrorWithCause(err, "failed to calculate file checksum")
	}

	actualHash := fmt.Sprintf("%x", hasher.Sum(nil))
	if actualHash != expectedHash {
		return errors.DownloadError(fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedHash, actualHash))
	}

	return nil
}

// verifyChecksum 验证文件校验和（保持向后兼容）
func (ad *ArchiveDownloader) verifyChecksum(filePath, expectedHash string) error {
	return ad.verifyChecksumWithAlgorithm(filePath, expectedHash, "sha256")
}

// execCommand 执行命令的辅助函数
func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// isArchiveSupported 检查是否支持该压缩格式
func (ad *ArchiveDownloader) isArchiveSupported(filename string) bool {
	supportedFormats := []string{
		".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar.xz", ".txz",
		".tar", ".zip", ".gz", ".bz2", ".xz",
	}

	filename = strings.ToLower(filename)
	for _, format := range supportedFormats {
		if strings.HasSuffix(filename, format) {
			return true
		}
	}
	return false
}

// getArchiveType 获取压缩包类型
func (ad *ArchiveDownloader) getArchiveType(filename string) string {
	filename = strings.ToLower(filename)

	switch {
	case strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(filename, ".tar.bz2") || strings.HasSuffix(filename, ".tbz2"):
		return "tar.bz2"
	case strings.HasSuffix(filename, ".tar.xz") || strings.HasSuffix(filename, ".txz"):
		return "tar.xz"
	case strings.HasSuffix(filename, ".tar"):
		return "tar"
	case strings.HasSuffix(filename, ".zip"):
		return "zip"
	case strings.HasSuffix(filename, ".gz"):
		return "gz"
	case strings.HasSuffix(filename, ".bz2"):
		return "bz2"
	case strings.HasSuffix(filename, ".xz"):
		return "xz"
	default:
		return "unknown"
	}
}
