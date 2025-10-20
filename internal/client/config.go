package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config хранит конфигурацию CLI клиента
type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token,omitempty"`
}

// ConfigPath возвращает путь к файлу конфигурации
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".gophkeeper")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// Если файл не существует, возвращаем конфиг по умолчанию
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			ServerURL: "http://localhost:8080",
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig сохраняет конфигурацию в файл
func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SetToken устанавливает токен и сохраняет конфигурацию
func (c *Config) SetToken(token string) error {
	c.Token = token
	return c.Save()
}

// ClearToken очищает токен
func (c *Config) ClearToken() error {
	c.Token = ""
	return c.Save()
}
