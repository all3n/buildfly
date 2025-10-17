package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"buildfly/internal/errors"
	"buildfly/pkg/config"
	"buildfly/pkg/venv"
)

// BuildExecutor 构建执行器
type BuildExecutor struct {
	context     *config.VariableContext
	venvManager *venv.Manager
}

// TemplateData 模板数据结构
type TemplateData struct {
	Dependency     config.Dependency
	ProjectName    string
	ProjectVersion string
	Variables      map[string]string
	SourceDir      string
	BuildDir       string
	InstallDir     string
	BuildType      string
	CPUCount       int
	OS             string
	Arch           string
}

// NewBuildExecutor 创建构建执行器
func NewBuildExecutor(ctx *config.VariableContext) *BuildExecutor {
	executor := &BuildExecutor{
		context: ctx,
	}

	// 初始化虚拟环境管理器（如果配置了）
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.VEnv != nil && ctx.ProjectConfig.VEnv.Enabled {
		if manager, err := venv.NewManager(ctx.ProjectConfig.VEnv, ctx.ProjectRoot); err == nil {
			executor.venvManager = manager
		}
	}

	return executor
}

// Execute 执行构建
func (be *BuildExecutor) Execute(dep config.Dependency, sourceDir, buildDir, installDir string) error {
	// 设置构建上下文
	// dir to absolute path
	sourceDir, _ = filepath.Abs(sourceDir)
	buildDir, _ = filepath.Abs(buildDir)
	installDir, _ = filepath.Abs(installDir)
	be.context.SourceDir = sourceDir
	be.context.BuildDir = buildDir
	be.context.InstallDir = installDir

	fmt.Printf("ctx: %+v\n", be.context)

	// 根据构建系统执行构建
	switch dep.BuildSystem {
	case "cmake":
		return be.executeCMake(dep, sourceDir, buildDir, installDir)
	case "make":
		return be.executeMake(dep, sourceDir, buildDir, installDir)
	case "configure":
		return be.executeConfigure(dep, sourceDir, buildDir, installDir)
	case "custom":
		return be.executeCustom(dep, sourceDir, buildDir, installDir)
	default:
		return errors.BuildError(fmt.Sprintf("unsupported build system: %s", dep.BuildSystem))
	}
}

// executeCMake 执行 CMake 构建
func (be *BuildExecutor) executeCMake(dep config.Dependency, sourceDir, buildDir, installDir string) error {
	// 确保构建目录存在
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return errors.BuildErrorWithCause(err, "failed to create build directory")
	}

	// 配置阶段
	if dep.BuildCommands.Configure != "" {
		if err := be.executeCommand(dep.BuildCommands.Configure, buildDir); err != nil {
			return errors.BuildErrorWithCause(err, "CMake configure failed")
		}
	} else {
		// 默认 CMake 配置
		cmakeArgs := []string{
			"-B", buildDir,
			"-S", sourceDir,
			fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", installDir),
			fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", be.context.BuildType),
		}

		// 添加 CMake 选项
		for _, option := range dep.CMakeOptions {
			expandedOption, err := be.context.ExpandCommand(option)
			if err != nil {
				return errors.BuildErrorWithCause(err, fmt.Sprintf("failed to expand CMake option: %s", option))
			}
			cmakeArgs = append(cmakeArgs, expandedOption)
		}

		if err := be.runCommand("cmake", cmakeArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "CMake configure failed")
		}
	}

	// 构建阶段
	if dep.BuildCommands.Build != "" {
		if err := be.executeCommand(dep.BuildCommands.Build, buildDir); err != nil {
			return errors.BuildErrorWithCause(err, "CMake build failed")
		}
	} else {
		// 默认 CMake 构建
		buildArgs := []string{
			"--build", buildDir,
			"--config", be.context.BuildType,
			"--parallel", fmt.Sprintf("%d", be.context.CPUCount),
		}

		if err := be.runCommand("cmake", buildArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "CMake build failed")
		}
	}

	// 安装阶段
	if dep.BuildCommands.Install != "" {
		if err := be.executeCommand(dep.BuildCommands.Install, buildDir); err != nil {
			return errors.BuildErrorWithCause(err, "CMake install failed")
		}
	} else {
		// 默认 CMake 安装
		installArgs := []string{
			"--install", buildDir,
			"--config", be.context.BuildType,
		}

		if err := be.runCommand("cmake", installArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "CMake install failed")
		}
	}

	return nil
}

