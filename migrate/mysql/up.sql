-- 全量建表快照（幂等）。改 schema 直接改本文件，不考虑数据兼容；重建先跑 drop.sql。
-- 表清单：prompt_session prompt_message
-- 语句以分号结尾，字符串字面量里不要出现分号：migrate.Statements 按分号切分。
-- 依赖 MySQL 8.0.13 及以上，TEXT 列才能写表达式默认值。

CREATE TABLE IF NOT EXISTS prompt_session (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  summary TEXT NOT NULL DEFAULT (''),
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0,
  deleted_at BIGINT NOT NULL DEFAULT 0,
  INDEX idx_prompt_session_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS prompt_message (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT NOT NULL DEFAULT 0,
  role VARCHAR(32) NOT NULL DEFAULT '',
  content MEDIUMTEXT NOT NULL DEFAULT (''),
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0,
  deleted_at BIGINT NOT NULL DEFAULT 0,
  INDEX idx_prompt_message_session_id_id (session_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
