package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"buildfly/internal/errors"
	"buildfly/pkg/config"

	"github.com/schollz/progressbar/v3"
)

// Downloader 下载器接口
type Downloader interface {
	Download(ctx context.Context, dep config.Dependency, targetDir string, callback ProgressCallback) error
	Verify(dep config.Dependency, filePath string) error
}

// DownloadManager 下载管理器
type DownloadManager struct {
	client      *http.Client
	downloaders map[string]Downloader
	semaphore   chan struct{}
	mu          sync.RWMutex
	proxy       *config.ProxyConfig
}

// NewDownloadManager 创建下载管理器
func NewDownloadManager(maxConcurrent int) *DownloadManager {
	return NewDownloadManagerWithProxy(maxConcurrent, nil)
}

// NewDownloadManagerWithProxy 创建带代理配置的下载管理器
func NewDownloadManagerWithProxy(maxConcurrent int, proxy *config.ProxyConfig) *DownloadManager {
	dm := &DownloadManager{
		client:      createHTTPClient(proxy),
		downloaders: make(map[string]Downloader),
		semaphore:   make(chan struct{}, maxConcurrent),
		proxy:       proxy,
	}

	// 注册下载器
	dm.registerDownloaders()
	return dm
}

// NewDownloadManagerFromConfig 从配置创建下载管理器
func NewDownloadManagerFromConfig(maxConcurrent int, config *config.ProjectConfig) *DownloadManager {
	return NewDownloadManagerWithProxy(maxConcurrent, config.Proxy)
}

// createHTTPClient 创建 HTTP 客户端，支持代理配置
func createHTTPClient(proxy *config.ProxyConfig) *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	if proxy != nil {
		if proxy.HTTP != "" || proxy.HTTPS != "" {
			proxyURL, err := url.Parse(proxy.HTTP)
			if err != nil {
				// 如果解析失败，使用 HTTPS 代理
				if proxy.HTTPS != "" {
					proxyURL, err = url.Parse(proxy.HTTPS)
					if err != nil {
						// 如果都失败，返回不使用代理的客户端
						return client
					}
				} else {
					return client
				}
			}
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return client
}

// registerDownloaders 注册下载器
func (dm *DownloadManager) registerDownloaders() {
	dm.RegisterDownloader("git", &GitDownloader{})
	dm.RegisterDownloader("archive", &ArchiveDownloader{client: dm.client})
	dm.RegisterDownloader("direct", &DirectDownloader{client: dm.client})
}

// RegisterDownloader 注册下载器
func (dm *DownloadManager) RegisterDownloader(sourceType string, downloader Downloader) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.downloaders[sourceType] = downloader
}

// Download 下载单个依赖
func (dm *DownloadManager) Download(ctx context.Context, dep config.Dependency, targetDir string) error {
	return dm.DownloadWithProgress(ctx, dep, targetDir, nil)
}

// DownloadWithProgress 下载单个依赖（带进度回调）
func (dm *DownloadManager) DownloadWithProgress(ctx context.Context, dep config.Dependency, targetDir string, callback ProgressCallback) error {
	// 获取信号量
	dm.semaphore <- struct{}{}
	defer func() { <-dm.semaphore }()

	downloader, err := dm.getDownloader(dep.Source.Type)
	if err != nil {
		return err
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to create target directory %s", targetDir))
	}

	// 执行下载
	if err := downloader.Download(ctx, dep, targetDir, callback); err != nil {
		return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to download %s", dep.Name))
	}

	return nil
}

// DownloadAll 并发下载多个依赖
func (dm *DownloadManager) DownloadAll(ctx context.Context, deps []config.Dependency, targetDir string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(deps))

	for _, dep := range deps {
		wg.Add(1)
		go func(dep config.Dependency) {
			defer wg.Done()

			depTargetDir := filepath.Join(targetDir, dep.Name)
			if err := dm.Download(ctx, dep, depTargetDir); err != nil {
				errChan <- err
			}
		}(dep)
	}

	// 等待所有下载完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var downloadErrors []error
	for err := range errChan {
		downloadErrors = append(downloadErrors, err)
	}

	if len(downloadErrors) > 0 {
		return errors.DownloadError(fmt.Sprintf("download failed for %d dependencies", len(downloadErrors)))
	}

	return nil
}

