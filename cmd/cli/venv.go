package cli

import (
	"buildfly/pkg/venv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newVenvCmd 创建 venv 命令组
func newVenvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "venv",
		Short: "管理 C++ 虚拟环境",
		Long: `管理基于 UV 的 C++ 虚拟环境。

支持创建、激活、停用和管理隔离的 C++ 构建环境。`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 确保上下文已初始化
			if err := GlobalCLIContext.Initialize(); err != nil {
				fmt.Printf("Failed to initialize CLI context: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// 添加子命令
	cmd.AddCommand(newVenvInitCmd())
	cmd.AddCommand(newVenvActivateCmd())
	cmd.AddCommand(newVenvDeactivateCmd())
	cmd.AddCommand(newVenvStatusCmd())
	cmd.AddCommand(newVenvResetCmd())
	cmd.AddCommand(newVenvListCmd())
	cmd.AddCommand(newVenvRunCmd())

	return cmd
}

// newVenvInitCmd 创建 venv init 命令
func newVenvInitCmd() *cobra.Command {
	var (
		force     bool
		uvVersion string
		rootDir   string
		backend   string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化 C++ 虚拟环境",
		Long: `在项目中初始化基于 UV 或 Pixi 的 C++ 虚拟环境。

创建 .buildfly/root 目录并安装必要的构建工具。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvInit(force, uvVersion, rootDir, backend)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "强制重新初始化")
	cmd.Flags().StringVar(&uvVersion, "uv-version", "latest", "UV 版本")
	cmd.Flags().StringVar(&rootDir, "root-dir", "", "环境根目录")
	cmd.Flags().StringVar(&backend, "backend", "", "后端类型 (uv|pixi)")

	return cmd
}

// newVenvActivateCmd 创建 venv activate 命令
func newVenvActivateCmd() *cobra.Command {
	var shell bool

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "激活 C++ 虚拟环境",
		Long: `激活项目的 C++ 虚拟环境。

设置环境变量和工具路径，确保使用隔离的构建环境。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvActivate(shell)
		},
	}

	cmd.Flags().BoolVar(&shell, "shell", false, "输出激活脚本供 shell 执行")

	return cmd
}

// newVenvDeactivateCmd 创建 venv deactivate 命令
func newVenvDeactivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "停用 C++ 虚拟环境",
		Long: `停用当前激活的 C++ 虚拟环境。

清理环境变量，恢复系统默认状态。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvDeactivate()
		},
	}

	return cmd
}

// newVenvStatusCmd 创建 venv status 命令
func newVenvStatusCmd() *cobra.Command {
	var (
		verbose bool
		json    bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看虚拟环境状态",
		Long: `显示 C++ 虚拟环境的当前状态信息。

包括环境路径、激活状态、已安装工具等信息。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvStatus(verbose, json)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细信息")
	cmd.Flags().BoolVar(&json, "json", false, "以 JSON 格式输出")

	return cmd
}

// newVenvResetCmd 创建 venv reset 命令
func newVenvResetCmd() *cobra.Command {
	var (
		force bool
	)

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "重置 C++ 虚拟环境",
		Long: `完全重置 C++ 虚拟环境。

删除所有安装的工具和环境配置，重新开始。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvReset(force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "强制重置不询问确认")

	return cmd
}

// newVenvListCmd 创建 venv list 命令
func newVenvListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出已安装的工具",
		Long: `列出虚拟环境中已安装的 C++ 构建工具。

显示工具名称、版本和安装路径。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvList()
		},
	}

	return cmd
}

// runVenvInit 执行环境初始化
func runVenvInit(force bool, uvVersion string, rootDir string, backend string) error {
	projectConfig := GlobalCLIContext.ProjectConfig

	// 获取项目根目录
	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	// 创建或获取 VENV 配置
	venvConfig := projectConfig.VEnv
	if venvConfig == nil {
		venvConfig = &venv.VEnvConfig{
			Enabled:      true,
			UVVersion:    uvVersion,
			RootDir:      rootDir,
			AutoActivate: false,
			CPPTools: venv.CPPToolsConfig{
				CMake: venv.ToolConfig{
					Version: "3.28.0",
					Enabled: true,
				},
				Ninja: venv.ToolConfig{
					Version: "1.11.1",
					Enabled: true,
				},
			},
			Python: venv.PythonConfig{
				Version:  "3.11",
				Packages: []string{"cmake", "ninja"},
			},
		}
	} else {
		// 更新配置
		if uvVersion != "" && uvVersion != "latest" {
			venvConfig.UVVersion = uvVersion
		}
		if rootDir != "" {
			venvConfig.RootDir = rootDir
		}
		venvConfig.Enabled = true
	}

	// 设置 backend
	if backend != "" {
		venvConfig.Backend = venv.BackendType(backend)
	}

	fmt.Printf("Initializing virtual environment with backend %s in projectRoot:%s venvRoot:%s\n",
		venvConfig.Backend, projectRoot, venvConfig.RootDir)

	// 检查是否已初始化
	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	// 检查环境是否已存在
	if !force {
		status := manager.GetStatus()

		if status.RootDir != "" {
			// 检查环境目录是否存在
			if _, err := os.Stat(status.RootDir); err == nil {
				fmt.Printf("Virtual environment already exists at: %s\n", status.RootDir)
				fmt.Println("Use --force to reinitialize")
				return nil
			}
		}
	}

	// 初始化环境
	if err := manager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize virtual environment: %w", err)
	}

	// 显示激活脚本路径
	activationScript := manager.GetActivationScript()
	fmt.Printf("\nTo activate the environment, run:\n")
	fmt.Printf("  source %s\n", activationScript)
	fmt.Printf("Or use: buildfly venv activate\n")

	return nil
}

