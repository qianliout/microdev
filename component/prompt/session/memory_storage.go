package session

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	sessions map[string]*Session
	config   *StorageConfig
	mutex    sync.RWMutex
	metrics  *StorageMetrics
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage(config *StorageConfig) *MemoryStorage {
	if config == nil {
		config = &StorageConfig{
			Type:        "memory",
			TTL:         24 * time.Hour, // 默认24小时过期
			MaxSessions: 100,            // 默认最大100个会话
		}
	}
	
	return &MemoryStorage{
		sessions: make(map[string]*Session),
		config:   config,
		metrics: &StorageMetrics{
			LastCleanup: time.Now(),
		},
	}
}

// SaveSession 保存会话
func (m *MemoryStorage) SaveSession(ctx context.Context, session *Session) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 检查会话数量限制
	if len(m.sessions) >= m.config.MaxSessions && m.sessions[session.ID] == nil {
		// 清理最旧的会话
		m.cleanupOldestSession()
	}
	
	// 深拷贝会话以避免并发修改
	sessionCopy := m.copySession(session)
	m.sessions[session.ID] = sessionCopy
	
	m.updateMetrics()
	return nil
}

// LoadSession 加载会话
func (m *MemoryStorage) LoadSession(ctx context.Context, sessionID string) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, NewStorageError("not_found", "session not found: "+sessionID, nil)
	}
	
	// 检查会话是否过期
	if m.isSessionExpired(session) {
		return nil, NewStorageError("expired", "session expired: "+sessionID, nil)
	}
	
	// 返回深拷贝以避免并发修改
	return m.copySession(session), nil
}

// DeleteSession 删除会话
func (m *MemoryStorage) DeleteSession(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	delete(m.sessions, sessionID)
	m.updateMetrics()
	return nil
}

// ListSessions 列出所有会话
func (m *MemoryStorage) ListSessions(ctx context.Context, limit, offset int) ([]*SessionInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var infos []*SessionInfo
	for _, session := range m.sessions {
		if m.isSessionExpired(session) {
			continue
		}
		
		info := &SessionInfo{
			ID:           session.ID,
			CreatedAt:    session.Messages[0].Timestamp, // 假设第一条消息时间为创建时间
			UpdatedAt:    session.Messages[len(session.Messages)-1].Timestamp,
			MessageCount: len(session.Messages),
			HasSummary:   session.Summary != "",
		}
		infos = append(infos, info)
	}
	
	// 按更新时间排序
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].UpdatedAt.After(infos[j].UpdatedAt)
	})
	
	// 应用分页
	start := offset
	if start > len(infos) {
		start = len(infos)
	}
	
	end := start + limit
	if end > len(infos) {
		end = len(infos)
	}
	
	return infos[start:end], nil
}

// SaveMessage 保存单条消息
func (m *MemoryStorage) SaveMessage(ctx context.Context, sessionID string, message *Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return NewStorageError("not_found", "session not found: "+sessionID, nil)
	}
	
	session.Messages = append(session.Messages, *message)
	m.updateMetrics()
	return nil
}

// LoadMessages 加载会话的所有消息
func (m *MemoryStorage) LoadMessages(ctx context.Context, sessionID string, limit, offset int) ([]*Message, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, NewStorageError("not_found", "session not found: "+sessionID, nil)
	}
	
	messages := session.Messages
	
	// 应用分页
	start := offset
	if start > len(messages) {
		start = len(messages)
	}
	
	end := start + limit
	if end > len(messages) {
		end = len(messages)
	}
	
	var result []*Message
	for i := start; i < end; i++ {
		messageCopy := messages[i]
		result = append(result, &messageCopy)
	}
	
	return result, nil
}

// UpdateSessionSummary 更新会话摘要
func (m *MemoryStorage) UpdateSessionSummary(ctx context.Context, sessionID string, summary string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return NewStorageError("not_found", "session not found: "+sessionID, nil)
	}
	
	session.Summary = summary
	return nil
}

// CleanupExpiredSessions 清理过期会话
func (m *MemoryStorage) CleanupExpiredSessions(ctx context.Context, expiredBefore time.Time) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	var expiredSessions []string
	for sessionID, session := range m.sessions {
		if m.isSessionExpiredBefore(session, expiredBefore) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}
	
	for _, sessionID := range expiredSessions {
		delete(m.sessions, sessionID)
	}
	
	m.metrics.LastCleanup = time.Now()
	m.updateMetrics()
	
	return nil
}

