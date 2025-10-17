## virtual env uv backend
使用uv 做为c++ 环境管理
创建.buildfly/root 作为 项目环境目录
初始化则用 检查系统uv 是否安装了
如果uv未安装 则用下面方式安装
linux/macos
```
Use curl to download the script and execute it with sh:


curl -LsSf https://astral.sh/uv/install.sh | sh
If your system doesn't have curl, you can use wget:


wget -qO- https://astral.sh/uv/install.sh | sh
Request a specific version by including it in the URL:


curl -LsSf https://astral.sh/uv/0.9.3/install.sh | sh

```


windows:
```
Use irm to download the script and execute it with iex:


powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
Changing the execution policy allows running a script from the internet.

Request a specific version by including it in the URL:


powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/0.9.3/install.ps1 | iex"
```

venv 环境配置 yaml 可以配置