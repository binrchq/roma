package model

import (
	"time"

	"gorm.io/gorm"
)

// HostKey 表示存储主机密钥的数据模型
type HostKey struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PrivateKey []byte         `gorm:"type:text" json:"private_key"`
	PublicKey  []byte         `gorm:"type:text" json:"public_key"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// 设置 HostKey 对应的数据库表名
func (HostKey) TableName() string {
	return "host_keys"
}
