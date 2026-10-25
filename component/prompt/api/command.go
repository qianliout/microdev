package api

import (
	"fmt"

	"github.com/spf13/cobra"

	"microdev/component/config"
	"microdev/component/prompt/service"
	"microdev/component/prompt/store"
	"microdev/infra/consts"
	"microdev/infra/llm"
	promptmodel "microdev/infra/model/prompt"
	"microdev/infra/rdb"
	"microdev/infra/utils"
	"microdev/migrate"
)

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

func runPromptCommand(cmd *cobra.Command, args []string, interactive bool) error {
	if !interactive && len(args) == 0 {
		return fmt.Errorf("require input in non-interactive mode; use -i/--interactive")
	}
	ctx := cmd.Context()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	llmCli, err := llm.NewClient(cfg)
	if err != nil {
		return err
	}
	db, err := rdb.Open(cfg.SessionDriver, cfg.SessionDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.ExecAll(ctx, migrate.Statements(db.Driver())); err != nil {
		return fmt.Errorf("init session tables: %w", err)
	}

	sessionSrv := service.NewSessionSrv(store.NewSessionDao(db), store.NewMessageDao(db))
	optimizer := service.NewOptimizerSrv(sessionSrv, llmCli)
	if interactive {
		runner := newInteractiveRunner(cmd, optimizer, sessionSrv)
		return runner.run(ctx)
	}

	text, err := utils.ReadInput(args[0])
	if err != nil {
		return err
	}
	param := &promptmodel.OptimizeParam{
		OriginalPrompt: text,
		Language:       consts.DefaultPromptLanguage,
		Output:         cmd.OutOrStdout(),
	}
	_, err = optimizer.Optimize(ctx, param)
	return err
}
