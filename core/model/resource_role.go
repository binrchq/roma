package model

// ResourceRole 资源角色
type ResourceRole struct {
	ResourceID   int64  `gorm:"column:resource_id;primaryKey" json:"resource_id"` // 资源ID
	ResourceType string `gorm:"column:resource_type" json:"resource_type"`        // 资源类型
	RoleID       int64  `gorm:"column:role_id" json:"role_id"`                    // 角色ID
}
