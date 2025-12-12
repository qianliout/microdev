package main

import (
	"os"

	"github.com/spf13/cobra"

	fileCmd "microdev/component/file/cmd"
	promptCmd "microdev/component/prompt/cmd"
	translateCmd "microdev/component/translate/cmd"
)

func main() {
	// 创建根命令
	var rootCmd = &cobra.Command{
		Use:   "micro",
		Short: "微开发工具集 - 包含提示词优化和翻译功能",
		Long: `micro 是一个微开发工具集，提供以下功能：
- prompt (p): 优化并生成更好的提示词
- translate (t): 中英互译，支持文件、目录或直接文本`,
	}

	// 添加子命令
	rootCmd.AddCommand(promptCmd.NewPromptCommand())
	rootCmd.AddCommand(translateCmd.NewTranslateCommand())
	rootCmd.AddCommand(fileCmd.NewFileCommand())

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
