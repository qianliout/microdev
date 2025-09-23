package config

import (
	"os"

	"microdev/pkg/errors"
)

// Config 包含应用程序配置
type Config struct {
	// API配置
	DashScopeAPIKey  string
	AliBailianAPIKey string

	// 模型配置
	ModelName   string
	MaxTokens   int
	Temperature float64

	// 其他配置
	StreamOutput bool
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// 默认配置
		ModelName:    "qwen-plus-latest",
		MaxTokens:    2048,
		Temperature:  0.7,
		StreamOutput: true,
	}

	// 从环境变量读取API密钥
	cfg.DashScopeAPIKey = os.Getenv("DASHSCOPE_API_KEY")
	cfg.AliBailianAPIKey = os.Getenv("ALI_BAILIAN_API_KEY")

	// 验证至少有一个API密钥
	if cfg.DashScopeAPIKey == "" && cfg.AliBailianAPIKey == "" {
		return nil, errors.NewConfigError("至少需要设置 DASHSCOPE_API_KEY 或 ALI_BAILIAN_API_KEY 环境变量", nil)
	}

	return cfg, nil
}

// GetAPIKey 获取可用的API密钥
func (c *Config) GetAPIKey() string {
	if c.DashScopeAPIKey != "" {
		return c.DashScopeAPIKey
	}
	return c.AliBailianAPIKey
}

// HasDashScopeKey 检查是否有DashScope API密钥
func (c *Config) HasDashScopeKey() bool {
	return c.DashScopeAPIKey != ""
}

// HasBailianKey 检查是否有百炼API密钥
func (c *Config) HasBailianKey() bool {
	return c.AliBailianAPIKey != ""
}
