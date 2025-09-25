//go:build redis
// +build redis

package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage Redis 存储实现
type RedisStorage struct {
	client *redis.Client
	config *StorageConfig
	prefix string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr         string        `json:"addr"`           // Redis 地址
	Password     string        `json:"password"`       // 密码
	DB           int           `json:"db"`             // 数据库编号
	PoolSize     int           `json:"pool_size"`      // 连接池大小
	MinIdleConns int           `json:"min_idle_conns"` // 最小空闲连接数
	DialTimeout  time.Duration `json:"dial_timeout"`   // 连接超时
	ReadTimeout  time.Duration `json:"read_timeout"`   // 读取超时
	WriteTimeout time.Duration `json:"write_timeout"`  // 写入超时
	KeyPrefix    string        `json:"key_prefix"`     // 键前缀
}

// NewRedisStorage 创建 Redis 存储实例
func NewRedisStorage(config *StorageConfig) (*RedisStorage, error) {
	// 解析 Redis 配置
	redisConfig := &RedisConfig{
		Addr:         "localhost:6379",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		KeyPrefix:    "session:",
	}

	if config.Options != nil {
		if configBytes, err := json.Marshal(config.Options); err == nil {
			json.Unmarshal(configBytes, redisConfig)
		}
	}

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, NewStorageError("connection_failed", "failed to connect to Redis", err)
	}

	return &RedisStorage{
		client: client,
		config: config,
		prefix: redisConfig.KeyPrefix,
	}, nil
}

// SaveSession 保存会话
func (r *RedisStorage) SaveSession(ctx context.Context, session *Session) error {
	sessionKey := r.getSessionKey(session.ID)

	// 保存会话基本信息
	sessionData := map[string]interface{}{
		"id":         session.ID,
		"summary":    session.Summary,
		"updated_at": time.Now().Unix(),
	}

	if err := r.client.HMSet(ctx, sessionKey, sessionData).Err(); err != nil {
		return NewStorageError("save_failed", "failed to save session", err)
	}

	// 设置过期时间
	if r.config.TTL > 0 {
		r.client.Expire(ctx, sessionKey, r.config.TTL)
	}

	// 添加到会话列表
	r.client.ZAdd(ctx, r.getSessionListKey(), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: session.ID,
	})

	return nil
}

// LoadSession 加载会话
func (r *RedisStorage) LoadSession(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := r.getSessionKey(sessionID)

	// 检查会话是否存在
	exists, err := r.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return nil, NewStorageError("check_failed", "failed to check session existence", err)
	}
	if exists == 0 {
		return nil, NewStorageError("not_found", "session not found", nil)
	}

	// 加载会话基本信息
	sessionData, err := r.client.HMGet(ctx, sessionKey, "id", "summary").Result()
	if err != nil {
		return nil, NewStorageError("load_failed", "failed to load session", err)
	}

	session := &Session{
		ID:      sessionData[0].(string),
		Summary: "",
	}

	if sessionData[1] != nil {
		session.Summary = sessionData[1].(string)
	}

	// 加载消息
	messages, err := r.LoadMessages(ctx, sessionID, 1000, 0)
	if err != nil {
		return nil, err
	}

	session.Messages = make([]Message, len(messages))
	for i, msg := range messages {
		session.Messages[i] = *msg
	}

	return session, nil
}

// DeleteSession 删除会话
func (r *RedisStorage) DeleteSession(ctx context.Context, sessionID string) error {
	sessionKey := r.getSessionKey(sessionID)
	messagesKey := r.getMessagesKey(sessionID)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.Del(ctx, messagesKey)
	pipe.ZRem(ctx, r.getSessionListKey(), sessionID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return NewStorageError("delete_failed", "failed to delete session", err)
	}

	return nil
}

// ListSessions 列出所有会话
func (r *RedisStorage) ListSessions(ctx context.Context, limit, offset int) ([]*SessionInfo, error) {
	sessionListKey := r.getSessionListKey()

	// 按分数（时间戳）倒序获取会话ID
	sessionIDs, err := r.client.ZRevRange(ctx, sessionListKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, NewStorageError("list_failed", "failed to list sessions", err)
	}

	var sessions []*SessionInfo
	for _, sessionID := range sessionIDs {
		sessionKey := r.getSessionKey(sessionID)

		// 获取会话信息
		sessionData, err := r.client.HMGet(ctx, sessionKey, "summary", "updated_at").Result()
		if err != nil {
			continue // 跳过错误的会话
		}

		// 获取消息数量
		messagesKey := r.getMessagesKey(sessionID)
		messageCount, _ := r.client.LLen(ctx, messagesKey).Result()

		var updatedAt time.Time
		if sessionData[1] != nil {
			if timestamp, err := strconv.ParseInt(sessionData[1].(string), 10, 64); err == nil {
				updatedAt = time.Unix(timestamp, 0)
			}
		}

		info := &SessionInfo{
			ID:           sessionID,
			CreatedAt:    updatedAt, // 简化实现，使用更新时间
			UpdatedAt:    updatedAt,
			MessageCount: int(messageCount),
			HasSummary:   sessionData[0] != nil && sessionData[0].(string) != "",
		}

		sessions = append(sessions, info)
	}

	return sessions, nil
}

