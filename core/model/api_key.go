package model

import "time"

// Apikey 访问凭证结构体
type Apikey struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                  // 凭据的唯一标识，作为主键
	Apikey      string    `gorm:"column:api_key;unique;not null" json:"api_key"`                      // 访问凭证据标识符
	Description string    `gorm:"type:varchar(1024);column:description" json:"description,omitempty"` // 凭据描述
	ExpiresAt   time.Time `gorm:"not null;default:'2129-09-09 09:09:09'" json:"expires_at"`           // 凭据的过期日期
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`                 // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                 // 更新时间
}
