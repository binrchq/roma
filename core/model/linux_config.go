package model

import (
	"fmt"
	"time"

	"bitrec.ai/roma/core/constants"
	"gorm.io/gorm"
)

type LinuxConfig struct {
	ID          int64          `gorm:"primary_key;column:id" json:"id"`                          // Linux配置的唯一标识，作为主键
	Hostname    string         `gorm:"type:varchar(255);column:hostname" json:"hostname"`        // Linux机器的主机名
	Port        int            `gorm:"type:int(11);column:port" json:"port"`                     // SSH端口号
	IPv4Pub     string         `gorm:"type:varchar(255);column:ipv4_pub" json:"ipv4_pub"`        // 公网IPv4地址
	PortActual  int            `gorm:"type:int(11);column:port_actual" json:"port_actual"`       // 实际使用的SSH端口号
	IPv4Priv    string         `gorm:"type:varchar(255);column:ipv4_priv" json:"ipv4_priv"`      // 内网IPv4地址
	IPv6        string         `gorm:"type:varchar(255);column:ipv6" json:"ipv6"`                // IPv6地址
	PortIPv6    int            `gorm:"type:int(11);column:port_ipv6" json:"port_ipv6"`           // IPv6连接的SSH端口号
	Password    string         `gorm:"type:varchar(255);column:password" json:"password"`        // SSH身份验证密码
	Username    string         `gorm:"type:varchar(255);column:username" json:"username"`        // SSH身份验证用户名
	PrivateKey  string         `gorm:"type:text;column:private_key" json:"private_key"`          // SSH身份验证私钥
	Description string         `gorm:"type:varchar(1024);column:description" json:"description"` // Linux配置描述
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (r *LinuxConfig) GetResource() Resource {
	return r
}

func (l *LinuxConfig) GetConnect() []map[string]interface{} {
	return []map[string]interface{}{
		{constants.ConnectSSH: map[string]interface{}{
			"host":       l.IPv4Pub,
			"port":       l.Port,
			"username":   l.Username,
			"password":   l.Password,
			"privateKey": l.PrivateKey,
		}}, {constants.ConnectSSH: map[string]interface{}{
			"host":       l.IPv4Pub,
			"port":       l.PortActual,
			"username":   l.Username,
			"password":   l.Password,
			"privateKey": l.PrivateKey,
		}},
		{constants.ConnectSSH: map[string]interface{}{
			"host":       l.IPv4Priv,
			"port":       l.PortActual,
			"username":   l.Username,
			"password":   l.Password,
			"privateKey": l.PrivateKey,
		}},
		{constants.ConnectSSH: map[string]interface{}{
			"host":       l.IPv6,
			"port":       l.PortIPv6,
			"username":   l.Username,
			"password":   l.Password,
			"privateKey": l.PrivateKey,
		}},
	}
}

// ID
func (r *LinuxConfig) GetID() int64 {
	return r.ID
}

// Name
func (r *LinuxConfig) GetName() string {
	return r.Hostname
}

func (r *LinuxConfig) GetTitle() []string {
	return []string{"ID", "Hostname", "Port", "IPv4Pub", "PortActual", "IPv4Priv", "IPv6", "PortIPv6", "Username", "Description", "CreatedAt", "UpdatedAt"}
}

func (r *LinuxConfig) GetLine() []string {
	return []string{
		fmt.Sprintf("%d", r.ID),
		r.Hostname,
		fmt.Sprintf("%d", r.Port),
		r.IPv4Pub,
		fmt.Sprintf("%d", r.PortActual),
		r.IPv4Priv,
		r.IPv6,
		fmt.Sprintf("%d", r.PortIPv6),
		r.Username,
		r.Description,
		r.CreatedAt.String(),
		r.UpdatedAt.String(),
	}
}
