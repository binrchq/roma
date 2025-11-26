package model

import "time"

// ResourceRole 资源角色关联
type ResourceRole struct {
	ID           uint      `gorm:"column:id;primaryKey" json:"id"`
	ResourceID   int64     `gorm:"column:resource_id;index" json:"resource_id"`     // 资源ID
	ResourceType string    `gorm:"column:resource_type;index" json:"resource_type"` // 资源类型
	RoleID       uint      `gorm:"column:role_id;index" json:"role_id"`             // 角色ID
	SpaceID      *uint     `gorm:"column:space_id;index" json:"space_id,omitempty"` // 空间ID（可选，nil表示全局资源）
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// 关联关系
	Role  *Role  `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Space *Space `gorm:"foreignKey:SpaceID" json:"space,omitempty"`
}
