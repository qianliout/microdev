package config

import (
	"os"
	"strconv"

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

	// 并发配置
	Concurrency int // 并发调用大模型的数量
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// 默认配置
		ModelName:    "qwen3-max",
		MaxTokens:    2048,
		Temperature:  0.7,
		StreamOutput: true,
		Concurrency:  getConcurrencyFromEnv(),
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

// getConcurrencyFromEnv 从环境变量获取并发数配置
func getConcurrencyFromEnv() int {
	concurrencyStr := os.Getenv("MICRO_CONCURRENCY")
	if concurrencyStr == "" {
		return 3 // 默认并发数为3
	}

	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		return 3 // 解析失败或无效值时使用默认值
	}

	// 限制最大并发数为10，避免过度并发
	if concurrency > 10 {
		return 10
	}

	return concurrency
}
