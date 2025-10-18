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
	_, err := u.userRepo.GetByEmail(ctx, email)
	if err == nil {
		u.logger.Error("failed to get user by email", zap.Error(err))
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

	return u.userRepo.Create(ctx, user)
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		u.logger.Error("failed to get user by email", zap.Error(err))
		return "", ErrInvalidCredentials
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		u.logger.Error("failed to check password", zap.Error(err))
		return "", ErrInvalidCredentials
	}

	token, err := u.jwtManager.GenerateToken(user.ID)
	if err != nil {
		u.logger.Error("failed to generate token", zap.Error(err))
		return "", err
	}

	return token, nil
}

// Вспомогательная функция
func generateUUID() string {
	return uuid.New().String()
}
