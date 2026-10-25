package service

import (
	"context"
	"fmt"

	"microdev/infra/consts"
	"microdev/infra/llm"
	"microdev/infra/logger"
	promptmodel "microdev/infra/model/prompt"
)

type OptimizerSrv struct {
	sessionSrv *SessionSrv
	llmCli     *llm.Client
	log        *logger.Logger
}

func NewOptimizerSrv(sessionSrv *SessionSrv, llmCli *llm.Client) *OptimizerSrv {
	return &OptimizerSrv{
		sessionSrv: sessionSrv,
		llmCli:     llmCli,
		log:        logger.New(logger.WithModule("prompt"), logger.WithSubModule("optimizer")),
	}
}

// Optimize 优化提示词，结果写入 param.Output，同时返回完整文本。
func (s *OptimizerSrv) Optimize(ctx context.Context, param *promptmodel.OptimizeParam) (string, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return "", err
	}

	generateParam := &llm.GenerateParam{
		System: consts.PromptOptimize,
		User:   s.buildUserMessage(param),
		Output: param.Output,
	}
	content, err := s.llmCli.Generate(ctx, generateParam)
	if err != nil {
		s.log.Error(ctx).Err(err).
			Int("promptLen", len(param.OriginalPrompt)).
			Str("language", param.Language).
			Msg("optimize prompt failed")
		return "", err
	}
	return content, nil
}

// OptimizeWithSession 带入会话上下文优化提示词，并把用户提示与助手回复写回会话。
func (s *OptimizerSrv) OptimizeWithSession(
	ctx context.Context,
	param *promptmodel.OptimizeWithSessionParam,
) (string, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return "", err
	}

	userMessage := &promptmodel.AddMessageParam{SessionID: param.SessionID, Role: consts.RoleUser, Content: param.OriginalPrompt}
	if err := s.sessionSrv.AddMessage(ctx, userMessage); err != nil {
		return "", err
	}
	sessionContext, err := s.sessionSrv.GetContext(ctx, &promptmodel.SessionIDParam{SessionID: param.SessionID})
	if err != nil {
		return "", err
	}

	optimizeParam := &promptmodel.OptimizeParam{
		OriginalPrompt: param.OriginalPrompt,
		Context:        param.Context,
		Language:       param.Language,
		SessionContext: sessionContext,
		Output:         param.Output,
	}
	content, err := s.Optimize(ctx, optimizeParam)
	if err != nil {
		return "", err
	}

	assistantMessage := &promptmodel.AddMessageParam{SessionID: param.SessionID, Role: consts.RoleAssistant, Content: content}
	if err := s.sessionSrv.AddMessage(ctx, assistantMessage); err != nil {
		return "", err
	}
	return content, nil
}

func (s *OptimizerSrv) buildUserMessage(param *promptmodel.OptimizeParam) string {
	message := fmt.Sprintf(consts.PromptOptimizeUserFormat, param.OriginalPrompt)
	if param.Context != "" {
		message += fmt.Sprintf(consts.PromptOptimizeContextFormat, param.Context)
	}
	if param.SessionContext != "" {
		message += fmt.Sprintf(consts.PromptOptimizeSessionFormat, param.SessionContext)
	}
	if param.Language != "" {
		message += fmt.Sprintf(consts.PromptOptimizeLangFormat, param.Language)
	}
	return message
}
