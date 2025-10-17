package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DetectBuildTag 自动检测系统的构建标签
func DetectBuildTag() (*BuildTag, error) {
	bt := &BuildTag{}

	// 检测架构
	if arch, err := detectArchitecture(); err == nil {
		bt.Arch = arch
	}

	// 检测平台
	if platform, err := detectPlatform(); err == nil {
		bt.Platform = platform
	}

	// 检测运行时
	if runtimeVer, err := detectRuntime(); err == nil {
		bt.Runtime = runtimeVer
	}

	// 检测编译器
	if compiler, err := detectCompiler(); err == nil {
		bt.Compiler = compiler
	}

	// 检测 C++ 标准
	if std, err := detectCPPStandard(); err == nil {
		bt.Std = std
	}

	// 检测 ABI
	if abi, err := detectABI(); err == nil {
		bt.ABI = abi
	}

	// 检测 GPU
	if gpu, err := detectGPU(); err == nil {
		bt.GPU = gpu
	}

	// 检测目标（特定于某些平台）
	if target, err := detectTarget(); err == nil {
		bt.Target = target
	}

	return bt, nil
}

// detectArchitecture 检测系统架构
func detectArchitecture() (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "x86_64", nil
	case "arm64":
		return "arm64", nil
	case "386":
		return "i386", nil
	case "arm":
		return "aarch64", nil
	default:
		return arch, nil
	}
}

// detectPlatform 检测操作系统平台
func detectPlatform() (string, error) {
	platform := runtime.GOOS
	switch platform {
	case "linux":
		return "linux", nil
	case "darwin":
		return "darwin", nil
	case "windows":
		return "windows", nil
	default:
		return platform, nil
	}
}

// detectRuntime 检测运行时环境
func detectRuntime() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// 检测 glibc 版本
		if glibcVer, err := detectGLibcVersion(); err == nil {
			return fmt.Sprintf("glibc_%s", glibcVer), nil
		}
		// 检测 musl
		if isMusl() {
			return "musl", nil
		}
		return "glibc", nil
	case "darwin":
		// macOS 使用 libc++
		if libcxxVer, err := detectLibcxxVersion(); err == nil {
			return fmt.Sprintf("libcxx_%s", libcxxVer), nil
		}
		return "libcxx", nil
	case "windows":
		return "msvcrt", nil
	default:
		return "", fmt.Errorf("unsupported platform for runtime detection: %s", runtime.GOOS)
	}
}

// detectGLibcVersion 检测 glibc 版本
func detectGLibcVersion() (string, error) {
	// 尝试从 ldd --version 获取
	if cmd := exec.Command("ldd", "--version"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "glibc") {
				// 解析版本号，例如 "ldd (Ubuntu GLIBC 2.35-0ubuntu3.1) 2.35"
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, "glibc") || strings.Contains(line, "GLIBC") {
						// 提取版本号
						parts := strings.Fields(line)
						for _, part := range parts {
							if strings.HasPrefix(part, "2.") {
								return strings.TrimSpace(part), nil
							}
						}
					}
				}
			}
		}
	}

	// 尝试从 /lib/x86_64-linux-gnu/libc.so.6 获取
	if cmd := exec.Command("/lib/x86_64-linux-gnu/libc.so.6"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "GNU C Library") {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, "version") {
						parts := strings.Fields(line)
						for i, part := range parts {
							if part == "version" && i+1 < len(parts) {
								return strings.TrimSuffix(parts[i+1], ","), nil
							}
						}
					}
				}
			}
		}
	}

	return "2.35", nil // 默认版本
}

// detectLibcxxVersion 检测 libc++ 版本 (macOS)
func detectLibcxxVersion() (string, error) {
	// macOS 版本对应的 libc++ 版本
	if cmd := exec.Command("sw_vers", "-productVersion"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			macosVer := strings.TrimSpace(string(output))
			// macOS 版本到 libc++ 版本的映射
			versionMap := map[string]string{
				"14.0":  "15",
				"13.0":  "14",
				"12.0":  "13",
				"11.0":  "12",
				"10.15": "11",
				"10.14": "10",
				"10.13": "9",
			}

			// 提取主版本号
			parts := strings.Split(macosVer, ".")
			if len(parts) >= 2 {
				key := fmt.Sprintf("%s.%s", parts[0], parts[1])
				if libcxxVer, exists := versionMap[key]; exists {
					return libcxxVer, nil
				}
			}
		}
	}

	return "15", nil // 默认版本
}

// isMusl 检测是否使用 musl
func isMusl() bool {
	// 检查 ldd 输出
	if cmd := exec.Command("ldd", "--version"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			return strings.Contains(outputStr, "musl")
		}
	}
	return false
}

