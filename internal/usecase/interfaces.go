package usecase

import (
	"context"
	"errors"

	"github.com/m-molecula741/gophkeeper/internal/domain"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrSecretNotFound = errors.New("secret not found")
	ErrUnauthorized   = errors.New("unauthorized access to secret")
)

// UserRepository определяет операции над пользователями.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// SecretRepository определяет операции над секретами.
type SecretRepository interface {
	Create(ctx context.Context, secret *domain.Secret) error
	GetByUserID(ctx context.Context, userID string) ([]*domain.Secret, error)
	GetByID(ctx context.Context, id, userID string) (*domain.Secret, error)
	Update(ctx context.Context, secret *domain.Secret) error
	Delete(ctx context.Context, id, userID string) error
}