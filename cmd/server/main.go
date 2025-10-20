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
	"github.com/m-molecula741/gophkeeper/internal/delivery/api"
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

	// JWT менеджер
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)

	// Use case
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager, logr)

	// HTTP хендлеры
	authHandler := api.NewAuthHandler(authUsecase, logr)

	// Роутер (можно использовать net/http или Gin/Echo)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

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
