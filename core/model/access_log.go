package model

import "time"

// AccessLog 访问日志结构体
type AccessLog struct {
	ID           uint      `gorm:"column:id;primaryKey" json:"id"`                                       // 访问日志的唯一标识，作为主键
	UserID       uint      `gorm:"column:user_id;index" json:"user_id"`                                  // 用户ID，索引字段
	ResourceType string    `gorm:"column:resource_type;type:varchar(255);not null" json:"resource_type"` // 被访问资源的类型
	ResourceID   uint      `gorm:"column:resource_id;index" json:"resource_id"`                          // 被访问资源的ID，索引字段
	Action       string    `gorm:"column:action;type:varchar(255);not null" json:"action"`               // 对资源执行的操作
	ActionLevel  string    `gorm:"column:action_level;type:varchar(255);not null" json:"action_level"`   // 执行操作的级别
	Source       string    `gorm:"column:source;type:varchar(255);not null" json:"source"`               // 访问来源
	IPPub        string    `gorm:"column:ip_pub;type:varchar(15);not null" json:"ip_pub"`                // 公网IP地址
	IPPriv       string    `gorm:"column:ip_priv;type:varchar(15);not null" json:"ip_priv"`              // 内网IP地址
	Status       string    `gorm:"column:status;type:varchar(15);not null" json:"status"`                // 访问状态（例如，成功，失败）
	Timestamp    time.Time `gorm:"column:timestamp;autoCreateTime" json:"timestamp"`                     // 访问时间戳
}
