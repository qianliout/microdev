-- 删全部表（保留数据库）。重建 schema 前执行，之后执行 up.sql。
DROP TABLE IF EXISTS prompt_message;
DROP TABLE IF EXISTS prompt_session;
