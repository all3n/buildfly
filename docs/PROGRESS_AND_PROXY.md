# 下载进度回调和代理支持

本文档介绍了 Buildfly 下载器新增的进度回调和代理支持功能。

## 功能概述

### 进度回调
- 支持实时显示下载进度
- 显示下载速度、已下载字节数、总字节数
- 预估剩余时间（ETA）
- 支持 Archive 和 Direct 下载器

### 代理支持
- 支持 HTTP/HTTPS 代理
- 自动从环境变量读取代理配置
- 支持代理绕过列表（NO_PROXY）
- 可通过代码配置代理

## 使用方法

### 1. 基本使用

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "buildfly/pkg/downloader"
    "buildfly/pkg/config"
)

func main() {
    // 创建下载管理器
    dm := downloader.NewDownloadManager(5) // 最大并发数 5
    
    // 创建依赖配置
    dep := config.Dependency{
        Name: "example",
        Version: "1.0.0",
        Source: config.SourceInfo{
            Type: "archive",
            URL: "https://example.com/file.tar.gz",
        },
    }
    
    // 定义进度回调函数
    progressCallback := func(progress downloader.DownloadProgress) {
        percent := float64(progress.DownloadedBytes) / float64(progress.TotalBytes) * 100
        fmt.Printf("\r下载进度: %.2f%% (%d/%d bytes) 速度: %d bytes/s ETA: %v",
            percent, progress.DownloadedBytes, progress.TotalBytes, 
            progress.Speed, progress.ETA)
    }
    
    // 带进度下载
    err := dm.DownloadWithProgress(context.Background(), dep, "./target", progressCallback)
    if err != nil {
        fmt.Printf("下载失败: %v\n", err)
        return
    }
    
    fmt.Println("\n下载完成!")
}
```

### 2. 代理配置

#### 环境变量方式
```bash
# 设置 HTTP 代理
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080

# 设置代理绕过列表
export NO_PROXY=localhost,127.0.0.1,.example.com

# 运行 buildfly
./buildfly install
```

#### 代码配置方式
```go
package main

import (
    "buildfly/pkg/downloader"
)

func main() {
    // 创建代理配置
    proxy := &downloader.ProxyConfig{
        HTTP:  "http://proxy.example.com:8080",
        HTTPS: "http://proxy.example.com:8080",
        NoProxy: []string{"localhost", "127.0.0.1", ".example.com"},
    }
    
    // 创建带代理的下载管理器
    dm := downloader.NewDownloadManagerWithProxy(5, proxy)
    
    // 使用下载管理器...
}
```

## API 参考

### ProxyConfig 结构体

```go
type ProxyConfig struct {
    HTTP    string   `yaml:"http" json:"http"`       // HTTP 代理 URL
    HTTPS   string   `yaml:"https" json:"https"`     // HTTPS 代理 URL
    NoProxy []string `yaml:"no_proxy" json:"no_proxy"` // 代理绕过列表
}
```

### DownloadProgress 结构体

```go
type DownloadProgress struct {
    TotalBytes      int64         // 总字节数
    DownloadedBytes int64         // 已下载字节数
    Speed           int64         // 下载速度（字节/秒）
    ETA             time.Duration // 预估剩余时间
}
```

### ProgressCallback 类型

```go
type ProgressCallback func(progress DownloadProgress)
```

### DownloadManager 方法

```go
// 创建下载管理器
func NewDownloadManager(maxConcurrent int) *DownloadManager

// 创建带代理配置的下载管理器
func NewDownloadManagerWithProxy(maxConcurrent int, proxy *ProxyConfig) *DownloadManager

// 基本下载（无进度回调）
func (dm *DownloadManager) Download(ctx context.Context, dep config.Dependency, targetDir string) error

// 带进度回调的下载
func (dm *DownloadManager) DownloadWithProgress(ctx context.Context, dep config.Dependency, targetDir string, callback ProgressCallback) error
```

## 支持的下载器

### ArchiveDownloader
- ✅ 支持进度回调
- ✅ 支持代理
- 支持格式：tar.gz, tar.bz2, tar.xz, tar, zip

### DirectDownloader  
- ✅ 支持进度回调
- ✅ 支持代理
- 支持直接文件下载

### GitDownloader
- ❌ 不支持进度回调（Git 操作本身不提供进度信息）
- ✅ 支持代理（通过 Git 配置）
- 支持 Git 仓库克隆

## 配置示例

### buildfly.yaml 配置

```yaml
project:
  name: "my-project"
  version: "1.0.0"

dependencies:
  # 大文件下载（显示进度）
  boost:
    version: "1.89.0"
    source:
      type: "archive"
      url: "https://archives.boost.io/release/1.89.0/source/boost_1_89_0.tar.gz"
      sha256: "6e808e01ed7bc6e928f158511352e415c23f9339d8795e651b9e8e226e5486c5"
    
  # Git 仓库下载
  fmt:
    version: "10.2.1"
    source:
      type: "git"
      url: "https://github.com/fmtlib/fmt.git"
      tag: "10.2.1"

  # 直接文件下载
  zlib:
    version: "1.3.1"
    source:
      type: "direct"
      url: "https://zlib.net/zlib-1.3.1.tar.gz"
      sha256: "9a93b2b7dfdac76ce66290e712c63ffcaef1b341dcaf6bb836c2f5b0660615e2"
```

## 实现细节

### 进度回调实现
- 使用 `ProgressWriter` 包装 `io.Writer`
- 限制回调频率（每秒最多一次）避免性能问题
- 自动计算下载速度和预估时间

### 代理实现
- 使用标准库 `http.Transport` 的 `Proxy` 功能
- 自动解析代理 URL
- 支持 HTTP 和 HTTPS 代理
- 处理代理解析错误

### 错误处理
- 代理配置错误时自动回退到直连
- 进度回调错误不影响下载过程
- 提供详细的错误信息

## 性能考虑

1. **回调频率限制**：进度回调每秒最多触发一次，避免频繁回调影响性能
2. **内存使用**：进度计算使用固定内存，不会随文件大小增长
3. **并发下载**：支持多个文件并发下载，每个文件独立显示进度
4. **代理连接池**：HTTP 客户端自动管理连接池

## 故障排除

### 代理不工作
1. 检查代理 URL 格式是否正确
2. 确认代理服务器是否可访问
3. 检查 NO_PROXY 配置是否正确
4. 查看错误日志获取详细信息

### 进度显示异常
1. 确认服务器返回 Content-Length 头
2. 检查回调函数是否正确处理进度数据
3. 大文件下载时进度更新可能有延迟

### 下载速度慢
1. 检查网络连接
2. 尝试禁用代理测试直连速度
3. 调整并发下载数量
4. 检查磁盘 I/O 性能

## 示例代码

完整的使用示例请参考：
- `examples/progress-demo/buildfly.yaml` - 配置示例
- `examples/progress-demo/test-progress.sh` - 演示脚本

## 更新日志

- **v1.0.0** - 初始版本，支持基本的进度回调和代理功能
- 支持 Archive 和 Direct 下载器的进度显示
- 支持 HTTP/HTTPS 代理配置
- 添加进度回调频率限制
- 完善错误处理和日志记录
