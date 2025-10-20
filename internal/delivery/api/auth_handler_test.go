package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-molecula741/gophkeeper/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAuthUsecase — мок для AuthUsecaseInterface
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) Register(ctx context.Context, email, password string) error {
	args := m.Called(ctx, email, password)
	return args.Error(0)
}

func (m *MockAuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		setupMocks         func(*MockAuthUsecase)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "успешная регистрация",
			requestBody: loginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Register", mock.Anything, "test@example.com", "password123").Return(nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedError:      "",
		},
		{
			name: "пользователь уже существует",
			requestBody: loginRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Register", mock.Anything, "existing@example.com", "password123").Return(usecase.ErrUserExists)
			},
			expectedStatusCode: http.StatusConflict,
			expectedError:      "user already exists",
		},
		{
			name: "невалидный email",
			requestBody: loginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Register", mock.Anything, "invalid-email", "password123").Return(usecase.ErrInvalidEmail)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "invalid email format",
		},
		{
			name: "слабый пароль",
			requestBody: loginRequest{
				Email:    "test@example.com",
				Password: "123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Register", mock.Anything, "test@example.com", "123").Return(usecase.ErrWeakPassword)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "password must be at least 6 characters",
		},
		{
			name:               "невалидный JSON",
			requestBody:        `{"invalid json`,
			setupMocks:         func(m *MockAuthUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "invalid json",
		},
		{
			name: "внутренняя ошибка сервера",
			requestBody: loginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Register", mock.Anything, "test@example.com", "password123").Return(errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockAuthUsecase)
			logger := newTestLogger()
			handler := NewAuthHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			// Подготовка тела запроса
			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/register", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.Register(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedError != "" {
				responseBody := w.Body.String()
				assert.Contains(t, responseBody, tt.expectedError)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		setupMocks         func(*MockAuthUsecase)
		expectedStatusCode int
		expectedToken      string
		expectedError      string
	}{
		{
			name: "успешный логин",
			requestBody: loginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Login", mock.Anything, "user@example.com", "password123").Return("fake.jwt.token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedToken:      "fake.jwt.token",
			expectedError:      "",
		},
		{
			name: "неправильные учетные данные",
			requestBody: loginRequest{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Login", mock.Anything, "user@example.com", "wrongpassword").Return("", usecase.ErrInvalidCredentials)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedToken:      "",
			expectedError:      "invalid credentials",
		},
		{
			name: "пользователь не найден",
			requestBody: loginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Login", mock.Anything, "notfound@example.com", "password123").Return("", usecase.ErrInvalidCredentials)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedToken:      "",
			expectedError:      "invalid credentials",
		},
		{
			name:               "невалидный JSON",
			requestBody:        `{"invalid json`,
			setupMocks:         func(m *MockAuthUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedToken:      "",
			expectedError:      "invalid json",
		},
		{
			name: "внутренняя ошибка сервера при генерации токена",
			requestBody: loginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Login", mock.Anything, "user@example.com", "password123").Return("", errors.New("jwt generation failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedToken:      "",
			expectedError:      "internal error",
		},
		{
			name: "невалидный email формат",
			requestBody: loginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Login", mock.Anything, "invalid-email", "password123").Return("", usecase.ErrInvalidEmail)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedToken:      "",
			expectedError:      "invalid email format",
		},
		{
			name: "пустой пароль",
			requestBody: loginRequest{
				Email:    "user@example.com",
				Password: "",
			},
			setupMocks: func(m *MockAuthUsecase) {
				m.On("Login", mock.Anything, "user@example.com", "").Return("", usecase.ErrEmptyPassword)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedToken:      "",
			expectedError:      "password cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockAuthUsecase)
			logger := newTestLogger()
			handler := NewAuthHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			// Подготовка тела запроса
			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/login", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.Login(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedToken != "" {
				var response tokenResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, response.Token)
			}

			if tt.expectedError != "" {
				responseBody := w.Body.String()
				assert.Contains(t, responseBody, tt.expectedError)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}
