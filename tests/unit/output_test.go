package unit

import (
	"os"
	"path/filepath"
	"testing"

	"microdev/pkg/logger"
	"microdev/pkg/output"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputProcessor_ConsoleOutput(t *testing.T) {
	log := logger.NewLogger()
	processor := output.NewProcessor("", log)

	assert.True(t, processor.IsConsoleOutput())
	assert.Equal(t, "", processor.GetOutputPath())

	// Test writing to console (we can't easily test stdout, but we can ensure no error)
	n, err := processor.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	// Test WriteString
	n, err = processor.WriteString("test string")
	assert.NoError(t, err)
	assert.Equal(t, 11, n)

	// Test Flush (should be no-op for console)
	err = processor.Flush()
	assert.NoError(t, err)

	// Test Close (should be no-op for console)
	err = processor.Close()
	assert.NoError(t, err)
}

func TestOutputProcessor_FileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	log := logger.NewLogger()
	processor := output.NewProcessor(outputFile, log)

	assert.False(t, processor.IsConsoleOutput())
	assert.Equal(t, outputFile, processor.GetOutputPath())

	// Test writing to file
	testData := "测试输出内容"
	n, err := processor.Write([]byte(testData))
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)

	// Test WriteString
	moreData := "\n更多内容"
	n, err = processor.WriteString(moreData)
	require.NoError(t, err)
	assert.Equal(t, len(moreData), n)

	// Test Flush
	err = processor.Flush()
	require.NoError(t, err)

	// Test Close
	err = processor.Close()
	require.NoError(t, err)

	// Verify file contents
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	expected := testData + moreData
	assert.Equal(t, expected, string(content))

	// Test closing again (should be no-op)
	err = processor.Close()
	assert.NoError(t, err)
}

func TestOutputProcessor_FileAppend(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "append.txt")

	// Write initial content
	initialContent := "初始内容\n"
	err := os.WriteFile(outputFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	log := logger.NewLogger()
	processor := output.NewProcessor(outputFile, log)

	// Write additional content
	additionalContent := "追加内容\n"
	n, err := processor.Write([]byte(additionalContent))
	require.NoError(t, err)
	assert.Equal(t, len(additionalContent), n)

	err = processor.Close()
	require.NoError(t, err)

	// Verify file contents include both initial and additional content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	expected := initialContent + additionalContent
	assert.Equal(t, expected, string(content))
}

func TestOutputProcessor_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "subdir", "nested", "output.txt")

	log := logger.NewLogger()
	processor := output.NewProcessor(outputFile, log)

	// Write to file (should create directories)
	testData := "测试目录创建"
	n, err := processor.Write([]byte(testData))
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)

	err = processor.Close()
	require.NoError(t, err)

	// Verify file exists and has correct content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, testData, string(content))

	// Verify directories were created
	info, err := os.Stat(filepath.Dir(outputFile))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestOutputProcessor_WriteAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.txt")

	log := logger.NewLogger()
	processor := output.NewProcessor(outputFile, log)

	// Write initial content
	_, err := processor.Write([]byte("initial"))
	require.NoError(t, err)

	// Close the processor
	err = processor.Close()
	require.NoError(t, err)

	// Try to write after close (should reopen file)
	_, err = processor.Write([]byte("after close"))
	require.NoError(t, err)

	err = processor.Close()
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "initialafter close", string(content))
}
