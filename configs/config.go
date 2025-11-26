package configs

type Config struct {
	Title               string                  `mapstructure:"title"`
	Api                 *ApiConfig              `mapstructure:"api"`
	Common              *CommonConfig           `mapstructure:"common"`
	Database            *DatabaseConfig         `mapstructure:"database"`
	Log                 *LogConfig              `mapstructure:"log"`
	ApiKey              *ApiKeyConfig           `mapstructure:"apikey"`
	User1st             *UserFirstConfig        `mapstructure:"user_1st"`
	Roles               []*RoleConfig           `mapstructure:"roles"`
	Spaces              []*SpaceConfig          `mapstructure:"spaces"`
	PermissionPolicy    *PermissionPolicyConfig `mapstructure:"permission_policy"`
	ControlPassport     *ControlPassportConfig  `mapstructure:"control_passport"`
	Banner              *BannerConfig           `mapstructure:"banner"`
	PermissionBlueprint []*PermissionTarget     `mapstructure:"permissions"`
}

func NewConfig() *Config {
	return &Config{}
}

type ApiConfig struct {
	GinMode          string `mapstructure:"gin_mode"`
	Host             string `mapstructure:"host"`
	Port             string `mapstructure:"port"`
	CorsAllowOrigins string `mapstructure:"cors_allow_origins"` // CORS 允许的域名列表，多个用逗号分隔
}

type CommonConfig struct {
	HistoryTmpDir     string `mapstructure:"history_tmp_dir"`
	HistoryTmpMaxLine int    `mapstructure:"history_tmp_max_line"`
	HistoryTmpMaxSize int    `mapstructure:"history_tmp_max_size"`
	Language          string `mapstructure:"language"`
	Port              string `mapstructure:"port"`
	Prompt            string `mapstructure:"prompt"`
}

type DatabaseConfig struct {
	CdbUrl    string `mapstructure:"cdb_url"`
	RdbPasswd string `mapstructure:"rdb_passwd"`
	RdbUrl    string `mapstructure:"rdb_url"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type ApiKeyConfig struct {
	Prefix string `mapstructure:"prefix"`
	Key    string `mapstructure:"key"`
}

type UserFirstConfig struct {
	Email     string `mapstructure:"email"`
	Name      string `mapstructure:"name"`
	Nickname  string `mapstructure:"nickname"`
	Password  string `mapstructure:"password"`
	PublicKey string `mapstructure:"public_key"`
	Username  string `mapstructure:"username"`
	Roles     string `mapstructure:"roles"`
}

type RoleConfig struct {
	Name            string                  `mapstructure:"name"`
	Desc            string                  `mapstructure:"desc"` // legacy textual format
	Description     string                  `mapstructure:"description"`
	IsDefaultSuper  bool                    `mapstructure:"is_default_super"`
	Permissions     []*RolePermissionConfig `mapstructure:"permissions"`
	PermissionScope []*RoleScopeConfig      `mapstructure:"scopes"` // optional legacy support
}

type RolePermissionConfig struct {
	Target  string               `mapstructure:"target"`
	Actions []string             `mapstructure:"actions"`
	Scope   *RolePermissionScope `mapstructure:"scope"`
}

type RolePermissionScope struct {
	Type  string `mapstructure:"type"`
	Value string `mapstructure:"value"`
}

type RoleScopeConfig struct {
	Target string `mapstructure:"target"`
	Type   string `mapstructure:"type"`
	Value  string `mapstructure:"value"`
}

type PermissionTarget struct {
	Name    string   `mapstructure:"name"`
	Actions []string `mapstructure:"actions"`
}

type BannerConfig struct {
	Show   bool   `mapstructure:"show"`
	Banner string `mapstructure:"banner"`
}

type SpaceConfig struct {
	Name        string   `mapstructure:"name"`
	Description string   `mapstructure:"description"`
	Members     []string `mapstructure:"members"`      // 用户名列表
	DefaultRole string   `mapstructure:"default_role"` // 默认空间角色
}

// PermissionPolicyConfig 权限策略配置
type PermissionPolicyConfig struct {
	// 是否启用资源角色检查
	EnableResourceRole bool `mapstructure:"enable_resource_role"`
	// 是否启用空间隔离
	EnableSpaceIsolation bool `mapstructure:"enable_space_isolation"`
	// 是否要求用户角色和资源角色完全匹配
	RequireExactRoleMatch bool `mapstructure:"require_exact_role_match"`
	// 是否允许 super 角色绕过所有限制
	SuperBypassAll bool `mapstructure:"super_bypass_all"`
	// 默认空间（全局资源所属的空间，nil表示无空间归属）
	DefaultSpace *string `mapstructure:"default_space"`
}

type ControlPassportConfig struct {
	ServiceUser  string `mapstructure:"service_user"`
	Password     string `mapstructure:"password"`
	ResourceType string `mapstructure:"resource_type"`
	PassportPub  string `mapstructure:"passport_pub"`
	Passport     string `mapstructure:"passport"`
	Description  string `mapstructure:"description"`
}
