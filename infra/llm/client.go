package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"microdev/component/config"
	"microdev/infra/consts"
)

type Client struct {
	model llms.Model
	cfg   *config.Config
}

// NewClient 创建对接通义千问的 OpenAI 兼容客户端。
func NewClient(cfg *config.Config) (*Client, error) {
	model, err := openai.New(
		openai.WithToken(cfg.GetAPIKey()),
		openai.WithBaseURL(consts.DashScopeBaseURL),
		openai.WithModel(cfg.ModelName),
	)
	if err != nil {
		return nil, fmt.Errorf("create llm client: %w", err)
	}
	return &Client{model: model, cfg: cfg}, nil
}

// Generate 调用大模型，结果写入 param.Output，同时返回完整文本。
func (c *Client) Generate(ctx context.Context, param *GenerateParam) (string, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return "", err
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, param.System),
		llms.TextParts(llms.ChatMessageTypeHuman, param.User),
	}
	options := []llms.CallOption{
		llms.WithTemperature(c.cfg.Temperature),
		llms.WithMaxTokens(int(c.cfg.MaxTokens)),
	}

	var content strings.Builder
	if c.cfg.StreamOutput {
		streamFunc := func(_ context.Context, chunk []byte) error {
			content.Write(chunk)
			_, err := param.Output.Write(chunk)
			return err
		}
		options = append(options, llms.WithStreamingFunc(streamFunc))
	}

	resp, err := c.model.GenerateContent(ctx, messages, options...)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}
	if c.cfg.StreamOutput {
		return content.String(), nil
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response content")
	}

	text := strings.TrimSpace(resp.Choices[0].Content)
	if _, err := param.Output.Write([]byte(text)); err != nil {
		return "", fmt.Errorf("write response: %w", err)
	}
	return text, nil
}

// TranslateText 按方向翻译文本，并按固定结构输出译文、音标与例句。
func (c *Client) TranslateText(ctx context.Context, param *TranslateTextParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}

	system := consts.PromptTranslateEn2Zh
	if param.Direction == consts.DirectionZh2En {
		system = consts.PromptTranslateZh2En
	}
	_, err := c.Generate(ctx, &GenerateParam{System: system, User: param.Text, Output: param.Output})
	return err
}

// OptimizePrompt 用简版系统提示词优化单条提示词。
func (c *Client) OptimizePrompt(ctx context.Context, param *OptimizePromptParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}

	user := fmt.Sprintf(consts.PromptOptimizeUserFormat, param.Prompt)
	_, err := c.Generate(ctx, &GenerateParam{System: consts.PromptOptimizeSimple, User: user, Output: param.Output})
	return err
}
