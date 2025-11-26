package initialize

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var cdb *gorm.DB

func InitCDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Determine the database type based on the URL
	cdbUrl := global.CONFIG.Database.CdbUrl
	if strings.Contains(cdbUrl, ".db") {
		db, err = initSQLite()
	} else if strings.Contains(cdbUrl, "postgres://") || strings.Contains(cdbUrl, "host=") {
		db, err = initPostgreSQL()
	} else {
		db, err = initMySQL()
	}

	if err != nil {
		return nil, err
	}

	if err := migrateTables(db, &model.HostKey{}, &model.User{}, &model.Passport{}, &model.Role{}, &model.Apikey{}, &model.LinuxConfig{}, &model.WindowsConfig{}, &model.DatabaseConfig{}, &model.RouterConfig{}, &model.SwitchConfig{}, &model.ResourceRole{}, &model.Space{}, &model.SpaceMember{}, &model.ResourceSpace{}, &model.Tag{}, &model.CredentialAccessLog{}, &model.AccessLog{}, &model.DockerConfig{}, &model.AuditLog{}); err != nil {
		return nil, err
	}

	cdb = db
	return cdb, nil
}

func initSQLite() (*gorm.DB, error) {
	if _, err := os.Stat(global.CONFIG.Database.CdbUrl); os.IsNotExist(err) {
		directory := filepath.Dir(global.CONFIG.Database.CdbUrl)
		if err := os.MkdirAll(directory, 0755); err != nil {
			// 创建目录失败，输出错误信息并退出程序
			log.Error().Err(err).Msgf("无法创建数据库目录")
			os.Exit(1)
		}
		if _, err := os.Create(global.CONFIG.Database.CdbUrl); err != nil {
			return nil, fmt.Errorf("failed to create SQLite database: %s", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(global.CONFIG.Database.CdbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %s", err)
	}
	return db, nil
}

func initMySQL() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(global.CONFIG.Database.CdbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		if strings.Contains(err.Error(), "Unknown database") || strings.Contains(err.Error(), "1049") {
			if err := createMySQLDatabase(global.CONFIG.Database.CdbUrl); err != nil {
				return nil, fmt.Errorf("failed to create MySQL database: %s", err)
			}
			db, err = gorm.Open(mysql.Open(global.CONFIG.Database.CdbUrl), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to open MySQL database after creation: %s", err)
			}
		} else {
			return nil, fmt.Errorf("failed to open MySQL database: %s", err)
		}
	}
	return db, nil
}

func initPostgreSQL() (*gorm.DB, error) {
	cdbUrl := global.CONFIG.Database.CdbUrl

	// 验证连接字符串是否完整
	if cdbUrl == "" {
		return nil, fmt.Errorf("PostgreSQL connection string is empty")
	}

	// 记录原始连接字符串（隐藏密码）
	log.Info().Str("original_url", maskPassword(cdbUrl)).Msg("Original PostgreSQL connection string")

	// 构建 DSN，如果没有 sslmode 则默认添加 sslmode=disable
	dsn := buildPostgresDSN(cdbUrl)

	// 验证 DSN 是否包含必要的参数
	if !isValidPostgresDSN(dsn) {
		log.Error().Str("dsn", maskPassword(dsn)).Msg("PostgreSQL DSN is incomplete, missing required parameters (host, port, user, password)")
		return nil, fmt.Errorf("PostgreSQL connection string is incomplete, missing required parameters")
	}

	// 验证 URL 格式是否包含密码
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsedURL, err := url.Parse(dsn)
		if err == nil && parsedURL.User != nil {
			_, hasPassword := parsedURL.User.Password()
			if !hasPassword {
				log.Error().Str("dsn", maskPassword(dsn)).Msg("PostgreSQL connection string is missing password")
				return nil, fmt.Errorf("PostgreSQL connection string is missing password")
			}
		}
	}

	log.Info().Str("dsn", maskPassword(dsn)).Msg("Attempting to connect to PostgreSQL database")

	// 先尝试直接连接到目标数据库
	// 对于 Pgpool 代理，添加连接超时和重试机制
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now()
		},
	})

	// 如果连接成功，配置连接池参数
	if err == nil {
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			// 设置连接池参数，适配 Pgpool 代理
			sqlDB.SetMaxOpenConns(10)
			sqlDB.SetMaxIdleConns(5)
			sqlDB.SetConnMaxLifetime(time.Hour)
			sqlDB.SetConnMaxIdleTime(time.Minute * 10)

			// 测试连接
			if pingErr := sqlDB.Ping(); pingErr != nil {
				log.Error().Err(pingErr).Msg("PostgreSQL connection ping failed")
				return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", pingErr)
			}
		}
	}

	// 如果连接失败，记录详细错误信息
	if err != nil {
		log.Error().Err(err).Str("dsn", maskPassword(dsn)).Msg("Failed to connect to PostgreSQL database")
	}

	// 如果连接失败且指定了数据库名，尝试创建数据库
	if err != nil {
		// 从连接字符串中提取数据库名
		dbName := extractDatabaseName(cdbUrl)
		if dbName != "" {
			// 连接到 postgres 默认数据库
			dsnWithoutDB := buildPostgresDSNWithoutDatabase(cdbUrl)
			tempDB, tempErr := gorm.Open(postgres.Open(dsnWithoutDB), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
			if tempErr != nil {
				// 如果连默认数据库都连不上，返回原始错误
				return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
			}

			// 检查数据库是否存在，如果不存在则创建
			var count int64
			checkDB := "SELECT COUNT(*) FROM pg_database WHERE datname = $1"
			if tempErr := tempDB.Raw(checkDB, dbName).Scan(&count).Error; tempErr != nil {
				sqlTempDB, _ := tempDB.DB()
				sqlTempDB.Close()
				return nil, fmt.Errorf("failed to check database existence: %w", tempErr)
			}

			if count == 0 {
				// 创建数据库
				createDB := fmt.Sprintf("CREATE DATABASE %s", dbName)
				if tempErr := tempDB.Exec(createDB).Error; tempErr != nil {
					sqlTempDB, _ := tempDB.DB()
					sqlTempDB.Close()
					return nil, fmt.Errorf("failed to create database %s: %w", dbName, tempErr)
				}
				log.Info().Msgf("PostgreSQL database '%s' created successfully", dbName)
			}

			// 关闭临时连接
			sqlTempDB, _ := tempDB.DB()
			sqlTempDB.Close()

			// 重新尝试连接到目标数据库
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
		}
	}

	return db, nil
}

