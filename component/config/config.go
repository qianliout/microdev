package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"

	"microdev/infra/consts"
)

type Config struct {
	DashScopeAPIKey  string
	AliBailianAPIKey string

	ModelName   string
	MaxTokens   int64
	Temperature float64

	StreamOutput bool

	// 并发调用大模型的数量。
	Concurrency int64

	// 会话存储使用的数据库驱动与连接串。
	SessionDriver string
	SessionDSN    string
}

// LoadConfig 从环境变量加载配置，非法值回退到默认值。
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DashScopeAPIKey:  os.Getenv(consts.EnvDashScopeAPIKey),
		AliBailianAPIKey: os.Getenv(consts.EnvAliBailianAPIKey),
		ModelName:        getModelNameFromEnv(),
		MaxTokens:        getInt64FromEnv(consts.EnvMaxTokens, consts.DefaultMaxTokens, consts.MaxMaxTokens),
		Temperature:      getTemperatureFromEnv(),
		StreamOutput:     getStreamOutputFromEnv(),
		Concurrency:      getInt64FromEnv(consts.EnvConcurrency, consts.DefaultConcurrency, consts.MaxConcurrency),
		SessionDriver:    getStringFromEnv(consts.EnvSessionDriver, consts.DefaultSessionDriver),
		SessionDSN:       getStringFromEnv(consts.EnvSessionDSN, consts.DefaultSessionDSN),
	}

	hasKey := cfg.DashScopeAPIKey != "" || cfg.AliBailianAPIKey != ""
	if !hasKey {
		return nil, fmt.Errorf("require env %s or %s", consts.EnvDashScopeAPIKey, consts.EnvAliBailianAPIKey)
	}
	return cfg, nil
}

// GetAPIKey 优先返回 DashScope 密钥，其次是百炼密钥。
func (c *Config) GetAPIKey() string {
	if c.DashScopeAPIKey != "" {
		return c.DashScopeAPIKey
	}
	return c.AliBailianAPIKey
}

func (c *Config) HasDashScopeKey() bool {
	return c.DashScopeAPIKey != ""
}

func (c *Config) HasBailianKey() bool {
	return c.AliBailianAPIKey != ""
}

func GetSupportedModels() []string {
	return slices.Clone(consts.SupportedModels)
}

func IsModelSupported(modelName string) bool {
	return slices.Contains(consts.SupportedModels, modelName)
}

func getStringFromEnv(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func getModelNameFromEnv() string {
	modelName := os.Getenv(consts.EnvModelName)
	if !IsModelSupported(modelName) {
		return consts.DefaultModelName
	}
	return modelName
}

// getInt64FromEnv 读取正整数配置，解析失败或小于 1 时返回默认值，超过上限时按上限截断。
func getInt64FromEnv(name string, fallback, upper int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || value < 1 {
		return fallback
	}
	return min(value, upper)
}

func getTemperatureFromEnv() float64 {
	temperature, err := strconv.ParseFloat(os.Getenv(consts.EnvTemperature), 64)
	outOfRange := temperature < 0 || temperature > consts.MaxTemperature
	if err != nil || outOfRange {
		return consts.DefaultTemperature
	}
	return temperature
}

func getStreamOutputFromEnv() bool {
	streamOutput, err := strconv.ParseBool(os.Getenv(consts.EnvStreamOutput))
	if err != nil {
		return true
	}
	return streamOutput
}
