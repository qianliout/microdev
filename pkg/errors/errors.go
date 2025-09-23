package errors

import (
	"fmt"
)

// ErrorType 错误类型
type ErrorType int

const (
	// ConfigError 配置错误
	ConfigError ErrorType = iota
	// InputError 输入错误
	InputError
	// OutputError 输出错误
	OutputError
	// LLMError LLM相关错误
	LLMError
	// NetworkError 网络错误
	NetworkError
	// FileError 文件操作错误
	FileError
)

// AppError 应用程序错误
type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewConfigError 创建配置错误
func NewConfigError(message string, cause error) *AppError {
	return &AppError{
		Type:    ConfigError,
		Message: message,
		Cause:   cause,
	}
}

// NewInputError 创建输入错误
func NewInputError(message string, cause error) *AppError {
	return &AppError{
		Type:    InputError,
		Message: message,
		Cause:   cause,
	}
}

// NewOutputError 创建输出错误
func NewOutputError(message string, cause error) *AppError {
	return &AppError{
		Type:    OutputError,
		Message: message,
		Cause:   cause,
	}
}

// NewLLMError 创建LLM错误
func NewLLMError(message string, cause error) *AppError {
	return &AppError{
		Type:    LLMError,
		Message: message,
		Cause:   cause,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, cause error) *AppError {
	return &AppError{
		Type:    NetworkError,
		Message: message,
		Cause:   cause,
	}
}

// NewFileError 创建文件错误
func NewFileError(message string, cause error) *AppError {
	return &AppError{
		Type:    FileError,
		Message: message,
		Cause:   cause,
	}
}

// GetUserFriendlyMessage 获取用户友好的错误消息
func GetUserFriendlyMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ConfigError:
			return fmt.Sprintf("配置错误: %s", appErr.Message)
		case InputError:
			return fmt.Sprintf("输入错误: %s", appErr.Message)
		case OutputError:
			return fmt.Sprintf("输出错误: %s", appErr.Message)
		case LLMError:
			return fmt.Sprintf("AI模型错误: %s", appErr.Message)
		case NetworkError:
			return fmt.Sprintf("网络错误: %s", appErr.Message)
		case FileError:
			return fmt.Sprintf("文件操作错误: %s", appErr.Message)
		default:
			return fmt.Sprintf("未知错误: %s", appErr.Message)
		}
	}
	return err.Error()
}
