package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Analysis AnalysisConfig `mapstructure:"analysis"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxConnections  int           `mapstructure:"max_connections"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
}

type AnalysisConfig struct {
	RequestTimeout    time.Duration `mapstructure:"request_timeout"`
	MaxContentLength  int64         `mapstructure:"max_content_length"`
	CacheTTL          time.Duration `mapstructure:"cache_ttl"`
	RateLimitPerIP    int           `mapstructure:"rate_limit_per_ip"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`
	MaxConcurrentJobs int           `mapstructure:"max_concurrent_jobs"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	setDefaults()

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "webpage_analyzer")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 50)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 100)
	viper.SetDefault("redis.min_idle_conns", 10)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")

	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.development", false)

	viper.SetDefault("analysis.request_timeout", "30s")
	viper.SetDefault("analysis.max_content_length", 10485760)
	viper.SetDefault("analysis.cache_ttl", "1h")
	viper.SetDefault("analysis.rate_limit_per_ip", 100)
	viper.SetDefault("analysis.rate_limit_window", "1m")
	viper.SetDefault("analysis.max_concurrent_jobs", 50)

	_ = viper.BindEnv("server.port", "PORT")
	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.name", "DB_NAME")
	_ = viper.BindEnv("database.ssl_mode", "DB_SSL_MODE")
	_ = viper.BindEnv("database.max_connections", "DB_MAX_CONNECTIONS")
	_ = viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	_ = viper.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "REDIS_DB")
	_ = viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	_ = viper.BindEnv("redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")
	_ = viper.BindEnv("redis.dial_timeout", "REDIS_DIAL_TIMEOUT")
	_ = viper.BindEnv("redis.read_timeout", "REDIS_READ_TIMEOUT")
	_ = viper.BindEnv("redis.write_timeout", "REDIS_WRITE_TIMEOUT")

	_ = viper.BindEnv("logger.level", "LOG_LEVEL")
	_ = viper.BindEnv("logger.development", "LOG_DEVELOPMENT")

	_ = viper.BindEnv("analysis.request_timeout", "ANALYSIS_REQUEST_TIMEOUT")
	_ = viper.BindEnv("analysis.max_content_length", "ANALYSIS_MAX_CONTENT_LENGTH")
	_ = viper.BindEnv("analysis.cache_ttl", "ANALYSIS_CACHE_TTL")
	_ = viper.BindEnv("analysis.rate_limit_per_ip", "ANALYSIS_RATE_LIMIT_PER_IP")
	_ = viper.BindEnv("analysis.rate_limit_window", "ANALYSIS_RATE_LIMIT_WINDOW")
	_ = viper.BindEnv("analysis.max_concurrent_jobs", "ANALYSIS_MAX_CONCURRENT_JOBS")
}
