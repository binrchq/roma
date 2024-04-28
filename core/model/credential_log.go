package model

import "time"

// CredentialAccessLog 访问凭证访问记录结构体
type CredentialAccessLog struct {
	ID           uint      `gorm:"column:id;primaryKey" json:"id"`                         // 凭据访问日志的唯一标识，作为主键
	CredentialID uint      `gorm:"column:credential_id;index" json:"credential_id"`        // 外键，关联的凭据ID
	UserID       uint      `gorm:"column:user_id;index" json:"user_id"`                    // 外键，关联的用户ID
	Action       string    `gorm:"column:action;not null;type:varchar(255)" json:"action"` // 对凭据执行的操作（例如，'login'，'logout'）
	IP           string    `gorm:"column:ip;not null;type:varchar(15)" json:"ip"`          // 访问来源的IP地址
	Status       string    `gorm:"column:status;not null;type:varchar(255)" json:"status"` // 访问状态（例如，成功，失败）
	Timestamp    time.Time `gorm:"column:timestamp;autoCreateTime" json:"timestamp"`       // 访问的时间戳
}
