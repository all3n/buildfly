package cli

import (
	"fmt"
	"os"

	"buildfly/pkg/config"

	"github.com/spf13/cobra"
)

// newDetectCmd 创建 detect 命令
func newDetectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "检测系统构建信息",
		Long: `自动检测系统的构建信息，包括架构、平台、编译器、GPU等信息。

这个命令会扫描当前系统环境，生成适合的构建标签。`,
		RunE: runDetect,
	}

	return cmd
}

// runDetect 执行检测
func runDetect(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 检测系统构建信息...")
	fmt.Println()

	// 检测系统信息
	buildTag, err := config.DetectBuildTag()
	if err != nil {
		return fmt.Errorf("检测失败: %w", err)
	}

	// 显示检测结果
	fmt.Println("📋 检测结果:")
	fmt.Printf("架构 (arch):     %s\n", buildTag.Arch)
	fmt.Printf("平台 (platform): %s\n", buildTag.Platform)
	fmt.Printf("运行时 (runtime): %s\n", buildTag.Runtime)
	fmt.Printf("编译器 (compiler): %s\n", buildTag.Compiler)
	fmt.Printf("C++标准 (std):    %s\n", buildTag.Std)
	fmt.Printf("ABI (abi):        %s\n", buildTag.ABI)
	if buildTag.Target != "" {
		fmt.Printf("目标 (target):    %s\n", buildTag.Target)
	}

	// 显示 GPU 信息
	if buildTag.GPU != nil {
		fmt.Printf("GPU 后端:        %s\n", buildTag.GPU.Backend)
		switch buildTag.GPU.Backend {
		case "cuda":
			if buildTag.GPU.CUDA != nil {
				fmt.Printf("  CUDA 版本:    %s\n", buildTag.GPU.CUDA.Version)
				if len(buildTag.GPU.CUDA.Arch) > 0 {
					fmt.Printf("  GPU 架构:     %v\n", buildTag.GPU.CUDA.Arch)
				}
				fmt.Printf("  启用状态:      %t\n", buildTag.GPU.CUDA.Enabled)
			}
		case "rocm":
			if buildTag.GPU.ROCm != nil {
				fmt.Printf("  ROCm 版本:    %s\n", buildTag.GPU.ROCm.Version)
				if len(buildTag.GPU.ROCm.Arch) > 0 {
					fmt.Printf("  GPU 架构:     %v\n", buildTag.GPU.ROCm.Arch)
				}
				fmt.Printf("  启用状态:      %t\n", buildTag.GPU.ROCm.Enabled)
			}
		case "opencl":
			if buildTag.GPU.OpenCL != nil {
				fmt.Printf("  OpenCL 版本:  %s\n", buildTag.GPU.OpenCL.Version)
				fmt.Printf("  启用状态:      %t\n", buildTag.GPU.OpenCL.Enabled)
			}
		}
	} else {
		fmt.Println("GPU 后端:        未检测到 GPU")
	}

	fmt.Println()

	// 显示构建标签字符串
	fmt.Println("🏷️  构建标签:")
	fmt.Printf("完整标签: %s\n", buildTag.String())
	fmt.Printf("目录名:   %s\n", buildTag.ToDirName())

	fmt.Println()

	// 显示使用示例
	fmt.Println("💡 使用示例:")
	fmt.Printf("命令行: ./buildfly install --build-tag \"%s\"\n", buildTag.String())
	fmt.Printf("环境变量: export BUILDFLY_BUILD_TAG=\"%s\"\n", buildTag.String())

	fmt.Println()

	// 显示配置文件示例
	fmt.Println("📄 配置文件示例:")
	fmt.Println("project:")
	fmt.Println("  name: \"my-project\"")
	fmt.Println("  build_tag:")
	fmt.Printf("    arch: \"%s\"\n", buildTag.Arch)
	fmt.Printf("    platform: \"%s\"\n", buildTag.Platform)
	fmt.Printf("    runtime: \"%s\"\n", buildTag.Runtime)
	fmt.Printf("    compiler: \"%s\"\n", buildTag.Compiler)
	fmt.Printf("    std: \"%s\"\n", buildTag.Std)
	fmt.Printf("    abi: \"%s\"\n", buildTag.ABI)
	if buildTag.Target != "" {
		fmt.Printf("    target: \"%s\"\n", buildTag.Target)
	}

	if buildTag.GPU != nil {
		fmt.Println("    gpu:")
		fmt.Printf("      backend: \"%s\"\n", buildTag.GPU.Backend)
		switch buildTag.GPU.Backend {
		case "cuda":
			if buildTag.GPU.CUDA != nil {
				fmt.Printf("      cuda:\n")
				fmt.Printf("        version: \"%s\"\n", buildTag.GPU.CUDA.Version)
				fmt.Printf("        enabled: %t\n", buildTag.GPU.CUDA.Enabled)
				if len(buildTag.GPU.CUDA.Arch) > 0 {
					fmt.Printf("        arch: [%q]\n", buildTag.GPU.CUDA.Arch)
				}
			}
		case "rocm":
			if buildTag.GPU.ROCm != nil {
				fmt.Printf("      rocm:\n")
				fmt.Printf("        version: \"%s\"\n", buildTag.GPU.ROCm.Version)
				fmt.Printf("        enabled: %t\n", buildTag.GPU.ROCm.Enabled)
				if len(buildTag.GPU.ROCm.Arch) > 0 {
					fmt.Printf("        arch: [%q]\n", buildTag.GPU.ROCm.Arch)
				}
			}
		case "opencl":
			if buildTag.GPU.OpenCL != nil {
				fmt.Printf("      opencl:\n")
				fmt.Printf("        version: \"%s\"\n", buildTag.GPU.OpenCL.Version)
				fmt.Printf("        enabled: %t\n", buildTag.GPU.OpenCL.Enabled)
			}
		}
	}

	// 检查环境变量
	fmt.Println()
	fmt.Println("🌍 环境变量检查:")
	if envTag := os.Getenv("BUILDFLY_BUILD_TAG"); envTag != "" {
		fmt.Printf("BUILDFLY_BUILD_TAG: %s\n", envTag)
	} else {
		fmt.Println("BUILDFLY_BUILD_TAG: 未设置")
	}

	return nil
}
