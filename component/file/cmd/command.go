package cmd

import (
	"github.com/spf13/cobra"

	"microdev/component/file/service"
	"microdev/pkg/logger"
)

func NewFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "file <path>",
		Aliases: []string{"f"},
		Short:   "检测二进制文件信息",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			path := args[0]

			analyzer := service.NewAnalyzerService()
			res, err := analyzer.Analyze(path)
			if err != nil {
				log.Err(err).Str("path", path).Msg("parse file failed")
				return err
			}
			analyzer.Output(&res)
			return nil
		},
	}

	return cmd
}
