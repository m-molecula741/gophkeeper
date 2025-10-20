// cmd/server/main.go

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/m-molecula741/gophkeeper/internal/auth"
	"github.com/m-molecula741/gophkeeper/internal/config"
	"github.com/m-molecula741/gophkeeper/internal/crypto"
	"github.com/m-molecula741/gophkeeper/internal/delivery/api"
	"github.com/m-molecula741/gophkeeper/internal/delivery/middleware"
	"github.com/m-molecula741/gophkeeper/internal/repository/postgres"
	"github.com/m-molecula741/gophkeeper/internal/usecase"
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

	// Логгер
	logr, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("❌ Failed to create logger: %v", err)
	}
	defer logr.Sync()

	// Подключение к БД
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgClient, err := postgres.New(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	defer pgClient.Close()

	// Репозитории
	userRepo := postgres.NewUserRepository(pgClient)
	secretRepo := postgres.NewSecretRepository(pgClient)

	// JWT менеджер
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)

	// Encryption (используем JWT secret как encryption key - в продакшене лучше отдельный ключ)
	encryptionKey := []byte(cfg.Auth.JWTSecret)
	if len(encryptionKey) != 32 {
		// Дополняем или обрезаем до 32 байт для AES-256
		key := make([]byte, 32)
		copy(key, encryptionKey)
		encryptionKey = key
	}
	encryptor, err := crypto.NewAESCipher(encryptionKey)
	if err != nil {
		log.Fatalf("❌ Failed to create encryptor: %v", err)
	}

	// Use cases
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager, logr)
	secretUsecase := usecase.NewSecretUsecase(secretRepo, encryptor, logr)

	// HTTP хендлеры
	authHandler := api.NewAuthHandler(authUsecase, logr)
	secretHandler := api.NewSecretHandler(secretUsecase, logr)

	// Middleware
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// Роутер
	mux := http.NewServeMux()

	// Публичные роуты (без авторизации)
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	// Защищенные роуты (требуют авторизации)
	mux.Handle("POST /api/v1/secrets", authMiddleware(http.HandlerFunc(secretHandler.CreateSecret)))
	mux.Handle("GET /api/v1/secrets", authMiddleware(http.HandlerFunc(secretHandler.GetAllSecrets)))
	mux.Handle("GET /api/v1/secrets/{id}", authMiddleware(http.HandlerFunc(secretHandler.GetSecret)))
	mux.Handle("PUT /api/v1/secrets/{id}", authMiddleware(http.HandlerFunc(secretHandler.UpdateSecret)))
	mux.Handle("DELETE /api/v1/secrets/{id}", authMiddleware(http.HandlerFunc(secretHandler.DeleteSecret)))

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		log.Printf("✅ Starting GophKeeper server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Ждём SIGINT или SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⏳ Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
