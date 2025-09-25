package session

import (
	"context"
	"time"
)

// SessionStorage 会话存储接口
// 抽象化会话存储操作，支持多种存储后端（内存、MySQL、Redis等）
type SessionStorage interface {
	// SaveSession 保存会话
	SaveSession(ctx context.Context, session *Session) error

	// LoadSession 加载会话
	LoadSession(ctx context.Context, sessionID string) (*Session, error)

	// DeleteSession 删除会话
	DeleteSession(ctx context.Context, sessionID string) error

	// ListSessions 列出所有会话（支持分页）
	ListSessions(ctx context.Context, limit, offset int) ([]*SessionInfo, error)

	// SaveMessage 保存单条消息
	SaveMessage(ctx context.Context, sessionID string, message *Message) error

	// LoadMessages 加载会话的所有消息
	LoadMessages(ctx context.Context, sessionID string, limit, offset int) ([]*Message, error)

	// UpdateSessionSummary 更新会话摘要
	UpdateSessionSummary(ctx context.Context, sessionID string, summary string) error

	// CleanupExpiredSessions 清理过期会话
	CleanupExpiredSessions(ctx context.Context, expiredBefore time.Time) error

	// GetSessionStats 获取会话统计信息
	GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error)

	// Close 关闭存储连接
	Close() error
}

// SessionInfo 会话基本信息
type SessionInfo struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
	HasSummary   bool      `json:"has_summary"`
}

// SessionStats 会话统计信息
type SessionStats struct {
	SessionActive     bool   `json:"session_active"`
	SessionID         string `json:"session_id,omitempty"`
	TotalMessages     int    `json:"total_messages"`
	UserMessages      int    `json:"user_messages"`
	AssistantMessages int    `json:"assistant_messages"`
	HasSummary        bool   `json:"has_summary"`
	SummaryLength     int    `json:"summary_length"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type        string                 `json:"type"`         // 存储类型：memory, mysql, redis
	Options     map[string]interface{} `json:"options"`      // 存储特定配置
	TTL         time.Duration          `json:"ttl"`          // 会话过期时间
	MaxSessions int                    `json:"max_sessions"` // 最大会话数
}

// StorageFactory 存储工厂接口
type StorageFactory interface {
	CreateStorage(config *StorageConfig) (SessionStorage, error)
	GetSupportedTypes() []string
}

// DefaultStorageFactory 默认存储工厂
type DefaultStorageFactory struct{}

// CreateStorage 创建存储实例
func (f *DefaultStorageFactory) CreateStorage(config *StorageConfig) (SessionStorage, error) {
	switch config.Type {
	case "memory", "":
		return NewMemoryStorage(config), nil
	case "mysql":
		return f.createMySQLStorage(config)
	case "redis":
		return f.createRedisStorage(config)
	default:
		return nil, &StorageError{
			Type:    "invalid_type",
			Message: "unsupported storage type: " + config.Type,
		}
	}
}

// createMySQLStorage 创建 MySQL 存储（需要 mysql 构建标签）
func (f *DefaultStorageFactory) createMySQLStorage(config *StorageConfig) (SessionStorage, error) {
	return nil, &StorageError{
		Type:    "build_tag_required",
		Message: "MySQL storage requires build tag 'mysql'. Build with: go build -tags mysql",
	}
}

// createRedisStorage 创建 Redis 存储（需要 redis 构建标签）
func (f *DefaultStorageFactory) createRedisStorage(config *StorageConfig) (SessionStorage, error) {
	return nil, &StorageError{
		Type:    "build_tag_required",
		Message: "Redis storage requires build tag 'redis'. Build with: go build -tags redis",
	}
}

// GetSupportedTypes 获取支持的存储类型
func (f *DefaultStorageFactory) GetSupportedTypes() []string {
	return []string{"memory", "mysql", "redis"}
}

// StorageError 存储错误
type StorageError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e *StorageError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// NewStorageError 创建存储错误
func NewStorageError(errorType, message string, cause error) *StorageError {
	return &StorageError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// StorageMetrics 存储指标
type StorageMetrics struct {
	TotalSessions          int           `json:"total_sessions"`
	ActiveSessions         int           `json:"active_sessions"`
	TotalMessages          int           `json:"total_messages"`
	AverageSessionDuration time.Duration `json:"average_session_duration"`
	StorageSize            int64         `json:"storage_size"` // 字节
	LastCleanup            time.Time     `json:"last_cleanup"`
}

// MetricsProvider 指标提供者接口
type MetricsProvider interface {
	GetMetrics(ctx context.Context) (*StorageMetrics, error)
}