// detectCompiler 检测可用的编译器
func detectCompiler() (string, error) {
	// 检测 GCC
	if gccVer, err := detectGCCVersion(); err == nil {
		return fmt.Sprintf("gcc_%s", gccVer), nil
	}

	// 检测 Clang
	if clangVer, err := detectClangVersion(); err == nil {
		return fmt.Sprintf("clang_%s", clangVer), nil
	}

	// 检测 Apple Clang (macOS)
	if runtime.GOOS == "darwin" {
		if appleClangVer, err := detectAppleClangVersion(); err == nil {
			return fmt.Sprintf("apple-clang_%s", appleClangVer), nil
		}
	}

	// 检测 MSVC (Windows)
	if runtime.GOOS == "windows" {
		if msvcVer, err := detectMSVCVersion(); err == nil {
			return fmt.Sprintf("msvc_%s", msvcVer), nil
		}
	}

	return "", fmt.Errorf("no supported compiler found")
}

// detectGCCVersion 检测 GCC 版本
func detectGCCVersion() (string, error) {
	if cmd := exec.Command("gcc", "--version"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			lines := strings.Split(outputStr, "\n")
			if len(lines) > 0 {
				parts := strings.Fields(lines[0])
				for i, part := range parts {
					if part == "gcc" && i+2 < len(parts) {
						// 通常格式: gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0
						version := parts[i+2]
						// 提取主版本号
						if strings.Contains(version, ".") {
							versionParts := strings.Split(version, ".")
							return versionParts[0], nil
						}
						return version, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("gcc not found")
}

// detectClangVersion 检测 Clang 版本
func detectClangVersion() (string, error) {
	if cmd := exec.Command("clang", "--version"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			lines := strings.Split(outputStr, "\n")
			if len(lines) > 0 {
				parts := strings.Fields(lines[0])
				for i, part := range parts {
					if part == "clang" && i+1 < len(parts) {
						version := parts[i+1]
						// 提取主版本号
						if strings.Contains(version, ".") {
							versionParts := strings.Split(version, ".")
							return versionParts[0], nil
						}
						return version, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("clang not found")
}

// detectAppleClangVersion 检测 Apple Clang 版本
func detectAppleClangVersion() (string, error) {
	if cmd := exec.Command("clang", "--version"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "Apple clang") {
				lines := strings.Split(outputStr, "\n")
				if len(lines) > 0 {
					parts := strings.Fields(lines[0])
					for i, part := range parts {
						if part == "clang" && i+1 < len(parts) {
							version := parts[i+1]
							// Apple Clang 版本格式通常是完整的版本号
							return version, nil
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("apple clang not found")
}

// detectMSVCVersion 检测 MSVC 版本
func detectMSVCVersion() (string, error) {
	// 尝试从 Visual Studio 安装目录检测
	vsWhere := `C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe`
	if cmd := exec.Command(vsWhere, "-latest", "-property", "installationVersion"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			version := strings.TrimSpace(string(output))
			// 转换为 MSVC 版本号格式
			if strings.Contains(version, ".") {
				parts := strings.Split(version, ".")
				if len(parts) >= 2 {
					return fmt.Sprintf("%s.%s", parts[0], parts[1]), nil
				}
			}
			return version, nil
		}
	}
	return "", fmt.Errorf("msvc not found")
}

// detectCPPStandard 检测默认 C++ 标准
func detectCPPStandard() (string, error) {
	// 默认使用 cpp17，现代编译器都支持
	return "cpp17", nil
}

// detectABI 检测 ABI
func detectABI() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return "sysv", nil
	case "darwin":
		return "macho", nil
	case "windows":
		return "msabi", nil
	default:
		return "", fmt.Errorf("unsupported platform for ABI detection: %s", runtime.GOOS)
	}
}

// detectGPU 检测 GPU 和相关后端
func detectGPU() (*GPUInfo, error) {
	// 检测 CUDA
	if cudaInfo, err := detectCUDA(); err == nil {
		return cudaInfo, nil
	}

	// 检测 ROCm
	if rocmInfo, err := detectROCm(); err == nil {
		return rocmInfo, nil
	}

	// 检测 OpenCL
	if openclInfo, err := detectOpenCL(); err == nil {
		return openclInfo, nil
	}

	return nil, nil // 没有 GPU 或未检测到
}

// detectCUDA 检测 CUDA
func detectCUDA() (*GPUInfo, error) {
	// 检查 nvcc 命令
	if cmd := exec.Command("nvcc", "--version"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "Cuda compilation tools") {
				// 解析 CUDA 版本
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(line, "release") {
						parts := strings.Fields(line)
						for i, part := range parts {
							if part == "release" && i+2 < len(parts) {
								version := strings.TrimSuffix(parts[i+2], ",")

								// 检测 GPU 架构
								archs, _ := detectCUDAArchs()

								return &GPUInfo{
									Backend: "cuda",
									CUDA: &CUDABackend{
										Version: version,
										Arch:    archs,
										Enabled: true,
									},
								}, nil
							}
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("cuda not found")
}

// detectCUDAArchs 检测 CUDA GPU 架构
func detectCUDAArchs() ([]string, error) {
	var archs []string

	// 尝试从 nvidia-smi 获取
	if cmd := exec.Command("nvidia-smi", "--query-gpu=compute_cap", "--format=csv,noheader,nounits"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := strings.TrimSpace(string(output))
			if outputStr != "" {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						archs = append(archs, fmt.Sprintf("compute_%s", line))
					}
				}
				return archs, nil
			}
		}
	}

	// 默认常见的架构
	archs = []string{"compute_60", "compute_70", "compute_80", "compute_89"}
	return archs, nil
}

// detectROCm 检测 ROCm
func detectROCm() (*GPUInfo, error) {
	// 检查 rocminfo 命令
	if cmd := exec.Command("rocminfo"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "ROCm") {
				// 解析 ROCm 版本
				version := "5.0" // 默认版本

				// 检测 GPU 架构
				archs, _ := detectROCmArchs(outputStr)

				return &GPUInfo{
					Backend: "rocm",
					ROCm: &ROCmBackend{
						Version: version,
						Arch:    archs,
						Enabled: true,
					},
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("rocm not found")
}

// detectROCmArchs 检测 ROCm GPU 架构
func detectROCmArchs(rocminfoOutput string) ([]string, error) {
	var archs []string
	lines := strings.Split(rocminfoOutput, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Name:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "Name:" && i+1 < len(parts) {
					gpuName := parts[i+1]
					// 映射 GPU 名称到架构
					if strings.Contains(gpuName, "gfx") {
						archs = append(archs, gpuName)
					}
				}
			}
		}
	}

	if len(archs) == 0 {
		archs = []string{"gfx900", "gfx1030"} // 默认架构
	}

	return archs, nil
}

// detectOpenCL 检测 OpenCL
func detectOpenCL() (*GPUInfo, error) {
	// 检查 clinfo 命令
	if cmd := exec.Command("clinfo"); cmd != nil {
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "OpenCL") {
				version := "2.0" // 默认版本

				return &GPUInfo{
					Backend: "opencl",
					OpenCL: &OpenCLBackend{
						Version: version,
						Enabled: true,
					},
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("opencl not found")
}

// detectTarget 检测特定目标
func detectTarget() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// 检测 macOS 目标版本
		if cmd := exec.Command("sw_vers", "-productVersion"); cmd != nil {
			if output, err := cmd.CombinedOutput(); err == nil {
				macosVer := strings.TrimSpace(string(output))
				return fmt.Sprintf("macos%s", macosVer), nil
			}
		}
	}
	return "", nil
}

// GetDefaultBuildTag 获取默认构建标签（基于配置和系统检测）
func GetDefaultBuildTag(config *BuildTag) (*BuildTag, error) {
	var result *BuildTag

	// 如果有配置文件中的 build tag，使用它作为基础
	if config != nil {
		result = config.Clone()
	} else {
		result = &BuildTag{}
	}

	// 检测系统信息并填充缺失的字段
	detected, err := DetectBuildTag()
	if err != nil {
		return nil, fmt.Errorf("failed to detect build tag: %w", err)
	}

	// 合并检测到的信息（只填充空字段）
	if result.Arch == "" {
		result.Arch = detected.Arch
	}
	if result.Platform == "" {
		result.Platform = detected.Platform
	}
	if result.Runtime == "" {
		result.Runtime = detected.Runtime
	}
	if result.Compiler == "" {
		result.Compiler = detected.Compiler
	}
	if result.Std == "" {
		result.Std = detected.Std
	}
	if result.ABI == "" {
		result.ABI = detected.ABI
	}
	if result.Target == "" {
		result.Target = detected.Target
	}
	if result.GPU == nil {
		result.GPU = detected.GPU
	}

	return result, nil
}

// GetBuildTagFromEnv 从环境变量获取构建标签
func GetBuildTagFromEnv() (*BuildTag, error) {
	buildTagStr := os.Getenv("BUILDFLY_BUILD_TAG")
	if buildTagStr == "" {
		return nil, nil
	}

	return ParseBuildTag(buildTagStr)
}
