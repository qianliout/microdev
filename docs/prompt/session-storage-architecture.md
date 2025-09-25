# 会话存储架构设计

## 概述

根据您的反馈，我们已经将记忆功能进行了模块化重构，实现了可插拔的存储后端架构。现在您可以轻松地接入 MySQL、Redis 等存储系统，而无需修改核心业务逻辑。

## 架构设计

### 核心组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Manager       │    │ SessionStorage  │    │ StorageFactory  │
│   (会话管理器)   │────│   (存储接口)    │────│   (存储工厂)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │ MemoryStorage│ │ MySQLStorage│ │RedisStorage│
        │   (内存存储)  │ │  (MySQL存储) │ │ (Redis存储)│
        └──────────────┘ └─────────────┘ └────────────┘
```

### 接口抽象

#### SessionStorage 接口
```go
type SessionStorage interface {
    SaveSession(ctx context.Context, session *Session) error
    LoadSession(ctx context.Context, sessionID string) (*Session, error)
    DeleteSession(ctx context.Context, sessionID string) error
    ListSessions(ctx context.Context, limit, offset int) ([]*SessionInfo, error)
    SaveMessage(ctx context.Context, sessionID string, message *Message) error
    LoadMessages(ctx context.Context, sessionID string, limit, offset int) ([]*Message, error)
    UpdateSessionSummary(ctx context.Context, sessionID string, summary string) error
    CleanupExpiredSessions(ctx context.Context, expiredBefore time.Time) error
    GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error)
    Close() error
}
```

## 存储实现

### 1. 内存存储 (MemoryStorage)
- **用途**: 开发和测试环境
- **特点**: 高性能，无持久化
- **配置**: 无需外部依赖

```go
storage := session.NewMemoryStorage(&session.StorageConfig{
    Type: "memory",
    TTL:  24 * time.Hour,
    MaxSessions: 100,
})
```

### 2. MySQL 存储 (MySQLStorage)
- **用途**: 生产环境，需要持久化
- **特点**: 事务支持，数据持久化
- **构建**: 需要 `mysql` 构建标签

```bash
go build -tags mysql -o micro ./cmd/micro
```

```go
config := &session.StorageConfig{
    Type: "mysql",
    Options: map[string]interface{}{
        "dsn": "user:password@tcp(localhost:3306)/database",
        "max_open_conns": 25,
        "table_prefix": "session_",
    },
}
```

### 3. Redis 存储 (RedisStorage)
- **用途**: 高性能，分布式部署
- **特点**: 高并发，支持集群
- **构建**: 需要 `redis` 构建标签

```bash
go build -tags redis -o micro ./cmd/micro
```

```go
config := &session.StorageConfig{
    Type: "redis",
    Options: map[string]interface{}{
        "addr": "localhost:6379",
        "pool_size": 10,
        "key_prefix": "session:",
    },
}
```

## 使用方式

### 默认使用（内存存储）
```go
manager := session.NewManager(logger)
```

### 自定义存储
```go
// 创建存储配置
config := &session.StorageConfig{
    Type: "mysql", // 或 "redis"
    TTL:  7 * 24 * time.Hour,
    Options: map[string]interface{}{
        // 存储特定配置
    },
}

// 创建存储实例
factory := &session.DefaultStorageFactory{}
storage, err := factory.CreateStorage(config)
if err != nil {
    log.Fatal(err)
}
defer storage.Close()

// 创建管理器
manager := session.NewManagerWithStorage(logger, storage)
```

## 扩展新存储

### 1. 实现存储接口
```go
type CustomStorage struct {
    // 自定义字段
}

func (c *CustomStorage) SaveSession(ctx context.Context, session *Session) error {
    // 实现保存逻辑
}

// 实现其他接口方法...
```

### 2. 扩展工厂
```go
func (f *DefaultStorageFactory) CreateStorage(config *StorageConfig) (SessionStorage, error) {
    switch config.Type {
    case "custom":
        return NewCustomStorage(config)
    // 其他存储类型...
    }
}
```

## 构建标签系统

为了避免不必要的依赖，我们使用 Go 的构建标签系统：

### 文件结构
```
session/
├── storage.go              # 核心接口和默认实现
├── memory_storage.go       # 内存存储（无构建标签）
├── mysql_storage.go        # MySQL存储（需要mysql标签）
├── redis_storage.go        # Redis存储（需要redis标签）
├── storage_mysql.go        # MySQL工厂覆盖（需要mysql标签）
└── storage_redis.go        # Redis工厂覆盖（需要redis标签）
```

### 构建命令
```bash
# 默认构建（仅内存存储）
go build -o micro ./cmd/micro

# 包含MySQL支持
go build -tags mysql -o micro ./cmd/micro

# 包含Redis支持
go build -tags redis -o micro ./cmd/micro

# 包含所有存储支持
go build -tags "mysql redis" -o micro ./cmd/micro
```

## 配置管理

### 存储配置结构
```go
type StorageConfig struct {
    Type        string                 // 存储类型
    Options     map[string]interface{} // 存储特定配置
    TTL         time.Duration          // 会话过期时间
    MaxSessions int                    // 最大会话数
}
```

### 配置示例
参见 `docs/storage-config-examples.md` 获取详细的配置示例。

## 迁移策略

### 从内存存储迁移到持久化存储
1. 实现数据导出功能
2. 配置新的存储后端
3. 导入现有数据
4. 切换存储配置

### 存储类型切换
```go
// 旧存储
oldStorage := session.NewMemoryStorage(nil)
oldManager := session.NewManagerWithStorage(logger, oldStorage)

// 新存储
newConfig := &session.StorageConfig{Type: "mysql", ...}
newStorage, _ := factory.CreateStorage(newConfig)
newManager := session.NewManagerWithStorage(logger, newStorage)

// 数据迁移逻辑...
```

## 性能考虑

### 内存存储
- 读写性能最高
- 无网络开销
- 内存使用需要控制

### MySQL 存储
- 适合复杂查询
- 事务保证数据一致性
- 需要优化数据库连接池

### Redis 存储
- 高并发性能优秀
- 支持分布式部署
- 需要考虑内存使用

## 监控和指标

### 存储指标
```go
type StorageMetrics struct {
    TotalSessions          int
    ActiveSessions         int
    TotalMessages          int
    AverageSessionDuration time.Duration
    StorageSize            int64
    LastCleanup            time.Time
}
```

### 获取指标
```go
if provider, ok := storage.(session.MetricsProvider); ok {
    metrics, err := provider.GetMetrics(ctx)
    // 处理指标数据
}
```

## 最佳实践

### 开发环境
- 使用内存存储
- 设置较短的TTL
- 启用详细日志

### 生产环境
- 选择合适的持久化存储
- 配置连接池参数
- 设置监控告警
- 定期清理过期数据

### 高可用部署
- 使用Redis集群或MySQL主从
- 配置健康检查
- 实现故障转移机制

## 总结

通过模块化的存储架构设计，我们实现了：

1. **可插拔性**: 轻松切换不同存储后端
2. **扩展性**: 方便添加新的存储实现
3. **灵活性**: 支持不同环境的不同需求
4. **性能**: 针对不同场景优化性能
5. **维护性**: 清晰的接口和实现分离

这个架构为后续接入更多存储系统（如 MongoDB、PostgreSQL、Elasticsearch 等）提供了坚实的基础。
