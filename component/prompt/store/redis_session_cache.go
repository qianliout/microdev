//go:build redis

package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"microdev/infra/consts"
	promptmodel "microdev/infra/model/prompt"
)

// RedisSessionCache 用 Redis 缓存会话与消息，需要 `-tags redis` 构建。
type RedisSessionCache struct {
	client *redis.Client
}

func NewRedisSessionCache(client *redis.Client) *RedisSessionCache {
	return &RedisSessionCache{client: client}
}

// SaveSession 写入会话；ID 为 0 时自动分配。
func (dal *RedisSessionCache) SaveSession(ctx context.Context, data *promptmodel.Session) error {
	data = data.Serialize()
	if err := data.Check(); err != nil {
		return err
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	if data.ID <= 0 {
		id, err := dal.client.Incr(cancelCtx, consts.RedisSessionIDKey).Result()
		if err != nil {
			return err
		}
		data.ID = id
	}
	now := time.Now().UTC().UnixMilli()
	data.UpdatedAt = now
	if data.CreatedAt == 0 {
		data.CreatedAt = now
	}

	key := fmt.Sprintf(consts.RedisSessionKeyFormat, data.ID)
	fields := map[string]any{"id": data.ID, "summary": data.Summary, "createdAt": data.CreatedAt, "updatedAt": now}
	pipe := dal.client.Pipeline()
	pipe.HSet(cancelCtx, key, fields)
	pipe.Expire(cancelCtx, key, consts.SessionTTL)
	pipe.ZAdd(cancelCtx, consts.RedisSessionIndexKey, redis.Z{Score: float64(now), Member: data.ID})
	if _, err := pipe.Exec(cancelCtx); err != nil {
		return err
	}
	return nil
}

// SaveMessage 追加一条消息，并刷新会话的最近活动时间。
func (dal *RedisSessionCache) SaveMessage(ctx context.Context, data *promptmodel.Message) error {
	data = data.Serialize()
	if err := data.Check(); err != nil {
		return err
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	id, err := dal.client.Incr(cancelCtx, consts.RedisMessageIDKey).Result()
	if err != nil {
		return err
	}
	now := time.Now().UTC().UnixMilli()
	data.ID = id
	data.CreatedAt = now
	data.UpdatedAt = now
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal message %d: %w", id, err)
	}

	messageKey := fmt.Sprintf(consts.RedisMessageKeyFormat, data.SessionID)
	sessionKey := fmt.Sprintf(consts.RedisSessionKeyFormat, data.SessionID)
	pipe := dal.client.Pipeline()
	pipe.RPush(cancelCtx, messageKey, raw)
	pipe.Expire(cancelCtx, messageKey, consts.SessionTTL)
	pipe.HSet(cancelCtx, sessionKey, "updatedAt", now)
	pipe.ZAdd(cancelCtx, consts.RedisSessionIndexKey, redis.Z{Score: float64(now), Member: data.SessionID})
	if _, err := pipe.Exec(cancelCtx); err != nil {
		return err
	}
	return nil
}

// SearchMessage 按写入顺序读取会话消息，Filter 的 Offset/Limit 控制分页。
func (dal *RedisSessionCache) SearchMessage(
	ctx context.Context,
	param *promptmodel.SearchMessageParam,
) ([]*promptmodel.Message, int64, error) {
	result := make([]*promptmodel.Message, 0)
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return result, 0, err
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	key := fmt.Sprintf(consts.RedisMessageKeyFormat, param.SessionID)
	total, err := dal.client.LLen(cancelCtx, key).Result()
	if err != nil {
		return result, 0, err
	}
	start := int64(0)
	stop := int64(-1)
	if param.Filter != nil && param.Filter.Limit > 0 {
		start = param.Filter.Offset
		stop = param.Filter.Offset + param.Filter.Limit - 1
	}
	rows, err := dal.client.LRange(cancelCtx, key, start, stop).Result()
	if err != nil {
		return result, 0, err
	}

	errs := make([]error, 0, len(rows))
	for _, raw := range rows {
		row := &promptmodel.Message{}
		errs = append(errs, json.Unmarshal([]byte(raw), row))
		result = append(result, row.Deserialize())
	}
	if err := errors.Join(errs...); err != nil {
		return result, 0, fmt.Errorf("unmarshal message of session %d: %w", param.SessionID, err)
	}
	return result, total, nil
}

func (dal *RedisSessionCache) DeleteSession(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("id must be positive")
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	pipe := dal.client.Pipeline()
	pipe.Del(cancelCtx, fmt.Sprintf(consts.RedisSessionKeyFormat, id))
	pipe.Del(cancelCtx, fmt.Sprintf(consts.RedisMessageKeyFormat, id))
	pipe.ZRem(cancelCtx, consts.RedisSessionIndexKey, id)
	if _, err := pipe.Exec(cancelCtx); err != nil {
		return err
	}
	return nil
}

// CleanupExpiredSession 删除最近活动时间早于 expiredBefore（毫秒）的会话。
func (dal *RedisSessionCache) CleanupExpiredSession(ctx context.Context, expiredBefore int64) error {
	if expiredBefore <= 0 {
		return fmt.Errorf("expiredBefore must be positive")
	}
	cancelCtx, cancelFunc := context.WithTimeout(ctx, consts.DBQueryTimeout)
	defer cancelFunc()

	members, err := dal.client.ZRangeByScore(cancelCtx, consts.RedisSessionIndexKey, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(expiredBefore, 10),
	}).Result()
	if err != nil {
		return err
	}

	errs := make([]error, 0, len(members))
	for _, member := range members {
		id, parseErr := strconv.ParseInt(member, 10, 64)
		errs = append(errs, parseErr)
		errs = append(errs, dal.DeleteSession(ctx, id))
	}
	return errors.Join(errs...)
}
