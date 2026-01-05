-- 为passports表添加space_id和role_id字段
-- 这两个字段用于支持按空间和角色匹配Passport
-- 如果字段为空，表示通用凭证

-- 添加space_id字段（可选，可为NULL）
ALTER TABLE passports ADD COLUMN IF NOT EXISTS space_id INTEGER;
CREATE INDEX IF NOT EXISTS idx_passports_space_id ON passports(space_id);

-- 添加role_id字段（可选，可为NULL）
ALTER TABLE passports ADD COLUMN IF NOT EXISTS role_id INTEGER;
CREATE INDEX IF NOT EXISTS idx_passports_role_id ON passports(role_id);

-- 添加resource_type索引（如果还没有）
CREATE INDEX IF NOT EXISTS idx_passports_resource_type ON passports(resource_type);

-- 添加外键约束（可选，根据实际需求决定是否添加）
-- ALTER TABLE passports ADD CONSTRAINT fk_passports_space FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE SET NULL;
-- ALTER TABLE passports ADD CONSTRAINT fk_passports_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;

