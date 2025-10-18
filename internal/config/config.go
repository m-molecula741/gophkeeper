// Package config отвечает за загрузку и валидацию конфигурации приложения.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config представляет корневую структуру конфигурации GophKeeper.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig содержит настройки HTTP-сервера.
type ServerConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	TLSEnabled  bool   `mapstructure:"tls_enabled"`
	CertFile    string `mapstructure:"cert_file"`
	KeyFile     string `mapstructure:"key_file"`
}

// DatabaseConfig содержит параметры подключения к базе данных.
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// AuthConfig содержит параметры аутентификации.
type AuthConfig struct {
	JWTSecret string        `mapstructure:"jwt_secret"`
	TokenTTL  time.Duration `mapstructure:"token_ttl"`
}

// LoggingConfig содержит настройки логирования.
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// Load загружает конфигурацию из YAML-файла (по умолчанию "config.yaml") или переменных окружения.
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOPHKEEPER")

	viper.BindEnv("database.dsn", "GOPHKEEPER_DATABASE_DSN")
	viper.BindEnv("auth.jwt_secret", "GOPHKEEPER_AUTH_JWT_SECRET")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %s", configPath)
		}
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database.dsn is required")
	}
	if cfg.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("auth.jwt_secret is required")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("auth.jwt_secret must be at least 32 characters long for security")
	}

	return &cfg, nil
}