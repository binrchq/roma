-- 修复passports表的passport字段唯一索引问题
-- 问题：SSH私钥内容过长（超过2700字节），超过了PostgreSQL B-tree索引的大小限制
-- 解决方案：删除唯一索引，改为在应用层检查唯一性

-- 删除现有的唯一索引（如果存在）
DROP INDEX IF EXISTS uni_passports_passport;
DROP INDEX IF EXISTS passports_passport_key;
DROP INDEX IF EXISTS idx_passports_passport;

-- 如果需要，也可以删除其他可能的唯一约束
-- ALTER TABLE passports DROP CONSTRAINT IF EXISTS passports_passport_key;

-- 注意：删除索引后，passport字段的唯一性将在应用层（passport_operation.go）进行检查

