package configs

type Config struct {
	Title           string                 `mapstructure:"title"`
	Api             *ApiConfig             `mapstructure:"api"`
	Common          *CommonConfig          `mapstructure:"common"`
	Database        *DatabaseConfig        `mapstructure:"database"`
	Log             *LogConfig             `mapstructure:"log"`
	ApiKey          *ApiKeyConfig          `mapstructure:"apikey"`
	User1st         *UserFirstConfig       `mapstructure:"user_1st"`
	Roles           []*RoleConfig          `mapstructure:"roles"`
	ControlPassport *ControlPassportConfig `mapstructure:"control_passport"`
	Banner          *BannerConfig          `mapstructure:"banner"`
}

func NewConfig() *Config {
	return &Config{}
}

type ApiConfig struct {
	GinMode string `mapstructure:"gin_mode"`
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
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
	Name string `mapstructure:"name"`
	Desc string `mapstructure:"desc"`
}

type BannerConfig struct {
	Show   bool   `mapstructure:"show"`
	Banner string `mapstructure:"banner"`
}

type ControlPassportConfig struct {
	ServiceUser  string `mapstructure:"service_user"`
	Password     string `mapstructure:"password"`
	ResourceType string `mapstructure:"resource_type"`
	PassportPub  string `mapstructure:"passport_pub"`
	Passport     string `mapstructure:"passport"`
	Description  string `mapstructure:"description"`
}