// buildPostgresDSN 构建 PostgreSQL DSN，如果没有 sslmode 则默认添加 sslmode=disable
// 对于 Pgpool 代理，添加 connect_timeout 参数以提高连接稳定性
func buildPostgresDSN(cdbUrl string) string {
	hasSSLMode := strings.Contains(cdbUrl, "sslmode=")
	hasConnectTimeout := strings.Contains(cdbUrl, "connect_timeout")

	// 如果连接字符串已经包含 sslmode 和 connect_timeout，直接返回
	if hasSSLMode && hasConnectTimeout {
		return cdbUrl
	}

	// 构建参数字符串
	params := []string{}
	if !hasSSLMode {
		params = append(params, "sslmode=disable")
	}
	if !hasConnectTimeout {
		params = append(params, "connect_timeout=10")
	}
	paramStr := strings.Join(params, "&")

	// 添加参数
	if strings.HasPrefix(cdbUrl, "postgres://") || strings.HasPrefix(cdbUrl, "postgresql://") {
		// URL 格式：直接字符串拼接，避免 url.Parse 可能丢失密码的问题
		if strings.Contains(cdbUrl, "?") {
			return cdbUrl + "&" + paramStr
		}
		return cdbUrl + "?" + paramStr
	} else if strings.Contains(cdbUrl, "host=") {
		// 连接字符串格式
		return cdbUrl + " " + strings.ReplaceAll(paramStr, "&", " ")
	}
	// 如果格式无法识别，直接追加
	return cdbUrl + " " + strings.ReplaceAll(paramStr, "&", " ")
}

// buildPostgresDSNWithoutDatabase 构建不包含数据库名的 PostgreSQL DSN（用于创建数据库）
func buildPostgresDSNWithoutDatabase(cdbUrl string) string {
	var user, password, host, port string

	if strings.HasPrefix(cdbUrl, "postgres://") || strings.HasPrefix(cdbUrl, "postgresql://") {
		parsedURL, err := url.Parse(cdbUrl)
		if err != nil {
			// 解析失败，尝试添加 sslmode 后返回
			return buildPostgresDSN(cdbUrl)
		}

		user = parsedURL.User.Username()
		password, _ = parsedURL.User.Password()
		host = parsedURL.Hostname()
		port = parsedURL.Port()
		if port == "" {
			port = "5432"
		}

		// 构建连接到 postgres 默认数据库的 URL
		if password != "" {
			return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", user, password, host, port)
		}
		return fmt.Sprintf("postgres://%s@%s:%s/postgres?sslmode=disable", user, host, port)
	} else {
		// 连接字符串格式
		re := regexp.MustCompile(`host=([^\s]+)`)
		if matches := re.FindStringSubmatch(cdbUrl); len(matches) > 1 {
			host = matches[1]
		}

		re = regexp.MustCompile(`port=(\d+)`)
		if matches := re.FindStringSubmatch(cdbUrl); len(matches) > 1 {
			port = matches[1]
		} else {
			port = "5432"
		}

		re = regexp.MustCompile(`user=([^\s]+)`)
		if matches := re.FindStringSubmatch(cdbUrl); len(matches) > 1 {
			user = matches[1]
		}

		re = regexp.MustCompile(`password=([^\s]+)`)
		if matches := re.FindStringSubmatch(cdbUrl); len(matches) > 1 {
			password = matches[1]
		}

		// 构建连接到 postgres 默认数据库的连接字符串
		dsn := fmt.Sprintf("host=%s port=%s user=%s", host, port, user)
		if password != "" {
			dsn += fmt.Sprintf(" password=%s", password)
		}
		dsn += " dbname=postgres sslmode=disable"
		return dsn
	}
}