// newVenvRunCmd 创建 venv run 命令
func newVenvRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command...]",
		Short: "在虚拟环境中运行命令",
		Long: `在虚拟环境中运行指定的命令。

如果虚拟环境未初始化，会自动初始化环境。
支持运行任何已安装的构建工具或命令。`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVenvRun(args)
		},
	}

	return cmd
}

// runVenvRun 执行命令运行
func runVenvRun(args []string) error {
	projectConfig := GlobalCLIContext.ProjectConfig

	// 获取项目根目录
	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	// 创建或获取 VENV 配置
	venvConfig := projectConfig.VEnv
	if venvConfig == nil {
		// 使用默认配置
		venvConfig = &venv.VEnvConfig{
			Enabled:      true,
			UVVersion:    "latest",
			RootDir:      "",
			AutoActivate: false,
			CPPTools: venv.CPPToolsConfig{
				CMake: venv.ToolConfig{
					Version: "3.28.0",
					Enabled: true,
				},
				Ninja: venv.ToolConfig{
					Version: "1.11.1",
					Enabled: true,
				},
			},
			Python: venv.PythonConfig{
				Version:  "3.11",
				Packages: []string{"cmake", "ninja"},
			},
		}
	}

	// 创建管理器
	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	// 运行命令
	if err := manager.Run(args); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}

// runVenvActivate 执行环境激活
func runVenvActivate(shell bool) error {
	projectConfig := GlobalCLIContext.ProjectConfig
	venvConfig := projectConfig.VEnv
	if venvConfig == nil || !venvConfig.Enabled {
		return fmt.Errorf("virtual environment not configured or disabled")
	}

	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	if shell {
		// 输出激活脚本内容
		activationScript := manager.GetActivationScript()
		content, err := os.ReadFile(activationScript)
		if err != nil {
			return fmt.Errorf("failed to read activation script: %w", err)
		}
		fmt.Print(string(content))
		return nil
	}

	// 直接激活环境
	if err := manager.Activate(); err != nil {
		return fmt.Errorf("failed to activate virtual environment: %w", err)
	}

	// 显示环境信息
	status := manager.GetStatus()
	fmt.Printf("Environment activated successfully!\n")
	fmt.Printf("Root: %s\n", status.RootDir)
	fmt.Printf("Backend: %s\n", venvConfig.Backend)
	if venvConfig.Backend == "uv" && status.UVInstalled {
		fmt.Printf("UV Version: %s\n", status.UVVersion)
	}

	return nil
}

// runVenvDeactivate 执行环境停用
func runVenvDeactivate() error {
	// 检查是否有激活的环境
	envRoot := os.Getenv("BUILDFLY_ENV_ROOT")
	if envRoot == "" {
		fmt.Println("No active virtual environment found")
		return nil
	}

	// 创建管理器来停用环境
	venvConfig := &venv.VEnvConfig{
		RootDir: envRoot,
	}

	manager, err := venv.NewManager(venvConfig, filepath.Dir(envRoot))
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	if err := manager.Deactivate(); err != nil {
		return fmt.Errorf("failed to deactivate virtual environment: %w", err)
	}

	return nil
}

