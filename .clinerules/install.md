以下是为大模型AI编辑器优化的提示词，用于生成依赖安装流程：

```markdown
# 依赖安装流程优化提示词

## 配置读取阶段
1. **读取和分析配置文件**
   - 优先读取 `~/.config/buildfly.yaml` 全局配置
   - 然后读取项目根目录 `buildfly.yaml` 本地配置
   - 执行配置合并，项目配置优先级高于全局配置

## 依赖分析阶段
2. **生成依赖构建标签**
   - 解析 `dependency_list` 中的每个依赖项
   - 为每个依赖生成唯一的 `build_tag`
   - 默认使用全局 `build_tag` 模板
   - 支持为特定依赖配置自定义 `build_tag`

## 下载管理阶段
3. **依赖源码获取**
   - **压缩包类型**：
     - 下载到标准化缓存路径：`{cache_dir}/buildfly/{name}/{version}/{filename}.{ext}`
     - 启用文件校验和检查，checksum匹配时跳过重复下载
   
   - **Git仓库类型**：
     - 直接克隆到构建目录

## 构建目录结构
4. **目录路径标准化**
   ```bash
   # 构建目录（包含build_tag确保环境隔离）
   DEP_BUILD_DIR=~/.buildfly/build/{name}/{version}/{build_tag}
   
   # 安装目录（与构建目录分离）
   DEP_INSTALL_DIR=~/.buildfly/install/{name}/{version}/{build_tag}
   ```

## 源码处理
5. **源码准备**
   - 压缩包：解压到 `DEP_BUILD_DIR` 并执行 `--strip-components=1`
   - Git仓库：直接克隆到 `DEP_BUILD_DIR`

## 构建安装流程
6. **标准构建过程**
   - 在 `DEP_BUILD_DIR` 中执行构建命令
   - 将构建结果安装到 `DEP_INSTALL_DIR`

## 项目集成
7. **依赖链接管理**
   - 创建符号链接：将 `DEP_INSTALL_DIR` 内容链接到项目 `.buildfly/install/`
   - 维护安装清单：在 `.buildfly/install.txt` 中记录所有链接项
   - 确保链接操作的原子性和可追溯性