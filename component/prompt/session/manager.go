package session

import (
	"context"
	"fmt"
	"strings"
	"time"

	"microdev/pkg/logger"
)

// Message 表示对话中的一条消息
type Message struct {
	Role      string    `json:"role"`      // "user" 或 "assistant"
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// Session 表示一个对话会话
type Session struct {
	ID       string    `json:"id"`       // 会话ID
	Messages []Message `json:"messages"` // 消息历史
	Summary  string    `json:"summary"`  // 会话摘要（用于记忆压缩）
	logger   *logger.Logger
}

// Manager 会话管理器
type Manager struct {
	currentSession *Session
	storage        SessionStorage
	logger         *logger.Logger
	maxMessages    int // 最大消息数，超过后进行压缩
	ctx            context.Context
}

// NewManager 创建新的会话管理器
func NewManager(log *logger.Logger) *Manager {
	return NewManagerWithStorage(log, nil)
}

// NewManagerWithStorage 使用指定存储创建会话管理器
func NewManagerWithStorage(log *logger.Logger, storage SessionStorage) *Manager {
	if storage == nil {
		// 使用默认内存存储
		storage = NewMemoryStorage(nil)
	}

	return &Manager{
		storage:     storage,
		logger:      log,
		maxMessages: 8, // 默认保留最近8条消息（4轮对话）
		ctx:         context.Background(),
	}
}

// StartSession 开始新的会话
func (m *Manager) StartSession() *Session {
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	m.currentSession = &Session{
		ID:       sessionID,
		Messages: make([]Message, 0),
		logger:   m.logger,
	}

	// 保存到存储
	if err := m.storage.SaveSession(m.ctx, m.currentSession); err != nil {
		m.logger.Error().Err(err).Str("session_id", sessionID).Msg("保存会话失败")
	}

	m.logger.Info().Str("session_id", sessionID).Msg("开始新的对话会话")
	return m.currentSession
}

// AddUserMessage 添加用户消息
func (m *Manager) AddUserMessage(content string) {
	if m.currentSession == nil {
		m.StartSession()
	}

	message := Message{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}

	m.currentSession.Messages = append(m.currentSession.Messages, message)

	// 保存消息到存储
	if err := m.storage.SaveMessage(m.ctx, m.currentSession.ID, &message); err != nil {
		m.logger.Error().Err(err).Str("session_id", m.currentSession.ID).Msg("保存用户消息失败")
	}

	// 详细日志：记录用户输入
	m.logger.Info().
		Str("session_id", m.currentSession.ID).
		Str("role", "user").
		Str("content", content).
		Int("content_length", len(content)).
		Int("total_messages", len(m.currentSession.Messages)).
		Msg("📝 记录用户消息")

	// 检查是否需要压缩记忆
	m.checkAndCompressMemory()
}

// AddAssistantMessage 添加助手消息
func (m *Manager) AddAssistantMessage(content string) {
	if m.currentSession == nil {
		return
	}

	message := Message{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}

	m.currentSession.Messages = append(m.currentSession.Messages, message)

	// 保存消息到存储
	if err := m.storage.SaveMessage(m.ctx, m.currentSession.ID, &message); err != nil {
		m.logger.Error().Err(err).Str("session_id", m.currentSession.ID).Msg("保存助手消息失败")
	}

	// 详细日志：记录助手回复
	m.logger.Info().
		Str("session_id", m.currentSession.ID).
		Str("role", "assistant").
		Str("content", content).
		Int("content_length", len(content)).
		Int("total_messages", len(m.currentSession.Messages)).
		Msg("🤖 记录助手回复")
}

// GetCurrentSession 获取当前会话
func (m *Manager) GetCurrentSession() *Session {
	return m.currentSession
}

// GetContextForOptimization 获取用于优化的上下文信息
func (m *Manager) GetContextForOptimization() string {
	if m.currentSession == nil || len(m.currentSession.Messages) == 0 {
		m.logger.Debug().Msg("🔍 无会话或消息，返回空上下文")
		return ""
	}

	var contextParts []string

	// 如果有会话摘要，先添加摘要
	if m.currentSession.Summary != "" {
		contextParts = append(contextParts, fmt.Sprintf("**会话摘要：**\n%s", m.currentSession.Summary))
		m.logger.Info().
			Str("session_id", m.currentSession.ID).
			Str("summary", m.currentSession.Summary).
			Int("summary_length", len(m.currentSession.Summary)).
			Msg("📋 包含会话摘要")
	}

	// 添加最近的对话历史
	recentMessages := m.getRecentMessages(6) // 最近3轮对话
	if len(recentMessages) > 0 {
		contextParts = append(contextParts, "**最近对话：**")
		for _, msg := range recentMessages {
			role := "用户"
			if msg.Role == "assistant" {
				role = "助手"
			}
			contextParts = append(contextParts, fmt.Sprintf("%s: %s", role, msg.Content))
		}

		m.logger.Info().
			Str("session_id", m.currentSession.ID).
			Int("recent_messages_count", len(recentMessages)).
			Int("total_messages", len(m.currentSession.Messages)).
			Msg("💬 包含最近对话历史")
	}

	if len(contextParts) == 0 {
		m.logger.Debug().Msg("🔍 无可用上下文信息")
		return ""
	}

	context := strings.Join(contextParts, "\n\n")

	// 详细日志：显示构建的上下文
	m.logger.Info().
		Str("session_id", m.currentSession.ID).
		Int("context_length", len(context)).
		Bool("has_summary", m.currentSession.Summary != "").
		Int("recent_messages", len(recentMessages)).
		Str("context_preview", context[:minInt(200, len(context))]).
		Msg("🧠 构建优化上下文")

	return context
}

// getRecentMessages 获取最近的消息
func (m *Manager) getRecentMessages(count int) []Message {
	if m.currentSession == nil || len(m.currentSession.Messages) == 0 {
		return nil
	}

	messages := m.currentSession.Messages
	if len(messages) <= count {
		return messages
	}

	return messages[len(messages)-count:]
}

// checkAndCompressMemory 检查并压缩记忆
func (m *Manager) checkAndCompressMemory() {
	if m.currentSession == nil || len(m.currentSession.Messages) <= m.maxMessages {
		return
	}

	m.logger.Info().
		Str("session_id", m.currentSession.ID).
		Int("current_messages", len(m.currentSession.Messages)).
		Int("max_messages", m.maxMessages).
		Msg("🗜️ 触发记忆压缩")

	// 获取需要压缩的消息（保留最近的消息）
	messagesToCompress := m.currentSession.Messages[:len(m.currentSession.Messages)-m.maxMessages/2]
	recentMessages := m.currentSession.Messages[len(m.currentSession.Messages)-m.maxMessages/2:]

	// 记录压缩前的详细信息
	m.logger.Info().
		Str("session_id", m.currentSession.ID).
		Int("messages_to_compress", len(messagesToCompress)).
		Int("messages_to_keep", len(recentMessages)).
		Msg("📦 准备压缩消息")

	// 生成摘要
	summary := m.generateSummary(messagesToCompress)

	// 更新会话
	oldSummary := m.currentSession.Summary
	m.currentSession.Summary = summary
	m.currentSession.Messages = recentMessages

	// 保存更新的摘要到存储
	if err := m.storage.UpdateSessionSummary(m.ctx, m.currentSession.ID, summary); err != nil {
		m.logger.Error().Err(err).Str("session_id", m.currentSession.ID).Msg("更新会话摘要失败")
	}

	// 详细日志：显示压缩结果
	m.logger.Info().
		Str("session_id", m.currentSession.ID).
		Int("compressed_messages", len(messagesToCompress)).
		Int("remaining_messages", len(recentMessages)).
		Bool("had_previous_summary", oldSummary != "").
		Int("new_summary_length", len(summary)).
		Str("summary_preview", summary[:minInt(100, len(summary))]).
		Msg("✅ 记忆压缩完成")
}

// generateSummary 生成对话摘要
func (m *Manager) generateSummary(messages []Message) string {
	if len(messages) == 0 {
		return ""
	}

	var summaryParts []string
	var currentTopic string
	var keyPoints []string

	for _, msg := range messages {
		if msg.Role == "user" {
			// 提取用户的主要需求
			if len(msg.Content) > 50 {
				currentTopic = msg.Content[:50] + "..."
			} else {
				currentTopic = msg.Content
			}
		} else if msg.Role == "assistant" {
			// 提取助手回复的关键点
			if strings.Contains(msg.Content, "优化后的提示词") {
				keyPoints = append(keyPoints, "提供了优化建议")
			}
			if strings.Contains(msg.Content, "主要改进点") {
				keyPoints = append(keyPoints, "说明了改进要点")
			}
		}
	}

	if currentTopic != "" {
		summaryParts = append(summaryParts, fmt.Sprintf("讨论主题：%s", currentTopic))
	}

	if len(keyPoints) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("关键活动：%s", strings.Join(keyPoints, "、")))
	}

	summaryParts = append(summaryParts, fmt.Sprintf("对话轮次：%d条消息", len(messages)))

	return strings.Join(summaryParts, "\n")
}

