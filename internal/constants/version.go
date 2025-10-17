package constants

// Version 应用版本号
var Version = "0.1.0"

// GitCommit Git 提交哈希，构建时通过 ldflags 设置
var GitCommit = "unknown"

// BuildDate 构建日期，构建时通过 ldflags 设置
var BuildDate = "unknown"

// GoVersion Go 版本，构建时通过 ldflags 设置
var GoVersion = "unknown"

// BuildInfo 构建信息
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
}

// GetBuildInfo 获取构建信息
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
	}
}
