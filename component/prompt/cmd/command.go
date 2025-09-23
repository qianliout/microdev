package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"microdev/component/prompt/model"
	"microdev/component/prompt/service"
	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/input"
	"microdev/pkg/logger"
	"microdev/pkg/output"
)

// NewPromptCommand 创建 prompt 子命令
func NewPromptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prompt <input>",
		Aliases: []string{"p"},
		Short:   "优化并生成更好的提示词",
		Long: `优化提示词工具，帮助您将原始提示词优化成更清晰、更有效的版本。

支持的输入方式：
- 直接文本：micro p "请帮我写一个函数"
- 文件路径：micro p ./prompt.txt
- 标准输入：echo "提示词" | micro p -

示例：
  micro p "请帮我写一个排序算法"
  micro p ./my-prompt.txt
  micro p - < prompt.txt`,
		Args: cobra.ExactArgs(1),
		RunE: runPromptCommand,
	}

	return cmd
}

// runPromptCommand 执行 prompt 命令的逻辑
func runPromptCommand(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	inputText := args[0]

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	// 处理输入
	inputProcessor := input.NewProcessor(log)
	content, err := inputProcessor.Process(inputText)
	if err != nil {
		log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	// 初始化优化服务
	optimizer, err := service.NewOptimizerService(cfg, log)
	if err != nil {
		log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	// 初始化输出处理器（输出到控制台）
	outputProcessor := output.NewProcessor("", log)
	defer outputProcessor.Close()

	// 创建优化请求
	request := &model.OptimizationRequest{
		OriginalPrompt: content,
		Language:       "中文", // 默认中文
	}

	// 执行优化
	if err := optimizer.Optimize(request, outputProcessor); err != nil {
		log.Error().Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	return nil
}
