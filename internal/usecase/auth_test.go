package usecase

import (
	"context"
	"errors"
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

func TestAuthUsecase_Register(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*MockUserRepository, *MockJWTManager)
		expectedError error
	}{
		{
			name:     "успешная регистрация",
			email:    "test@example.com",
			password: "securepassword123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Пользователь не найден (еще не существует)
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), ErrUserNotFound)
				// Create выполняется успешно
				repo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Email == "test@example.com" && len(u.PasswordHash) > 0 && u.ID != ""
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "пользователь уже существует",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Пользователь уже существует
				existingUser := &domain.User{ID: "123", Email: "existing@example.com"}
				repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: ErrUserExists,
		},
		{
			name:     "пустой email",
			email:    "",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedError: ErrEmptyEmail,
		},
		{
			name:     "невалидный email",
			email:    "invalid-email",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedError: ErrInvalidEmail,
		},
		{
			name:     "пустой пароль",
			email:    "test@example.com",
			password: "",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedError: ErrEmptyPassword,
		},
		{
			name:     "слишком короткий пароль",
			email:    "test@example.com",
			password: "12345",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedError: ErrWeakPassword,
		},
		{
			name:     "ошибка при создании пользователя",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), ErrUserNotFound)
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			logger := newTestLogger()

			tt.setupMocks(mockRepo, mockJWT)

			usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

			// Act
			err := usecase.Register(context.Background(), tt.email, tt.password)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrUserExists) ||
					errors.Is(tt.expectedError, ErrEmptyEmail) ||
					errors.Is(tt.expectedError, ErrInvalidEmail) ||
					errors.Is(tt.expectedError, ErrEmptyPassword) ||
					errors.Is(tt.expectedError, ErrWeakPassword) {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	// Подготовка: хешируем пароль для тестов
	validPassword := "mypassword123"
	validHash, _ := auth.HashPassword(validPassword)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*MockUserRepository, *MockJWTManager)
		expectedToken string
		expectedError error
	}{
		{
			name:     "успешный логин",
			email:    "user@example.com",
			password: validPassword,
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				user := &domain.User{
					ID:           "user-123",
					Email:        "user@example.com",
					PasswordHash: validHash,
				}
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				jwt.On("GenerateToken", "user-123").Return("fake.jwt.token", nil)
			},
			expectedToken: "fake.jwt.token",
			expectedError: nil,
		},
		{
			name:     "пользователь не найден",
			email:    "notfound@example.com",
			password: "anypassword",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				repo.On("GetByEmail", mock.Anything, "notfound@example.com").Return((*domain.User)(nil), ErrUserNotFound)
			},
			expectedToken: "",
			expectedError: ErrInvalidCredentials,
		},
		{
			name:     "неправильный пароль",
			email:    "user@example.com",
			password: "wrongpassword",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				user := &domain.User{
					ID:           "user-123",
					Email:        "user@example.com",
					PasswordHash: validHash,
				}
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
			},
			expectedToken: "",
			expectedError: ErrInvalidCredentials,
		},
		{
			name:     "пустой email",
			email:    "",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedToken: "",
			expectedError: ErrEmptyEmail,
		},
		{
			name:     "невалидный email",
			email:    "invalid-email",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedToken: "",
			expectedError: ErrInvalidEmail,
		},
		{
			name:     "пустой пароль",
			email:    "user@example.com",
			password: "",
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedToken: "",
			expectedError: ErrEmptyPassword,
		},
		{
			name:     "ошибка генерации токена",
			email:    "user@example.com",
			password: validPassword,
			setupMocks: func(repo *MockUserRepository, jwt *MockJWTManager) {
				user := &domain.User{
					ID:           "user-123",
					Email:        "user@example.com",
					PasswordHash: validHash,
				}
				repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				jwt.On("GenerateToken", "user-123").Return("", errors.New("jwt generation failed"))
			},
			expectedToken: "",
			expectedError: errors.New("jwt generation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			logger := newTestLogger()

			tt.setupMocks(mockRepo, mockJWT)

			usecase := NewAuthUsecase(mockRepo, mockJWT, logger)

			// Act
			token, err := usecase.Login(context.Background(), tt.email, tt.password)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrInvalidCredentials) ||
					errors.Is(tt.expectedError, ErrEmptyEmail) ||
					errors.Is(tt.expectedError, ErrInvalidEmail) ||
					errors.Is(tt.expectedError, ErrEmptyPassword) {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}
