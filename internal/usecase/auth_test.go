package usecase

import (
	"context"
	"testing"

	"github.com/m-molecula741/gophkeeper/internal/auth"
	"github.com/m-molecula741/gophkeeper/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserRepository — мок репозитория пользователя
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

// MockJWTManager — мок JWT-менеджера
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) VerifyToken(tokenStr string) (*auth.UserClaims, error) {
	args := m.Called(tokenStr)
	if claims, ok := args.Get(0).(*auth.UserClaims); ok {
		return claims, args.Error(1)
	}
	return nil, args.Error(1)
}

// Вспомогательная функция для создания логгера в тестах
func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	logger := newTestLogger()
	usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

	email := "test@example.com"
	password := "securepassword"

	// GetByEmail возвращает ошибку — пользователя нет
	mockRepo.On("GetByEmail", mock.Anything, email).Return((*domain.User)(nil), ErrUserNotFound)
	// Create вызывается успешно
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == email && len(u.PasswordHash) > 0
	})).Return(nil)

	err := usecase.Register(context.Background(), email, password)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	logger := newTestLogger()
	usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

	email := "existing@example.com"
	password := "password"

	// Пользователь уже существует
	existingUser := &domain.User{ID: "123", Email: email}
	mockRepo.On("GetByEmail", mock.Anything, email).Return(existingUser, nil)

	err := usecase.Register(context.Background(), email, password)

	assert.ErrorIs(t, err, ErrUserExists)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	logger := newTestLogger()
	usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

	email := "user@example.com"
	password := "mypassword"
	userID := "user-123"
	hashedPassword, _ := auth.HashPassword(password)
	token := "fake.jwt.token"

	user := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	mockRepo.On("GetByEmail", mock.Anything, email).Return(user, nil)
	mockJWT.On("GenerateToken", userID).Return(token, nil)

	resultToken, err := usecase.Login(context.Background(), email, password)

	assert.NoError(t, err)
	assert.Equal(t, token, resultToken)
	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	logger := newTestLogger()
	usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

	email := "notfound@example.com"
	password := "any"

	mockRepo.On("GetByEmail", mock.Anything, email).Return((*domain.User)(nil), ErrUserNotFound)

	_, err := usecase.Login(context.Background(), email, password)

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)
	logger := newTestLogger()
	usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

	email := "user@example.com"
	wrongPassword := "wrong"
	correctHash, _ := auth.HashPassword("correct")

	user := &domain.User{
		ID:           "123",
		Email:        email,
		PasswordHash: correctHash,
	}

	mockRepo.On("GetByEmail", mock.Anything, email).Return(user, nil)

	_, err := usecase.Login(context.Background(), email, wrongPassword)

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}