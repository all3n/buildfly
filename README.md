# buildfly
buildfly 是c++基于github管理包管理编译工具.

## 配置
1. 需要项目目录下存在buildfly.yml

## target
1. 包含type: bin/lib 会被自动识别为target
1. 属性
    1. type: string
    1. srcs: array 支持glob 匹配
    1. cflags: string
    1. libs: array[string]
    1. includes: array[string] 
    1. deps: array[dict|string]
        1. 依赖本项目target lib,需要加//前缀 like "//my-lib"
        1. dict:
            1. {"name": "link_type:lib_name"}
            1. name为dependency中定义名称
            1. link_type: 支持shared/static
            1. lib_name: g++ -l后面加lib名称
1. lib 独有属性
    1. lib_type: static/shared/all
1. 说明:
    1. libs,includes 不需要包含deps涉及的libs,include



## build
1. build all target
    1. bfly build
1. build target
    1. bfly build target


## get 
1. github dep
    1. bfly get owner/repo@tag 
        1. example: bfly get open-source-parsers/jsoncpp@1.8.4
    1. bfly get owner/repo@branch


## dependency
1. 定义dict
1. name为依赖名称,方便在target中引用
1. value: string|dict
    1. string:  目前支持github 依赖  等价bfly get
        1. owner/repo 获取master代码
        1. owner/repo@tag 获取对应tag
    1. dict:
        1. url: 目前支持tar.gz
        1. modules: array 可选
        1. cmds: array
            1. 环境变量
                1. INSTALL_MODULES: modules 定义才会有 ,分割
                1. INSTALL_PREFIX: 包指定安装目录


## build support
1. auto detact
    1. cmake
        1. if CMakeList.txt exist
    1. configure
1. 通过dependency: dict cmds 配置


## Proxy
1. pip install requests[socks]
1. bfly config proxy.http sock5://127.0.0.1:1086
1. bfly config proxy.https sock5://127.0.0.1:1086

[ ] cmake generate
[ ] docker support
[ ] global config
