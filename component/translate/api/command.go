package api

import (
	"github.com/spf13/cobra"

	"microdev/component/config"
	"microdev/component/translate/service"
	"microdev/infra/llm"
	translatemodel "microdev/infra/model/translate"
	"microdev/infra/utils"
)

func NewTranslateCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "translate <input>",
		Aliases: []string{"t"},
		Short:   "中英互译，支持流式输出",
		Long: `中英互译：
- 若输入英文，则译为中文
- 若输入中文，则译为英文
- 输入可以是直接文本、文件路径，或用 '-' 读取标准输入`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := utils.ReadInput(args[0])
			if err != nil {
				return err
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			llmCli, err := llm.NewClient(cfg)
			if err != nil {
				return err
			}

			translator := service.NewTranslatorSrv(llmCli)
			param := &translatemodel.TranslateParam{Text: text, Output: cmd.OutOrStdout()}
			return translator.Translate(cmd.Context(), param)
		},
	}
}
