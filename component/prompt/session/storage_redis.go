//go:build redis
// +build redis

package session

// createRedisStorage 创建 Redis 存储（覆盖默认实现）
func (f *DefaultStorageFactory) createRedisStorage(config *StorageConfig) (SessionStorage, error) {
	return NewRedisStorage(config)
}
