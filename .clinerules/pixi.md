## virtual env pixi backend
Linux & macOS

```
curl -fsSL https://pixi.sh/install.sh | sh
If your system doesn't have curl, you can use wget:
wget -qO- https://pixi.sh/install.sh | sh



envs:
PIXI_VERSION	The version of Pixi getting installed, can be used to up- or down-grade.	latest
PIXI_HOME	The location of the binary folder.	$HOME/.pixi
PIXI_ARCH	The architecture the Pixi version was built for.	uname -m
PIXI_NO_PATH_UPDATE	If set the $PATH will not be updated to add pixi to it.	
TMP_DIR	The temporary directory the script uses to download to and unpack the binary from.	/tmp



curl -fsSL https://pixi.sh/install.sh | PIXI_ARCH=x86_64 bash

curl -fsSL https://pixi.sh/install.sh | PIXI_VERSION=v0.18.0 bash
```

windows:
```
powershell -ExecutionPolicy ByPass -c "irm -useb https://pixi.sh/install.ps1 | iex"


envs:
Environment variable	Description	Default Value
PIXI_VERSION	The version of Pixi getting installed, can be used to up- or down-grade.	latest
PIXI_HOME	The location of the installation.	$Env:USERPROFILE\.pixi
PIXI_NO_PATH_UPDATE	If set, the $PATH will not be updated to add pixi to it.	false



$env:PIXI_VERSION='v0.18.0'; powershell -ExecutionPolicy Bypass -Command "iwr -useb https://pixi.sh/install.ps1 | iex"
```


pixi update
```
pixi self-update
```


use brew
```
brew install pixi
```

## 对于非windows 可以用 wrapper 脚本
pixiw linux/macos 可以用这个wrapper
```bash
#!/bin/sh

# Pixi Wrapper (pixiw) - 类似 Maven Wrapper 的 Pixi 包装器
# 自动安装和管理 pixi 版本

PIXI_VERSION="0.35.1"
PIXI_HOME="$HOME/.pixi"
PIXI_EXE="$PIXI_HOME/bin/pixi"
PIXI_URL="https://pixi.sh/install.sh"

# 检查 pixi 是否已安装
check_pixi_installed() {
    if [ -x "$PIXI_EXE" ]; then
        return 0
    else
        return 1
    fi
}

# 安装 pixi
install_pixi() {
    echo "正在安装 pixi 到 $PIXI_HOME..."
    export PIXI_REPOURL="https://hproxy.all3n.top/github.com/prefix-dev/pixi"
    
    # 使用官方安装脚本
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL $PIXI_URL | sh -s -- --dir $PIXI_HOME
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- $PIXI_URL | sh -s -- --dir $PIXI_HOME
    else
        echo "错误: 需要 curl 或 wget 来下载安装脚本"
        exit 1
    fi
    
    # 检查安装是否成功
    if [ -x "$PIXI_EXE" ]; then
        echo "✓ pixi 安装成功!"
    else
        echo "✗ pixi 安装失败!"
        exit 1
    fi
}

# 更新 pixi
update_pixi() {
    echo "正在更新 pixi..."
    $PIXI_EXE self-update
}

# 卸载 pixi
uninstall_pixi() {
    echo "正在卸载 pixi..."
    if [ -f "$PIXI_EXE" ]; then
        $PIXI_EXE self-uninstall
        rm -rf "$PIXI_HOME"
        echo "✓ pixi 已卸载"
    else
        echo "pixi 未安装"
    fi
}

# 显示帮助信息
show_help() {
    echo "Pixi Wrapper (pixiw) - 自动管理 pixi 安装"
    echo ""
    echo "用法: ./pixiw [命令]"
    echo ""
    echo "命令:"
    echo "  install    安装 pixi"
    echo "  update     更新 pixi"
    echo "  uninstall  卸载 pixi"
    echo "  help       显示此帮助信息"
    echo "  *          任何其他命令将传递给 pixi 执行"
    echo ""
    echo "示例:"
    echo "  ./pixiw install    # 安装 pixi"
    echo "  ./pixiw init       # 在当前目录初始化 pixi 项目"
    echo "  ./pixiw add python # 添加 python 包"
}

# 主逻辑
main() {
    case "$1" in
        "install")
            if check_pixi_installed; then
                echo "pixi 已经安装在 $PIXI_HOME"
            else
                install_pixi
            fi
            ;;
        "update")
            if check_pixi_installed; then
                update_pixi
            else
                echo "pixi 未安装，请先运行: ./pixiw install"
                exit 1
            fi
            ;;
        "uninstall")
            uninstall_pixi
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            # 如果 pixi 未安装，先安装
            if ! check_pixi_installed; then
                echo "pixi 未安装，正在自动安装..."
                install_pixi
            fi
            
            # 执行 pixi 命令
            "$PIXI_EXE" "$@"
            ;;
    esac
}

# 运行主函数
main "$@"
```


## 项目初始化
```
pixi init xxx



```

简单基本配置
```toml
[workspace]
authors = ["all3n <wanghch8398@163.com>"]
channels = ["conda-forge"]
name = "xxx"
platforms = ["osx-arm64"]
version = "0.1.0"

[tasks]

[dependencies]
```

添加依赖
```
pixi add numpy==2.2.6 pytest==8.3.5
或者
pixi add --pypi httpx
```



## 项目配置文件
pixi.toml
```toml
[project]
name = "test"
version = "0.1.0"
description = "test"
authors = ["all3n <wanghch8398@163.com>"]
channels = ["conda-forge"]
platforms = ["linux-64", "osx-64"]

[dependencies]
python = "3.12.*"
numpy = ">=1.21"
pip = "*"
gtest = ">=1.17.0,<2"
protobuf = ">=6.32.1,<7"

# 构建工具和编译器
[build-dependencies]
cmake = ">=3.20"
make = "*"  # 或者 ninja
pkg-config = "*"


[target.linux-64.dependencies]
# Linux 下的编译工具
gcc_linux-64 = ">=11"
gxx_linux-64 = ">=11"
gfortran_linux-64 = ">=11"

[target.osx-64.dependencies]
# macOS 下的编译工具
clang_osx-64 = ">=11"
clangxx_osx-64 = ">=11"
gfortran_osx-64 = ">=11"

[tasks]
# 构建任务
configure = "cmake -B build ."
build = "cmake --build build"
install = "cmake --install build"
test = "cd build && ctest"

# Python 相关任务
python-shell = "python"
pip-install = "pip install -e ."
dev-install = "pip install -e .[dev]"

# 清理任务
clean = "rm -rf build dist *.egg-info"

# 开发环境任务
dev = { depends-on = ["configure", "build", "pip-install"] }


# 激活环境时自动设置的环境变量
[activation]
scripts = ["setup_env.sh"]
```

