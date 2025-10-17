package errors

import (
	"fmt"
)

// 错误类型定义
var (
	// 配置错误
	ErrConfigNotFound      = New("config file not found")
	ErrConfigInvalid       = New("invalid config")
	ErrConfigValidation    = New("config validation failed")
	ErrProjectNameRequired = New("project name is required")
	ErrVersionRequired     = New("version is required")

	// 依赖错误
	ErrDependencyNotFound = New("dependency not found")
	ErrDependencyExists   = New("dependency already exists")
	ErrDependencyConflict = New("dependency conflict")
	ErrVersionConflict    = New("version conflict")
	ErrSourceURLRequired  = New("source URL is required")
	ErrUnsupportedSource  = New("unsupported source type")
	ErrUnsupportedBuild   = New("unsupported build system")

	// 下载错误
	ErrDownloadFailed   = New("download failed")
	ErrInvalidURL       = New("invalid URL")
	ErrNetworkError     = New("network error")
	ErrChecksumMismatch = New("checksum mismatch")
	ErrFileNotFound     = New("file not found")
	ErrPermissionDenied = New("permission denied")

	// 构建错误
	ErrBuildFailed     = New("build failed")
	ErrConfigureFailed = New("configure failed")
	ErrInstallFailed   = New("install failed")
	ErrTestFailed      = New("test failed")
	ErrCommandNotFound = New("command not found")
	ErrScriptExecution = New("script execution failed")

	// 缓存错误
	ErrCacheNotFound   = New("cache not found")
	ErrCacheCorrupted  = New("cache corrupted")
	ErrCachePermission = New("cache permission denied")

	// 系统错误
	ErrOSNotSupported   = New("operating system not supported")
	ErrArchNotSupported = New("architecture not supported")
	ErrCompilerNotFound = New("compiler not found")
)

// Error 自定义错误类型
type Error struct {
	Code    string
	Message string
	Cause   error
}

// New 创建新的错误
func New(message string) *Error {
	return &Error{
		Message: message,
	}
}

// NewWithCode 创建带错误码的错误
func NewWithCode(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(err error, message string) *Error {
	return &Error{
		Message: message,
		Cause:   err,
	}
}

// WrapWithCode 包装错误并添加错误码
func WrapWithCode(err error, code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 支持错误链
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is 支持错误比较
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Message == t.Message && e.Code == t.Code
	}
	return false
}

// ConfigError 配置错误
func ConfigError(message string) *Error {
	return WrapWithCode(nil, "CONFIG_ERROR", message)
}

// ConfigErrorWithCause 带原因的配置错误
func ConfigErrorWithCause(err error, message string) *Error {
	return WrapWithCode(err, "CONFIG_ERROR", message)
}

// DependencyError 依赖错误
func DependencyError(message string) *Error {
	return WrapWithCode(nil, "DEPENDENCY_ERROR", message)
}

// DependencyErrorWithCause 带原因的依赖错误
func DependencyErrorWithCause(err error, message string) *Error {
	return WrapWithCode(err, "DEPENDENCY_ERROR", message)
}

// DownloadError 下载错误
func DownloadError(message string) *Error {
	return WrapWithCode(nil, "DOWNLOAD_ERROR", message)
}

// DownloadErrorWithCause 带原因的下载错误
func DownloadErrorWithCause(err error, message string) *Error {
	return WrapWithCode(err, "DOWNLOAD_ERROR", message)
}

// BuildError 构建错误
func BuildError(message string) *Error {
	return WrapWithCode(nil, "BUILD_ERROR", message)
}

// BuildErrorWithCause 带原因的构建错误
func BuildErrorWithCause(err error, message string) *Error {
	return WrapWithCode(err, "BUILD_ERROR", message)
}

// CacheError 缓存错误
func CacheError(message string) *Error {
	return WrapWithCode(nil, "CACHE_ERROR", message)
}

// CacheErrorWithCause 带原因的缓存错误
func CacheErrorWithCause(err error, message string) *Error {
	return WrapWithCode(err, "CACHE_ERROR", message)
}

// SystemError 系统错误
func SystemError(message string) *Error {
	return WrapWithCode(nil, "SYSTEM_ERROR", message)
}

// SystemErrorWithCause 带原因的系统错误
func SystemErrorWithCause(err error, message string) *Error {
	return WrapWithCode(err, "SYSTEM_ERROR", message)
}
