package model

import "time"

// Passport 访问凭证结构体
type Passport struct {
	ID           uint      `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                  // 凭据的唯一标识，作为主键
	ServiceUser  string    `gorm:"column:service_user;not null" json:"service_user"`                   // 服务用户名
	Password     string    `gorm:"column:password;not null" json:"-"`                                  // 凭据密码，不在 JSON 输出中显示
	ResourceType string    `gorm:"column:resource_type;not null" json:"type"`                          // 凭据类型（例如，'database'，'linux'，'windows'）
	PassportPub  string    `gorm:"column:passport_pub;not null" json:"passport_pub"`                   // 凭据的公共部分（如果适用）
	Passport     string    `gorm:"column:passport;unique;not null" json:"passport"`                    // 凭据标识符
	Description  string    `gorm:"type:varchar(1024);column:description" json:"description,omitempty"` // 凭据描述
	ExpiresAt    time.Time `gorm:"not null;default:'2129-09-09 09:09:09'" json:"expires_at"`           // 凭据的过期日期
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`                 // 创建时间
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                 // 更新时间
}
