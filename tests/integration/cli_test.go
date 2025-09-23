package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey = "test-api-key-for-integration-tests"
)

func TestMainCommand(t *testing.T) {
	// Build the binary first
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	t.Run("no arguments", func(t *testing.T) {
		cmd := exec.Command(binaryPath)
		output, err := cmd.CombinedOutput()

		// Should fail due to missing input argument
		assert.Error(t, err)
		outputStr := string(output)
		assert.Contains(t, outputStr, "使用方法")
		assert.Contains(t, outputStr, "input_content")
	})
}

func TestPromptWithoutAPIKey(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Ensure no API keys are set
	os.Unsetenv("DASHSCOPE_API_KEY")
	os.Unsetenv("ALI_BAILIAN_API_KEY")

	t.Run("missing API key", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "test input")
		cmd.Env = []string{} // Empty environment
		output, err := cmd.CombinedOutput()

		// Should fail due to missing API key
		assert.Error(t, err)
		outputStr := string(output)
		assert.Contains(t, outputStr, "配置错误")
		assert.Contains(t, outputStr, "API")
	})
}

func TestPromptWithMockAPI(t *testing.T) {
	// Skip this test if we don't want to test against real API
	// This test would need a mock server or test API key
	t.Skip("Skipping test that requires real API key")

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Set test API key
	os.Setenv("DASHSCOPE_API_KEY", testAPIKey)
	defer os.Unsetenv("DASHSCOPE_API_KEY")

	t.Run("direct text input", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "请优化这个提示词")
		output, err := cmd.CombinedOutput()

		if err != nil {
			// If it fails due to network/API issues, that's expected in tests
			t.Logf("Command failed (expected in test environment): %v", err)
			t.Logf("Output: %s", string(output))
			return
		}

		outputStr := string(output)
		assert.NotEmpty(t, outputStr)
	})
}

func TestFileInputOutput(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	tmpDir := t.TempDir()

	t.Run("empty file input", func(t *testing.T) {
		// Create empty input file
		inputFile := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(inputFile, []byte(""), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binaryPath, inputFile)
		cmd.Env = []string{"DASHSCOPE_API_KEY=" + testAPIKey}
		output, err := cmd.CombinedOutput()

		// Should fail due to empty file
		assert.Error(t, err)
		outputStr := string(output)
		assert.Contains(t, outputStr, "为空")
	})
}

// buildBinary builds the outback binary for testing
func buildBinary(t *testing.T) string {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "outback")

	// Get the project root directory
	projectRoot, err := filepath.Abs("../..")
	require.NoError(t, err)

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/prompt")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}

	return binaryPath
}

// testFileExists checks if a file exists
func testFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// createTestFile creates a test file with given content
func createTestFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}
