package model

import "time"

// AuditLog 审计日志结构体
type AuditLog struct {
	ID           uint      `gorm:"column:id;primaryKey" json:"id"`                                  // 审计日志的唯一标识，作为主键
	UserID       uint      `gorm:"column:user_id;index" json:"user_id"`                             // 执行操作的用户ID，索引字段
	Username     string    `gorm:"column:username;type:varchar(255);not null" json:"username"`      // 执行操作的用户名
	Action       string    `gorm:"column:action;type:varchar(255);not null" json:"action"`          // 执行的操作（例如，'delete_resource', 'update_user', 'delete_user'）
	ActionType   string    `gorm:"column:action_type;type:varchar(50);not null" json:"action_type"` // 操作类型：高危操作(high_risk)、普通操作(normal)
	ResourceType string    `gorm:"column:resource_type;type:varchar(255)" json:"resource_type"`     // 资源类型（如果适用）
	ResourceID   uint      `gorm:"column:resource_id;index" json:"resource_id"`                     // 资源ID（如果适用），索引字段
	ResourceName string    `gorm:"column:resource_name;type:varchar(255)" json:"resource_name"`     // 资源名称（如果适用）
	Description  string    `gorm:"column:description;type:text" json:"description"`                 // 操作描述
	IPAddress    string    `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`            // 操作来源IP地址
	Status       string    `gorm:"column:status;type:varchar(20);not null" json:"status"`           // 操作状态：成功(success)、失败(failed)
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`             // 错误信息（如果操作失败）
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`              // 操作时间戳
}
