package connector

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/utils"
	"binrc.com/roma/core/utils/logger"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseConnector 数据库连接器
type DatabaseConnector struct {
	Config *model.DatabaseConfig
}

// NewDatabaseConnector 创建数据库连接器
func NewDatabaseConnector(config *model.DatabaseConfig) *DatabaseConnector {
	return &DatabaseConnector{
		Config: config,
	}
}

// resolveHostForConnection 用途: 解析数据库主机地址（支持域名和IP）
// 输入: 无（使用配置中的地址）
// 输出: string - 用于连接的主机地址
// 必要性: 允许资源配置中填写域名，但连接时需要解析为IP
func (d *DatabaseConnector) resolveHostForConnection() string {
	candidates := []string{d.Config.IPv4Pub, d.Config.IPv4Priv, d.Config.IPv6}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if resolved, err := utils.ResolveHostName(candidate); err != nil {
			logger.Logger.Warning(fmt.Sprintf("Resolve host failed for %s: %v", candidate, err))
			return candidate
		} else if resolved != "" {
			if resolved != candidate {
				logger.Logger.Debug(fmt.Sprintf("Resolve host %s -> %s", candidate, resolved))
			}
			return resolved
		}
	}
	return ""
}

// displayHost 返回用于展示的原始主机值
func (d *DatabaseConnector) displayHost() string {
	candidates := []string{d.Config.IPv4Pub, d.Config.IPv4Priv, d.Config.IPv6}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

// Connect 连接数据库
func (d *DatabaseConnector) Connect() (interface{}, error) {
	dbType := strings.ToLower(d.Config.DatabaseType)

	switch dbType {
	case "mysql":
		return d.connectMySQL()
	case "postgresql", "postgres":
		return d.connectPostgreSQL()
	case "mongodb", "mongo":
		return d.connectMongoDB()
	case "redis":
		return d.connectRedis()
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", d.Config.DatabaseType)
	}
}

// MySQL 连接
func (d *DatabaseConnector) connectMySQL() (*sql.DB, error) {
	host := d.resolveHostForConnection()
	if host == "" {
		return nil, fmt.Errorf("缺少数据库连接地址")
	}

	// 解密密码
	decryptedPassword, err := utils.DecryptPassword(d.Config.Password)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %v", err)
	}
	// DSN格式: username:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.Config.Username,
		decryptedPassword,
		host,
		d.Config.Port,
		d.Config.DatabaseName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("MySQL 连接测试失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// PostgreSQL 连接
func (d *DatabaseConnector) connectPostgreSQL() (*sql.DB, error) {
	host := d.resolveHostForConnection()
	if host == "" {
		return nil, fmt.Errorf("缺少数据库连接地址")
	}

	// 解密密码
	decryptedPassword, err := utils.DecryptPassword(d.Config.Password)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %v", err)
	}
	// DSN格式: postgres://username:password@host:port/database?sslmode=disable
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.Config.Username,
		decryptedPassword,
		host,
		d.Config.Port,
		d.Config.DatabaseName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接 PostgreSQL 失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("PostgreSQL 连接测试失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// MongoDB 连接
func (d *DatabaseConnector) connectMongoDB() (*mongo.Client, error) {
	host := d.resolveHostForConnection()
	if host == "" {
		return nil, fmt.Errorf("缺少数据库连接地址")
	}

	// 解密密码
	decryptedPassword, err := utils.DecryptPassword(d.Config.Password)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %v", err)
	}
	// MongoDB URI格式
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		d.Config.Username,
		decryptedPassword,
		host,
		d.Config.Port,
		d.Config.DatabaseName,
	)

	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("连接 MongoDB 失败: %v", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("MongoDB 连接测试失败: %v", err)
	}

	return client, nil
}

// Redis 连接
func (d *DatabaseConnector) connectRedis() (interface{}, error) {
	host := d.resolveHostForConnection()
	if host == "" {
		return nil, fmt.Errorf("缺少数据库连接地址")
	}

	// 解密密码
	decryptedPassword, err := utils.DecryptPassword(d.Config.Password)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %v", err)
	}
	return map[string]interface{}{
		"type":     "redis",
		"host":     host,
		"port":     d.Config.Port,
		"password": decryptedPassword,
		"message":  "Redis 连接需要使用 redis-cli 或客户端库",
	}, nil
}

// ExecuteQuery 执行数据库查询
func (d *DatabaseConnector) ExecuteQuery(query string) (interface{}, error) {
	conn, err := d.Connect()
	if err != nil {
		return nil, err
	}

	dbType := strings.ToLower(d.Config.DatabaseType)

	switch dbType {
	case "mysql", "postgresql", "postgres":
		db := conn.(*sql.DB)
		defer db.Close()

		rows, err := db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %v", err)
		}
		defer rows.Close()

		// 获取列名
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		// 读取数据
		var results []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			rowMap := make(map[string]interface{})
			for i, col := range columns {
				rowMap[col] = values[i]
			}
			results = append(results, rowMap)
		}

		return results, nil

	case "mongodb", "mongo":
		client := conn.(*mongo.Client)
		defer client.Disconnect(context.Background())

		// MongoDB 查询需要特殊处理
		return map[string]string{
			"message": "MongoDB 查询请使用 MongoDB 查询语法",
			"example": "db.collection.find({})",
		}, nil

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", d.Config.DatabaseType)
	}
}

// GetConnectionInfo 获取连接信息（用于显示）
func (d *DatabaseConnector) GetConnectionInfo() map[string]interface{} {
	host := d.displayHost()

	var connectionString string
	dbType := strings.ToLower(d.Config.DatabaseType)

	switch dbType {
	case "mysql":
		connectionString = fmt.Sprintf("mysql -h %s -P %d -u %s -p %s",
			host, d.Config.Port, d.Config.Username, d.Config.DatabaseName)
	case "postgresql", "postgres":
		connectionString = fmt.Sprintf("psql -h %s -p %d -U %s -d %s",
			host, d.Config.Port, d.Config.Username, d.Config.DatabaseName)
	case "mongodb", "mongo":
		connectionString = fmt.Sprintf("mongo --host %s --port %d -u %s -p *** %s",
			host, d.Config.Port, d.Config.Username, d.Config.DatabaseName)
	case "redis":
		connectionString = fmt.Sprintf("redis-cli -h %s -p %d -a ***",
			host, d.Config.Port)
	}

	return map[string]interface{}{
		"type":              d.Config.DatabaseType,
		"name":              d.Config.DatabaseNick,
		"database":          d.Config.DatabaseName,
		"host":              host,
		"port":              d.Config.Port,
		"username":          d.Config.Username,
		"connection_string": connectionString,
	}
}
