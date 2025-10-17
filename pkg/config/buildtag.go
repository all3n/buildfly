package config

import (
	"fmt"
	"regexp"
	"strings"
)

// BuildTag 构建标签结构体
type BuildTag struct {
	Arch     string `json:"arch,omitempty"`     // x86_64, arm64, x64
	Platform string `json:"platform,omitempty"` // linux, darwin, windows
	Runtime  string `json:"runtime,omitempty"`  // glibc, libcxx, msvcrt, musl
	Compiler string `json:"compiler,omitempty"` // gcc, clang, msvc
	Std      string `json:"std,omitempty"`      // cpp11, cpp17, cpp20
	ABI      string `json:"abi,omitempty"`      // sysv, macho, msabi
	Target   string `json:"target,omitempty"`   // 平台特定目标

	// GPU 计算框架 - 互斥字段组
	GPU *GPUInfo `json:"gpu,omitempty"`
}

// GPUInfo GPU 相关信息，多个后端互斥
type GPUInfo struct {
	// 后端类型 - 只能选择一个
	Backend string `json:"backend"` // cuda, rocm, opencl, none

	// 具体后端配置 - 根据 Backend 选择填写
	CUDA   *CUDABackend   `json:"cuda,omitempty"`
	ROCm   *ROCmBackend   `json:"rocm,omitempty"`
	OpenCL *OpenCLBackend `json:"opencl,omitempty"`
}

// CUDABackend CUDA 配置
type CUDABackend struct {
	Version string   `json:"version"`        // 11.0, 12.0
	Arch    []string `json:"arch,omitempty"` // compute_50, compute_80
	Enabled bool     `json:"enabled"`        // true, false
}

// ROCmBackend ROCm 配置
type ROCmBackend struct {
	Version string   `json:"version"`        // 5.0, 6.0
	Arch    []string `json:"arch,omitempty"` // gfx900, gfx1030
	Enabled bool     `json:"enabled"`        // true, false
}

// OpenCLBackend OpenCL 配置
type OpenCLBackend struct {
	Version string `json:"version"` // 2.0, 3.0
	Enabled bool   `json:"enabled"` // true, false
}

// ParseBuildTag 解析构建标签字符串
// 示例: "arch=x86_64,platform=linux,runtime=glibc_2.35+,compiler=gcc_11+,std=cpp17,abi=sysv"
func ParseBuildTag(tagStr string) (*BuildTag, error) {
	if tagStr == "" {
		return nil, fmt.Errorf("build tag string cannot be empty")
	}

	bt := &BuildTag{}

	// 分割键值对
	pairs := strings.Split(tagStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// 分割键和值
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid build tag format: %s", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 根据键设置相应的字段
		switch key {
		case "arch":
			bt.Arch = value
		case "platform":
			bt.Platform = value
		case "runtime":
			bt.Runtime = value
		case "compiler":
			bt.Compiler = value
		case "std":
			bt.Std = value
		case "abi":
			bt.ABI = value
		case "target":
			bt.Target = value
		case "cuda", "cuda_version":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{Backend: "cuda"}
			}
			if bt.GPU.CUDA == nil {
				bt.GPU.CUDA = &CUDABackend{}
			}
			if key == "cuda" {
				bt.GPU.CUDA.Version = value
				bt.GPU.CUDA.Enabled = true
			} else {
				bt.GPU.CUDA.Version = value
			}
		case "cuda_arch":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{Backend: "cuda"}
			}
			if bt.GPU.CUDA == nil {
				bt.GPU.CUDA = &CUDABackend{Enabled: true}
			}
			// 分割多个架构
			archs := strings.Split(value, "|")
			bt.GPU.CUDA.Arch = archs
		case "rocm", "rocm_version":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{Backend: "rocm"}
			}
			if bt.GPU.ROCm == nil {
				bt.GPU.ROCm = &ROCmBackend{}
			}
			if key == "rocm" {
				bt.GPU.ROCm.Version = value
				bt.GPU.ROCm.Enabled = true
			} else {
				bt.GPU.ROCm.Version = value
			}
		case "rocm_arch":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{Backend: "rocm"}
			}
			if bt.GPU.ROCm == nil {
				bt.GPU.ROCm = &ROCmBackend{Enabled: true}
			}
			// 分割多个架构
			archs := strings.Split(value, "|")
			bt.GPU.ROCm.Arch = archs
		case "opencl", "opencl_version":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{Backend: "opencl"}
			}
			if bt.GPU.OpenCL == nil {
				bt.GPU.OpenCL = &OpenCLBackend{}
			}
			if key == "opencl" {
				bt.GPU.OpenCL.Version = value
				bt.GPU.OpenCL.Enabled = true
			} else {
				bt.GPU.OpenCL.Version = value
			}
		case "gpu_backend":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{}
			}
			bt.GPU.Backend = value
		case "gpu_enabled":
			if bt.GPU == nil {
				bt.GPU = &GPUInfo{}
			}
			if value == "true" || value == "1" {
				// 根据现有后端设置启用状态
				if bt.GPU.CUDA != nil {
					bt.GPU.CUDA.Enabled = true
				}
				if bt.GPU.ROCm != nil {
					bt.GPU.ROCm.Enabled = true
				}
				if bt.GPU.OpenCL != nil {
					bt.GPU.OpenCL.Enabled = true
				}
			}
		default:
			return nil, fmt.Errorf("unknown build tag key: %s", key)
		}
	}

	return bt, nil
}

