package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"microdev/component/file/service"
	"microdev/pkg/errors"
	"microdev/pkg/logger"
	"microdev/pkg/output"
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

			if _, err := os.Stat(path); err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				log.UserError(fmt.Sprintf("错误: %s", errors.GetUserFriendlyMessage(err)))
				return err
			}

			analyzer := service.NewAnalyzerService(log)
			res, err := analyzer.Analyze(path)
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				log.UserError(fmt.Sprintf("错误: %s", errors.GetUserFriendlyMessage(err)))
				return err
			}

			out := output.NewProcessor("", log)
			defer out.Close()

			fmt.Fprintf(out, "文件: %s\n", res.Path)
			fmt.Fprintf(out, "二进制格式: %s\n", res.Format)
			fmt.Fprintf(out, "可执行格式: %t\n", res.IsExecutable)
			fmt.Fprintf(out, "执行权限: %t\n", res.HasExecPerms)
			fmt.Fprintf(out, "动态链接库: %t\n", res.HasDynamicLibs)
			if res.GOOS != "" || res.GOARCH != "" {
				fmt.Fprintf(out, "平台: %s/%s\n", res.GOOS, res.GOARCH)
			} else {
				fmt.Fprintf(out, "平台: 未知\n")
			}
			if res.IsGoBinary {
				fmt.Fprintf(out, "Go构建: true\n")
				fmt.Fprintf(out, "Go版本: %s\n", res.GoVersion)
				if res.CGOEnabled != "" {
					fmt.Fprintf(out, "CGO_ENABLED: %s\n", res.CGOEnabled)
				}
			} else {
				fmt.Fprintf(out, "Go构建: false\n")
			}

			return nil
		},
	}

	return cmd
}
