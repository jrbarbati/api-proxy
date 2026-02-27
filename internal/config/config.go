package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultServerPort = "8080"
)

type Config struct {
	Server    *ServerConfig `yaml:"server"`
	DB        *DBConfig     `yaml:"db"`
	JWTConfig *JWTConfig    `yaml:"jwt"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
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

	return applyEnvOverrides(config), nil
}

func applyEnvOverrides(config *Config) *Config {
	if config.Server == nil {
		config.Server = &ServerConfig{}
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

	if config.Server.Port == "" {
		config.Server.Port = DefaultServerPort
	}

	return config
}
