package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"microdev/component/translate/model"
	"microdev/component/translate/service"
	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/input"
	"microdev/pkg/logger"
)

// NewTranslateCommand 创建 translate 子命令
func NewTranslateCommand() *cobra.Command {
	var (
		flagDir         bool
		flagFile        bool
		flagType        string
		flagConcurrency int
		flagForce       bool
	)

	cmd := &cobra.Command{
		Use:     "translate <input>",
		Aliases: []string{"t"},
		Short:   "中英互译，支持文件、目录或直接文本",
		Long: `智能翻译工具，支持中英文互译，自动检测语言方向。

支持的输入方式：
- 直接文本：micro t "Hello World"
- 文件路径：micro t ./document.md
- 目录路径：micro t ./docs/ -d
- 标准输入：echo "Hello" | micro t -

功能特性：
- 自动语言检测（中文↔英文）
- Markdown格式智能处理
- 并发翻译提升效率
- 翻译幂等性（避免重复翻译）
- 智能输出路径生成

示例：
  micro t "你好世界"                    # 直接翻译文本
  micro t ./README.md                  # 翻译单个文件
  micro t ./docs/ -d                   # 翻译目录下所有md文件
  micro t ./docs/ -d -t txt            # 翻译目录下所有txt文件
  micro t ./file.md -f                 # 强制重新翻译
  micro t ./docs/ -d -c 5              # 使用5个并发协程`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTranslateCommand(args[0], flagDir, flagFile, flagType, flagConcurrency, flagForce)
		},
	}

	// 添加命令行标志
	cmd.Flags().BoolVarP(&flagDir, "directory", "d", false, "处理目录模式")
	cmd.Flags().BoolVarP(&flagFile, "file", "f", false, "处理文件模式")
	cmd.Flags().StringVarP(&flagType, "type", "t", "md", "文件类型过滤（仅在目录模式下有效）")
	cmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 0, "并发数（0表示使用配置默认值）")
	cmd.Flags().BoolVar(&flagForce, "force", false, "强制重新翻译（忽略已存在的翻译文件）")

	return cmd
}

// runTranslateCommand 执行翻译命令的逻辑
func runTranslateCommand(target string, flagDir, flagFile bool, flagType string, flagConcurrency int, flagForce bool) error {
	log := logger.NewLogger()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	// 如果指定了并发数，覆盖配置
	if flagConcurrency > 0 {
		cfg.Concurrency = flagConcurrency
	}

	// 处理输入
	inputProcessor := input.NewProcessor(log)
	content, err := inputProcessor.Process(target)
	if err != nil {
		log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	// 初始化翻译服务
	translator, err := service.NewTranslatorService(cfg, log)
	if err != nil {
		log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	// 初始化处理服务
	processor := service.NewProcessorService(translator, log)

	// 根据标志和输入类型决定处理方式
	return processInput(target, content, processor, log, &model.ProcessingOptions{
		Directory:   flagDir,
		File:        flagFile,
		FileType:    flagType,
		Force:       flagForce,
		Concurrency: flagConcurrency,
	})
}

// processInput 处理输入
func processInput(target, content string, processor *service.ProcessorService, log *logger.Logger, options *model.ProcessingOptions) error {
	// 当指定目录翻译时，仅在 -d=true 下启用类型过滤
	if options.Directory {
		if stat, err := os.Stat(target); err == nil && stat.IsDir() {
			return processor.ProcessDirectory(target, options)
		}
	}

	// 文件模式：明确 -f 或目标是文件
	if options.File {
		if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
			return processor.ProcessSingleFileWithForce(target, content, options.Force)
		}
	}

	// 自动判断：目录/文件/直接文本
	if stat, err := os.Stat(target); err == nil {
		if stat.IsDir() {
			return processor.ProcessDirectory(target, options)
		}
		return processor.ProcessSingleFileWithForce(target, content, options.Force)
	}

	// 直接文本
	if err := processor.ProcessDirectText(content); err != nil {
		log.Error().Msg(errors.GetUserFriendlyMessage(err))
		fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
		return err
	}

	return nil
}
