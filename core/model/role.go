package model

// Role 角色结构体
type Role struct {
	ID        uint    `gorm:"column:id;primaryKey" json:"id"`          // 角色ID，作为主键
	Name      string  `gorm:"column:name;unique;not null" json:"name"` // 角色名称，唯一且不为空
	Desc      string  `gorm:"column:desc" json:"desc"`                 // 角色描述
	CreatedAt int64   `gorm:"column:created_at" json:"created_at"`     // 创建时间
	UpdatedAt int64   `gorm:"column:updated_at" json:"updated_at"`     // 更新时间
	Users     []*User `gorm:"many2many:user_roles;" json:"users"`      // 角色与用户之间的多对多关联关系
}
