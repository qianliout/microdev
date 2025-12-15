package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"microdev/component/translate/model"
	"microdev/component/translate/service"
	"microdev/pkg/config"
	"microdev/pkg/llm"
	"microdev/pkg/logger"
)

// NewTranslateCommand 创建 translate 子命令
func NewTranslateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "translate [input]",
		Aliases: []string{"t"},
		Short:   "中英互译，支持流式输出",
		Long: `中英互译：
- 若输入英文，则译为中文
- 若输入中文，则译为英文 `,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()

			// 加载配置
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Err(err).Msg(err.Error())
				return err
			}

			// 创建 LLM 客户端
			cli, err := llm.NewClient(cfg)
			if err != nil {
				log.Err(err).Msg(err.Error())
				return err
			}

			// 创建服务
			svc := service.NewTranslatorService(cli, cfg)

			// 非交互模式，需要输入参数或stdin标记
			if len(args) == 0 {
				return fmt.Errorf("require input; use '-' for stdin")
			}
			inArg := args[0]

			// 构建输入 ReadCloser
			inputRC, err := svc.Input(inArg)
			if err != nil {
				log.Err(err).Msg(err.Error())
				return err
			}

			// 构建输出 ReadWriteCloser（管道，支持流式）
			pipe := model.NewRWPipe()

			// 构造请求
			req := &model.Translate{
				Input:  inputRC,
				Direct: "",
				Output: pipe,
			}

			// 并发执行：翻译写管道 + 控制台读管道
			done := make(chan error, 1)
			go func() {
				// 翻译写入管道
				err := svc.Translate(req)
				// 写端关闭，触发读端EOF
				_ = pipe.CloseWrite()
				done <- err
			}()

			// 控制台输出（流式滚动）
			if err := svc.Output(req); err != nil {
				log.Err(err).Msg(err.Error())
				return err
			}

			// 等待翻译完成并检查错误
			if err := <-done; err != nil {
				log.Err(err).Msg("translate failed")
				return err
			}

			return nil
		},
	}

	return cmd
}
