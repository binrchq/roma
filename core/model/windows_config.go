package model

import (
	"fmt"
	"time"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/types"
	"gorm.io/gorm"
)

// WindowsConfig 配置文件
type WindowsConfig struct {
	ID          int64          `gorm:"primary_key;column:id" json:"id"`                         // Windows配置的唯一标识，作为主键
	Hostname    string         `gorm:"type:varchar(255);column:hostname" json:"hostname"`       // Windows主机名
	Port        int            `gorm:"type:int(11);column:port" json:"port"`                    // RDP端口
	IPv4Pub     string         `gorm:"type:varchar(255);column:ipv4_pub" json:"ipv4_pub"`       // 公网IPv4地址
	IPv4Priv    string         `gorm:"type:varchar(255);column:ipv4_priv" json:"ipv4_priv"`     // 内网IPv4地址
	IPv6        string         `gorm:"type:varchar(255);column:ipv6" json:"ipv6"`               // IPv6地址
	PortIPv6    int            `gorm:"type:int(11);column:port_ipv6" json:"port_ipv6"`          // RDP端口IPv6
	Password    string         `gorm:"type:varchar(255);column:password" json:"password"`       // RDP密码
	Username    string         `gorm:"type:varchar(255);column:username" json:"username"`       // RDP用户名
	Description string         `gorm:"type:varchar(255);column:description" json:"description"` // Windows配置描述
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (w *WindowsConfig) GetResource() Resource {
	return w
}

func (w *WindowsConfig) GetConnect() []*types.Connection {
	connection := []*types.Connection{}
	connection = append(connection, types.NewConnection(constants.ConnectRDP, w.IPv4Priv, w.Port, w.Username, w.Password))
	connection = append(connection, types.NewConnection(constants.ConnectVNC, w.IPv4Priv, w.Port, w.Username, w.Password))
	connection = append(connection, types.NewConnection(constants.ConnectSSH, w.IPv4Priv, w.Port, w.Username, w.Password))
	return connection
}

// ID
func (w *WindowsConfig) GetID() int64 {

	return w.ID
}

// Name
func (w *WindowsConfig) GetName() string {

	return w.Hostname
}
func (w *WindowsConfig) GetTitle() []string {
	return []string{"ID", "Hostname", "Port", "IPv4Pub", "IPv4Priv", "IPv6", "PortIPv6", "Password", "Username", "Description", "CreatedAt", "UpdatedAt"}
}

func (w *WindowsConfig) GetLine() []string {
	return []string{
		fmt.Sprintf("%d", w.ID),
		w.Hostname,
		fmt.Sprintf("%d", w.Port),
		w.IPv4Pub,
		w.IPv4Priv,
		w.IPv6,
		fmt.Sprintf("%d", w.PortIPv6),
		w.Password,
		w.Username,
		w.Description,
		w.CreatedAt.String(),
		w.UpdatedAt.String(),
	}
}
