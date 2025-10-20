package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPath(t *testing.T) {
	// Act
	configPath, err := ConfigPath()
	
	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, configPath)
	assert.Contains(t, configPath, ".gophkeeper")
	assert.Contains(t, configPath, "config.json")
}

func TestLoadConfig_NewConfig(t *testing.T) {
	// Arrange - используем временную директорию
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Act
	cfg, err := LoadConfig()
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "http://localhost:8080", cfg.ServerURL)
	assert.Empty(t, cfg.Token)
}

func TestLoadConfig_ExistingConfig(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Создаем конфиг
	cfg := &Config{
		ServerURL: "http://example.com:9090",
		Token:     "test-token",
	}
	err := cfg.Save()
	require.NoError(t, err)
	
	// Act
	loadedCfg, err := LoadConfig()
	
	// Assert
	require.NoError(t, err)
	assert.Equal(t, "http://example.com:9090", loadedCfg.ServerURL)
	assert.Equal(t, "test-token", loadedCfg.Token)
}

func TestConfig_Save(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	cfg := &Config{
		ServerURL: "http://test.com",
		Token:     "my-token",
	}
	
	// Act
	err := cfg.Save()
	
	// Assert
	require.NoError(t, err)
	
	// Проверяем что файл создан
	configPath, _ := ConfigPath()
	assert.FileExists(t, configPath)
	
	// Проверяем содержимое
	loadedCfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, cfg.ServerURL, loadedCfg.ServerURL)
	assert.Equal(t, cfg.Token, loadedCfg.Token)
}

func TestConfig_SetToken(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	cfg := &Config{
		ServerURL: "http://localhost:8080",
	}
	
	// Act
	err := cfg.SetToken("new-token-123")
	
	// Assert
	require.NoError(t, err)
	assert.Equal(t, "new-token-123", cfg.Token)
	
	// Проверяем что токен сохранен в файл
	loadedCfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "new-token-123", loadedCfg.Token)
}

func TestConfig_ClearToken(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	cfg := &Config{
		ServerURL: "http://localhost:8080",
		Token:     "token-to-clear",
	}
	err := cfg.Save()
	require.NoError(t, err)
	
	// Act
	err = cfg.ClearToken()
	
	// Assert
	require.NoError(t, err)
	assert.Empty(t, cfg.Token)
	
	// Проверяем что токен очищен в файле
	loadedCfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Empty(t, loadedCfg.Token)
}

func TestConfigPath_CreateDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Act
	configPath, err := ConfigPath()
	
	// Assert
	require.NoError(t, err)
	
	// Проверяем что директория создана
	configDir := filepath.Dir(configPath)
	assert.DirExists(t, configDir)
	
	// Проверяем права доступа
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Создаем невалидный JSON
	configPath, _ := ConfigPath()
	err := os.WriteFile(configPath, []byte("invalid json {"), 0600)
	require.NoError(t, err)
	
	// Act
	cfg, err := LoadConfig()
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestConfig_SaveAndLoad_Multiple(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		token     string
	}{
		{
			name:      "с токеном",
			serverURL: "http://server1.com",
			token:     "token1",
		},
		{
			name:      "без токена",
			serverURL: "http://server2.com",
			token:     "",
		},
		{
			name:      "длинный токен",
			serverURL: "https://production.example.com:8443",
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", originalHome)
			
			cfg := &Config{
				ServerURL: tt.serverURL,
				Token:     tt.token,
			}
			
			// Act
			err := cfg.Save()
			require.NoError(t, err)
			
			loadedCfg, err := LoadConfig()
			
			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.serverURL, loadedCfg.ServerURL)
			assert.Equal(t, tt.token, loadedCfg.Token)
		})
	}
}