// SaveMessage 保存单条消息
func (r *RedisStorage) SaveMessage(ctx context.Context, sessionID string, message *Message) error {
	messagesKey := r.getMessagesKey(sessionID)

	// 序列化消息
	messageData, err := json.Marshal(message)
	if err != nil {
		return NewStorageError("serialize_failed", "failed to serialize message", err)
	}

	// 添加到消息列表
	if err := r.client.RPush(ctx, messagesKey, messageData).Err(); err != nil {
		return NewStorageError("save_message_failed", "failed to save message", err)
	}

	// 设置过期时间
	if r.config.TTL > 0 {
		r.client.Expire(ctx, messagesKey, r.config.TTL)
	}

	// 更新会话的最后活动时间
	sessionKey := r.getSessionKey(sessionID)
	r.client.HSet(ctx, sessionKey, "updated_at", time.Now().Unix())
	r.client.ZAdd(ctx, r.getSessionListKey(), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: sessionID,
	})

	return nil
}

// LoadMessages 加载会话的所有消息
func (r *RedisStorage) LoadMessages(ctx context.Context, sessionID string, limit, offset int) ([]*Message, error) {
	messagesKey := r.getMessagesKey(sessionID)

	// 获取消息列表
	messageData, err := r.client.LRange(ctx, messagesKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, NewStorageError("load_messages_failed", "failed to load messages", err)
	}

	var messages []*Message
	for _, data := range messageData {
		var message Message
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			continue // 跳过无法解析的消息
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

// UpdateSessionSummary 更新会话摘要
func (r *RedisStorage) UpdateSessionSummary(ctx context.Context, sessionID string, summary string) error {
	sessionKey := r.getSessionKey(sessionID)

	err := r.client.HMSet(ctx, sessionKey, map[string]interface{}{
		"summary":    summary,
		"updated_at": time.Now().Unix(),
	}).Err()

	if err != nil {
		return NewStorageError("update_failed", "failed to update session summary", err)
	}

	return nil
}

// CleanupExpiredSessions 清理过期会话
func (r *RedisStorage) CleanupExpiredSessions(ctx context.Context, expiredBefore time.Time) error {
	sessionListKey := r.getSessionListKey()

	// 获取过期的会话ID
	expiredSessionIDs, err := r.client.ZRangeByScore(ctx, sessionListKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", expiredBefore.Unix()),
	}).Result()

	if err != nil {
		return NewStorageError("cleanup_failed", "failed to find expired sessions", err)
	}

	// 删除过期会话
	for _, sessionID := range expiredSessionIDs {
		r.DeleteSession(ctx, sessionID)
	}

	return nil
}

// GetSessionStats 获取会话统计信息
func (r *RedisStorage) GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error) {
	sessionKey := r.getSessionKey(sessionID)
	messagesKey := r.getMessagesKey(sessionID)

	// 检查会话是否存在
	exists, err := r.client.Exists(ctx, sessionKey).Result()
	if err != nil || exists == 0 {
		return &SessionStats{SessionActive: false}, nil
	}

	// 获取消息统计
	messages, err := r.LoadMessages(ctx, sessionID, 1000, 0)
	if err != nil {
		return nil, err
	}

	userMessages := 0
	assistantMessages := 0
	for _, msg := range messages {
		if msg.Role == "user" {
			userMessages++
		} else if msg.Role == "assistant" {
			assistantMessages++
		}
	}

	// 获取摘要信息
	summary, _ := r.client.HGet(ctx, sessionKey, "summary").Result()

	return &SessionStats{
		SessionActive:     true,
		SessionID:         sessionID,
		TotalMessages:     len(messages),
		UserMessages:      userMessages,
		AssistantMessages: assistantMessages,
		HasSummary:        summary != "",
		SummaryLength:     len(summary),
	}, nil
}

// GetMetrics 获取存储指标
func (r *RedisStorage) GetMetrics(ctx context.Context) (*StorageMetrics, error) {
	sessionListKey := r.getSessionListKey()

	// 获取总会话数
	totalSessions, err := r.client.ZCard(ctx, sessionListKey).Result()
	if err != nil {
		return nil, NewStorageError("metrics_failed", "failed to get total sessions", err)
	}

	// 获取活跃会话数（最近24小时）
	activeSessions, err := r.client.ZCount(ctx, sessionListKey,
		fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()), "+inf").Result()
	if err != nil {
		return nil, NewStorageError("metrics_failed", "failed to get active sessions", err)
	}

	return &StorageMetrics{
		TotalSessions:  int(totalSessions),
		ActiveSessions: int(activeSessions),
		LastCleanup:    time.Now(), // 简化实现
	}, nil
}

// Close 关闭存储连接
func (r *RedisStorage) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// 辅助方法

// getSessionKey 获取会话键
func (r *RedisStorage) getSessionKey(sessionID string) string {
	return r.prefix + "session:" + sessionID
}

// getMessagesKey 获取消息键
func (r *RedisStorage) getMessagesKey(sessionID string) string {
	return r.prefix + "messages:" + sessionID
}

// getSessionListKey 获取会话列表键
func (r *RedisStorage) getSessionListKey() string {
	return r.prefix + "sessions"
}
