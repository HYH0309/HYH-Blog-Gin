package config

// PostgresConfig 定义 PostgreSQL 数据库连接配置。
type PostgresConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	AutoMigrate bool // 是否在启动时执行 AutoMigrate（建议生产禁用）
}

func loadPostgres() PostgresConfig {

	return PostgresConfig{
		Host:        getEnv("DB_HOST", "localhost"),
		Port:        getEnv("DB_PORT", "5432"),
		User:        getEnv("DB_USER", "postgres"),
		Password:    getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "postgres"),
		SSLMode:     getEnv("DB_SSLMODE", ""),
		AutoMigrate: getEnvBool("DB_AUTO_MIGRATE", true),
	}
}
