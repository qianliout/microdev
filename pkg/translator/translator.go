package translator

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/logger"
	"microdev/pkg/markdown"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Translator 翻译器
type Translator struct {
	llm    llms.Model
	config *config.Config
	logger *logger.Logger
}

// NewTranslator 创建新的翻译器
func NewTranslator(cfg *config.Config, log *logger.Logger) (*Translator, error) {
	// 创建OpenAI兼容的客户端，用于通义千问
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
		return nil, errors.NewConfigError("没有可用的API密钥", nil)
	}

	if err != nil {
		return nil, errors.NewLLMError("创建翻译器失败", err)
	}
	log.Info().Str("model", cfg.ModelName).Msg("翻译器初始化成功")

	return &Translator{
		llm:    llm,
		config: cfg,
		logger: log,
	}, nil
}

// Translate 执行翻译
func (t *Translator) Translate(input string, writer io.Writer) error {
	return t.TranslateWithPath(input, "", writer)
}

// TranslateWithPath 执行翻译，支持文件路径处理
func (t *Translator) TranslateWithPath(input string, inputPath string, writer io.Writer) error {
	t.logger.Info().Int("input_len", len(input)).Msg("开始翻译")

	// 检查是否需要切分
	if len(input) > 2000 {
		return t.translateInChunks(input, inputPath, writer)
	}

	// 直接翻译
	return t.translateSingle(input, inputPath, writer, 0)
}

// translateInChunks 分块翻译
func (t *Translator) translateInChunks(input string, inputPath string, writer io.Writer) error {
	chunks := t.splitContent(input)
	t.logger.Info().Int("chunks", len(chunks)).Msg("内容过长，进行分块处理")

	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		err := t.translateSingle(chunk, inputPath, writer, i+1)
		if err != nil {
			return err
		}

		// 在块之间添加分隔
		if i < len(chunks)-1 {
			writer.Write([]byte("\n---\n\n"))
		}
	}

	return nil
}

// translateSingle 翻译单个内容块
func (t *Translator) translateSingle(input string, inputPath string, writer io.Writer, chunkNum int) error {
	// 检测语言方向
	direction := t.detectLanguage(input)

	// 输出分块标识（如果是分块处理）
	if chunkNum > 0 {
		writer.Write([]byte(fmt.Sprintf("[分块 #%d]\n", chunkNum)))
	}

	// 输出原文
	writer.Write([]byte("[原文]\n"))
	writer.Write([]byte(input))
	writer.Write([]byte("\n\n[译文]\n"))

	// 构建翻译提示词
	systemPrompt := t.buildSystemPrompt(direction)
	userMessage := fmt.Sprintf("请翻译以下内容：\n\n%s", input)

	// 准备消息
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userMessage),
	}

	// 设置生成选项
	options := []llms.CallOption{
		llms.WithTemperature(0.3), // 翻译任务使用较低的温度
		llms.WithMaxTokens(t.config.MaxTokens),
	}

	// 执行翻译
	if t.config.StreamOutput {
		return t.translateStreaming(messages, options, writer)
	}

	return t.translateNonStreaming(messages, options, writer)
}

// detectLanguage 检测主要语言
func (t *Translator) detectLanguage(text string) string {
	chineseCount := 0
	englishCount := 0
	totalCount := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalCount++
			if t.isChinese(r) {
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
func (t *Translator) isChinese(r rune) bool {
	return r >= 0x4e00 && r <= 0x9fff
}

// buildSystemPrompt 构建系统提示词
func (t *Translator) buildSystemPrompt(direction string) string {
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

// splitContent 智能切分内容
func (t *Translator) splitContent(content string) []string {
	// 使用Markdown拆分器进行智能拆分
	splitter := markdown.NewSplitter(2000)

	// 检查是否是Markdown内容
	if t.isMarkdownContent(content) {
		chunks := splitter.SplitByHeadings(content)
		var result []string
		for i, chunk := range chunks {
			if chunk.Title != "" {
				t.logger.Debug().Int("index", i+1).Int("level", chunk.Level).Str("title", chunk.Title).Msg("分块")
			} else {
				t.logger.Debug().Int("index", i+1).Msg("分块: 内容块")
			}
			result = append(result, chunk.Content)
		}
		return result
	}

	// 对于非Markdown内容，使用简单的段落分割
	return t.splitByParagraphs(content)
}

// isMarkdownContent 检查是否是Markdown内容
func (t *Translator) isMarkdownContent(content string) bool {
	// 简单检查是否包含Markdown标记
	markdownPatterns := []string{
		"^#+ ",             // 标题
		"```",              // 代码块
		"\\*\\*",           // 加粗
		"\\[.*\\]\\(.*\\)", // 链接
		"^- ",              // 列表
		"^\\* ",            // 列表
		"^> ",              // 引用
	}

	for _, pattern := range markdownPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}

// splitByParagraphs 按段落分割（原有逻辑）
func (t *Translator) splitByParagraphs(content string) []string {
	// 按段落分割
	paragraphs := strings.Split(content, "\n\n")

	var chunks []string
	var currentChunk strings.Builder

	for _, paragraph := range paragraphs {
		// 如果当前块加上新段落超过2000字符，则开始新块
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph) > 2000 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
	}

	// 添加最后一块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// translateStreaming 流式翻译
func (t *Translator) translateStreaming(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	t.logger.Debug().Msg("开始流式翻译")

	// 添加流式回调
	options = append(options, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 将chunk写入writer
		_, err := writer.Write(chunk)
		return err
	}))

	// 调用LLM
	_, err := t.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return errors.NewLLMError("流式翻译失败", err)
	}

	t.logger.Debug().Msg("流式翻译完成")
	return nil
}

// translateNonStreaming 非流式翻译
func (t *Translator) translateNonStreaming(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	t.logger.Debug().Msg("开始非流式翻译")

	// 调用LLM
	resp, err := t.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return errors.NewLLMError("翻译失败", err)
	}

	// 提取响应内容
	if len(resp.Choices) == 0 {
		return errors.NewLLMError("没有收到翻译结果", nil)
	}

	content := resp.Choices[0].Content

	// 写入响应
	_, err = writer.Write([]byte(strings.TrimSpace(content)))
	if err != nil {
		return errors.NewOutputError("写入翻译结果失败", err)
	}

	t.logger.Debug().Msg("非流式翻译完成")
	return nil
}