// executeMake 执行 Make 构建
func (be *BuildExecutor) executeMake(dep config.Dependency, sourceDir, buildDir, installDir string) error {
	// 确保构建目录存在
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return errors.BuildErrorWithCause(err, "failed to create build directory")
	}

	// 配置阶段（如果需要）
	if dep.BuildCommands.Configure != "" {
		if err := be.executeCommand(dep.BuildCommands.Configure, sourceDir); err != nil {
			return errors.BuildErrorWithCause(err, "Make configure failed")
		}
	}

	// 构建阶段
	if dep.BuildCommands.Build != "" {
		if err := be.executeCommand(dep.BuildCommands.Build, sourceDir); err != nil {
			return errors.BuildErrorWithCause(err, "Make build failed")
		}
	} else {
		// 默认 Make 构建
		makeArgs := []string{
			fmt.Sprintf("-j%d", be.context.CPUCount),
		}

		// 添加 Make 选项
		for _, option := range dep.MakeOptions {
			expandedOption, err := be.context.ExpandCommand(option)
			if err != nil {
				return errors.BuildErrorWithCause(err, fmt.Sprintf("failed to expand Make option: %s", option))
			}
			makeArgs = append(makeArgs, expandedOption)
		}

		if err := be.runCommandInDir("make", sourceDir, makeArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "Make build failed")
		}
	}

	// 安装阶段
	if dep.BuildCommands.Install != "" {
		if err := be.executeCommand(dep.BuildCommands.Install, sourceDir); err != nil {
			return errors.BuildErrorWithCause(err, "Make install failed")
		}
	} else {
		// 默认 Make 安装
		installArgs := []string{"install"}

		if err := be.runCommandInDir("make", sourceDir, installArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "Make install failed")
		}
	}

	return nil
}

// executeConfigure 执行 Configure 构建
func (be *BuildExecutor) executeConfigure(dep config.Dependency, sourceDir, buildDir, installDir string) error {
	// 确保构建目录存在
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return errors.BuildErrorWithCause(err, "failed to create build directory")
	}

	// 配置阶段
	if dep.BuildCommands.Configure != "" {
		if err := be.executeCommand(dep.BuildCommands.Configure, sourceDir); err != nil {
			return errors.BuildErrorWithCause(err, "Configure failed")
		}
	} else {
		// 默认 Configure
		configurePath := filepath.Join(sourceDir, "configure")
		configureArgs := []string{
			fmt.Sprintf("--prefix=%s", installDir),
		}

		// 添加 Configure 选项
		for _, option := range dep.ConfigureOptions {
			expandedOption, err := be.context.ExpandCommand(option)
			if err != nil {
				return errors.BuildErrorWithCause(err, fmt.Sprintf("failed to expand Configure option: %s", option))
			}
			configureArgs = append(configureArgs, expandedOption)
		}

		if err := be.runCommandInDir(configurePath, sourceDir, configureArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "Configure failed")
		}
	}

	// 构建阶段
	if dep.BuildCommands.Build != "" {
		if err := be.executeCommand(dep.BuildCommands.Build, sourceDir); err != nil {
			return errors.BuildErrorWithCause(err, "Make build failed")
		}
	} else {
		// 默认 Make 构建
		makeArgs := []string{
			fmt.Sprintf("-j%d", be.context.CPUCount),
		}

		if err := be.runCommandInDir("make", sourceDir, makeArgs...); err != nil {
			return errors.BuildErrorWithCause(err, "Make build failed")
		}
	}

	// 安装阶段
	if dep.BuildCommands.Install != "" {
		if err := be.executeCommand(dep.BuildCommands.Install, sourceDir); err != nil {
			return errors.BuildErrorWithCause(err, "Make install failed")
		}
	} else {
		// 默认 Make 安装
		if err := be.runCommandInDir("make", sourceDir, "install"); err != nil {
			return errors.BuildErrorWithCause(err, "Make install failed")
		}
	}

	return nil
}

