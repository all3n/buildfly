# 代理配置演示

本示例展示了如何在 buildfly 中配置和使用代理服务器来下载依赖。

## 代理配置说明

### 配置结构

```yaml
project:
  name: "your-project"
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
  # 你的依赖配置...
```

### 配置字段说明

- **http**: HTTP 代理服务器地址
- **https**: HTTPS 代理服务器地址
- **no_proxy**: 不使用代理的域名或 IP 地址列表

### 常见代理配置示例

#### 1. 公司代理服务器

```yaml
proxy:
  http: "http://proxy.company.com:8080"
  https: "https://proxy.company.com:8080"
  no_proxy:
    - "localhost"
    - "127.0.0.1"
    - "*.local"
    - "git.company.com"
```

#### 2. 需要认证的代理

```yaml
proxy:
  http: "http://username:password@proxy.company.com:8080"
  https: "https://username:password@proxy.company.com:8080"
```

#### 3. SOCKS 代理

```yaml
proxy:
  http: "socks5://proxy.company.com:1080"
  https: "socks5://proxy.company.com:1080"
```

#### 4. 仅 HTTPS 代理

```yaml
proxy:
  https: "https://proxy.company.com:8080"
  no_proxy:
    - "github.com"
    - "gitlab.com"
```

## 使用方法

### 1. 在项目配置中添加代理设置

编辑你的 `buildfly.yaml` 文件，添加 `proxy` 配置段。

### 2. 运行安装命令

```bash
# 安装所有依赖（使用配置的代理）
buildfly install

# 安装特定依赖
buildfly install fmt zlib
```

### 3. 验证代理配置

buildfly 会在下载依赖时自动使用配置的代理服务器。你可以在日志中看到下载过程。

## 全局配置

你也可以在全局配置文件中设置代理，这样所有项目都会使用相同的代理设置：

### 创建全局配置文件

```bash
mkdir -p ~/.config/buildfly
cat > ~/.config/buildfly/config.yaml << EOF
proxy:
  http: "http://proxy.company.com:8080"
  https: "https://proxy.company.com:8080"
  no_proxy:
    - "localhost"
    - "127.0.0.1"
    - "*.local"
EOF
```

### 配置优先级

1. 项目本地配置 (`./buildfly.yaml`)
2. 全局配置 (`~/.config/buildfly/config.yaml`)

项目配置会覆盖全局配置中的相同设置。

## 环境变量支持

buildfly 也支持通过环境变量设置代理：

```bash
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="https://proxy.company.com:8080"
export NO_PROXY="localhost,127.0.0.1,*.local"

buildfly install
```

## 故障排除

### 1. 代理连接失败

检查代理服务器地址和端口是否正确：

```bash
curl -x http://proxy.company.com:8080 https://www.google.com
```

### 2. 认证失败

确保用户名和密码正确，并且代理服务器支持认证。

### 3. 某些域名无法访问

检查 `no_proxy` 配置，确保不需要代理的域名在列表中。

### 4. SSL 证书问题

如果代理使用自签名证书，可能需要配置信任证书。

## 测试示例

运行本示例：

```bash
cd examples/proxy-demo
buildfly install
```

这将演示如何使用代理配置下载 fmt 和 zlib 依赖。
