package service

import (
	"context"
	"fmt"
	"io"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"microdev/component/prompt/model"
	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/logger"
)

// OptimizerService 提示词优化服务
type OptimizerService struct {
	llm    llms.Model
	config *config.Config
	logger *logger.Logger
}

// NewOptimizerService 创建新的优化服务
func NewOptimizerService(cfg *config.Config, log *logger.Logger) (*OptimizerService, error) {
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
		return nil, errors.NewConfigError("没有可用的API密钥", nil)
	}

	if err != nil {
		return nil, errors.NewLLMError("创建优化服务失败", err)
	}

	log.Info().Str("model", cfg.ModelName).Msg("提示词优化服务初始化成功")

	return &OptimizerService{
		llm:    llm,
		config: cfg,
		logger: log,
	}, nil
}

// Optimize 优化提示词
func (s *OptimizerService) Optimize(request *model.OptimizationRequest, writer io.Writer) error {
	s.logger.Info().Str("original", request.OriginalPrompt).Msg("开始优化提示词")

	// 构建系统提示词
	systemPrompt := s.buildSystemPrompt()

	// 构建用户消息
	userMessage := s.buildUserMessage(request)

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
func (s *OptimizerService) buildSystemPrompt() string {
	return `你是一个专业的提示词优化专家，擅长将用户的原始提示词优化成更清晰、更有效的版本。

## 优化原则：
1. **清晰性**：使提示词更加明确和具体
2. **结构化**：合理组织信息层次
3. **完整性**：补充必要的上下文信息
4. **可操作性**：确保AI能够准确理解和执行

## 优化策略：
- 明确任务目标和期望输出
- 提供必要的背景信息和约束条件
- 使用结构化的格式（如分点、分段）
- 添加示例或模板（如果有助于理解）
- 优化语言表达，使其更加专业和准确

## 输出格式：
请按以下格式输出：

**优化后的提示词：**
[在这里提供优化后的完整提示词]

**主要改进点：**
[简要说明主要的优化改进]

**使用建议：**
[提供使用这个优化提示词的建议]`
}

// buildUserMessage 构建用户消息
func (s *OptimizerService) buildUserMessage(request *model.OptimizationRequest) string {
	message := fmt.Sprintf("请优化以下提示词：\n\n%s", request.OriginalPrompt)

	if request.Context != "" {
		message += fmt.Sprintf("\n\n**上下文信息：**\n%s", request.Context)
	}

	if request.Language != "" {
		message += fmt.Sprintf("\n\n**目标语言：**%s", request.Language)
	}

	return message
}

// generateStreamingResponse 生成流式响应
func (s *OptimizerService) generateStreamingResponse(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	s.logger.Debug().Msg("开始流式优化")

	// 添加流式回调
	options = append(options, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 将chunk写入writer
		_, err := writer.Write(chunk)
		return err
	}))

	// 调用LLM
	_, err := s.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return errors.NewLLMError("\n ❎ 流式优化失败", err)
	}

	s.logger.Info().Msg("✅ 流式优化完成")
	return nil
}

// generateResponse 生成非流式响应
func (s *OptimizerService) generateResponse(messages []llms.MessageContent, options []llms.CallOption, writer io.Writer) error {
	ctx := context.Background()

	s.logger.Debug().Msg("开始非流式优化")

	// 调用LLM
	response, err := s.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return errors.NewLLMError("\n ❎ 优化失败", err)
	}

	// 写入响应
	for _, choice := range response.Choices {
		_, err := writer.Write([]byte(choice.Content))
		if err != nil {
			return errors.NewOutputError("写入优化结果失败", err)
		}
	}

	s.logger.Info().Msg("✅ 优化完成")
	return nil
}
