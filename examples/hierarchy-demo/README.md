# 层次化配置演示

这个示例演示了 buildfly 的层次化配置功能，支持全局配置和本地配置的合并。

## 配置层次

1. **全局配置** (`~/.config/buildfly/config.yaml`)
   - 系统级别的默认配置
   - 可以被环境变量 `BUILDFLY_CONFIG_FILE` 覆盖

2. **本地配置** (`./buildfly.yaml`)
   - 项目特定的配置
   - 优先级高于全局配置，会覆盖同名配置

## 运行演示

```bash
# 设置全局配置
mkdir -p ~/.config/buildfly
cp global-config.yaml ~/.config/buildfly/config.yaml

# 在项目目录中运行
cd examples/hierarchy-demo
buildfly install

# 查看合并后的配置
buildfly config show
```

## 配置合并规则

- 项目信息：本地配置覆盖全局配置
- 变量：本地配置覆盖全局配置的同名变量
- 依赖项：合并所有依赖项，本地配置覆盖同名依赖
- 构建配置文件：合并所有配置文件
- 目录配置：本地配置优先
