package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/input"
	"microdev/pkg/logger"
	"microdev/pkg/output"
	"microdev/pkg/translator"
)

var (
	flagDir  bool
	flagFile bool
	flagType string
)

func main() {
	log := logger.NewLogger()

	// 定义 translate 命令及参数
	var rootCmd = &cobra.Command{
		Use:   "translate <target>",
		Short: "中英互译，支持文件、目录或直接文本",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			// 同时传 -d 与 -f 时给出警告
			if flagDir && flagFile {
				log.Warn().Msg("警告：-d 和 -f 同时为 true，仅会根据上下文选择其一")
			}

			// 加载配置
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			// 处理输入
			inputProcessor := input.NewProcessor(log)
			content, err := inputProcessor.Process(target)
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			// 初始化翻译器
			trans, err := translator.NewTranslator(cfg, log)
			if err != nil {
				log.Err(err).Msg(errors.GetUserFriendlyMessage(err))
				return err
			}

			// 当指定目录翻译时，仅在 -d=true 下启用类型过滤
			if flagDir {
				if stat, err := os.Stat(target); err == nil && stat.IsDir() {
					return processDirectoryWithFilter(target, flagType, trans, log)
				}
			}

			// 文件模式：明确 -f 或目标是文件
			if flagFile {
				if stat, err := os.Stat(target); err == nil && !stat.IsDir() {
					return processSingleFile(target, content, trans, log)
				}
			}

			// 自动判断：目录/文件/直接文本
			if stat, err := os.Stat(target); err == nil {
				if stat.IsDir() {
					return processDirectoryWithFilter(target, flagType, trans, log)
				}
				return processSingleFile(target, content, trans, log)
			}

			// 直接文本
			if err := processDirectText(content, trans, log); err != nil {
				log.Error().Msg(errors.GetUserFriendlyMessage(err))
				fmt.Fprintf(os.Stderr, "\n错误: %s\n", errors.GetUserFriendlyMessage(err))
				return err
			}
			fmt.Println()
			return nil
		},
	}

	// 绑定 flags
	rootCmd.Flags().BoolVarP(&flagDir, "dir", "d", false, "翻译目录。无值等同于 true")
	rootCmd.Flags().BoolVarP(&flagFile, "file", "f", false, "翻译单个文件。无值等同于 true")
	rootCmd.Flags().StringVarP(&flagType, "type", "t", "md", "当 -d=true 时生效的文件类型过滤，如 md")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// generateOutputPath 根据输入文件路径和内容生成输出文件路径
func generateOutputPath(inputPath string, content string) string {
	// 检测内容主要语言
	direction := detectLanguage(content)

	// 获取文件扩展名和基础名
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(inputPath, ext)

	// 根据翻译方向添加后缀
	var suffix string
	if direction == "zh2en" {
		suffix = "_en"
	} else {
		suffix = "_zh"
	}

	return baseName + suffix + ext
}

// detectLanguage 检测主要语言
func detectLanguage(text string) string {
	chineseCount := 0
	englishCount := 0
	totalCount := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalCount++
			if isChinese(r) {
				chineseCount++
			} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				englishCount++
			}
		}
	}

	if totalCount == 0 {
		return "zh2en" // 默认中译英
	}

	if float64(chineseCount)/float64(totalCount) > 0.3 {
		return "zh2en" // 中译英
	}
	return "en2zh" // 英译中
}

// isChinese 判断是否为中文字符
func isChinese(r rune) bool {
	return r >= 0x4e00 && r <= 0x9fff
}

// processDirectory 处理目录，遍历所有Markdown文件
// 在目录模式下，支持通过 -t 过滤文件类型（默认 md）
func processDirectoryWithFilter(dirPath string, typ string, translator *translator.Translator, log *logger.Logger) error {
	if typ == "" {
		typ = "md"
	}
	log.Info().Str("dir", dirPath).Str("type", typ).Msg("开始处理目录")

	// 遍历目录下的所有文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理匹配类型的文件
		if !isTypeFile(path, typ) {
			return nil
		}

		log.Info().Str("path", path).Msg("发现文件")

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("读取文件失败")
			return nil // 继续处理其他文件
		}

		contentStr := string(content)

		// 检测语言，只处理英文文件
		if detectLanguage(contentStr) != "en2zh" {
			log.Info().Str("path", path).Msg("跳过非英文文件")
			return nil
		}

		log.Info().Str("path", path).Msg("开始翻译英文文件")

		// 生成输出文件路径
		outputPath := generateOutputPath(path, contentStr)

		// 处理翻译
		err = processSingleFileWithPath(path, contentStr, outputPath, translator, log)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("翻译文件失败")
			return nil // 继续处理其他文件
		}

		return nil
	})

	if err != nil {
		return errors.NewFileError("遍历目录失败", err)
	}

	log.Info().Msg("目录处理完成")
	return nil
}

// processSingleFile 处理单个文件
func processSingleFile(filePath string, content string, translator *translator.Translator, log *logger.Logger) error {
	outputPath := generateOutputPath(filePath, content)
	log.Info().Str("output", outputPath).Msg("检测到文件输入")

	return processSingleFileWithPath(filePath, content, outputPath, translator, log)
}

// processSingleFileWithPath 使用指定输出路径处理单个文件
func processSingleFileWithPath(filePath string, content string, outputPath string, translator *translator.Translator, log *logger.Logger) error {
	// 初始化输出处理器
	outputProcessor := output.NewProcessor(outputPath, log)
	defer outputProcessor.Close()

	// 执行翻译
	return translator.TranslateWithPath(content, filePath, outputProcessor)
}

// processDirectText 处理直接文本输入
func processDirectText(content string, translator *translator.Translator, log *logger.Logger) error {
	// 初始化输出处理器（输出到控制台）
	outputProcessor := output.NewProcessor("", log)
	defer outputProcessor.Close()

	// 执行翻译
	return translator.TranslateWithPath(content, "", outputProcessor)
}

// isMarkdownFile 判断是否为Markdown文件
func isTypeFile(filename string, typ string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	return ext == strings.ToLower(typ)
}
