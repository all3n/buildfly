package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"buildfly/pkg/config"
	"buildfly/pkg/downloader"
)

func main() {
	// 创建下载管理器
	dm := downloader.NewDownloadManager(3)

	// 创建测试依赖
	dep := config.Dependency{
		Name: "test-file.tar.gz",
		Source: config.SourceInfo{
			Type: "archive",
			URLS: []string{"https://github.com/schollz/progressbar/archive/refs/tags/v3.18.0.tar.gz"},
		},
	}

	// 创建目标目录
	targetDir := "./test-download"
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(targetDir)

	// 使用进度条回调下载
	callback := downloader.CreateProgressBarCallback(dep.Name)

	fmt.Println("开始下载测试文件...")
	ctx := context.Background()

	start := time.Now()
	err := dm.DownloadWithProgress(ctx, dep, targetDir, callback)
	if err != nil {
		log.Fatalf("下载失败: %v", err)
	}

	fmt.Printf("下载完成，耗时: %v\n", time.Since(start))

	// 验证下载的文件
	archivePath := fmt.Sprintf("%s/%s", targetDir, dep.Name)
	if err := dm.Verify(dep, archivePath); err != nil {
		fmt.Printf("验证失败: %v\n", err)
	} else {
		fmt.Println("文件验证成功")
	}
}
