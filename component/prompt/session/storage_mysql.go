//go:build mysql
// +build mysql

package session

// createMySQLStorage 创建 MySQL 存储（覆盖默认实现）
func (f *DefaultStorageFactory) createMySQLStorage(config *StorageConfig) (SessionStorage, error) {
	return NewMySQLStorage(config)
}
