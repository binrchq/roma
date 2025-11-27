package constants

// DatabaseTypeOption 描述一个可选的数据库类型
type DatabaseTypeOption struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	DefaultPort int    `json:"default_port"`
	Description string `json:"description"`
}

// DefaultDatabaseTypes 内置数据库类型列表
var DefaultDatabaseTypes = []DatabaseTypeOption{
	{Key: "mysql", Label: "MySQL", DefaultPort: 3306, Description: "MySQL / MariaDB 兼容数据库"},
	{Key: "postgresql", Label: "PostgreSQL", DefaultPort: 5432, Description: "PostgreSQL 关系型数据库"},
	{Key: "mongodb", Label: "MongoDB", DefaultPort: 27017, Description: "MongoDB 文档型数据库"},
	{Key: "redis", Label: "Redis", DefaultPort: 6379, Description: "Redis 内存数据库"},
	{Key: "sqlserver", Label: "SQLServer", DefaultPort: 1433, Description: "Microsoft SQL Server"},
	{Key: "clickhouse", Label: "ClickHouse", DefaultPort: 9000, Description: "ClickHouse 列式数据库"},
	{Key: "elasticsearch", Label: "Elasticsearch", DefaultPort: 9200, Description: "Elasticsearch 搜索引擎"},
}
