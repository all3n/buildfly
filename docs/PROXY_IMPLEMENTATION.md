# 代理配置实现文档

## 概述

本文档描述了 buildfly 中代理配置功能的实现细节和使用方法。

## 功能特性

### ✅ 已实现的功能

1. **完整的代理配置支持**
   - HTTP 代理配置
   - HTTPS 代理配置
   - no_proxy 列表支持
   - 支持认证代理（用户名密码）

2. **配置层级支持**
   - 项目本地配置 (`./buildfly.yaml`)
   - 全局配置 (`~/.config/buildfly/config.yaml`)
   - 配置合并和覆盖机制

3. **下载器集成**
   - HTTP/HTTPS 下载器支持代理
   - Git 下载器支持代理
   - 自动环境变量检测

4. **测试覆盖**
   - 单元测试覆盖配置加载
   - 配置合并测试
   - 实际下载测试

## 配置结构

### YAML 配置示例

```yaml
project:
  name: "my-project"
  version: "1.0.0"

# 代理配置
proxy:
  # HTTP 代理服务器
  http: "http://proxy.example.com:8080"
  
  # HTTPS 代理服务器
  https: "https://proxy.example.com:8080"
  
  # 不使用代理的域名列表
  no_proxy:
    - "localhost"
    - "127.0.0.1"
    - "*.local"
    - "*.internal"

dependencies:
  # 依赖配置...
```

### 配置字段说明

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `http` | string | 否 | HTTP 代理服务器地址 |
| `https` | string | 否 | HTTPS 代理服务器地址 |
| `no_proxy` | []string | 否 | 不使用代理的域名列表 |

## 实现细节

### 1. 数据结构

```go
type ProxyConfig struct {
    HTTP    string   `yaml:"http"`     // HTTP 代理服务器
    HTTPS   string   `yaml:"https"`    // HTTPS 代理服务器
    NoProxy []string `yaml:"no_proxy"` // 不使用代理的域名列表
}

type ProjectConfig struct {
    // ... 其他字段
    Proxy *ProxyConfig `yaml:"proxy,omitempty"` // 代理配置
}
```

### 2. 配置加载

配置加载器支持：
- 从 YAML 文件加载代理配置
- 全局和本地配置合并
- 配置验证和默认值

### 3. 下载器集成

#### HTTP/HTTPS 下载器

```go
func (d *Downloader) createHTTPClient() *http.Client {
    // 配置代理
    if d.config.Proxy != nil {
        // 设置 HTTP_PROXY 和 HTTPS_PROXY 环境变量
        // 创建代理 URL
        // 配置 http.Transport
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
}
```

#### Git 下载器

```go
func (gd *GitDownloader) cloneWithProxy(ctx context.Context, repoURL, targetDir string) error {
    // 配置 Git 代理
    if gd.config.Proxy != nil {
        // 设置 git config http.proxy
        // 设置 git config https.proxy
    }
    
    // 执行 git clone
}
```

### 4. 环境变量支持

系统自动检测以下环境变量：
- `HTTP_PROXY`
- `HTTPS_PROXY`
- `NO_PROXY`

如果配置文件中没有设置代理，系统会尝试使用环境变量中的代理设置。

## 使用示例

### 1. 基本代理配置

```yaml
proxy:
  http: "http://proxy.company.com:8080"
  https: "https://proxy.company.com:8080"
```

### 2. 带认证的代理

```yaml
proxy:
  http: "http://username:password@proxy.company.com:8080"
  https: "https://username:password@proxy.company.com:8080"
```

### 3. SOCKS 代理

```yaml
proxy:
  http: "socks5://proxy.company.com:1080"
  https: "socks5://proxy.company.com:1080"
```

### 4. 排除特定域名

```yaml
proxy:
  http: "http://proxy.company.com:8080"
  https: "https://proxy.company.com:8080"
  no_proxy:
    - "localhost"
    - "127.0.0.1"
    - "*.local"
    - "github.com"
```

## 测试验证

### 单元测试

```bash
go test ./pkg/config -v -run TestProxy
```

测试覆盖：
- 配置加载验证
- 配置合并逻辑
- 边界情况处理

### 集成测试

```bash
cd examples/proxy-demo
../../bin/buildfly --config test-real-proxy.yaml install test-lib
```

验证：
- 实际下载功能
- 代理配置生效
- 校验和验证

## 故障排除

### 1. 代理连接失败

**症状**：下载超时或连接被拒绝

**解决方案**：
- 检查代理服务器地址和端口
- 验证代理服务器是否运行
- 检查网络连接

```bash
curl -x http://proxy.company.com:8080 https://www.google.com
```

### 2. 认证失败

**症状**：407 Proxy Authentication Required

**解决方案**：
- 验证用户名和密码
- 检查代理服务器认证配置
- 确保认证信息格式正确

### 3. 特定域名无法访问

**症状**：某些域名下载失败

**解决方案**：
- 检查 `no_proxy` 配置
- 添加不需要代理的域名到 `no_proxy` 列表
- 验证域名匹配规则

### 4. SSL/TLS 问题

**症状**：证书验证失败

**解决方案**：
- 检查代理服务器的 SSL 证书
- 考虑使用 HTTP 代理而不是 HTTPS
- 配置证书信任

## 最佳实践

1. **配置管理**
   - 在全局配置中设置通用代理
   - 在项目配置中覆盖特定设置
   - 使用环境变量进行动态配置

2. **安全考虑**
   - 避免在配置文件中硬编码密码
   - 使用环境变量存储敏感信息
   - 定期更新代理认证信息

3. **性能优化**
   - 合理配置 `no_proxy` 列表
   - 使用本地代理服务器减少延迟
   - 启用缓存机制

## 未来扩展

### 计划中的功能

1. **代理认证增强**
   - 支持更多认证方式（NTLM、Digest）
   - 认证信息加密存储

2. **代理路由**
   - 基于域名的智能代理选择
   - 负载均衡支持

3. **监控和日志**
   - 代理使用统计
   - 详细的连接日志

4. **配置验证**
   - 代理连接测试
   - 配置语法验证

## 总结

代理配置功能已经完全实现并通过测试验证，支持：

- ✅ HTTP/HTTPS 代理配置
- ✅ 认证支持
- ✅ no_proxy 列表
- ✅ 全局和项目配置
- ✅ 环境变量支持
- ✅ 完整的测试覆盖

该功能可以满足企业在网络受限环境中的依赖下载需求，提供了灵活且安全的代理配置方案。