// GetSessionStats 获取会话统计信息
func (m *MemoryStorage) GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return &SessionStats{SessionActive: false}, nil
	}
	
	userMessages := 0
	assistantMessages := 0
	
	for _, msg := range session.Messages {
		if msg.Role == "user" {
			userMessages++
		} else if msg.Role == "assistant" {
			assistantMessages++
		}
	}
	
	return &SessionStats{
		SessionActive:     true,
		SessionID:         session.ID,
		TotalMessages:     len(session.Messages),
		UserMessages:      userMessages,
		AssistantMessages: assistantMessages,
		HasSummary:        session.Summary != "",
		SummaryLength:     len(session.Summary),
	}, nil
}

// GetMetrics 获取存储指标
func (m *MemoryStorage) GetMetrics(ctx context.Context) (*StorageMetrics, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.copyMetrics(), nil
}

// Close 关闭存储连接
func (m *MemoryStorage) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.sessions = make(map[string]*Session)
	return nil
}

// 辅助方法

// copySession 深拷贝会话
func (m *MemoryStorage) copySession(session *Session) *Session {
	copy := &Session{
		ID:      session.ID,
		Summary: session.Summary,
		logger:  session.logger,
	}
	
	copy.Messages = make([]Message, len(session.Messages))
	for i, msg := range session.Messages {
		copy.Messages[i] = Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
	}
	
	return copy
}

// copyMetrics 深拷贝指标
func (m *MemoryStorage) copyMetrics() *StorageMetrics {
	return &StorageMetrics{
		TotalSessions:    m.metrics.TotalSessions,
		ActiveSessions:   m.metrics.ActiveSessions,
		TotalMessages:    m.metrics.TotalMessages,
		AverageSessionDuration: m.metrics.AverageSessionDuration,
		StorageSize:      m.metrics.StorageSize,
		LastCleanup:      m.metrics.LastCleanup,
	}
}

// isSessionExpired 检查会话是否过期
func (m *MemoryStorage) isSessionExpired(session *Session) bool {
	if len(session.Messages) == 0 {
		return false
	}
	
	lastMessage := session.Messages[len(session.Messages)-1]
	return time.Since(lastMessage.Timestamp) > m.config.TTL
}

// isSessionExpiredBefore 检查会话是否在指定时间前过期
func (m *MemoryStorage) isSessionExpiredBefore(session *Session, before time.Time) bool {
	if len(session.Messages) == 0 {
		return false
	}
	
	lastMessage := session.Messages[len(session.Messages)-1]
	return lastMessage.Timestamp.Before(before)
}

// cleanupOldestSession 清理最旧的会话
func (m *MemoryStorage) cleanupOldestSession() {
	var oldestID string
	var oldestTime time.Time
	
	for sessionID, session := range m.sessions {
		if len(session.Messages) == 0 {
			continue
		}
		
		lastMessage := session.Messages[len(session.Messages)-1]
		if oldestID == "" || lastMessage.Timestamp.Before(oldestTime) {
			oldestID = sessionID
			oldestTime = lastMessage.Timestamp
		}
	}
	
	if oldestID != "" {
		delete(m.sessions, oldestID)
	}
}

// updateMetrics 更新指标
func (m *MemoryStorage) updateMetrics() {
	totalSessions := len(m.sessions)
	activeSessions := 0
	totalMessages := 0
	var totalDuration time.Duration
	
	for _, session := range m.sessions {
		if !m.isSessionExpired(session) {
			activeSessions++
		}
		
		totalMessages += len(session.Messages)
		
		if len(session.Messages) >= 2 {
			first := session.Messages[0].Timestamp
			last := session.Messages[len(session.Messages)-1].Timestamp
			totalDuration += last.Sub(first)
		}
	}
	
	m.metrics.TotalSessions = totalSessions
	m.metrics.ActiveSessions = activeSessions
	m.metrics.TotalMessages = totalMessages
	
	if totalSessions > 0 {
		m.metrics.AverageSessionDuration = totalDuration / time.Duration(totalSessions)
	}
	
	// 估算存储大小（简单估算）
	m.metrics.StorageSize = int64(totalMessages * 100) // 假设每条消息平均100字节
}
