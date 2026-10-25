package api

import (
	"fmt"

	"github.com/spf13/cobra"

	"microdev/component/file/service"
	"microdev/infra/console"
	filemodel "microdev/infra/model/file"
)

func NewFileCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "file <path>",
		Aliases: []string{"f"},
		Short:   "检测二进制文件信息",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			analyzer := service.NewAnalyzerSrv()
			res, err := analyzer.Analyze(cmd.Context(), &filemodel.AnalyzeParam{Path: args[0]})
			if err != nil {
				return err
			}

			out := console.New()
			out.Info(fmt.Sprintf("File Path: %s", res.Path))
			out.Info(fmt.Sprintf("File Format: %s", res.Format))
			out.Info(fmt.Sprintf("Executable: %s", res.IsExecutable))
			out.Info(fmt.Sprintf("Has Exec Perms: %s", res.HasExecPerms))
			out.Info(fmt.Sprintf("Has Dynamic Libs: %s", res.HasDynamicLibs))
			out.Info(fmt.Sprintf("GOOS: %s", res.GOOS))
			out.Info(fmt.Sprintf("GOARCH: %s", res.GOARCH))
			out.Info(fmt.Sprintf("Is Go Binary: %s", res.IsGoBinary))
			out.Info(fmt.Sprintf("Go Version: %s", res.GoVersion))
			out.Info(fmt.Sprintf("CGO Enabled: %s", res.CGOEnabled))
			return nil
		},
	}
}