// executeCustom 执行自定义构建脚本
func (be *BuildExecutor) executeCustom(dep config.Dependency, sourceDir, buildDir, installDir string) error {
	if dep.CustomScript == "" {
		return errors.BuildError("custom build system requires custom_script")
	}

	// 自定义构建脚本在构建目录执行
	if err := be.executeScriptWithDependency(dep.CustomScript, buildDir, dep); err != nil {
		return errors.BuildErrorWithCause(err, "custom script execution failed")
	}

	return nil
}

// executeCommand 执行单个命令
func (be *BuildExecutor) executeCommand(commandLine, workDir string) error {
	// 展开命令中的变量
	expandedCommand, err := be.context.ExpandCommand(commandLine)
	if err != nil {
		return errors.BuildErrorWithCause(err, "failed to expand command variables")
	}

	// 使用 shell 执行命令（支持管道和重定向）
	cmd := exec.Command("sh", "-c", expandedCommand)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量
	if be.context.CustomVars != nil {
		env := os.Environ()
		for key, value := range be.context.CustomVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	return cmd.Run()
}

// executeScript 执行脚本（支持Go template）
func (be *BuildExecutor) executeScript(script, workDir string) error {
	return be.executeScriptWithDependency(script, workDir, config.Dependency{})
}

// executeScriptWithDependency 执行脚本（支持Go template，包含依赖信息）
func (be *BuildExecutor) executeScriptWithDependency(script, workDir string, dep config.Dependency) error {
	// 创建模板数据 转成绝对路径
	templateData := TemplateData{
		Dependency:     dep,
		ProjectName:    be.context.ProjectName,
		ProjectVersion: be.context.ProjectVersion,
		Variables:      be.context.CustomVars,
		SourceDir:      be.context.SourceDir,
		BuildDir:       be.context.BuildDir,
		InstallDir:     be.context.InstallDir,
		BuildType:      be.context.BuildType,
		CPUCount:       be.context.CPUCount,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
	}

	// 解析并执行模板
	tmpl, err := template.New("buildscript").Parse(script)
	if err != nil {
		return errors.BuildErrorWithCause(err, "failed to parse script template")
	}

	// 生成脚本内容
	var scriptContent bytes.Buffer
	if err := tmpl.Execute(&scriptContent, templateData); err != nil {
		return errors.BuildErrorWithCause(err, "failed to execute script template")
	}

	// 创建脚本文件
	scriptName := ".buildfly_build_script.sh"
	scriptPath := filepath.Join(workDir, scriptName)
	if err := os.WriteFile(scriptPath, scriptContent.Bytes(), 0755); err != nil {
		return errors.BuildErrorWithCause(err, "failed to write script file")
	}

	// 执行脚本
	cmd := exec.Command("bash", scriptName)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量
	if be.context.CustomVars != nil {
		env := os.Environ()
		for key, value := range be.context.CustomVars {
			fmt.Printf("Run With Env: %s=%s\n", key, value)
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	return cmd.Run()
}

// runCommand 运行命令
func (be *BuildExecutor) runCommand(name string, args ...string) error {
	return be.runCommandWithEnv(name, "", args...)
}

// runCommandInDir 在指定目录运行命令
func (be *BuildExecutor) runCommandInDir(name, dir string, args ...string) error {
	return be.runCommandWithEnv(name, dir, args...)
}

// runCommandWithEnv 在指定目录和环境变量下运行命令
func (be *BuildExecutor) runCommandWithEnv(name, dir string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量
	env := os.Environ()

	// 添加自定义变量
	if be.context.CustomVars != nil {
		for key, value := range be.context.CustomVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// 如果有虚拟环境管理器，添加虚拟环境变量
	if be.venvManager != nil {
		if venvVars, err := be.venvManager.GetEnvironmentVars(); err == nil {
			for key, value := range venvVars {
				env = append(env, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	cmd.Env = env
	return cmd.Run()
}

// ValidateBuildSystem 验证构建系统
func (be *BuildExecutor) ValidateBuildSystem(dep config.Dependency) error {
	supportedSystems := map[string]bool{
		"cmake":     true,
		"make":      true,
		"configure": true,
		"custom":    true,
	}

	if !supportedSystems[dep.BuildSystem] {
		return errors.BuildError(fmt.Sprintf("unsupported build system: %s", dep.BuildSystem))
	}

	// 验证自定义构建系统
	if dep.BuildSystem == "custom" && dep.CustomScript == "" {
		return errors.BuildError("custom build system requires custom_script")
	}

	return nil
}

// GetBuildOutputPath 获取构建输出路径
func (be *BuildExecutor) GetBuildOutputPath(dep config.Dependency) string {
	switch dep.BuildSystem {
	case "cmake":
		return be.context.BuildDir
	case "make", "configure":
		return be.context.SourceDir
	case "custom":
		// 自定义脚本可能输出到任意位置，默认返回构建目录
		return be.context.BuildDir
	default:
		return be.context.BuildDir
	}
}

// GetInstallPath 获取安装路径
func (be *BuildExecutor) GetInstallPath() string {
	return be.context.InstallDir
}

// CheckBuildTools 检查构建工具是否可用
func (be *BuildExecutor) CheckBuildTools(dep config.Dependency) error {
	// 如果有虚拟环境管理器，先检查虚拟环境中的工具
	if be.venvManager != nil {
		if be.venvManager.IsActivated() {
			switch dep.BuildSystem {
			case "cmake":
				if be.venvManager.IsToolInstalled("cmake") {
					return nil // 在虚拟环境中找到cmake
				}
			case "make", "configure":
				if be.venvManager.IsToolInstalled("make") {
					return nil // 在虚拟环境中找到make
				}
			}
		}
	}

	// 回退到系统PATH检查
	switch dep.BuildSystem {
	case "cmake":
		if _, err := exec.LookPath("cmake"); err != nil {
			return errors.BuildError("cmake not found in PATH or virtual environment")
		}
	case "make", "configure":
		if _, err := exec.LookPath("make"); err != nil {
			return errors.BuildError("make not found in PATH or virtual environment")
		}
	case "custom":
		// 自定义脚本可能使用任意工具，无法预先检查
	}

	return nil
}

// GetBuildCommands 获取构建命令
func (be *BuildExecutor) GetBuildCommands(dep config.Dependency) (configure, build, install string) {
	if dep.BuildCommands.Configure != "" {
		configure, _ = be.context.ExpandCommand(dep.BuildCommands.Configure)
	}
	if dep.BuildCommands.Build != "" {
		build, _ = be.context.ExpandCommand(dep.BuildCommands.Build)
	}
	if dep.BuildCommands.Install != "" {
		install, _ = be.context.ExpandCommand(dep.BuildCommands.Install)
	}
	return
}

// ExecuteCommandWithOutput 执行命令并捕获输出
func (be *BuildExecutor) ExecuteCommandWithOutput(commandLine, workDir string) (string, error) {
	// 展开命令中的变量
	expandedCommand, err := be.context.ExpandCommand(commandLine)
	if err != nil {
		return "", errors.BuildErrorWithCause(err, "failed to expand command variables")
	}

	// 使用 shell 执行命令
	cmd := exec.Command("sh", "-c", expandedCommand)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.BuildErrorWithCause(err, fmt.Sprintf("command failed: %s\nstderr: %s", expandedCommand, stderr.String()))
	}

	return stdout.String(), nil
}
