package model

import (
	"fmt"
	"time"

	"bitrec.ai/roma/core/constants"
	"gorm.io/gorm"
)

type RouterConfig struct {
	ID          int64          `gorm:"primary_key;column:id" json:"id"`                           // 路由器配置的唯一标识，作为主键
	RouterName  string         `gorm:"type:varchar(255);column:router_name" json:"router_name"`   // 路由器名称
	WebPort     int            `gorm:"type:int(11);column:web_port" json:"web_port"`              // Web管理端口
	WebUsername string         `gorm:"type:varchar(255);column:web_username" json:"web_username"` // Web管理用户名
	WebPassword string         `gorm:"type:varchar(255);column:web_password" json:"web_password"` // Web管理密码
	Port        int            `gorm:"type:int(11);column:port" json:"port"`                      // SSH端口
	IPv4Pub     string         `gorm:"type:varchar(255);column:ipv4_pub" json:"ipv4_pub"`         // 公网IPv4地址
	IPv4Priv    string         `gorm:"type:varchar(255);column:ipv4_priv" json:"ipv4_priv"`       // 内网IPv4地址
	IPv6        string         `gorm:"type:varchar(255);column:ipv6" json:"ipv6"`                 // IPv6地址
	Password    string         `gorm:"type:varchar(255);column:password" json:"password"`         // SSH密码
	Username    string         `gorm:"type:varchar(255);column:username" json:"username"`         // SSH用户名
	PrivateKey  string         `gorm:"type:text;column:private_key" json:"private_key"`           // SSH私钥
	Description string         `gorm:"type:varchar(255);column:description" json:"description"`   // 路由器配置描述
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (r *RouterConfig) GetResource() Resource {
	return r
}

func (r *RouterConfig) GetConnect() []map[string]interface{} {
	return []map[string]interface{}{
		{constants.ConnectSSH: map[string]interface{}{
			"host":       r.IPv4Pub,
			"port":       r.Port,
			"username":   r.Username,
			"password":   r.Password,
			"privateKey": r.PrivateKey,
		}}, {constants.ConnectSSH: map[string]interface{}{
			"host":       r.IPv4Priv,
			"port":       r.Port,
			"username":   r.Username,
			"password":   r.Password,
			"privateKey": r.PrivateKey,
		}},
		{constants.ConnectSSH: map[string]interface{}{
			"host":       r.IPv6,
			"port":       r.Port,
			"username":   r.Username,
			"password":   r.Password,
			"privateKey": r.PrivateKey,
		}}, {constants.ConnectHTTP: map[string]interface{}{
			"host":     r.IPv4Pub,
			"port":     r.WebPort,
			"username": r.WebUsername,
			"password": r.WebPassword,
		}}, {constants.ConnectHTTP: map[string]interface{}{
			"host":     r.IPv4Priv,
			"port":     r.WebPort,
			"username": r.WebUsername,
			"password": r.WebPassword,
		}},
		{constants.ConnectHTTP: map[string]interface{}{
			"host":     r.IPv6,
			"port":     r.WebPort,
			"username": r.WebUsername,
			"password": r.WebPassword,
		}},
	}
}

// ID
func (r *RouterConfig) GetID() int64 {
	return r.ID
}

// Name
func (r *RouterConfig) GetName() string {
	return r.RouterName
}

func (r *RouterConfig) GetTitle() []string {
	return []string{"ID", "RouterName", "WebPort", "WebUsername", "WebPassword", "Port", "IPv4Pub", "IPv4Priv", "IPv6", "Password", "Username", "PrivateKey", "Description", "CreatedAt", "UpdatedAt"}
}

func (r *RouterConfig) GetLine() []string {
	return []string{
		fmt.Sprintf("%d", r.ID),
		r.RouterName,
		fmt.Sprintf("%d", r.WebPort),
		r.WebUsername,
		r.WebPassword,
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
