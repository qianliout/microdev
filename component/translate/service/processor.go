package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"microdev/component/translate/model"
	"microdev/pkg/logger"
	"microdev/pkg/output"
	"microdev/pkg/utils"
)

// ProcessorService 文件处理服务
type ProcessorService struct {
	translator *TranslatorService
	logger     *logger.Logger
}

// NewProcessorService 创建新的文件处理服务
func NewProcessorService(translator *TranslatorService, logger *logger.Logger) *ProcessorService {
	return &ProcessorService{
		translator: translator,
		logger:     logger,
	}
}

// ProcessConcurrentTasks 并发处理翻译任务
func (p *ProcessorService) ProcessConcurrentTasks(tasks []model.TranslationTask) error {
	if len(tasks) == 0 {
		return nil
	}

	// 获取并发数配置
	concurrency := p.translator.GetConcurrency()
	if concurrency <= 0 {
		concurrency = 3 // 默认并发数
	}

	// 限制并发数不超过任务数
	if concurrency > len(tasks) {
		concurrency = len(tasks)
	}

	p.logger.Info().Int("concurrency", concurrency).Int("tasks", len(tasks)).Msg("开始并发翻译")

	// 创建任务通道和错误通道
	taskChan := make(chan model.TranslationTask, len(tasks))
	errChan := make(chan error, len(tasks))
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskChan {
				p.logger.Debug().Int("worker", workerID).Str("file", task.InputPath).Msg("处理翻译任务")
				if err := p.processTask(task); err != nil {
					errChan <- fmt.Errorf("工作协程 %d 处理文件 %s 失败: %w", workerID, task.InputPath, err)
				}
			}
		}(i + 1)
	}

	// 发送任务
	go func() {
		defer close(taskChan)
		for _, task := range tasks {
			taskChan <- task
		}
	}()

	// 等待所有任务完成
	wg.Wait()
	close(errChan)

	// 检查错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		// 返回第一个错误
		return errors[0]
	}

	p.logger.Info().Int("completed", len(tasks)).Msg("并发翻译完成")
	return nil
}

// processTask 处理单个翻译任务
func (p *ProcessorService) processTask(task model.TranslationTask) error {
	// 初始化输出处理器
	outputProcessor := output.NewProcessor(task.OutputPath, p.logger)
	defer outputProcessor.Close()

	// 创建翻译请求
	request := &model.TranslationRequest{
		Content:    task.Content,
		InputPath:  task.InputPath,
		OutputPath: task.OutputPath,
	}

	// 执行翻译
	return p.translator.Translate(request, outputProcessor)
}

// ProcessDirectory 处理目录，遍历所有指定类型的文件
func (p *ProcessorService) ProcessDirectory(dirPath string, options *model.ProcessingOptions) error {
	if options.FileType == "" {
		options.FileType = "md"
	}

	p.logger.Info().Str("dir", dirPath).Str("type", options.FileType).Msg("开始处理目录")

	var tasks []model.TranslationTask

	// 遍历目录
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名
		if !strings.HasSuffix(strings.ToLower(path), "."+options.FileType) {
			return nil
		}

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			p.logger.Warn().Str("file", path).Err(err).Msg("读取文件失败，跳过")
			return nil
		}

		contentStr := string(content)

		// 幂等性检查：如果翻译文件已存在且未强制重新翻译，则跳过
		if !options.Force && utils.IsTranslationExists(path, contentStr) {
			outputPath := utils.GetTranslationOutputPath(path, contentStr)
			p.logger.Info().Str("file", path).Str("output", outputPath).Msg("翻译文件已存在，跳过")
			return nil
		}

		// 生成输出路径
		outputPath := utils.GetTranslationOutputPath(path, contentStr)

		// 添加到任务列表
		tasks = append(tasks, model.TranslationTask{
			InputPath:  path,
			Content:    contentStr,
			OutputPath: outputPath,
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	if len(tasks) == 0 {
		p.logger.Info().Msg("没有找到需要翻译的文件")
		return nil
	}

	p.logger.Info().Int("count", len(tasks)).Msg("开始并发翻译文件")

	// 使用并发翻译
	return p.ProcessConcurrentTasks(tasks)
}

// ProcessSingleFile 处理单个文件
func (p *ProcessorService) ProcessSingleFile(filePath string, content string) error {
	outputPath := utils.GenerateOutputPath(filePath, content)
	p.logger.Info().Str("output", outputPath).Msg("检测到文件输入")

	return p.ProcessSingleFileWithPath(filePath, content, outputPath)
}

// ProcessSingleFileWithForce 处理单个文件，支持强制重新翻译
func (p *ProcessorService) ProcessSingleFileWithForce(filePath string, content string, force bool) error {
	// 幂等性检查：如果翻译文件已存在且未强制重新翻译，则跳过
	if !force && utils.IsTranslationExists(filePath, content) {
		outputPath := utils.GetTranslationOutputPath(filePath, content)
		p.logger.Info().Str("file", filePath).Str("output", outputPath).Msg("翻译文件已存在，跳过")
		return nil
	}

	outputPath := utils.GetTranslationOutputPath(filePath, content)
	p.logger.Info().Str("output", outputPath).Msg("检测到文件输入")

	return p.ProcessSingleFileWithPath(filePath, content, outputPath)
}

// ProcessSingleFileWithPath 使用指定输出路径处理单个文件
func (p *ProcessorService) ProcessSingleFileWithPath(filePath string, content string, outputPath string) error {
	// 初始化输出处理器
	outputProcessor := output.NewProcessor(outputPath, p.logger)
	defer outputProcessor.Close()

	// 创建翻译请求
	request := &model.TranslationRequest{
		Content:    content,
		InputPath:  filePath,
		OutputPath: outputPath,
	}

	// 执行翻译
	return p.translator.Translate(request, outputProcessor)
}

// ProcessDirectText 处理直接文本输入
func (p *ProcessorService) ProcessDirectText(content string) error {
	// 初始化输出处理器（输出到控制台）
	outputProcessor := output.NewProcessor("", p.logger)
	defer outputProcessor.Close()

	// 创建翻译请求
	request := &model.TranslationRequest{
		Content: content,
	}

	// 执行翻译
	return p.translator.Translate(request, outputProcessor)
}
