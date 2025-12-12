package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"microdev/component/prompt/model"
	"microdev/component/prompt/service"
	"microdev/pkg/config"
	"microdev/pkg/input"
	"microdev/pkg/logger"
	"microdev/pkg/output"
)

// NewPromptCommand 创建 prompt 子命令
func NewPromptCommand() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:     "prompt [input]",
		Aliases: []string{"p"},
		Short:   "优化并生成更好的提示词",
		Long: `优化提示词工具，帮助您将原始提示词优化成更清晰、更有效的版本。

支持的输入方式：
- 直接文本：micro p "请帮我写一个函数"
- 文件路径：micro p ./prompt.txt
- 标准输入：echo "提示词" | micro p -
- 交互模式：micro p -i 或 micro p --interactive

交互模式特性：
- 保持对话历史和上下文
- 智能记忆压缩
- 输入 'exit'、'quit' 或 '退出' 结束会话

示例：
  micro p "请帮我写一个排序算法"
  micro p ./my-prompt.txt
  micro p - < prompt.txt
  micro p -i  # 进入交互模式`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPromptCommand(cmd, args, interactive)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "启用交互式会话模式")

	return cmd
}

// runPromptCommand 执行 prompt 命令的逻辑
func runPromptCommand(cmd *cobra.Command, args []string, interactive bool) error {
	log := logger.NewLogger()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Err(err).Msg(err.Error())
		log.Error().Msg(err.Error())
		return err
	}

	// 初始化优化服务
	optimizer, err := service.NewOptimizerService(cfg, log)
	if err != nil {
		log.Err(err).Msg(err.Error())
		log.Error().Msg(err.Error())
		return err
	}

	// 如果是交互模式
	if interactive {
		return runInteractiveMode(optimizer, log)
	}

	// 非交互模式，需要输入参数
	if len(args) == 0 {
		return fmt.Errorf("require input in non-interactive mode; use -i/--interactive")
	}

	inputText := args[0]

	// 处理输入
	inputProcessor := input.NewProcessor(log)
	content, err := inputProcessor.Process(inputText)
	if err != nil {
		log.Err(err).Msg(err.Error())
		log.Error().Msg(err.Error())
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
		log.Error().Msg(err.Error())
		return err
	}

	return nil
}

// runInteractiveMode 运行交互式模式
func runInteractiveMode(optimizer *service.OptimizerService, log *logger.Logger) error {
	// 开始会话
	optimizer.StartSession()
	defer optimizer.EndSession()

	// 显示欢迎信息
	log.Info().Msg("🤖 欢迎使用提示词优化交互模式！")
	log.Info().Msg("💡 特性：保持对话历史、智能记忆压缩、上下文感知优化")
	log.Info().Msg("📝 输入您的提示词，我将为您优化。输入 'exit'、'quit' 或 '退出' 结束会话。")
	log.Info().Msg(strings.Repeat("-", 60))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// 显示输入提示
		log.Info().Msg("\n💬 请输入提示词: ")

		// 读取用户输入
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		// 检查退出命令
		if isExitCommand(userInput) {
			log.Info().Msg("\n👋 感谢使用！会话已结束。")
			break
		}

		// 跳过空输入
		if userInput == "" {
			log.Warn().Msg("⚠️  输入不能为空，请重新输入。")
			continue
		}

		// 显示处理状态
		log.Info().Msg("\n🔄 正在优化提示词...")

		// 初始化输出处理器
		outputProcessor := output.NewProcessor("", log)

		// 创建优化请求
		request := &model.OptimizationRequest{
			OriginalPrompt: userInput,
			Language:       "中文",
		}

		// 执行会话模式优化
		if err := optimizer.OptimizeWithSession(request, outputProcessor); err != nil {
			log.Error().Err(err).Msg("优化失败")
			log.Error().Msg(fmt.Sprintf("❌ 优化失败: %s", err.Error()))
			continue
		}

		outputProcessor.Close()

		// 显示会话统计
		stats := optimizer.GetSessionStats()
		if stats["session_active"].(bool) {
			statsMsg := fmt.Sprintf("\n📊 会话统计: %d条消息", stats["total_messages"])
			if stats["has_summary"].(bool) {
				statsMsg += " (已压缩记忆)"
			}
			log.Info().Msg(statsMsg)
		}

		log.Info().Msg(strings.Repeat("-", 60))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	return nil
}

// isExitCommand 检查是否为退出命令
func isExitCommand(input string) bool {
	exitCommands := []string{"exit", "quit", "退出", "q", "bye", "再见"}
	lowerInput := strings.ToLower(strings.TrimSpace(input))

	for _, cmd := range exitCommands {
		if lowerInput == cmd {
			return true
		}
	}

	return false
}