// extractDatabaseName 从连接字符串中提取数据库名
func extractDatabaseName(cdbUrl string) string {
	if strings.HasPrefix(cdbUrl, "postgres://") || strings.HasPrefix(cdbUrl, "postgresql://") {
		parsedURL, err := url.Parse(cdbUrl)
		if err != nil {
			return ""
		}
		return strings.TrimPrefix(parsedURL.Path, "/")
	} else {
		re := regexp.MustCompile(`dbname=([^\s]+)`)
		if matches := re.FindStringSubmatch(cdbUrl); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// isValidPostgresDSN 验证 PostgreSQL DSN 是否包含必要的参数
func isValidPostgresDSN(dsn string) bool {
	// URL 格式检查
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsedURL, err := url.Parse(dsn)
		if err != nil {
			return false
		}
		// 检查是否有 host 和 user
		if parsedURL.Hostname() == "" || parsedURL.User == nil || parsedURL.User.Username() == "" {
			return false
		}
		return true
	}

	// 连接字符串格式检查
	hasHost := strings.Contains(dsn, "host=")
	hasUser := strings.Contains(dsn, "user=")

	// host 和 user 是必需的，password 可能为空（如果使用 trust 认证）
	return hasHost && hasUser
}

// maskPassword 隐藏连接字符串中的密码
func maskPassword(dsn string) string {
	// URL 格式
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsedURL, err := url.Parse(dsn)
		if err != nil {
			return "***"
		}
		if parsedURL.User != nil {
			username := parsedURL.User.Username()
			parsedURL.User = url.User(username)
			return parsedURL.String()
		}
		return dsn
	}

	// 连接字符串格式
	re := regexp.MustCompile(`password=([^\s]+)`)
	return re.ReplaceAllString(dsn, "password=***")
}

func getTableName(model interface{}) string {
	// 通过反射获取结构体的类型
	t := reflect.TypeOf(model)
	// 如果是指针类型，取其指向的类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// 返回结构体类型的名称作为表名
	return t.Name()
}

func migrateTables(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		tableName := getTableName(model)
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate table[%s]: %s", tableName, err)
		}
	}
	return nil
}

func createMySQLDatabase(dsn string) error {
	re := regexp.MustCompile(`^([^:]+):([^@]+)@tcp\(([^:]+):(\d+)\)/([^?]+)`)
	matches := re.FindStringSubmatch(dsn)
	if len(matches) != 6 {
		return fmt.Errorf("invalid MySQL DSN format")
	}

	user := matches[1]
	password := matches[2]
	host := matches[3]
	port := matches[4]
	dbName := matches[5]

	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %s", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		return fmt.Errorf("failed to create database: %s", err)
	}

	log.Info().Msgf("MySQL database '%s' created successfully", dbName)
	return nil
}

func createPostgreSQLDatabase(dsn string) error {
	var user, password, host, port, dbName string

	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsedURL, err := url.Parse(dsn)
		if err != nil {
			return fmt.Errorf("invalid PostgreSQL URL format: %s", err)
		}

		user = parsedURL.User.Username()
		password, _ = parsedURL.User.Password()
		host = parsedURL.Hostname()
		port = parsedURL.Port()
		if port == "" {
			port = "5432"
		}
		dbName = strings.TrimPrefix(parsedURL.Path, "/")
	} else {
		re := regexp.MustCompile(`host=([^\s]+)`)
		if matches := re.FindStringSubmatch(dsn); len(matches) > 1 {
			host = matches[1]
		}

		re = regexp.MustCompile(`port=(\d+)`)
		if matches := re.FindStringSubmatch(dsn); len(matches) > 1 {
			port = matches[1]
		} else {
			port = "5432"
		}

		re = regexp.MustCompile(`user=([^\s]+)`)
		if matches := re.FindStringSubmatch(dsn); len(matches) > 1 {
			user = matches[1]
		}

		re = regexp.MustCompile(`password=([^\s]+)`)
		if matches := re.FindStringSubmatch(dsn); len(matches) > 1 {
			password = matches[1]
		}

		re = regexp.MustCompile(`dbname=([^\s]+)`)
		if matches := re.FindStringSubmatch(dsn); len(matches) > 1 {
			dbName = matches[1]
		}
	}

	if dbName == "" {
		return fmt.Errorf("database name not found in DSN")
	}

	serverDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)
	db, err := sql.Open("postgres", serverDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %s", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Info().Msgf("PostgreSQL database '%s' already exists", dbName)
			return nil
		}
		return fmt.Errorf("failed to create database: %s", err)
	}

	log.Info().Msgf("PostgreSQL database '%s' created successfully", dbName)
	return nil
}
