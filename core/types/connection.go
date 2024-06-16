package types

// Connection 结构体表示连接配置
type Connection struct {
	Type         string // 连接类型，使用连接类型的常量
	Host         string // 主机地址
	Port         int    // 端口号
	Username     string // 用户名
	Password     string // 密码
	PrivateKey   string // 私钥路径或内容
	DatabaseType string // 数据库类型（仅对数据库连接有效）
	Database     string // 数据库名称（仅对数据库连接有效）
	Timeout      int    // 连接超时时间（单位：秒）
	UseSSL       bool   // 是否使用SSL/TLS加密
	MaxRetries   int    // 最大重试次数
	Compression  bool   // 是否启用压缩
	Certificate  string // 证书路径（仅对HTTPS连接有效）
	ProxyAddress string // 代理服务器地址（仅对HTTP连接有效）
}

// NewConnection 创建一个新的连接配置实例
func NewConnection(connType, host string, port int, username, password string, options ...string) *Connection {
	var privateKey, databaseType, database string

	if len(options) > 0 {
		privateKey = options[0]
	}
	if len(options) > 1 {
		databaseType = options[1]
	}
	if len(options) > 2 {
		database = options[2]
	}
	return &Connection{
		Type:         connType,
		Host:         host,
		Port:         port,
		Username:     username,
		Password:     password,
		PrivateKey:   privateKey,
		DatabaseType: databaseType,
		Database:     database,
	}
}
