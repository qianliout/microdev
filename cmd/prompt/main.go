package main

import (
	"os"

	"github.com/spf13/cobra"

	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/input"
	"microdev/pkg/llm"
	"microdev/pkg/logger"
	"microdev/pkg/output"
)

func main() {
	log := logger.NewLogger()

	// 使用 Cobra 构建 prompt 命令
	var rootCmd = &cobra.Command{
		Use:   "prompt <input>",
		Short: "优化并生成更好的提示词",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputContent := args[0]

			// 加载配置
			cfg, err := config.LoadConfig()

			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			// 处理输入
			inputProcessor := input.NewProcessor(log)
			content, err := inputProcessor.Process(inputContent)
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			// 初始化LLM客户端
			llmClient, err := llm.NewClient(cfg, log)
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			// 初始化输出处理器（输出到控制台）
			outputProcessor := output.NewProcessor("", log)
			defer func() { _ = outputProcessor.Close() }()

			// 生成prompt并处理输出
			if err := llmClient.GeneratePrompt(content, outputProcessor); err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}
			return nil
		},
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