// String 返回构建标签的字符串表示
func (bt *BuildTag) String() string {
	if bt == nil {
		return ""
	}

	var parts []string

	if bt.Arch != "" {
		parts = append(parts, fmt.Sprintf("arch=%s", bt.Arch))
	}
	if bt.Platform != "" {
		parts = append(parts, fmt.Sprintf("platform=%s", bt.Platform))
	}
	if bt.Runtime != "" {
		parts = append(parts, fmt.Sprintf("runtime=%s", bt.Runtime))
	}
	if bt.Compiler != "" {
		parts = append(parts, fmt.Sprintf("compiler=%s", bt.Compiler))
	}
	if bt.Std != "" {
		parts = append(parts, fmt.Sprintf("std=%s", bt.Std))
	}
	if bt.ABI != "" {
		parts = append(parts, fmt.Sprintf("abi=%s", bt.ABI))
	}
	if bt.Target != "" {
		parts = append(parts, fmt.Sprintf("target=%s", bt.Target))
	}

	// 添加 GPU 信息
	if bt.GPU != nil {
		switch bt.GPU.Backend {
		case "cuda":
			if bt.GPU.CUDA != nil && bt.GPU.CUDA.Enabled {
				if bt.GPU.CUDA.Version != "" {
					parts = append(parts, fmt.Sprintf("cuda=%s", bt.GPU.CUDA.Version))
				}
				if len(bt.GPU.CUDA.Arch) > 0 {
					parts = append(parts, fmt.Sprintf("cuda_arch=%s", strings.Join(bt.GPU.CUDA.Arch, "|")))
				}
			}
		case "rocm":
			if bt.GPU.ROCm != nil && bt.GPU.ROCm.Enabled {
				if bt.GPU.ROCm.Version != "" {
					parts = append(parts, fmt.Sprintf("rocm=%s", bt.GPU.ROCm.Version))
				}
				if len(bt.GPU.ROCm.Arch) > 0 {
					parts = append(parts, fmt.Sprintf("rocm_arch=%s", strings.Join(bt.GPU.ROCm.Arch, "|")))
				}
			}
		case "opencl":
			if bt.GPU.OpenCL != nil && bt.GPU.OpenCL.Enabled {
				if bt.GPU.OpenCL.Version != "" {
					parts = append(parts, fmt.Sprintf("opencl=%s", bt.GPU.OpenCL.Version))
				}
			}
		}
	}

	return strings.Join(parts, ",")
}

// ToDirName 将构建标签转换为文件系统安全的目录名
func (bt *BuildTag) ToDirName() string {
	if bt == nil {
		return "default"
	}

	dirName := bt.String()

	// 替换文件系统不安全的字符，但保持逗号分隔
	dirName = strings.ReplaceAll(dirName, "=", "-")
	dirName = strings.ReplaceAll(dirName, "+", "plus")
	dirName = strings.ReplaceAll(dirName, "|", "or")

	// 确保目录名不以点开头
	if strings.HasPrefix(dirName, ".") {
		dirName = "_" + dirName[1:]
	}

	return dirName
}

