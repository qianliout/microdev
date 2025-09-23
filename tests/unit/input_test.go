package unit

import (
	"os"
	"path/filepath"
	"testing"

	"microdev/pkg/input"
	"microdev/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputProcessor_Process(t *testing.T) {
	log := logger.NewLogger()
	processor := input.NewProcessor(log)

	t.Run("empty input", func(t *testing.T) {
		result, err := processor.Process("")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "输入内容不能为空")
	})

	t.Run("direct text input", func(t *testing.T) {
		input := "这是一个测试文本"
		result, err := processor.Process(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("text with whitespace", func(t *testing.T) {
		input := "  \n这是一个测试文本\n  "
		expected := "这是一个测试文本"
		result, err := processor.Process(input)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("valid file path", func(t *testing.T) {
		// 创建临时文件
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "这是文件内容\n测试文本"

		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := processor.Process(testFile)
		require.NoError(t, err)
		assert.Equal(t, "这是文件内容\n测试文本", result)
	})

	t.Run("empty file", func(t *testing.T) {
		// 创建空的临时文件
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "empty.txt")

		err := os.WriteFile(testFile, []byte(""), 0644)
		require.NoError(t, err)

		result, err := processor.Process(testFile)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "为空")
	})

	t.Run("directory path", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 传入目录路径应该作为文本处理，而不是文件
		result, err := processor.Process(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, tmpDir, result)
	})

	t.Run("non-existent file path", func(t *testing.T) {
		nonExistentPath := "/path/that/does/not/exist.txt"

		// 不存在的路径应该作为文本处理
		result, err := processor.Process(nonExistentPath)
		require.NoError(t, err)
		assert.Equal(t, nonExistentPath, result)
	})

	t.Run("file with whitespace content", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "whitespace.txt")
		content := "  \n\t文件内容  \n\t"

		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := processor.Process(testFile)
		require.NoError(t, err)
		assert.Equal(t, "文件内容", result)
	})
}
