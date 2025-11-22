package model

import (
	"fmt"
	"time"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/types"
	"gorm.io/gorm"
)

type SwitchConfig struct {
	ID          int64          `gorm:"primary_key;column:id" json:"id"`                         // 交换机配置的唯一标识，作为主键
	SwitchName  string         `gorm:"type:varchar(255);column:switch_name" json:"switch_name"` // 交换机名称
	Port        int            `gorm:"type:int(11);column:port" json:"port"`                    // SSH端口
	IPv4Pub     string         `gorm:"type:varchar(255);column:ipv4_pub" json:"ipv4_pub"`       // 公网IPv4地址
	PortActual  int            `gorm:"type:int(11);column:port_actual" json:"port_actual"`      // 实际SSH端口ipv4"
	IPv4Priv    string         `gorm:"type:varchar(255);column:ipv4_priv" json:"ipv4_priv"`     // 内网IPv4地址
	IPv6        string         `gorm:"type:varchar(255);column:ipv6" json:"ipv6"`               // IPv6地址
	PortIPv6    int            `gorm:"type:int(11);column:port_ipv6" json:"port_ipv6"`          // SSH端口IPv6
	Password    string         `gorm:"type:varchar(255);column:password" json:"password"`       // SSH密码
	Username    string         `gorm:"type:varchar(255);column:username" json:"username"`       // SSH用户名
	Description string         `gorm:"type:varchar(255);column:description" json:"description"` // 交换机配置描述
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (r *SwitchConfig) GetResource() Resource {
	return r
}

func (r *SwitchConfig) GetConnect() []*types.Connection {
	connection := []*types.Connection{}
	connection = append(connection, types.NewConnection(constants.ConnectSSH, r.IPv4Pub, r.Port, r.Username, r.Password))
	connection = append(connection, types.NewConnection(constants.ConnectSSH, r.IPv4Priv, r.PortActual, r.Username, r.Password))
	connection = append(connection, types.NewConnection(constants.ConnectSSH, r.IPv6, r.PortIPv6, r.Username, r.Password))
	return connection
}

// ID
func (r *SwitchConfig) GetID() int64 {
	return r.ID
}

// Name
func (r *SwitchConfig) GetName() string {
	return r.SwitchName
}

func (r *SwitchConfig) GetTitle() []string {
	return []string{"ID", "SwitchName", "Port", "IPv4Pub", "PortActual", "IPv4Priv", "IPv6", "PortIPv6", "Password", "Username", "Description", "CreatedAt", "UpdatedAt"}
}

func (r *SwitchConfig) GetLine() []string {
	return []string{
		fmt.Sprintf("%d", r.ID),
		r.SwitchName,
		fmt.Sprintf("%d", r.Port),
		r.IPv4Pub,
		fmt.Sprintf("%d", r.PortActual),
		r.IPv4Priv,
		r.IPv6,
		fmt.Sprintf("%d", r.PortIPv6),
		r.Password,
		r.Username,
		r.Description,
		r.CreatedAt.String(),
		r.UpdatedAt.String(),
	}
}
