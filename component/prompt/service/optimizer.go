package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"microdev/component/prompt/model"
	"microdev/component/prompt/session"
	"microdev/pkg/config"
	"microdev/pkg/errors"
	"microdev/pkg/logger"
)

// OptimizerService 提示词优化服务
type OptimizerService struct {
	llm            llms.Model
	config         *config.Config
	logger         *logger.Logger
	sessionManager *session.Manager
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
		llm:            llm,
		config:         cfg,
		logger:         log,
		sessionManager: session.NewManager(log),
	}, nil
}

// Optimize 优化提示词
func (s *OptimizerService) Optimize(request *model.OptimizationRequest, writer io.Writer) error {
	s.logger.Info().
		Str("original_prompt", request.OriginalPrompt).
		Int("prompt_length", len(request.OriginalPrompt)).
		Bool("has_context", request.Context != "").
		Bool("has_session_context", request.SessionContext != "").
		Str("language", request.Language).
		Msg("🎯 开始优化提示词")

	// 构建系统提示词
	systemPrompt := s.buildSystemPrompt()
	s.logger.Debug().
		Int("system_prompt_length", len(systemPrompt)).
		Msg("📋 构建系统提示词")

	// 构建用户消息
	userMessage := s.buildUserMessage(request)
	s.logger.Info().
		Int("user_message_length", len(userMessage)).
		Bool("includes_context", request.Context != "").
		Bool("includes_session_context", request.SessionContext != "").
		Str("user_message_preview", userMessage[:minInt(200, len(userMessage))]).
		Msg("📝 构建用户消息")

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

	s.logger.Info().
		Float64("temperature", s.config.Temperature).
		Int("max_tokens", s.config.MaxTokens).
		Bool("stream_output", s.config.StreamOutput).
		Msg("⚙️ 配置LLM参数")

	// 记录发送给LLM的完整内容
	s.logger.Info().
		Str("system_prompt", systemPrompt).
		Str("user_message", userMessage).
		Msg("📤 发送给LLM的完整内容")

	// 如果支持流式输出
	if s.config.StreamOutput {
		s.logger.Info().Msg("🌊 使用流式输出模式")
		return s.generateStreamingResponse(messages, options, writer)
	}

	// 非流式输出
	s.logger.Info().Msg("📄 使用非流式输出模式")
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

	// 添加会话上下文
	if request.SessionContext != "" {
		message += fmt.Sprintf("\n\n**对话上下文：**\n%s", request.SessionContext)
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

	s.logger.Info().Msg("🤖 调用LLM生成响应")
	startTime := time.Now()

	// 调用LLM
	response, err := s.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		s.logger.Error().Err(err).Msg("❌ LLM生成响应失败")
		return errors.NewLLMError("\n ❎ 优化失败", err)
	}

	duration := time.Since(startTime)
	s.logger.Info().
		Dur("duration", duration).
		Int("choices_count", len(response.Choices)).
		Msg("⏱️ LLM响应生成完成")

	// 写入响应
	for i, choice := range response.Choices {
		s.logger.Info().
			Int("choice_index", i).
			Int("content_length", len(choice.Content)).
			Str("content_preview", choice.Content[:minInt(200, len(choice.Content))]).
			Msg("📥 收到LLM响应内容")

		_, err := writer.Write([]byte(choice.Content))
		if err != nil {
			s.logger.Error().Err(err).Msg("❌ 写入响应失败")
			return errors.NewOutputError("写入优化结果失败", err)
		}
	}

	s.logger.Info().Msg("✅ 非流式优化完成")
	return nil
}

// StartSession 开始新的会话
func (s *OptimizerService) StartSession() {
	s.sessionManager.StartSession()
	s.logger.Info().Msg("开始交互式会话模式")
}

// EndSession 结束当前会话
func (s *OptimizerService) EndSession() {
	s.sessionManager.EndSession()
	s.logger.Info().Msg("结束交互式会话模式")
}

// OptimizeWithSession 在会话模式下优化提示词
func (s *OptimizerService) OptimizeWithSession(request *model.OptimizationRequest, writer io.Writer) error {
	s.logger.Info().
		Str("original_prompt", request.OriginalPrompt).
		Int("prompt_length", len(request.OriginalPrompt)).
		Msg("🚀 开始会话模式优化")

	// 添加用户消息到会话历史
	s.sessionManager.AddUserMessage(request.OriginalPrompt)

	// 获取会话上下文
	sessionContext := s.sessionManager.GetContextForOptimization()
	if sessionContext != "" {
		request.SessionContext = sessionContext
		s.logger.Info().
			Int("context_length", len(sessionContext)).
			Str("context_preview", sessionContext[:minInt(150, len(sessionContext))]).
			Msg("🧠 使用会话上下文")
	} else {
		s.logger.Info().Msg("🆕 首次对话，无历史上下文")
	}

	// 执行优化
	err := s.Optimize(request, writer)
	if err != nil {
		s.logger.Error().Err(err).Msg("❌ 优化失败")
		return err
	}

	// 这里需要获取实际的响应内容来添加到会话历史
	// 由于当前架构限制，我们先添加一个占位符
	s.sessionManager.AddAssistantMessage("已提供优化建议")

	s.logger.Info().Msg("✅ 会话模式优化完成")
	return nil
}

// GetSessionStats 获取会话统计信息
func (s *OptimizerService) GetSessionStats() map[string]interface{} {
	return s.sessionManager.GetSessionStats()
}

// minInt 返回两个整数中的较小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
