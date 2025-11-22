package model

import (
	"fmt"
	"time"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/types"
	"gorm.io/gorm"
)

type DockerConfig struct {
	ID            int64          `gorm:"primary_key;column:id" json:"id"`                          // Linux配置的唯一标识，作为主键
	ContainerName string         `gorm:"type:varchar(255);column:ContainerName" json:"hostname"`   // Linux机器的主机名
	Port          int            `gorm:"type:int(11);column:port" json:"port"`                     // SSH端口号
	IPv4Priv      string         `gorm:"type:varchar(255);column:ipv4_priv" json:"ipv4_priv"`      // 内网IPv4地址
	IPv6          string         `gorm:"type:varchar(255);column:ipv6" json:"ipv6"`                // IPv6地址
	PortIPv6      int            `gorm:"type:int(11);column:port_ipv6" json:"port_ipv6"`           // IPv6连接的SSH端口号
	Password      string         `gorm:"type:varchar(255);column:password" json:"password"`        // SSH身份验证密码
	Username      string         `gorm:"type:varchar(255);column:username" json:"username"`        // SSH身份验证用户名
	PrivateKey    string         `gorm:"type:text;column:private_key" json:"private_key"`          // SSH身份验证私钥
	Description   string         `gorm:"type:varchar(1024);column:description" json:"description"` // Linux配置描述
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (r *DockerConfig) GetResource() Resource {
	return r
}

func (r *DockerConfig) GetConnect() []*types.Connection {
	connection := []*types.Connection{}
	connection = append(connection, types.NewConnection(constants.ConnectSSH, r.IPv4Priv, r.Port, r.Username, r.Password, r.PrivateKey))
	connection = append(connection, types.NewConnection(constants.ConnectSSH, r.IPv6, r.PortIPv6, r.Username, r.Password, r.PrivateKey))
	return connection
}
func (r *DockerConfig) GetID() int64 {
	return r.ID
}

// Name
func (r *DockerConfig) GetName() string {
	return r.ContainerName
}

func (r *DockerConfig) GetTitle() []string {
	return []string{"ID", "ContainerName", "Port", "IPv4Priv", "IPv6", "PortIPv6", "Username", "Description", "CreatedAt", "UpdatedAt"}
}

func (r *DockerConfig) GetLine() []string {
	return []string{
		fmt.Sprintf("%d", r.ID),
		r.ContainerName,
		fmt.Sprintf("%d", r.Port),
		r.IPv4Priv,
		r.IPv6,
		fmt.Sprintf("%d", r.PortIPv6),
		r.Username,
		r.Description,
		r.CreatedAt.String(),
		r.UpdatedAt.String(),
	}
}
