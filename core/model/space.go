package model

import "time"

// Space 空间结构体
// Space represents a workspace/namespace for resource isolation
type Space struct {
	ID          uint      `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;unique;not null;size:100" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedBy   uint      `gorm:"column:created_by;index" json:"created_by"` // 创建者ID
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// 关联关系
	Members   []*SpaceMember   `gorm:"foreignKey:SpaceID" json:"members,omitempty"`
	Resources []*ResourceSpace `gorm:"foreignKey:SpaceID" json:"resources,omitempty"`
	Creator   *User            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// SpaceMember 空间成员
type SpaceMember struct {
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	SpaceID   uint      `gorm:"column:space_id;index" json:"space_id"`
	UserID    uint      `gorm:"column:user_id;index" json:"user_id"`
	RoleID    uint      `gorm:"column:role_id;index" json:"role_id"` // 空间内角色
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// 关联关系
	Space *Space `gorm:"foreignKey:SpaceID" json:"space,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role  *Role  `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// ResourceSpace 资源空间关联
type ResourceSpace struct {
	ID           uint      `gorm:"column:id;primaryKey" json:"id"`
	SpaceID      uint      `gorm:"column:space_id;index" json:"space_id"`
	ResourceID   int64     `gorm:"column:resource_id;index" json:"resource_id"`
	ResourceType string    `gorm:"column:resource_type;index" json:"resource_type"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// 关联关系
	Space *Space `gorm:"foreignKey:SpaceID" json:"space,omitempty"`
}
