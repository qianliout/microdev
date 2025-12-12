package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"microdev/component/translate/model"
	"microdev/pkg/config"
	"microdev/pkg/logger"
	"microdev/pkg/utils"
)

// TranslatorService 翻译服务
type TranslatorService struct {
	llm             llms.Model
	config          *config.Config
	logger          *logger.Logger
	markdownService *MarkdownService
}

// NewTranslatorService 创建新的翻译服务
func NewTranslatorService(cfg *config.Config, log *logger.Logger) (*TranslatorService, error) {
	// 创建OpenAI兼容的客户端
	var llm llms.Model
	var err error

	if cfg.HasDashScopeKey() {
		// 使用DashScope API
		llm, err = openai.New(
			openai.WithToken(cfg.DashScopeAPIKey),
			openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
			openai.WithModel(cfg.ModelName),
		)
	} else if cfg.HasBailianKey() {
		// 使用百炼API
		llm, err = openai.New(
			openai.WithToken(cfg.AliBailianAPIKey),
			openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
			openai.WithModel(cfg.ModelName),
		)
	} else {
		return nil, fmt.Errorf("not get llm api key")
	}

	if err != nil {
		return nil, fmt.Errorf("create translator service failed: %v", err)
	}

	log.Info().Str("model", cfg.ModelName).Msg("翻译服务初始化成功")

	return &TranslatorService{
		llm:             llm,
		config:          cfg,
		logger:          log,
		markdownService: NewMarkdownService(2000),
	}, nil
}

// GetConcurrency 获取并发数配置
func (s *TranslatorService) GetConcurrency() int {
	return s.config.Concurrency
}

// Translate 执行翻译
func (s *TranslatorService) Translate(request *model.TranslationRequest, writer io.Writer) error {
	s.logger.Info().Int("input_len", len(request.Content)).Msg("开始翻译")

	// 检查是否需要切分
	if len(request.Content) > 2000 {
		return s.translateInChunks(request, writer)
	}

	// 直接翻译
	return s.translateSingle(request, writer, 0)
}

// translateInChunks 分块翻译
func (s *TranslatorService) translateInChunks(request *model.TranslationRequest, writer io.Writer) error {
	chunks := s.markdownService.SplitContent(request.Content)
	s.logger.Info().Int("chunks", len(chunks)).Msg("内容过长，进行分块处理")

	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		chunkRequest := &model.TranslationRequest{
			Content:    chunk,
			InputPath:  request.InputPath,
			OutputPath: request.OutputPath,
			Language:   request.Language,
		}

		err := s.translateSingle(chunkRequest, writer, i+1)
		if err != nil {
			return err
		}

		// 在分块之间添加分隔符
		if i < len(chunks)-1 {
			writer.Write([]byte("\n---\n\n"))
		}
	}

	return nil
}

// translateSingle 翻译单个内容块
func (s *TranslatorService) translateSingle(request *model.TranslationRequest, writer io.Writer, chunkNum int) error {
	// 检测语言方向
	direction := utils.DetectLanguage(request.Content)
	if request.Language != "" {
		direction = request.Language
	}

	// 输出分块标识（如果是分块处理）
	if chunkNum > 0 {
		writer.Write([]byte(fmt.Sprintf("[分块 #%d]\n", chunkNum)))
	}

	// 输出原文
	writer.Write([]byte("[原文]\n"))
	writer.Write([]byte(request.Content))
	writer.Write([]byte("\n\n[译文]\n"))

	// 构建翻译提示词
	systemPrompt := s.buildSystemPrompt(direction)
	userMessage := fmt.Sprintf("请翻译以下内容：\n\n%s", request.Content)

	// 准备消息
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userMessage),
	}

	// 设置生成选项
	options := []llms.CallOption{
		llms.WithTemperature(s.config.Temperature),
		llms.WithMaxTokens(s.config.MaxTokens),
	}

	// 如果支持流式输出
	if s.config.StreamOutput {
		return s.generateStreamingResponse(messages, options, writer)
	}

	// 非流式输出
	return s.generateResponse(messages, options, writer)
}

// buildSystemPrompt 构建系统提示词
func (s *TranslatorService) buildSystemPrompt(direction string) string {
	basePrompt := `你是一个专业的翻译助手，专门处理中英互译任务。请严格按以下规则执行：

## 翻译规则：
1. 保持语气、语义、风格自然流畅
2. 如果内容是Markdown格式，请识别并区分以下元素：

### 需要翻译的内容：
- 普通段落文字
- 列表项中的文字
- 标题文字（只翻译标题内容）
- 引用块中的文字
- 行内强调文字（如加粗、斜体中的文字）

### 禁止翻译的内容：
- 代码块（三个反引号包裹的内容或单个反引号包裹的内容）
- URL链接地址
- 图片链接地址
- HTML标签或属性
- 表格结构符号

### 格式保留要求：
- 保持原有Markdown语法结构、缩进、换行、符号不变
- 翻译后的内容必须能直接替换原文，不影响渲染

请直接输出翻译结果，不要添加额外解释。`

	if direction == "zh2en" {
		return basePrompt + "\n\n当前任务：将中文翻译成英文。"
	}
	return basePrompt + "\n\n当前任务：将英文翻译成中文。"
}

// generateStreamingResponse 生成流式响应
func (s *TranslatorService) generateStreamingResponse(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	s.logger.Debug().Msg("开始流式翻译")

	// 添加流式回调
	options = append(options, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 将chunk写入writer
		_, err := writer.Write(chunk)
		return err
	}))

	// 调用LLM
	_, err := s.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return fmt.Errorf("streaming translation failed: %v", err)
	}

	s.logger.Info().Msg("✅ 流式翻译完成")
	return nil
}

// generateResponse 生成非流式响应
func (s *TranslatorService) generateResponse(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	s.logger.Debug().Msg("开始非流式翻译")

	// 调用LLM
	response, err := s.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return fmt.Errorf("translation failed: %v", err)
	}

	// 写入响应
	for _, choice := range response.Choices {
		_, err := writer.Write([]byte(choice.Content))
		if err != nil {
			return fmt.Errorf("write translation result failed: %v", err)
		}
	}

	s.logger.Info().Msg("✅ 翻译完成")
	return nil
}
