package config

import (
	"errors"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	DefaultServerPort           = "8080"
	defaultLogLevel             = "INFO"
	defaultRequestLogQueueSize  = 500
	defaultRequestRetentionDays = 7
)

var ErrInvalidLoggingRequestQueueSize = errors.New("invalid logging request queue size")
var ErrInvalidLoggingRequestRetention = errors.New("invalid logging request retention")

type Config struct {
	Server             *ServerConfig       `yaml:"server"`
	DB                 *DBConfig           `yaml:"db"`
	JWTConfig          *JWTConfig          `yaml:"jwt"`
	LoggingConfig      *LoggingConfig      `yaml:"logging"`
	RateLimitingConfig *RateLimitingConfig `yaml:"rate_limiting"`
}

type LoggingConfig struct {
	Level                string                `yaml:"level"`
	LoggingRequestConfig *LoggingRequestConfig `yaml:"request"`
	LoggingAuditConfig   *LoggingAuditConfig   `yaml:"audit"`
}

type LoggingRequestConfig struct {
	QueueSize     *int `yaml:"queue_size"`
	RetentionDays *int `yaml:"retention_days"`
}

type LoggingAuditConfig struct {
	QueueSize     *int `yaml:"queue_size"`
	RetentionDays *int `yaml:"retention_days"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type RateLimitingConfig struct {
	Backend string       `yaml:"backend"`
	Redis   *RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	URL string `yaml:"url"`
}

type DBConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"db_name"`
}

type JWTConfig struct {
	SigningSecret string          `yaml:"signing_secret"`
	Admin         *AdminJWTConfig `yaml:"admin"`
}

type AdminJWTConfig struct {
	SigningSecret string `yaml:"signing_secret"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(file, config)

	if err != nil {
		return nil, err
	}

	return applyEnvOverrides(config)
}

func applyEnvOverrides(config *Config) (*Config, error) {
	if config.Server == nil {
		config.Server = &ServerConfig{}
	}

	if config.RateLimitingConfig == nil {
		config.RateLimitingConfig = &RateLimitingConfig{}
	}

	if config.RateLimitingConfig.Redis == nil {
		config.RateLimitingConfig.Redis = &RedisConfig{}
	}

	if config.DB == nil {
		config.DB = &DBConfig{}
	}

	if config.JWTConfig == nil {
		config.JWTConfig = &JWTConfig{}
	}

	if config.JWTConfig.Admin == nil {
		config.JWTConfig.Admin = &AdminJWTConfig{}
	}

	if config.LoggingConfig == nil {
		config.LoggingConfig = &LoggingConfig{}
	}

	if config.LoggingConfig.LoggingRequestConfig == nil {
		config.LoggingConfig.LoggingRequestConfig = &LoggingRequestConfig{}
	}

	if val := os.Getenv("SERVER_PORT"); val != "" {
		config.Server.Port = val
	}

	if val := os.Getenv("DB_URL"); val != "" {
		config.DB.URL = val
	}

	if val := os.Getenv("DB_USERNAME"); val != "" {
		config.DB.Username = val
	}

	if val := os.Getenv("DB_PASSWORD"); val != "" {
		config.DB.Password = val
	}

	if val := os.Getenv("DB_PORT"); val != "" {
		config.DB.Port = val
	}

	if val := os.Getenv("DB_NAME"); val != "" {
		config.DB.DBName = val
	}

	if val := os.Getenv("JWT_SIGNING_SECRET"); val != "" {
		config.JWTConfig.SigningSecret = val
	}

	if val := os.Getenv("ADMIN_JWT_SIGNING_SECRET"); val != "" {
		config.JWTConfig.Admin.SigningSecret = val
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LoggingConfig.Level = val
	}

	if val := os.Getenv("LOG_REQUEST_QUEUE_SIZE"); val != "" {
		valInt, err := strconv.Atoi(val)

		if err != nil {
			return nil, ErrInvalidLoggingRequestQueueSize
		}

		config.LoggingConfig.LoggingRequestConfig.QueueSize = &valInt
	}

	if val := os.Getenv("LOG_REQUEST_RETENTION_DAYS"); val != "" {
		valInt, err := strconv.Atoi(val)

		if err != nil {
			return nil, ErrInvalidLoggingRequestRetention
		}

		config.LoggingConfig.LoggingRequestConfig.RetentionDays = &valInt
	}

	if val := os.Getenv("RATE_LIMIT_BACKEND"); val != "" {
		config.RateLimitingConfig.Backend = val
	}

	if val := os.Getenv("REDIS_URL"); val != "" {
		config.RateLimitingConfig.Redis.URL = val
	}

	if config.Server.Port == "" {
		config.Server.Port = DefaultServerPort
	}

	if config.LoggingConfig.Level == "" {
		config.LoggingConfig.Level = defaultLogLevel
	}

	if config.LoggingConfig.LoggingRequestConfig.QueueSize == nil {
		config.LoggingConfig.LoggingRequestConfig.QueueSize = new(defaultRequestLogQueueSize)
	}

	if config.LoggingConfig.LoggingRequestConfig.RetentionDays == nil {
		config.LoggingConfig.LoggingRequestConfig.RetentionDays = new(defaultRequestRetentionDays)
	}

	return config, nil
}
