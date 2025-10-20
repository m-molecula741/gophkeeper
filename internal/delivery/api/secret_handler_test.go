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
	"time"

	"github.com/m-molecula741/gophkeeper/internal/delivery/middleware"
	"github.com/m-molecula741/gophkeeper/internal/domain"
	"github.com/m-molecula741/gophkeeper/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecretUsecase —мок для SecretUsecaseInterface
type MockSecretUsecase struct {
	mock.Mock
}

func (m *MockSecretUsecase) Create(ctx context.Context, userID, secretType string, data []byte, metadata string) (*domain.Secret, error) {
	args := m.Called(ctx, userID, secretType, data, metadata)
	if secret, ok := args.Get(0).(*domain.Secret); ok {
		return secret, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSecretUsecase) GetAll(ctx context.Context, userID string) ([]*domain.Secret, error) {
	args := m.Called(ctx, userID)
	if secrets, ok := args.Get(0).([]*domain.Secret); ok {
		return secrets, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSecretUsecase) GetByID(ctx context.Context, secretID, userID string) (*domain.Secret, error) {
	args := m.Called(ctx, secretID, userID)
	if secret, ok := args.Get(0).(*domain.Secret); ok {
		return secret, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSecretUsecase) GetDecrypted(ctx context.Context, secretID, userID string) (*domain.Secret, []byte, error) {
	args := m.Called(ctx, secretID, userID)
	if secret, ok := args.Get(0).(*domain.Secret); ok {
		if data, ok := args.Get(1).([]byte); ok {
			return secret, data, args.Error(2)
		}
	}
	return nil, nil, args.Error(2)
}

func (m *MockSecretUsecase) Update(ctx context.Context, secretID, userID, secretType string, data []byte, metadata string) error {
	args := m.Called(ctx, secretID, userID, secretType, data, metadata)
	return args.Error(0)
}

func (m *MockSecretUsecase) Delete(ctx context.Context, secretID, userID string) error {
	args := m.Called(ctx, secretID, userID)
	return args.Error(0)
}

func TestSecretHandler_CreateSecret(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name               string
		requestBody        interface{}
		userID             string
		hasAuth            bool
		setupMocks         func(*MockSecretUsecase)
		expectedStatusCode int
		checkResponse      func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "успешное создание секрета",
			requestBody: createSecretRequest{
				Type:     "login",
				Data:     "username:password",
				Metadata: `{"website":"example.com"}`,
			},
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				secret := &domain.Secret{
					ID:        "secret-123",
					UserID:    "user-123",
					Type:      "login",
					Metadata:  `{"website":"example.com"}`,
					CreatedAt: now,
					UpdatedAt: now,
				}
				m.On("Create", mock.Anything, "user-123", "login", []byte("username:password"), `{"website":"example.com"}`).
					Return(secret, nil)
			},
			expectedStatusCode: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp secretResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "secret-123", resp.ID)
				assert.Equal(t, "login", resp.Type)
			},
		},
		{
			name: "невалидный тип секрета",
			requestBody: createSecretRequest{
				Type:     "invalid",
				Data:     "data",
				Metadata: `{}`,
			},
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Create", mock.Anything, "user-123", "invalid", []byte("data"), `{}`).
					Return((*domain.Secret)(nil), usecase.ErrInvalidSecretType)
			},
			expectedStatusCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid secret type")
			},
		},
		{
			name: "пустые данные",
			requestBody: createSecretRequest{
				Type:     "login",
				Data:     "",
				Metadata: `{}`,
			},
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Create", mock.Anything, "user-123", "login", []byte(""), `{}`).
					Return((*domain.Secret)(nil), usecase.ErrEmptySecretData)
			},
			expectedStatusCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "secret data cannot be empty")
			},
		},
		{
			name: "без авторизации",
			requestBody: createSecretRequest{
				Type:     "login",
				Data:     "data",
				Metadata: `{}`,
			},
			userID:             "",
			hasAuth:            false,
			setupMocks:         func(m *MockSecretUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "unauthorized")
			},
		},
		{
			name:               "невалидный JSON",
			requestBody:        `{"invalid json`,
			userID:             "user-123",
			hasAuth:            true,
			setupMocks:         func(m *MockSecretUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid json")
			},
		},
		{
			name: "внутренняя ошибка сервера",
			requestBody: createSecretRequest{
				Type:     "login",
				Data:     "data",
				Metadata: `{}`,
			},
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Create", mock.Anything, "user-123", "login", []byte("data"), `{}`).
					Return((*domain.Secret)(nil), errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "internal error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockSecretUsecase)
			logger := newTestLogger()
			handler := NewSecretHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			// Подготовка тела запроса
			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", body)
			req.Header.Set("Content-Type", "application/json")

			// Добавляем userID в контекст если есть авторизация
			if tt.hasAuth {
				ctx := context.WithValue(req.Context(), middleware.GetUserIDKey(), tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Act
			handler.CreateSecret(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestSecretHandler_GetAllSecrets(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name               string
		userID             string
		hasAuth            bool
		setupMocks         func(*MockSecretUsecase)
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:    "успешное получение нескольких секретов",
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				secrets := []*domain.Secret{
					{ID: "secret-1", Type: "login", CreatedAt: now, UpdatedAt: now},
					{ID: "secret-2", Type: "text", CreatedAt: now, UpdatedAt: now},
				}
				m.On("GetAll", mock.Anything, "user-123").Return(secrets, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedCount:      2,
		},
		{
			name:    "пользователь без секретов",
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("GetAll", mock.Anything, "user-123").Return([]*domain.Secret{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedCount:      0,
		},
		{
			name:               "без авторизации",
			userID:             "",
			hasAuth:            false,
			setupMocks:         func(m *MockSecretUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedCount:      0,
		},
		{
			name:    "внутренняя ошибка",
			userID:  "user-123",
			hasAuth: true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("GetAll", mock.Anything, "user-123").Return(([]*domain.Secret)(nil), errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockSecretUsecase)
			logger := newTestLogger()
			handler := NewSecretHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

			if tt.hasAuth {
				ctx := context.WithValue(req.Context(), middleware.GetUserIDKey(), tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Act
			handler.GetAllSecrets(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedStatusCode == http.StatusOK {
				var response []secretResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedCount)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestSecretHandler_GetSecret(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name               string
		secretID           string
		userID             string
		hasAuth            bool
		setupMocks         func(*MockSecretUsecase)
		expectedStatusCode int
	}{
		{
			name:     "успешное получение секрета",
			secretID: "secret-123",
			userID:   "user-123",
			hasAuth:  true,
			setupMocks: func(m *MockSecretUsecase) {
				secret := &domain.Secret{
					ID:        "secret-123",
					Type:      "login",
					Metadata:  `{}`,
					CreatedAt: now,
					UpdatedAt: now,
				}
				m.On("GetDecrypted", mock.Anything, "secret-123", "user-123").
					Return(secret, []byte("decrypted-data"), nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:     "секрет не найден",
			secretID: "secret-999",
			userID:   "user-123",
			hasAuth:  true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("GetDecrypted", mock.Anything, "secret-999", "user-123").
					Return((*domain.Secret)(nil), []byte(nil), usecase.ErrSecretNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "без авторизации",
			secretID:           "secret-123",
			userID:             "",
			hasAuth:            false,
			setupMocks:         func(m *MockSecretUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:     "внутренняя ошибка при расшифровке",
			secretID: "secret-123",
			userID:   "user-123",
			hasAuth:  true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("GetDecrypted", mock.Anything, "secret-123", "user-123").
					Return((*domain.Secret)(nil), []byte(nil), errors.New("decryption failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockSecretUsecase)
			logger := newTestLogger()
			handler := NewSecretHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/"+tt.secretID, nil)
			req.SetPathValue("id", tt.secretID)

			if tt.hasAuth {
				ctx := context.WithValue(req.Context(), middleware.GetUserIDKey(), tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Act
			handler.GetSecret(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestSecretHandler_UpdateSecret(t *testing.T) {
	tests := []struct {
		name               string
		secretID           string
		userID             string
		hasAuth            bool
		requestBody        interface{}
		setupMocks         func(*MockSecretUsecase)
		expectedStatusCode int
	}{
		{
			name:     "успешное обновление",
			secretID: "secret-123",
			userID:   "user-123",
			hasAuth:  true,
			requestBody: updateSecretRequest{
				Type:     "text",
				Data:     "new-data",
				Metadata: `{"updated":true}`,
			},
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Update", mock.Anything, "secret-123", "user-123", "text", []byte("new-data"), `{"updated":true}`).
					Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:     "секрет не найден",
			secretID: "secret-999",
			userID:   "user-123",
			hasAuth:  true,
			requestBody: updateSecretRequest{
				Type:     "text",
				Data:     "data",
				Metadata: `{}`,
			},
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Update", mock.Anything, "secret-999", "user-123", "text", []byte("data"), `{}`).
					Return(usecase.ErrSecretNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "невалидный тип",
			secretID: "secret-123",
			userID:   "user-123",
			hasAuth:  true,
			requestBody: updateSecretRequest{
				Type:     "invalid",
				Data:     "data",
				Metadata: `{}`,
			},
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Update", mock.Anything, "secret-123", "user-123", "invalid", []byte("data"), `{}`).
					Return(usecase.ErrInvalidSecretType)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "без авторизации",
			secretID:           "secret-123",
			userID:             "",
			hasAuth:            false,
			requestBody:        updateSecretRequest{},
			setupMocks:         func(m *MockSecretUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockSecretUsecase)
			logger := newTestLogger()
			handler := NewSecretHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/"+tt.secretID, bytes.NewBuffer(jsonBody))
			req.SetPathValue("id", tt.secretID)
			req.Header.Set("Content-Type", "application/json")

			if tt.hasAuth {
				ctx := context.WithValue(req.Context(), middleware.GetUserIDKey(), tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Act
			handler.UpdateSecret(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestSecretHandler_DeleteSecret(t *testing.T) {
	tests := []struct {
		name               string
		secretID           string
		userID             string
		hasAuth            bool
		setupMocks         func(*MockSecretUsecase)
		expectedStatusCode int
	}{
		{
			name:     "успешное удаление",
			secretID: "secret-123",
			userID:   "user-123",
			hasAuth:  true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Delete", mock.Anything, "secret-123", "user-123").Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:     "секрет не найден",
			secretID: "secret-999",
			userID:   "user-123",
			hasAuth:  true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Delete", mock.Anything, "secret-999", "user-123").Return(usecase.ErrSecretNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "без авторизации",
			secretID:           "secret-123",
			userID:             "",
			hasAuth:            false,
			setupMocks:         func(m *MockSecretUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:     "внутренняя ошибка",
			secretID: "secret-123",
			userID:   "user-123",
			hasAuth:  true,
			setupMocks: func(m *MockSecretUsecase) {
				m.On("Delete", mock.Anything, "secret-123", "user-123").Return(errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(MockSecretUsecase)
			logger := newTestLogger()
			handler := NewSecretHandler(mockUsecase, logger)

			tt.setupMocks(mockUsecase)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/"+tt.secretID, nil)
			req.SetPathValue("id", tt.secretID)

			if tt.hasAuth {
				ctx := context.WithValue(req.Context(), middleware.GetUserIDKey(), tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Act
			handler.DeleteSecret(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			mockUsecase.AssertExpectations(t)
		})
	}
}
