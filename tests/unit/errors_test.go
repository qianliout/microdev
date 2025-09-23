package unit

import (
	"errors"
	"testing"

	apperrors "microdev/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := apperrors.NewConfigError("config failed", cause)

		assert.Equal(t, apperrors.ConfigError, err.Type)
		assert.Equal(t, "config failed: underlying error", err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("error without cause", func(t *testing.T) {
		err := apperrors.NewInputError("input invalid", nil)

		assert.Equal(t, apperrors.InputError, err.Type)
		assert.Equal(t, "input invalid", err.Error())
		assert.Nil(t, err.Unwrap())
	})
}

func TestErrorConstructors(t *testing.T) {
	cause := errors.New("test cause")

	tests := []struct {
		name         string
		constructor  func(string, error) *apperrors.AppError
		expectedType apperrors.ErrorType
	}{
		{"config error", apperrors.NewConfigError, apperrors.ConfigError},
		{"input error", apperrors.NewInputError, apperrors.InputError},
		{"output error", apperrors.NewOutputError, apperrors.OutputError},
		{"llm error", apperrors.NewLLMError, apperrors.LLMError},
		{"network error", apperrors.NewNetworkError, apperrors.NetworkError},
		{"file error", apperrors.NewFileError, apperrors.FileError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message", cause)
			assert.Equal(t, tt.expectedType, err.Type)
			assert.Equal(t, "test message", err.Message)
			assert.Equal(t, cause, err.Cause)
		})
	}
}

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "config error",
			err:      apperrors.NewConfigError("missing key", nil),
			expected: "配置错误: missing key",
		},
		{
			name:     "input error",
			err:      apperrors.NewInputError("empty input", nil),
			expected: "输入错误: empty input",
		},
		{
			name:     "output error",
			err:      apperrors.NewOutputError("write failed", nil),
			expected: "输出错误: write failed",
		},
		{
			name:     "llm error",
			err:      apperrors.NewLLMError("model failed", nil),
			expected: "AI模型错误: model failed",
		},
		{
			name:     "network error",
			err:      apperrors.NewNetworkError("connection failed", nil),
			expected: "网络错误: connection failed",
		},
		{
			name:     "file error",
			err:      apperrors.NewFileError("file not found", nil),
			expected: "文件操作错误: file not found",
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: "standard error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := apperrors.GetUserFriendlyMessage(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
