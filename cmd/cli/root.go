package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "buildfly",
	Short: "C++ 依赖管理器",
	Long: `BuildFly 是一个用 Golang 开发的 C++ 依赖管理器，
支持 YAML 配置文件，可以管理 C++ 项目的依赖下载、构建和安装。

支持的功能：
- YAML 配置文件定义依赖
- 多种构建系统支持（CMake、Make、Configure、自定义脚本）
- 依赖缓存和版本管理
- 跨平台支持`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 初始化全局设置
		if GlobalCLIContext.GlobalOptions.Verbose {
			fmt.Println("Verbose mode enabled")
		}

		// 初始化上下文
		if err := GlobalCLIContext.Initialize(); err != nil {
			fmt.Printf("Failed to initialize CLI context: %v\n", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 绑定全局标志到上下文
	rootCmd.PersistentFlags().StringVar(&GlobalCLIContext.GlobalOptions.ConfigFile, "config", "", "配置文件路径 (默认为 ./buildfly.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&GlobalCLIContext.GlobalOptions.Verbose, "verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().StringVar(&GlobalCLIContext.GlobalOptions.CacheDir, "cache-dir", "", "缓存目录路径")
	rootCmd.PersistentFlags().StringVar(&GlobalCLIContext.GlobalOptions.MaxCacheAge, "max-cache-age", "7d", "最大缓存时间")

	// 添加子命令
	rootCmd.AddCommand(newInstallCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newCleanCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newDetectCmd())
	rootCmd.AddCommand(newVenvCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newVersionCmd())
}
