package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	fileCmd "microdev/component/file/api"
	promptCmd "microdev/component/prompt/api"
	translateCmd "microdev/component/translate/api"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "micro",
		Short: "微开发工具集 - 包含提示词优化和翻译功能",
		Long: `micro 是一个微开发工具集，提供以下功能：
- prompt (p): 优化并生成更好的提示词
- translate (t): 中英互译，支持文件、目录或直接文本
- file (f): 检测二进制文件信息`,
		SilenceUsage: true,
	}
	rootCmd.AddCommand(promptCmd.NewPromptCommand())
	rootCmd.AddCommand(translateCmd.NewTranslateCommand())
	rootCmd.AddCommand(fileCmd.NewFileCommand())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		stop()
		os.Exit(1)
	}
}
