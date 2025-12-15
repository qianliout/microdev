package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"microdev/pkg/config"
	"microdev/pkg/logger"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Client LLM客户端
type Client struct {
	llm    llms.Model
	config *config.Config
	logger *logger.Logger
}

// NewClient 创建新的LLM客户端
func NewClient(cfg *config.Config) (*Client, error) {
	// 创建OpenAI兼容的客户端，用于通义千问
	var llm llms.Model
	var err error
	key := cfg.DashScopeAPIKey
	if key == "" {
		key = cfg.AliBailianAPIKey
	}
	if key == "" {
		return nil, fmt.Errorf("not get llm api key")
	}
	if cfg.ModelName == "" {
		return nil, fmt.Errorf("not get llm model name")
	}

	llm, err = openai.New(
		openai.WithToken(key),
		openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		openai.WithModel(cfg.ModelName),
	)
	if err != nil {
		return nil, fmt.Errorf("create llm client failed: %v", err)
	}
	log.Info().Str("model", cfg.ModelName).Msg("LLM客户端初始化成功")
	log := logger.NewLogger(
		logger.WithModule("llm", ""),
	)
	return &Client{
		llm:    llm,
		config: cfg,
		logger: log,
	}, nil
}

// GenerateOptimizePrompt 生成优化的prompt
func (c *Client) GenerateOptimizePrompt(input string, writer io.Writer) error {
	c.logger.Info().Int("input_len", len(input)).Msg("开始生成prompt")

	// 构建系统提示词
	systemPrompt := `你是一个专业的prompt工程师，擅长优化和改进用户的提示词。请根据用户提供的原始内容，生成一个更加清晰、具体、有效的中文提示词。

优化原则：
1. 保持原意不变，但表达更清晰
2. 添加必要的上下文和约束条件
3. 使用更精确的词汇和表达
4. 确保指令明确、可执行
5. 适当添加输出格式要求

请直接输出优化后的提示词，不需要额外的解释。`

	// 构建用户消息
	userMessage := fmt.Sprintf("请优化以下提示词：\n\n%s", input)

	// 准备消息
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userMessage),
	}

	// 设置生成选项
	options := []llms.CallOption{
		llms.WithTemperature(c.config.Temperature),
		llms.WithMaxTokens(c.config.MaxTokens),
	}

	// 如果支持流式输出
	if c.config.StreamOutput {
		return c.generateStreamingResponse(messages, options, writer)
	}

	// 非流式输出
	return c.generateResponse(messages, options, writer)
}

// TranslateText 翻译文本
// direct: en2zh 或 zh2en
func (c *Client) TranslateText(input string, direct string, writer io.Writer) error {
	c.logger.Info().Str("direct", direct).Int("input_len", len(input)).Msg("start translate")

	var systemPrompt string
	if direct == "en2zh" {
		systemPrompt = `你是一个专业的中英互译助手。将英文文本精准翻译为中文，并严格按以下结构输出：

【原文】
<英文原文>

【译文】
<中文译文>

【音标】
如涉及关键词或短语，提供标准 IPA 音标（若为句子则给出核心词汇音标）。

【例句】
至少提供 2 个中英文对照例句，例句需贴合语义与常用表达。`
	}
	if direct == "zh2en" {
		systemPrompt = `你是一个专业的中英互译助手。将中文文本精准翻译为英文，并严格按以下结构输出：

【原文】
<中文原文>

【译文】
<英文译文>

【音标】
为英文译文中的关键词或核心短语提供标准 IPA 音标。

【例句】
至少提供 2 个中英文对照例句，英文自然地道，中文准确对应。`
	}

	userMessage := input

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userMessage),
	}

	options := []llms.CallOption{
		llms.WithTemperature(c.config.Temperature),
		llms.WithMaxTokens(c.config.MaxTokens),
	}

	if c.config.StreamOutput {
		return c.generateStreamingResponse(messages, options, writer)
	}
	return c.generateResponse(messages, options, writer)
}

// generateStreamingResponse 生成流式响应
func (c *Client) generateStreamingResponse(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	c.logger.Info().Msg("开始流式生成")

	// 添加流式回调
	options = append(options, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 将chunk写入writer
		_, err := writer.Write(chunk)
		return err
	}))

	// 调用LLM
	_, err := c.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return fmt.Errorf("streaming generation failed: %v", err)
	}

	c.logger.Info().Msg("✅ 流式生成完成")
	return nil
}

// generateResponse 生成非流式响应
func (c *Client) generateResponse(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	c.logger.Debug().Msg("开始非流式生成")

	// 调用LLM
	resp, err := c.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return fmt.Errorf("generate failed: %v", err)
	}

	// 提取响应内容
	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty response content")
	}

	content := resp.Choices[0].Content

	// 写入响应
	_, err = writer.Write([]byte(strings.TrimSpace(content)))
	if err != nil {
		return fmt.Errorf("write response failed: %v", err)
	}

	c.logger.Info().Msg("✅ 非流式生成完成")
	return nil
}
