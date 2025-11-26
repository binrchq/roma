-- 强制删除 roma 数据库的 SQL 脚本
-- 适用于 PostgreSQL

-- 1. 断开所有连接到 roma 数据库的会话
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'roma' AND pid <> pg_backend_pid();

-- 2. 等待一下确保所有连接都已断开（可选，在 psql 中执行时可能需要）
-- SELECT pg_sleep(1);

-- 3. 删除数据库
DROP DATABASE IF EXISTS roma;

-- 如果上面的命令失败，可以尝试：
-- DROP DATABASE roma WITH (FORCE);

