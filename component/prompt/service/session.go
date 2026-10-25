package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"microdev/component/prompt/store"
	"microdev/infra/consts"
	"microdev/infra/logger"
	"microdev/infra/model/filter"
	promptmodel "microdev/infra/model/prompt"
)

type SessionSrv struct {
	sessionDal *store.SessionDao
	messageDal *store.MessageDao
	log        *logger.Logger
}

// summaryAcc 累积被压缩消息中的主题与关键活动。
type summaryAcc struct {
	topic     string
	keyPoints []string
}

func NewSessionSrv(sessionDal *store.SessionDao, messageDal *store.MessageDao) *SessionSrv {
	return &SessionSrv{
		sessionDal: sessionDal,
		messageDal: messageDal,
		log:        logger.New(logger.WithModule("prompt"), logger.WithSubModule("session")),
	}
}

func (s *SessionSrv) CreateSession(ctx context.Context, param *promptmodel.CreateSessionParam) (*promptmodel.Session, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return nil, err
	}

	data := &promptmodel.Session{Summary: param.Summary}
	if err := s.sessionDal.CreateSession(ctx, data); err != nil {
		s.log.Error(ctx).Err(err).Msg("create session failed")
		return nil, err
	}
	return data, nil
}

// AddMessage 保存一条消息；用户消息保存后会检查是否需要压缩记忆。
func (s *SessionSrv) AddMessage(ctx context.Context, param *promptmodel.AddMessageParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}

	data := &promptmodel.Message{SessionID: param.SessionID, Role: param.Role, Content: param.Content}
	if err := s.messageDal.CreateMessage(ctx, data); err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Str("role", param.Role).Msg("add message failed")
		return err
	}
	if param.Role != consts.RoleUser {
		return nil
	}
	return s.compressMemory(ctx, &promptmodel.SessionIDParam{SessionID: param.SessionID})
}

// GetContext 拼接会话摘要与最近的对话，供优化请求带入上下文；会话无内容时返回空串。
func (s *SessionSrv) GetContext(ctx context.Context, param *promptmodel.SessionIDParam) (string, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return "", err
	}

	sessions, _, err := s.sessionDal.SearchSession(ctx, &promptmodel.SearchSessionParam{ID: param.SessionID})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("search session failed")
		return "", err
	}
	// 按时间升序取最后几条，因此先倒序取再翻转。
	messages, _, err := s.messageDal.SearchMessage(ctx, &promptmodel.SearchMessageParam{
		SessionID: param.SessionID,
		Filter:    &filter.Filter{SortFiled: "id", SortBy: filter.SortDesc, Limit: consts.SessionRecentMessages},
	})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("search message failed")
		return "", err
	}

	parts := make([]string, 0, len(messages)+2)
	for _, session := range sessions {
		parts = append(parts, summaryPart(session.Summary)...)
	}
	parts = append(parts, recentParts(messages)...)
	return strings.Join(parts, "\n\n"), nil
}

// GetSessionStats 统计会话中的消息数量与摘要情况，会话不存在时 SessionActive 为 false。
func (s *SessionSrv) GetSessionStats(ctx context.Context, param *promptmodel.SessionIDParam) (*promptmodel.SessionStats, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return nil, err
	}

	stats := &promptmodel.SessionStats{SessionActive: consts.FalseString, HasSummary: consts.FalseString}
	sessions, _, err := s.sessionDal.SearchSession(ctx, &promptmodel.SearchSessionParam{ID: param.SessionID})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("search session failed")
		return nil, err
	}
	if len(sessions) == 0 {
		return stats, nil
	}

	messages, total, err := s.messageDal.SearchMessage(ctx, &promptmodel.SearchMessageParam{
		SessionID: param.SessionID,
		Filter:    &filter.Filter{SortFiled: "id", SortBy: filter.SortAsc, Limit: consts.SessionLoadLimit},
	})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("search message failed")
		return nil, err
	}

	stats.SessionActive = consts.TrueString
	stats.SessionID = param.SessionID
	stats.TotalMessages = total
	stats.UserMessages = countRole(messages, consts.RoleUser)
	stats.AssistantMessages = countRole(messages, consts.RoleAssistant)
	stats.SummaryLength = int64(len(sessions[0].Summary))
	if sessions[0].Summary != "" {
		stats.HasSummary = consts.TrueString
	}
	return stats, nil
}

// SearchSession 分页列出会话及其消息数量。
func (s *SessionSrv) SearchSession(
	ctx context.Context,
	param *promptmodel.SearchSessionParam,
) ([]*promptmodel.SessionInfo, int64, error) {
	result := make([]*promptmodel.SessionInfo, 0)
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return result, 0, err
	}

	sessions, total, err := s.sessionDal.SearchSession(ctx, param)
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.ID).Msg("search session failed")
		return result, 0, err
	}
	errs := make([]error, 0, len(sessions))
	for _, session := range sessions {
		info, err := s.sessionInfo(ctx, session)
		errs = append(errs, err)
		result = append(result, info)
	}
	if err := errors.Join(errs...); err != nil {
		return result, 0, err
	}
	return result, total, nil
}

// DeleteSession 先删消息再删会话。
func (s *SessionSrv) DeleteSession(ctx context.Context, param *promptmodel.SessionIDParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}

	messages, _, err := s.messageDal.SearchMessage(ctx, &promptmodel.SearchMessageParam{SessionID: param.SessionID})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("search message failed")
		return err
	}
	errs := make([]error, 0, len(messages))
	for _, message := range messages {
		errs = append(errs, s.messageDal.DeleteMessage(ctx, message.ID))
	}
	if err := errors.Join(errs...); err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("delete message failed")
		return err
	}
	if err := s.sessionDal.DeleteSession(ctx, param.SessionID); err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("delete session failed")
		return err
	}
	return nil
}

