// Package api 提供有道云笔记同步的 cobra 命令。
package api

import (
	"fmt"

	"github.com/spf13/cobra"

	"microdev/component/config"
	"microdev/component/youdao/service"
	"microdev/component/youdao/store"
	"microdev/infra/console"
	"microdev/infra/consts"
	youdaomodel "microdev/infra/model/youdao"
	"microdev/infra/youdao"
)

func NewYoudaoCommand() *cobra.Command {
	var (
		folder string
		title  string
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:     "youdao [path]",
		Aliases: []string{"y"},
		Short:   "把本地 Markdown 同步到有道云笔记",
		Long: `把本地 Markdown 单向同步到有道云笔记，依赖官方 youdaonote CLI：
- 不带参数时同步默认目录（环境变量 ` + consts.EnvYoudaoDir + `，默认 ` + consts.DefaultYoudaoDir + `）
- path 可以是目录或单个 .md 文件
- 首次同步创建笔记，之后内容变化才更新，映射记录在同步目录的 ` + consts.YoudaoMappingFile,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.LoadYoudaoConfig()
			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			syncer := service.NewSyncerSrv(store.NewMappingStore(), youdao.NewClient(cfg.Bin))
			param := &youdaomodel.SyncParam{Path: path, Dir: cfg.Dir, Folder: folder, Title: title, DryRun: dryRun}
			summary, err := syncer.Sync(cmd.Context(), param)
			if err != nil {
				return err
			}

			printSummary(summary, dryRun)
			if summary.Failed > 0 {
				return fmt.Errorf("%d files failed to sync", summary.Failed)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&folder, "folder", "", "首次创建笔记时放入的云端目录 ID")
	cmd.Flags().StringVarP(&title, "name", "n", "", "笔记标题，仅单文件同步时可用，默认取文件名")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只打印将要执行的操作，不调用云端也不写映射")
	return cmd
}

func printSummary(summary *youdaomodel.SyncSummary, dryRun bool) {
	out := console.New()
	if len(summary.Results) == 0 {
		out.Println("no markdown files to sync")
		return
	}
	for _, res := range summary.Results {
		out.Printf("[%s] %s\n", res.Action, res.Path)
		if res.Note != "" {
			out.Printf("  note: %s\n", res.Note)
		}
		if res.Err != nil {
			out.Printf("  error: %v\n", res.Err)
		}
	}
	prefix := ""
	if dryRun {
		prefix = "(dry-run) "
	}
	out.Info(fmt.Sprintf("%screated %d, updated %d, skipped %d, failed %d",
		prefix, summary.Created, summary.Updated, summary.Skipped, summary.Failed))
}