// GetSessionStats 获取会话统计信息
func (m *Manager) GetSessionStats() map[string]interface{} {
	if m.currentSession == nil {
		return map[string]interface{}{
			"session_active": false,
		}
	}

	// 尝试从存储获取统计信息
	if stats, err := m.storage.GetSessionStats(m.ctx, m.currentSession.ID); err == nil {
		return map[string]interface{}{
			"session_active":     stats.SessionActive,
			"session_id":         stats.SessionID,
			"total_messages":     stats.TotalMessages,
			"user_messages":      stats.UserMessages,
			"assistant_messages": stats.AssistantMessages,
			"has_summary":        stats.HasSummary,
			"summary_length":     stats.SummaryLength,
		}
	}

	// 回退到本地计算
	userMessages := 0
	assistantMessages := 0

	for _, msg := range m.currentSession.Messages {
		if msg.Role == "user" {
			userMessages++
		} else if msg.Role == "assistant" {
			assistantMessages++
		}
	}

	return map[string]interface{}{
		"session_active":     true,
		"session_id":         m.currentSession.ID,
		"total_messages":     len(m.currentSession.Messages),
		"user_messages":      userMessages,
		"assistant_messages": assistantMessages,
		"has_summary":        m.currentSession.Summary != "",
		"summary_length":     len(m.currentSession.Summary),
	}
}

