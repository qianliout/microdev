package unit

import (
	"os"
	"testing"

	"microdev/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		dashScopeKey    string
		bailianKey      string
		expectError     bool
		expectModelName string
	}{
		{
			name:            "with dashscope key",
			dashScopeKey:    "test-dashscope-key",
			bailianKey:      "",
			expectError:     false,
			expectModelName: "qwen-plus-latest",
		},
		{
			name:            "with bailian key",
			dashScopeKey:    "",
			bailianKey:      "test-bailian-key",
			expectError:     false,
			expectModelName: "qwen-plus-latest",
		},
		{
			name:            "with both keys",
			dashScopeKey:    "test-dashscope-key",
			bailianKey:      "test-bailian-key",
			expectError:     false,
			expectModelName: "qwen-plus-latest",
		},
		{
			name:         "no keys",
			dashScopeKey: "",
			bailianKey:   "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境变量
			os.Setenv("DASHSCOPE_API_KEY", tt.dashScopeKey)
			os.Setenv("ALI_BAILIAN_API_KEY", tt.bailianKey)
			defer func() {
				os.Unsetenv("DASHSCOPE_API_KEY")
				os.Unsetenv("ALI_BAILIAN_API_KEY")
			}()

			cfg, err := config.LoadConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				assert.Equal(t, tt.expectModelName, cfg.ModelName)
				assert.Equal(t, 2048, cfg.MaxTokens)
				assert.Equal(t, 0.7, cfg.Temperature)
				assert.True(t, cfg.StreamOutput)
				assert.Equal(t, tt.dashScopeKey, cfg.DashScopeAPIKey)
				assert.Equal(t, tt.bailianKey, cfg.AliBailianAPIKey)
			}
		})
	}
}

func TestConfigMethods(t *testing.T) {
	cfg := &config.Config{
		DashScopeAPIKey:  "test-dashscope",
		AliBailianAPIKey: "test-bailian",
	}

	// Test GetAPIKey - should return DashScope key first
	assert.Equal(t, "test-dashscope", cfg.GetAPIKey())

	// Test HasDashScopeKey
	assert.True(t, cfg.HasDashScopeKey())

	// Test HasBailianKey
	assert.True(t, cfg.HasBailianKey())

	// Test with only Bailian key
	cfg.DashScopeAPIKey = ""
	assert.Equal(t, "test-bailian", cfg.GetAPIKey())
	assert.False(t, cfg.HasDashScopeKey())
	assert.True(t, cfg.HasBailianKey())
}