// runVenvStatus 执行状态查看
func runVenvStatus(verbose bool, jsonOutput bool) error {
	projectConfig := GlobalCLIContext.ProjectConfig
	venvConfig := projectConfig.VEnv

	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	// 如果没有配置，显示未配置状态
	if venvConfig == nil {
		if jsonOutput {
			fmt.Println(`{"status": "not_configured", "message": "Virtual environment not configured"}`)
		} else {
			fmt.Println("Virtual environment not configured")
			fmt.Println("Use 'buildfly venv init' to initialize")
		}
		return nil
	}

	// 在详细模式下显示配置信息
	if verbose {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Backend: %s\n", venvConfig.Backend)
		fmt.Printf("  Enabled: %t\n", venvConfig.Enabled)
		if venvConfig.RootDir != "" {
			fmt.Printf("  Root Directory: %s\n", venvConfig.RootDir)
		}
		if venvConfig.UVVersion != "" {
			fmt.Printf("  UV Version: %s\n", venvConfig.UVVersion)
		}
		fmt.Printf("  Auto Activate: %t\n", venvConfig.AutoActivate)

		// 显示 C++ 工具配置
		fmt.Printf("  C++ Tools:\n")
		fmt.Printf("    CMake: enabled=%t, version=%s\n", venvConfig.CPPTools.CMake.Enabled, venvConfig.CPPTools.CMake.Version)
		fmt.Printf("    Ninja: enabled=%t, version=%s\n", venvConfig.CPPTools.Ninja.Enabled, venvConfig.CPPTools.Ninja.Version)

		// 显示 Python 配置
		fmt.Printf("  Python: version=%s, packages=%v\n", venvConfig.Python.Version, venvConfig.Python.Packages)
		fmt.Printf("\n")
	}

	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	status := manager.GetStatus()

	if jsonOutput {
		// JSON 输出
		if venvConfig.Backend == "uv" {
			fmt.Printf(`{"status": "ok", "backend": "uv", "root_dir": "%s", "activated": %t, "uv_installed": %t, "uv_version": "%s"}`,
				status.RootDir, status.Activated, status.UVInstalled, status.UVVersion)
		} else if venvConfig.Backend == "pixi" {
			fmt.Printf(`{"status": "ok", "backend": "pixi", "root_dir": "%s", "activated": %t, "pixi_installed": %t, "pixi_version": "%s"}`,
				status.RootDir, status.Activated, status.PixiInstalled, status.PixiVersion)
		} else {
			fmt.Printf(`{"status": "ok", "backend": "%s", "root_dir": "%s", "activated": %t}`,
				venvConfig.Backend, status.RootDir, status.Activated)
		}
	} else {
		// 人类可读输出
		fmt.Printf("Virtual Environment Status:\n")
		fmt.Printf("  Backend: %s\n", venvConfig.Backend)
		fmt.Printf("  Root Directory: %s\n", status.RootDir)
		fmt.Printf("  Activated: %t\n", status.Activated)

		if venvConfig.Backend == "uv" {
			fmt.Printf("  UV Installed: %t\n", status.UVInstalled)
			if status.UVInstalled {
				fmt.Printf("  UV Version: %s\n", status.UVVersion)
			}
		} else if venvConfig.Backend == "pixi" {
			fmt.Printf("  Pixi Installed: %t\n", status.PixiInstalled)
			if status.PixiInstalled {
				fmt.Printf("  Pixi Version: %s\n", status.PixiVersion)
			}
		}

		fmt.Printf("  Python Version: %s\n", status.PythonVersion)
		fmt.Printf("  Created: %s\n", status.CreatedAt.Format("2006-01-02 15:04:05"))
		if status.Activated {
			fmt.Printf("  Last Activated: %s\n", status.LastActivated.Format("2006-01-02 15:04:05"))
		}

		if verbose && len(status.InstalledTools) > 0 {
			fmt.Printf("\nInstalled Tools:\n")
			for name, version := range status.InstalledTools {
				fmt.Printf("  %s: %s\n", name, version)
			}
		}
	}

	return nil
}

// runVenvReset 执行环境重置
func runVenvReset(force bool) error {
	projectConfig := GlobalCLIContext.ProjectConfig
	venvConfig := projectConfig.VEnv
	if venvConfig == nil {
		return fmt.Errorf("virtual environment not configured")
	}

	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	// 检查环境是否存在
	status := manager.GetStatus()

	if status.RootDir == "" {
		fmt.Println("No virtual environment to reset")
		return nil
	}

	// 确认重置
	if !force {
		fmt.Printf("Are you sure you want to reset the virtual environment at %s? [y/N]: ", status.RootDir)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Reset cancelled")
			return nil
		}
	}

	// 停用环境（如果已激活）
	if manager.IsActivated() {
		if err := manager.Deactivate(); err != nil {
			fmt.Printf("Warning: failed to deactivate environment: %v\n", err)
		}
	}

	// 重置环境
	if err := manager.Reset(); err != nil {
		return fmt.Errorf("failed to reset virtual environment: %w", err)
	}

	fmt.Println("Virtual environment reset successfully")
	return nil
}

// runVenvList 执行工具列表
func runVenvList() error {
	projectConfig := GlobalCLIContext.ProjectConfig
	venvConfig := projectConfig.VEnv
	if venvConfig == nil {
		return fmt.Errorf("virtual environment not configured")
	}

	projectRoot := GlobalCLIContext.ProjectConfig.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		projectRoot = cwd
	}

	manager, err := venv.NewManager(venvConfig, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create venv manager: %w", err)
	}

	status := manager.GetStatus()

	if len(status.InstalledTools) == 0 {
		fmt.Println("No tools installed yet")
		return nil
	}

	fmt.Println("Installed Tools:")
	for name, version := range status.InstalledTools {
		fmt.Printf("  %s: %s\n", name, version)
	}

	return nil
}
