package initialize

import (
	"gorm.io/gorm"
)

// MCPToken MCP 访问令牌表结构
type MCPToken struct {
	ID        uint      `gorm:"primarykey"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null;comment:令牌哈希"`
	UserID    int       `gorm:"not null;index;comment:用户ID"`
	Username  string    `gorm:"type:varchar(100);comment:用户名"`
	ExpiresAt int64     `gorm:"not null;comment:过期时间戳"`
	CreatedAt int64     `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt int64     `gorm:"autoUpdateTime;comment:更新时间"`
}

// InitMCPTokensTable 初始化 MCP 令牌表
func InitMCPTokensTable(db *gorm.DB) error {
	return db.AutoMigrate(&MCPToken{})
}


