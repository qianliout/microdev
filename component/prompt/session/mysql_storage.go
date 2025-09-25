//go:build mysql
// +build mysql

package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL 驱动
)

// MySQLStorage MySQL 存储实现
type MySQLStorage struct {
	db     *sql.DB
	config *StorageConfig
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	DSN             string        `json:"dsn"`               // 数据源名称
	MaxOpenConns    int           `json:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // 连接最大生命周期
	TablePrefix     string        `json:"table_prefix"`      // 表前缀
}

// NewMySQLStorage 创建 MySQL 存储实例
func NewMySQLStorage(config *StorageConfig) (*MySQLStorage, error) {
	// 解析 MySQL 配置
	mysqlConfig := &MySQLConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		TablePrefix:     "session_",
	}

	if config.Options != nil {
		if configBytes, err := json.Marshal(config.Options); err == nil {
			json.Unmarshal(configBytes, mysqlConfig)
		}
	}

	// 连接数据库
	db, err := sql.Open("mysql", mysqlConfig.DSN)
	if err != nil {
		return nil, NewStorageError("connection_failed", "failed to connect to MySQL", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	db.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	db.SetConnMaxLifetime(mysqlConfig.ConnMaxLifetime)

	storage := &MySQLStorage{
		db:     db,
		config: config,
	}

	// 初始化表结构
	if err := storage.initTables(mysqlConfig.TablePrefix); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// initTables 初始化数据库表
func (m *MySQLStorage) initTables(prefix string) error {
	// 会话表
	sessionTable := `
	CREATE TABLE IF NOT EXISTS ` + prefix + `sessions (
		id VARCHAR(255) PRIMARY KEY,
		summary TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_updated_at (updated_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 消息表
	messageTable := `
	CREATE TABLE IF NOT EXISTS ` + prefix + `messages (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		session_id VARCHAR(255) NOT NULL,
		role ENUM('user', 'assistant') NOT NULL,
		content TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_session_id (session_id),
		INDEX idx_timestamp (timestamp),
		FOREIGN KEY (session_id) REFERENCES ` + prefix + `sessions(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	if _, err := m.db.Exec(sessionTable); err != nil {
		return NewStorageError("table_creation_failed", "failed to create sessions table", err)
	}

	if _, err := m.db.Exec(messageTable); err != nil {
		return NewStorageError("table_creation_failed", "failed to create messages table", err)
	}

	return nil
}

// SaveSession 保存会话
func (m *MySQLStorage) SaveSession(ctx context.Context, session *Session) error {
	query := `INSERT INTO session_sessions (id, summary) VALUES (?, ?) 
			  ON DUPLICATE KEY UPDATE summary = VALUES(summary), updated_at = CURRENT_TIMESTAMP`

	_, err := m.db.ExecContext(ctx, query, session.ID, session.Summary)
	if err != nil {
		return NewStorageError("save_failed", "failed to save session", err)
	}

	return nil
}

// LoadSession 加载会话
func (m *MySQLStorage) LoadSession(ctx context.Context, sessionID string) (*Session, error) {
	// 加载会话基本信息
	var session Session
	query := `SELECT id, summary FROM session_sessions WHERE id = ?`

	err := m.db.QueryRowContext(ctx, query, sessionID).Scan(&session.ID, &session.Summary)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewStorageError("not_found", "session not found", err)
		}
		return nil, NewStorageError("load_failed", "failed to load session", err)
	}

	// 加载消息
	messages, err := m.LoadMessages(ctx, sessionID, 1000, 0) // 加载最近1000条消息
	if err != nil {
		return nil, err
	}

	session.Messages = make([]Message, len(messages))
	for i, msg := range messages {
		session.Messages[i] = *msg
	}

	return &session, nil
}

// DeleteSession 删除会话
func (m *MySQLStorage) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM session_sessions WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return NewStorageError("delete_failed", "failed to delete session", err)
	}
	return nil
}

// ListSessions 列出所有会话
func (m *MySQLStorage) ListSessions(ctx context.Context, limit, offset int) ([]*SessionInfo, error) {
	query := `
	SELECT s.id, s.created_at, s.updated_at, 
		   COALESCE(COUNT(m.id), 0) as message_count,
		   (s.summary IS NOT NULL AND s.summary != '') as has_summary
	FROM session_sessions s
	LEFT JOIN session_messages m ON s.id = m.session_id
	GROUP BY s.id, s.created_at, s.updated_at, s.summary
	ORDER BY s.updated_at DESC
	LIMIT ? OFFSET ?`

	rows, err := m.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, NewStorageError("query_failed", "failed to list sessions", err)
	}
	defer rows.Close()

	var sessions []*SessionInfo
	for rows.Next() {
		var info SessionInfo
		err := rows.Scan(&info.ID, &info.CreatedAt, &info.UpdatedAt, &info.MessageCount, &info.HasSummary)
		if err != nil {
			return nil, NewStorageError("scan_failed", "failed to scan session info", err)
		}
		sessions = append(sessions, &info)
	}

	return sessions, nil
}

// SaveMessage 保存单条消息
func (m *MySQLStorage) SaveMessage(ctx context.Context, sessionID string, message *Message) error {
	query := `INSERT INTO session_messages (session_id, role, content, timestamp) VALUES (?, ?, ?, ?)`
	_, err := m.db.ExecContext(ctx, query, sessionID, message.Role, message.Content, message.Timestamp)
	if err != nil {
		return NewStorageError("save_message_failed", "failed to save message", err)
	}
	return nil
}

// LoadMessages 加载会话的所有消息
func (m *MySQLStorage) LoadMessages(ctx context.Context, sessionID string, limit, offset int) ([]*Message, error) {
	query := `SELECT role, content, timestamp FROM session_messages 
			  WHERE session_id = ? ORDER BY timestamp ASC LIMIT ? OFFSET ?`

	rows, err := m.db.QueryContext(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, NewStorageError("query_failed", "failed to load messages", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Role, &msg.Content, &msg.Timestamp)
		if err != nil {
			return nil, NewStorageError("scan_failed", "failed to scan message", err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// UpdateSessionSummary 更新会话摘要
func (m *MySQLStorage) UpdateSessionSummary(ctx context.Context, sessionID string, summary string) error {
	query := `UPDATE session_sessions SET summary = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.ExecContext(ctx, query, summary, sessionID)
	if err != nil {
		return NewStorageError("update_failed", "failed to update session summary", err)
	}
	return nil
}

// CleanupExpiredSessions 清理过期会话
func (m *MySQLStorage) CleanupExpiredSessions(ctx context.Context, expiredBefore time.Time) error {
	query := `DELETE FROM session_sessions WHERE updated_at < ?`
	result, err := m.db.ExecContext(ctx, query, expiredBefore)
	if err != nil {
		return NewStorageError("cleanup_failed", "failed to cleanup expired sessions", err)
	}

	affected, _ := result.RowsAffected()
	if affected > 0 {
		// 日志记录清理的会话数量
	}

	return nil
}

// GetSessionStats 获取会话统计信息
func (m *MySQLStorage) GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error) {
	query := `
	SELECT 
		COUNT(CASE WHEN role = 'user' THEN 1 END) as user_messages,
		COUNT(CASE WHEN role = 'assistant' THEN 1 END) as assistant_messages,
		COUNT(*) as total_messages
	FROM session_messages WHERE session_id = ?`

	var stats SessionStats
	err := m.db.QueryRowContext(ctx, query, sessionID).Scan(
		&stats.UserMessages, &stats.AssistantMessages, &stats.TotalMessages)

	if err != nil {
		if err == sql.ErrNoRows {
			return &SessionStats{SessionActive: false}, nil
		}
		return nil, NewStorageError("stats_failed", "failed to get session stats", err)
	}

	// 检查会话是否存在
	var summary string
	summaryQuery := `SELECT summary FROM session_sessions WHERE id = ?`
	err = m.db.QueryRowContext(ctx, summaryQuery, sessionID).Scan(&summary)
	if err != nil {
		return &SessionStats{SessionActive: false}, nil
	}

	stats.SessionActive = true
	stats.SessionID = sessionID
	stats.HasSummary = summary != ""
	stats.SummaryLength = len(summary)

	return &stats, nil
}

// GetMetrics 获取存储指标
func (m *MySQLStorage) GetMetrics(ctx context.Context) (*StorageMetrics, error) {
	var metrics StorageMetrics

	// 获取会话统计
	sessionQuery := `SELECT COUNT(*) FROM session_sessions`
	m.db.QueryRowContext(ctx, sessionQuery).Scan(&metrics.TotalSessions)

	// 获取消息统计
	messageQuery := `SELECT COUNT(*) FROM session_messages`
	m.db.QueryRowContext(ctx, messageQuery).Scan(&metrics.TotalMessages)

	// 获取活跃会话数（最近24小时有活动）
	activeQuery := `SELECT COUNT(*) FROM session_sessions WHERE updated_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)`
	m.db.QueryRowContext(ctx, activeQuery).Scan(&metrics.ActiveSessions)

	metrics.LastCleanup = time.Now() // 简化实现

	return &metrics, nil
}

// Close 关闭存储连接
func (m *MySQLStorage) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