// CleanupExpiredSession 删除更新时间早于 ExpiredBefore 的会话。
func (s *SessionSrv) CleanupExpiredSession(ctx context.Context, param *promptmodel.CleanupSessionParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}

	sessions, _, err := s.sessionDal.SearchSession(ctx, &promptmodel.SearchSessionParam{UpdatedBefore: param.ExpiredBefore})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("expiredBefore", param.ExpiredBefore).Msg("search expired session failed")
		return err
	}
	errs := make([]error, 0, len(sessions))
	for _, session := range sessions {
		errs = append(errs, s.DeleteSession(ctx, &promptmodel.SessionIDParam{SessionID: session.ID}))
	}
	return errors.Join(errs...)
}

// sessionInfo 组装会话基本信息，消息数量取自消息表；统计失败时仍返回不含数量的基本信息。
func (s *SessionSrv) sessionInfo(ctx context.Context, session *promptmodel.Session) (*promptmodel.SessionInfo, error) {
	_, total, err := s.messageDal.SearchMessage(ctx, &promptmodel.SearchMessageParam{
		SessionID: session.ID,
		Filter:    &filter.Filter{Limit: 1},
	})
	info := &promptmodel.SessionInfo{
		ID:           session.ID,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
		MessageCount: total,
		HasSummary:   consts.FalseString,
	}
	if session.Summary != "" {
		info.HasSummary = consts.TrueString
	}
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", session.ID).Msg("count message failed")
		return info, err
	}
	return info, nil
}

// compressMemory 消息数超过上限时，把较早的消息压缩成摘要并删除原消息，只保留最近一半上限的消息。
func (s *SessionSrv) compressMemory(ctx context.Context, param *promptmodel.SessionIDParam) error {
	messages, total, err := s.messageDal.SearchMessage(ctx, &promptmodel.SearchMessageParam{
		SessionID: param.SessionID,
		Filter:    &filter.Filter{SortFiled: "id", SortBy: filter.SortAsc, Limit: consts.SessionLoadLimit},
	})
	if err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("search message failed")
		return err
	}
	if total <= consts.SessionMaxMessages {
		return nil
	}

	keep := int(consts.SessionMaxMessages / 2)
	compressed := messages[:max(len(messages)-keep, 0)]
	summary := buildSummary(compressed)
	updater := map[string]any{
		"summary":    summary,
		"updated_at": time.Now().UTC().UnixMilli(),
	}
	if err := s.sessionDal.UpdateSession(ctx, param.SessionID, updater); err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("update session summary failed")
		return err
	}

	errs := make([]error, 0, len(compressed))
	for _, message := range compressed {
		errs = append(errs, s.messageDal.DeleteMessage(ctx, message.ID))
	}
	if err := errors.Join(errs...); err != nil {
		s.log.Error(ctx).Err(err).Int64("sessionID", param.SessionID).Msg("delete compressed message failed")
		return err
	}
	s.log.Info(ctx).
		Int64("sessionID", param.SessionID).
		Int("compressed", len(compressed)).
		Int("kept", keep).
		Msg("session memory compressed")
	return nil
}

func (vi *summaryAcc) add(message *promptmodel.Message) {
	switch message.Role {
	case consts.RoleUser:
		vi.topic = truncateRunes(message.Content, int(consts.SessionTopicLength))
	case consts.RoleAssistant:
		vi.addAssistant(message.Content)
	}
}

func (vi *summaryAcc) addAssistant(content string) {
	if strings.Contains(content, "优化后的提示词") {
		vi.keyPoints = append(vi.keyPoints, "提供了优化建议")
	}
	if strings.Contains(content, "主要改进点") {
		vi.keyPoints = append(vi.keyPoints, "说明了改进要点")
	}
}

// buildSummary 取最后一条用户消息作主题，并汇总助手回复中的关键活动。
func buildSummary(messages []*promptmodel.Message) string {
	if len(messages) == 0 {
		return ""
	}
	acc := &summaryAcc{}
	for _, message := range messages {
		acc.add(message)
	}

	parts := make([]string, 0, 3)
	if acc.topic != "" {
		parts = append(parts, fmt.Sprintf("讨论主题：%s", acc.topic))
	}
	if len(acc.keyPoints) > 0 {
		parts = append(parts, fmt.Sprintf("关键活动：%s", strings.Join(acc.keyPoints, "、")))
	}
	parts = append(parts, fmt.Sprintf("对话轮次：%d条消息", len(messages)))
	return strings.Join(parts, "\n")
}

func summaryPart(summary string) []string {
	if summary == "" {
		return nil
	}
	return []string{fmt.Sprintf("**会话摘要：**\n%s", summary)}
}

// recentParts 把倒序取出的最近消息翻转成时间正序的对话文本。
func recentParts(messages []*promptmodel.Message) []string {
	if len(messages) == 0 {
		return nil
	}
	parts := make([]string, 0, len(messages)+1)
	parts = append(parts, "**最近对话：**")
	for i := len(messages) - 1; i >= 0; i-- {
		parts = append(parts, fmt.Sprintf("%s: %s", roleLabel(messages[i].Role), messages[i].Content))
	}
	return parts
}

func roleLabel(role string) string {
	if role == consts.RoleAssistant {
		return "助手"
	}
	return "用户"
}

func countRole(messages []*promptmodel.Message, role string) int64 {
	var count int64
	for _, message := range messages {
		matched := message.Role == role
		count += boolToInt64(matched)
	}
	return count
}

func boolToInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}
