# 存储配置示例

## 概述

会话管理器支持多种存储后端，包括内存、MySQL 和 Redis。每种存储类型都有其特定的配置选项和使用场景。

## 内存存储（默认）

内存存储是默认选项，适用于开发和测试环境。

```go
config := &session.StorageConfig{
    Type: "memory", // 可以省略，默认为 memory
    TTL:  24 * time.Hour, // 会话过期时间
    MaxSessions: 100, // 最大会话数
}

storage := session.NewMemoryStorage(config)
manager := session.NewManagerWithStorage(logger, storage)
```

### 特点
- 无需外部依赖
- 性能最佳
- 重启后数据丢失
- 适合开发和测试

## MySQL 存储

MySQL 存储适用于需要持久化和高可靠性的生产环境。

```go
config := &session.StorageConfig{
    Type: "mysql",
    TTL:  7 * 24 * time.Hour, // 7天过期
    Options: map[string]interface{}{
        "dsn": "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local",
        "max_open_conns": 25,
        "max_idle_conns": 25,
        "conn_max_lifetime": "5m",
        "table_prefix": "session_",
    },
}

factory := &session.DefaultStorageFactory{}
storage, err := factory.CreateStorage(config)
if err != nil {
    log.Fatal(err)
}

manager := session.NewManagerWithStorage(logger, storage)
```

### 数据库表结构

系统会自动创建以下表：

```sql
-- 会话表
CREATE TABLE session_sessions (
    id VARCHAR(255) PRIMARY KEY,
    summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 消息表
CREATE TABLE session_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    role ENUM('user', 'assistant') NOT NULL,
    content TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session_id (session_id),
    INDEX idx_timestamp (timestamp),
    FOREIGN KEY (session_id) REFERENCES session_sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 特点
- 数据持久化
- 支持事务
- 适合大规模部署
- 需要数据库维护

## Redis 存储

Redis 存储适用于需要高性能和分布式部署的场景。

```go
config := &session.StorageConfig{
    Type: "redis",
    TTL:  24 * time.Hour,
    Options: map[string]interface{}{
        "addr": "localhost:6379",
        "password": "", // 如果有密码
        "db": 0,
        "pool_size": 10,
        "min_idle_conns": 5,
        "dial_timeout": "5s",
        "read_timeout": "3s",
        "write_timeout": "3s",
        "key_prefix": "session:",
    },
}

factory := &session.DefaultStorageFactory{}
storage, err := factory.CreateStorage(config)
if err != nil {
    log.Fatal(err)
}

manager := session.NewManagerWithStorage(logger, storage)
```

### Redis 键结构

```
session:session:session_id  -> Hash (会话基本信息)
session:messages:session_id -> List (消息列表)
session:sessions           -> Sorted Set (会话列表，按时间排序)
```

### 特点
- 高性能
- 支持分布式
- 内存存储，重启数据丢失（除非配置持久化）
- 适合高并发场景

## 配置选项详解

### 通用配置

```go
type StorageConfig struct {
    Type        string                 // 存储类型：memory, mysql, redis
    Options     map[string]interface{} // 存储特定配置
    TTL         time.Duration          // 会话过期时间
    MaxSessions int                    // 最大会话数（仅内存存储）
}
```

### MySQL 配置选项

```go
type MySQLConfig struct {
    DSN             string        // 数据源名称
    MaxOpenConns    int           // 最大打开连接数
    MaxIdleConns    int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大生命周期
    TablePrefix     string        // 表前缀
}
```

### Redis 配置选项

```go
type RedisConfig struct {
    Addr         string        // Redis 地址
    Password     string        // 密码
    DB           int           // 数据库编号
    PoolSize     int           // 连接池大小
    MinIdleConns int           // 最小空闲连接数
    DialTimeout  time.Duration // 连接超时
    ReadTimeout  time.Duration // 读取超时
    WriteTimeout time.Duration // 写入超时
    KeyPrefix    string        // 键前缀
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "log"
    "time"
    
    "microdev/component/prompt/session"
    "microdev/pkg/logger"
)

func main() {
    logger := logger.NewLogger()
    
    // 使用默认内存存储
    manager := session.NewManager(logger)
    
    // 或者使用自定义存储
    config := &session.StorageConfig{
        Type: "mysql",
        TTL:  7 * 24 * time.Hour,
        Options: map[string]interface{}{
            "dsn": "user:password@tcp(localhost:3306)/database",
        },
    }
    
    factory := &session.DefaultStorageFactory{}
    storage, err := factory.CreateStorage(config)
    if err != nil {
        log.Fatal(err)
    }
    defer storage.Close()
    
    manager = session.NewManagerWithStorage(logger, storage)
    
    // 开始会话
    session := manager.StartSession()
    
    // 添加消息
    manager.AddUserMessage("Hello")
    manager.AddAssistantMessage("Hi there!")
    
    // 获取统计信息
    stats := manager.GetSessionStats()
    log.Printf("Session stats: %+v", stats)
    
    // 结束会话
    manager.EndSession()
}
```

### 会话管理

```go
// 列出所有会话
sessions, err := manager.ListSessions(10, 0) // 获取前10个会话
if err != nil {
    log.Printf("Failed to list sessions: %v", err)
}

// 加载特定会话
err = manager.LoadSession("session_123456")
if err != nil {
    log.Printf("Failed to load session: %v", err)
}

// 清理过期会话
expiredBefore := time.Now().Add(-7 * 24 * time.Hour)
err = manager.CleanupExpiredSessions(expiredBefore)
if err != nil {
    log.Printf("Failed to cleanup sessions: %v", err)
}

// 获取存储指标
metrics, err := manager.GetStorageMetrics()
if err != nil {
    log.Printf("Failed to get metrics: %v", err)
} else {
    log.Printf("Storage metrics: %+v", metrics)
}
```

## 性能对比

| 存储类型 | 读取性能 | 写入性能 | 持久化 | 分布式 | 维护成本 |
|---------|---------|---------|--------|--------|----------|
| Memory  | 极高     | 极高     | 否     | 否     | 极低     |
| Redis   | 高       | 高       | 可选   | 是     | 中       |
| MySQL   | 中       | 中       | 是     | 可选   | 高       |

## 最佳实践

### 开发环境
- 使用内存存储
- 设置较短的 TTL（1小时）
- 限制最大会话数

### 测试环境
- 使用 Redis 存储
- 模拟生产环境配置
- 定期清理测试数据

### 生产环境
- 根据需求选择 MySQL 或 Redis
- 配置合适的连接池大小
- 设置监控和告警
- 定期备份数据（MySQL）
- 配置持久化（Redis）

### 迁移策略
1. 实现数据导出/导入功能
2. 使用蓝绿部署
3. 逐步迁移会话数据
4. 验证数据完整性