// Verify 验证下载的文件
func (dm *DownloadManager) Verify(dep config.Dependency, filePath string) error {
	downloader, err := dm.getDownloader(dep.Source.Type)
	if err != nil {
		return err
	}

	return downloader.Verify(dep, filePath)
}

// getDownloader 获取下载器
func (dm *DownloadManager) getDownloader(sourceType string) (Downloader, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	downloader, exists := dm.downloaders[sourceType]
	if !exists {
		return nil, errors.DownloadError(fmt.Sprintf("unsupported source type: %s", sourceType))
	}

	return downloader, nil
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	TotalBytes      int64
	DownloadedBytes int64
	Speed           int64
	ETA             time.Duration
}

// ProgressCallback 进度回调函数
type ProgressCallback func(progress DownloadProgress)

// ProgressWriter 进度写入器
type ProgressWriter struct {
	total     int64
	written   int64
	callback  ProgressCallback
	startTime time.Time
	lastTime  time.Time
}

// NewProgressWriter 创建进度写入器
func NewProgressWriter(total int64, callback ProgressCallback) *ProgressWriter {
	return &ProgressWriter{
		total:     total,
		callback:  callback,
		startTime: time.Now(),
		lastTime:  time.Now(),
	}
}

// Write 实现 io.Writer 接口
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pw.written += int64(n)

	// 限制回调频率（每秒最多一次）
	now := time.Now()
	if now.Sub(pw.lastTime) >= time.Second {
		progress := pw.calculateProgress()
		if pw.callback != nil {
			pw.callback(progress)
		}
		pw.lastTime = now
	}

	return
}

// calculateProgress 计算进度
func (pw *ProgressWriter) calculateProgress() DownloadProgress {
	elapsed := time.Since(pw.startTime)
	speed := pw.written / int64(elapsed.Seconds())

	var eta time.Duration
	if speed > 0 {
		remaining := pw.total - pw.written
		eta = time.Duration(float64(remaining)/float64(speed)) * time.Second
	}

	return DownloadProgress{
		TotalBytes:      pw.total,
		DownloadedBytes: pw.written,
		Speed:           speed,
		ETA:             eta,
	}
}

// downloadFile 下载文件（带进度）
func downloadFile(client *http.Client, url string, targetPath string, callback ProgressCallback) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.DownloadError(fmt.Sprintf("HTTP error: %d", resp.StatusCode))
	}

	// 创建目标文件
	file, err := os.Create(targetPath)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

// isValidURL 检查 URL 是否有效
func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "git@")
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
		return filename
	}
	return "download"
}

// ProgressBarDownloader 带进度条的下载器
type ProgressBarDownloader struct {
	bar *progressbar.ProgressBar
}

// NewProgressBarDownloader 创建进度条下载器
func NewProgressBarDownloader(total int64, description string) *ProgressBarDownloader {
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	return &ProgressBarDownloader{bar: bar}
}

// Write 实现 io.Writer 接口
func (pbd *ProgressBarDownloader) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pbd.bar.Add64(int64(n))
	return
}

// Close 关闭进度条
func (pbd *ProgressBarDownloader) Close() error {
	pbd.bar.Finish()
	return nil
}

// CreateProgressBarCallback 创建进度条回调函数
func CreateProgressBarCallback(depName string) ProgressCallback {
	var pbd *ProgressBarDownloader

	return func(progress DownloadProgress) {
		if pbd == nil {
			// 第一次调用时创建进度条
			description := fmt.Sprintf("Downloading %s", depName)
			pbd = NewProgressBarDownloader(progress.TotalBytes, description)
		}

		// 更新进度条
		current := int64(pbd.bar.State().CurrentBytes)
		toAdd := progress.DownloadedBytes - current
		if toAdd > 0 {
			pbd.bar.Add64(toAdd)
		}

		// 如果下载完成，关闭进度条
		if progress.DownloadedBytes >= progress.TotalBytes {
			pbd.Close()
		}
	}
}

// DownloadWithProgressBar 使用进度条下载文件
func DownloadWithProgressBar(client *http.Client, url, targetPath, description string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.DownloadError(fmt.Sprintf("HTTP error: %d", resp.StatusCode))
	}

	// 创建目标文件
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建进度条
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	// 创建多写入器
	writer := io.MultiWriter(file, bar)

	// 复制数据
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	bar.Finish()
	return nil
}
