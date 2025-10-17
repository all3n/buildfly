## 编译目标artifact tag结构 
tag 用一个字符串标志 示例如果某个条件不限制可以用any，或可以用|,+代表以上
```
// 任意架构的 Linux
"arch=any,platform=linux,runtime=glibc_2.35+,compiler=gcc_11+,std=cpp17,abi=sysv"

// 特定编译器版本范围
"arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_9|10|11,std=cpp17,abi=sysv"

// 多个运行时选择
"arch=x86_64,platform=linux,runtime=glibc_2.35|musl_1.2,compiler=gcc_11+,std=cpp17,abi=sysv"


// Ubuntu 22.04 + GCC
"arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11.3,std=cpp17,abi=sysv"

// CentOS 7 + 较老 GCC
"arch=x86_64,platform=linux,runtime=glibc_2.17,compiler=gcc_4.8,std=cpp11,abi=sysv"

// Alpine Linux + musl
"arch=x86_64,platform=linux,runtime=musl_1.2,compiler=gcc_10.3,std=cpp17,abi=sysv"

// Arch Linux + Clang
"arch=x86_64,platform=linux,runtime=glibc_2.37,compiler=clang_16.0,std=cpp20,abi=sysv"

// C++ 标准最低要求
"arch=x86_64,platform=linux,runtime=glibc_2.35+,compiler=gcc_11+,std=cpp17+,abi=sysv"

// macOS 14 + ARM64
"arch=arm64,platform=darwin,runtime=libcxx_15,compiler=apple-clang_15.0,std=cpp20,abi=macho,target=macos14.0"

// 通用二进制 (Universal Binary)
"arch=universal,platform=darwin,runtime=libcxx_15,compiler=apple-clang_15.0,std=cpp20,abi=macho,target=macos13.0"


// CUDA 12.x + NVIDIA GPU
"arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=nvcc_12.0+gcc_11.4,std=cpp17,abi=sysv,cuda=12.0,cuda_arch=compute_80|compute_90,gpu_enabled=true,gpu_backend=cuda"

// CUDA 11.8 + 多架构
"arch=x86_64,platform=linux,runtime=glibc_2.31,compiler=nvcc_11.8+gcc_9.4,std=cpp14,abi=sysv,cuda=11.8,cuda_arch=compute_50|compute_60|compute_70|compute_80,gpu_enabled=true,gpu_backend=cuda"

// Windows + CUDA
"arch=x64,platform=windows,runtime=msvcrt,compiler=nvcc_12.0+msvc_19.38,std=cpp17,abi=msabi,cuda=12.0,cuda_arch=compute_80,gpu_enabled=true,gpu_backend=cuda"

// 仅 CPU 版本 (CUDA 工具链但无 GPU)
"arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=nvcc_12.0+gcc_11.4,std=cpp17,abi=sysv,cuda=12.0,cuda_arch=none,gpu_enabled=false,gpu_backend=none"
```


示例代码如下
```go
package main

// UniversalBuildInfo 构建信息结构体
type UniversalBuildInfo struct {
    // 基础字段
    Arch      string `json:"arch,omitempty"`      // x86_64, arm64, x64
    Platform  string `json:"platform,omitempty"`  // linux, darwin, windows
    Runtime   string `json:"runtime,omitempty"`   // glibc, libcxx, msvcrt, musl
    Compiler  string `json:"compiler,omitempty"`  // gcc, clang, msvc
    Std       string `json:"std,omitempty"`       // cpp11, cpp17, cpp20
    ABI       string `json:"abi,omitempty"` // sysv, macho, msabi
    Target    string `json:"target,omitempty"`    // 平台特定目标
    
    // GPU 计算框架 - 互斥字段组
    GPU *GPUInfo `json:"gpu,omitempty"`
}

// GPUInfo GPU 相关信息，多个后端互斥
type GPUInfo struct {
    // 后端类型 - 只能选择一个
    Backend string `json:"backend"` // cuda, rocm, opencl, none
    
    // 具体后端配置 - 根据 Backend 选择填写
    CUDA  *CUDABackend  `json:"cuda,omitempty"`
    ROCm  *ROCmBackend  `json:"rocm,omitempty"`
    OpenCL *OpenCLBackend `json:"opencl,omitempty"`
}

// CUDABackend CUDA 配置
type CUDABackend struct {
    Version string   `json:"version"`           // 11.0, 12.0
    Arch    []string `json:"arch,omitempty"`    // compute_50, compute_80
    Enabled bool     `json:"enabled"`           // true, false
}

// ROCmBackend ROCm 配置
type ROCmBackend struct {
    Version string   `json:"version"`           // 5.0, 6.0
    Arch    []string `json:"arch,omitempty"`    // gfx900, gfx1030
    Enabled bool     `json:"enabled"`           // true, false
}

// OpenCLBackend OpenCL 配置
type OpenCLBackend struct {
    Version string   `json:"version"`           // 2.0, 3.0
    Enabled bool     `json:"enabled"`           // true, false
}
```