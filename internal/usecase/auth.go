package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/m-molecula741/gophkeeper/internal/auth"
	"github.com/m-molecula741/gophkeeper/internal/domain"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
)

type AuthUsecase struct {
	userRepo   UserRepository
	jwtManager AuthJWTManager
	logger     *zap.Logger
}

func NewAuthUsecase(userRepo UserRepository, jwtManager AuthJWTManager, logger *zap.Logger) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, email, password string) error {
	// Валидация входных данных
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if err := ValidatePassword(password); err != nil {
		return err
	}

	// Проверка существования пользователя
	_, err := u.userRepo.GetByEmail(ctx, email)
	if err == nil {
		// Пользователь уже существует - это нормальная бизнес-логика, не ошибка системы
		return ErrUserExists
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		u.logger.Error("failed to hash password", zap.Error(err))
		return err
	}

	userID := generateUUID() // используй github.com/google/uuid

	user := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		u.logger.Error("failed to create user", zap.Error(err), zap.String("email", email))
		return err
	}

	u.logger.Info("user registered successfully", zap.String("email", email), zap.String("user_id", userID))
	return nil
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	// Валидация входных данных
	if err := ValidateEmail(email); err != nil {
		return "", err
	}
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Не логируем как ошибку - это может быть просто неправильный email
		u.logger.Debug("user not found", zap.String("email", email))
		return "", ErrInvalidCredentials
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		// Неправильный пароль - это нормальная ситуация, не системная ошибка
		u.logger.Debug("invalid password attempt", zap.String("email", email))
		return "", ErrInvalidCredentials
	}

	token, err := u.jwtManager.GenerateToken(user.ID)
	if err != nil {
		u.logger.Error("failed to generate token", zap.Error(err))
		return "", err
	}

	u.logger.Info("user logged in successfully", zap.String("email", email), zap.String("user_id", user.ID))
	return token, nil
}

// Вспомогательная функция
func generateUUID() string {
	return uuid.New().String()
}
