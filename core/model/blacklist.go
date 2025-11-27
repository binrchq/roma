package model

import (
	"time"

	"gorm.io/gorm"
)

// Blacklist IP黑名单模型
type Blacklist struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	IP        string         `gorm:"type:varchar(45);column:ip;uniqueIndex;not null" json:"ip"` // IPv4或IPv6地址
	Reason    string         `gorm:"type:varchar(512);column:reason" json:"reason"`             // 封禁原因
	Source    string         `gorm:"type:varchar(50);column:source;not null" json:"source"`     // 封禁来源：api_auth_failure, ssh_auth_failure, manual
	BanUntil  *time.Time     `gorm:"column:ban_until" json:"ban_until"`                         // 解封时间（nil表示永久封禁）
	IPInfo    string         `gorm:"type:text;column:ip_info" json:"ip_info"`                   // IP信息（JSON格式，从API获取）
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Blacklist) TableName() string {
	return "blacklists"
}

// IsPermanent 检查是否为永久封禁
func (b *Blacklist) IsPermanent() bool {
	return b.BanUntil == nil
}

// IsExpired 检查是否已过期
func (b *Blacklist) IsExpired() bool {
	if b.IsPermanent() {
		return false
	}
	return time.Now().After(*b.BanUntil)
}