// Validate 验证构建标签的有效性
func (bt *BuildTag) Validate() error {
	if bt == nil {
		return fmt.Errorf("build tag is nil")
	}

	// 验证架构
	if bt.Arch != "" {
		validArchs := []string{"x86_64", "arm64", "x64", "i386", "aarch64"}
		if !contains(validArchs, bt.Arch) {
			return fmt.Errorf("invalid architecture: %s", bt.Arch)
		}
	}

	// 验证平台
	if bt.Platform != "" {
		validPlatforms := []string{"linux", "darwin", "windows"}
		if !contains(validPlatforms, bt.Platform) {
			return fmt.Errorf("invalid platform: %s", bt.Platform)
		}
	}

	// 验证 C++ 标准
	if bt.Std != "" {
		validStds := []string{"cpp11", "cpp14", "cpp17", "cpp20", "cpp23"}
		if !contains(validStds, bt.Std) {
			return fmt.Errorf("invalid C++ standard: %s", bt.Std)
		}
	}

	// 验证 ABI
	if bt.ABI != "" {
		validABIs := []string{"sysv", "macho", "msabi"}
		if !contains(validABIs, bt.ABI) {
			return fmt.Errorf("invalid ABI: %s", bt.ABI)
		}
	}

	// 验证 GPU 配置
	if bt.GPU != nil {
		if err := bt.GPU.Validate(); err != nil {
			return fmt.Errorf("invalid GPU configuration: %w", err)
		}
	}

	return nil
}

// Validate 验证 GPU 配置
func (gpu *GPUInfo) Validate() error {
	if gpu == nil {
		return nil
	}

	validBackends := []string{"cuda", "rocm", "opencl", "none"}
	if !contains(validBackends, gpu.Backend) {
		return fmt.Errorf("invalid GPU backend: %s", gpu.Backend)
	}

	switch gpu.Backend {
	case "cuda":
		if gpu.CUDA == nil {
			return fmt.Errorf("CUDA backend requires CUDA configuration")
		}
		if gpu.CUDA.Version == "" {
			return fmt.Errorf("CUDA backend requires version")
		}
	case "rocm":
		if gpu.ROCm == nil {
			return fmt.Errorf("ROCm backend requires ROCm configuration")
		}
		if gpu.ROCm.Version == "" {
			return fmt.Errorf("ROCm backend requires version")
		}
	case "opencl":
		if gpu.OpenCL == nil {
			return fmt.Errorf("OpenCL backend requires OpenCL configuration")
		}
		if gpu.OpenCL.Version == "" {
			return fmt.Errorf("OpenCL backend requires version")
		}
	}

	return nil
}

// Equals 比较两个构建标签是否相等
func (bt *BuildTag) Equals(other *BuildTag) bool {
	if bt == nil && other == nil {
		return true
	}
	if bt == nil || other == nil {
		return false
	}

	return bt.String() == other.String()
}

// Clone 克隆构建标签
func (bt *BuildTag) Clone() *BuildTag {
	if bt == nil {
		return nil
	}

	clone := *bt

	// 深度复制 GPU 配置
	if bt.GPU != nil {
		gpuClone := *bt.GPU
		if bt.GPU.CUDA != nil {
			cudaClone := *bt.GPU.CUDA
			if bt.GPU.CUDA.Arch != nil {
				cudaClone.Arch = make([]string, len(bt.GPU.CUDA.Arch))
				copy(cudaClone.Arch, bt.GPU.CUDA.Arch)
			}
			gpuClone.CUDA = &cudaClone
		}
		if bt.GPU.ROCm != nil {
			rocmClone := *bt.GPU.ROCm
			if bt.GPU.ROCm.Arch != nil {
				rocmClone.Arch = make([]string, len(bt.GPU.ROCm.Arch))
				copy(rocmClone.Arch, bt.GPU.ROCm.Arch)
			}
			gpuClone.ROCm = &rocmClone
		}
		if bt.GPU.OpenCL != nil {
			openclClone := *bt.GPU.OpenCL
			gpuClone.OpenCL = &openclClone
		}
		clone.GPU = &gpuClone
	}

	return &clone
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsValidBuildTagString 检查构建标签字符串格式是否有效
func IsValidBuildTagString(tagStr string) bool {
	if tagStr == "" {
		return false
	}

	// 简单的正则验证
	// 格式: key=value,key=value,...
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*=[^,]+(,[a-zA-Z_][a-zA-Z0-9_]*=[^,]+)*$`
	matched, err := regexp.MatchString(pattern, tagStr)
	if err != nil {
		return false
	}

	return matched
}
