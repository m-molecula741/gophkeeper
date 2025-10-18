// cmd/server/main.go

package main

import (
	"log"
	"os"

	"github.com/m-molecula741/gophkeeper/internal/config"
	"github.com/m-molecula741/gophkeeper/pkg/logger"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Инициализируем логгер
	logr, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("❌ Не удалось создать логгер: %v", err)
	}
	defer logr.Sync()

	log.Printf("✅ Starting GophKeeper server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	// ... запуск HTTP-сервера, БД, роутов и т.д.
}
