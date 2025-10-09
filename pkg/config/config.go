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
		ModelName:    getModelNameFromEnv(),
		MaxTokens:    getMaxTokensFromEnv(),
		Temperature:  getTemperatureFromEnv(),
		StreamOutput: getStreamOutputFromEnv(),
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

// GetSupportedModels 获取支持的模型列表
func GetSupportedModels() []string {
	return []string{
		"qwen3-max", "qwen-plus", "qwen-plus-latest",
		"qwen-turbo", "qwen-turbo-latest",
		"qwen-max", "qwen-max-latest",
		"qwen2.5-72b-instruct", "qwen2.5-32b-instruct",
		"qwen2.5-14b-instruct", "qwen2.5-7b-instruct",
		"qwen-flash",
	}
}

// IsModelSupported 检查模型是否被支持
func IsModelSupported(modelName string) bool {
	supportedModels := GetSupportedModels()
	for _, supported := range supportedModels {
		if modelName == supported {
			return true
		}
	}
	return false
}

// getModelNameFromEnv 从环境变量获取模型名称配置
func getModelNameFromEnv() string {
	modelName := os.Getenv("MICRO_MODEL_NAME")
	if modelName == "" {
		return "qwen3-max" // 默认模型
	}

	// 验证模型名称是否在支持列表中
	if IsModelSupported(modelName) {
		return modelName
	}

	// 如果不在支持列表中，返回默认值
	return "qwen3-max"
}

// getMaxTokensFromEnv 从环境变量获取最大Token数配置
func getMaxTokensFromEnv() int {
	maxTokensStr := os.Getenv("MICRO_MAX_TOKENS")
	if maxTokensStr == "" {
		return 2048 // 默认最大Token数
	}

	maxTokens, err := strconv.Atoi(maxTokensStr)
	if err != nil || maxTokens < 1 {
		return 2048 // 解析失败或无效值时使用默认值
	}

	// 限制最大Token数在合理范围内
	if maxTokens > 8192 {
		return 8192
	}

	return maxTokens
}

// getTemperatureFromEnv 从环境变量获取温度配置
func getTemperatureFromEnv() float64 {
	temperatureStr := os.Getenv("MICRO_TEMPERATURE")
	if temperatureStr == "" {
		return 0.7 // 默认温度
	}

	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil || temperature < 0 || temperature > 2 {
		return 0.7 // 解析失败或无效值时使用默认值
	}

	return temperature
}

// getStreamOutputFromEnv 从环境变量获取流式输出配置
func getStreamOutputFromEnv() bool {
	streamOutputStr := os.Getenv("MICRO_STREAM_OUTPUT")
	if streamOutputStr == "" {
		return true // 默认启用流式输出
	}

	streamOutput, err := strconv.ParseBool(streamOutputStr)
	if err != nil {
		return true // 解析失败时使用默认值
	}

	return streamOutput
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
