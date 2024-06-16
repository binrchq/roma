package model

import (
	"fmt"
	"time"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/types"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	ID           int64          `gorm:"primary_key;column:id" json:"id"`                             // 数据库配置的唯一标识，作为主键
	DatabaseNick string         `gorm:"type:varchar(255);column:database_nick" json:"database_nick"` // 数据库配置的名称
	DatabaseName string         `gorm:"type:varchar(255);column:database_name" json:"database_name"` // 数据库名称
	DatabaseType string         `gorm:"type:varchar(255);column:database_type" json:"database_type"` // 数据库类型（例如，'MySQL'，'PostgreSQL'）
	Port         int            `gorm:"type:int(11);column:port" json:"port"`                        // 数据库连接端口号
	IPv4Pub      string         `gorm:"type:varchar(15);column:ipv4_pub" json:"ipv4_pub"`            // 公网IPv4地址
	IPv4Priv     string         `gorm:"type:varchar(15);column:ipv4_priv" json:"ipv4_priv"`          // 内网IPv4地址
	IPv6         string         `gorm:"type:varchar(39);column:ipv6" json:"ipv6"`                    // IPv6地址
	Password     string         `gorm:"type:varchar(255);column:password" json:"password"`           // 数据库认证密码
	Username     string         `gorm:"type:varchar(255);column:username" json:"username"`           // 数据库认证用户名
	PrivateKey   string         `gorm:"type:varchar(1024);column:private_key" json:"private_key"`    // 数据库认证私钥
	Description  string         `gorm:"type:varchar(1024);column:description" json:"description"`    // 数据库配置描述
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (r *DatabaseConfig) GetResource() Resource {
	return r
}

func (r *DatabaseConfig) GetConnect() []*types.Connection {
	connection := []*types.Connection{}
	connection = append(connection, types.NewConnection(constants.ConnectDatabase, r.IPv4Pub, r.Port, r.Username, r.Password, r.PrivateKey, r.DatabaseType, r.DatabaseName))
	connection = append(connection, types.NewConnection(constants.ConnectDatabase, r.IPv4Priv, r.Port, r.Username, r.Password, r.PrivateKey, r.DatabaseType, r.DatabaseName))
	connection = append(connection, types.NewConnection(constants.ConnectDatabase, r.IPv6, r.Port, r.Username, r.Password, r.PrivateKey, r.DatabaseType, r.DatabaseName))
	return connection
}

func (r *DatabaseConfig) GetID() int64 {
	return r.ID
}

func (r *DatabaseConfig) GetName() string {
	return r.DatabaseNick
}

func (r *DatabaseConfig) GetTitle() []string {
	return []string{"ID", "DatabaseNick", "DatabaseName", "DatabaseType", "Port", "IPv4Pub", "IPv4Priv", "IPv6", "Password", "Username", "PrivateKey", "Description", "CreatedAt", "UpdatedAt"}
}

func (r *DatabaseConfig) GetLine() []string {
	return []string{
		fmt.Sprintf("%d", r.ID),
		r.DatabaseNick,
		r.DatabaseName,
		r.DatabaseType,
		fmt.Sprintf("%d", r.Port),
		r.IPv4Pub,
		r.IPv4Priv,
		r.IPv6,
		r.Password,
		r.Username,
		r.PrivateKey,
		r.Description,
		r.CreatedAt.String(),
		r.UpdatedAt.String(),
	}
}