// EndSession 结束当前会话
func (m *Manager) EndSession() {
	if m.currentSession != nil {
		stats := m.GetSessionStats()
		m.logger.Info().Interface("stats", stats).Msg("结束对话会话")
		m.currentSession = nil
	}
}

// LoadSession 加载指定会话
func (m *Manager) LoadSession(sessionID string) error {
	session, err := m.storage.LoadSession(m.ctx, sessionID)
	if err != nil {
		return err
	}

	m.currentSession = session
	m.currentSession.logger = m.logger
	m.logger.Info().Str("session_id", sessionID).Msg("加载会话成功")
	return nil
}

// ListSessions 列出所有会话
func (m *Manager) ListSessions(limit, offset int) ([]*SessionInfo, error) {
	return m.storage.ListSessions(m.ctx, limit, offset)
}

// CleanupExpiredSessions 清理过期会话
func (m *Manager) CleanupExpiredSessions(expiredBefore time.Time) error {
	return m.storage.CleanupExpiredSessions(m.ctx, expiredBefore)
}

// GetStorageMetrics 获取存储指标
func (m *Manager) GetStorageMetrics() (*StorageMetrics, error) {
	if provider, ok := m.storage.(MetricsProvider); ok {
		return provider.GetMetrics(m.ctx)
	}
	return nil, NewStorageError("unsupported", "storage does not support metrics", nil)
}

// Close 关闭管理器
func (m *Manager) Close() error {
	if m.storage != nil {
		return m.storage.Close()
	}
	return nil
}

// minInt 返回两个整数中的较小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
