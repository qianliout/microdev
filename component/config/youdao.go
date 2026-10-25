package config

import (
	"os"
	"path/filepath"
	"strings"

	"microdev/infra/consts"
)

// YoudaoConfig 不依赖大模型密钥，与 Config 分开加载。
type YoudaoConfig struct {
	// 默认同步目录，已展开开头的 ~。
	Dir string
	// 官方 youdaonote CLI 的可执行文件名或绝对路径。
	Bin string
}

// LoadYoudaoConfig 从环境变量加载有道同步配置，空值回退到默认值。
func LoadYoudaoConfig() *YoudaoConfig {
	return &YoudaoConfig{
		Dir: expandHome(getTrimmedFromEnv(consts.EnvYoudaoDir, consts.DefaultYoudaoDir)),
		Bin: getTrimmedFromEnv(consts.EnvYoudaoBin, consts.DefaultYoudaoBin),
	}
}

func getTrimmedFromEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

// expandHome 展开开头的 ~，取不到用户主目录时去掉 ~ 按相对路径处理。
func expandHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return strings.TrimPrefix(path, "~/")
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
