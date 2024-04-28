package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户结构体
type User struct {
	ID        uint           `gorm:"column:id;primaryKey" json:"id"`                          // 用户的唯一标识，作为主键
	Username  string         `gorm:"column:username;unique;not null;size:50" json:"username"` // 用户名，唯一且不为空的字符串，最大长度为 50
	Name      string         `gorm:"column:name;not null;size:50" json:"name"`                // 用户姓名，不为空
	Nickname  string         `gorm:"column:nickname;not null" json:"nickname"`                // 用户昵称，不为空
	Password  string         `gorm:"column:password" json:"-"`                                // 用户密码，不为空，不在 JSON 输出中显示
	PublicKey string         `gorm:"column:public_key" json:"public_key"`                     // 用户公钥，不为空，不在 JSON 输出中显示
	Email     string         `gorm:"column:email;unique;not null" json:"email"`               // 用户邮箱，唯一且不为空
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`               // 用户状态，默认为 0
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`      // 用户创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`      // 用户更新时间
	Roles     []Role         `gorm:"many2many:user_roles;" json:"roles"`                      // 用户拥有的角色，多对多关联
}
