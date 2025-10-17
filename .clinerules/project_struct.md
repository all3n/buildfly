# 目录结构 示例
```
cmd/                 # 命令行入口
  main.go            # 主程序入口
pkg/
  dependency/        # 依赖管理核心逻辑
  resolver/          # 依赖解析器
  downloader/        # 下载管理器
  cache/             # 缓存管理
  config/            # 配置管理
  utils/             # 工具函数
internal/            # 内部包（不对外暴露）
  types/             # 数据类型定义
  constants/         # 常量定义
configs/             # 配置文件模板
  global.yaml
  templates/
    cmake.yaml
    configure.yaml
    makefile.yaml
examples/            # 使用示例
  example-1/
  example-2/
docs                  # 文档markdown 相关放到这里
data/                 # 数据文件
tests/                # 测试脚本 或者 shell 测试文件
manager               # 项目根管理脚本 包含编译运行测试 fmt lint 等命令封装
bin/                  # 脚本shell目录
```
不要在根目录创建测试shell文件
