-- 清空全部数据并保留表结构，同时重置自增主键。
DELETE FROM prompt_message;
DELETE FROM prompt_session;
DELETE FROM sqlite_sequence WHERE name IN ('prompt_message', 'prompt_session');
