-- 全量建表快照（幂等）。改 schema 直接改本文件，不考虑数据兼容；重建先跑 drop.sql。
-- 表清单：prompt_session prompt_message
-- 语句以分号结尾，字符串字面量里不要出现分号：migrate.Statements 按分号切分。

CREATE TABLE IF NOT EXISTS prompt_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  summary TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL DEFAULT 0,
  deleted_at INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_prompt_session_updated_at ON prompt_session (updated_at);

CREATE TABLE IF NOT EXISTS prompt_message (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL DEFAULT 0,
  role TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL DEFAULT 0,
  deleted_at INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_prompt_message_session_id_id ON prompt_message (session_id, id);
