package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m-molecula741/gophkeeper/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecretRepository — мок репозитория секретов
type MockSecretRepository struct {
	mock.Mock
}

func (m *MockSecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *MockSecretRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Secret, error) {
	args := m.Called(ctx, userID)
	if secrets, ok := args.Get(0).([]*domain.Secret); ok {
		return secrets, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSecretRepository) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	args := m.Called(ctx, id, userID)
	if secret, ok := args.Get(0).(*domain.Secret); ok {
		return secret, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSecretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *MockSecretRepository) Delete(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// MockEncryptor — мок для шифрования/дешифрования
type MockEncryptor struct {
	mock.Mock
}

func (m *MockEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	args := m.Called(plaintext)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	args := m.Called(ciphertext)
	if data, ok := args.Get(0).([]byte); ok {
		return data, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestSecretUsecase_Create(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		secretType    string
		data          []byte
		metadata      string
		setupMocks    func(*MockSecretRepository, *MockEncryptor)
		expectedError error
	}{
		{
			name:       "успешное создание секрета login",
			userID:     "user-123",
			secretType: "login",
			data:       []byte("username:password"),
			metadata:   `{"website":"example.com"}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				enc.On("Encrypt", []byte("username:password")).Return([]byte("encrypted-data"), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Secret) bool {
					return s.UserID == "user-123" && s.Type == "login" && len(s.ID) > 0
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "успешное создание секрета text",
			userID:     "user-456",
			secretType: "text",
			data:       []byte("my secret note"),
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				enc.On("Encrypt", []byte("my secret note")).Return([]byte("encrypted-note"), nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "невалидный тип секрета",
			userID:     "user-123",
			secretType: "invalid-type",
			data:       []byte("some data"),
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedError: ErrInvalidSecretType,
		},
		{
			name:       "пустые данные",
			userID:     "user-123",
			secretType: "login",
			data:       []byte{},
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				// Моки не нужны - валидация произойдет раньше
			},
			expectedError: ErrEmptySecretData,
		},
		{
			name:       "ошибка шифрования",
			userID:     "user-123",
			secretType: "login",
			data:       []byte("data"),
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				enc.On("Encrypt", []byte("data")).Return([]byte{}, errors.New("encryption failed"))
			},
			expectedError: errors.New("encryption failed"),
		},
		{
			name:       "ошибка сохранения в БД",
			userID:     "user-123",
			secretType: "card",
			data:       []byte("4111111111111111"),
			metadata:   `{"bank":"Test Bank"}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				enc.On("Encrypt", mock.Anything).Return([]byte("encrypted"), nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockSecretRepository)
			mockEnc := new(MockEncryptor)
			logger := newTestLogger()

			tt.setupMocks(mockRepo, mockEnc)

			usecase := NewSecretUsecase(mockRepo, mockEnc, logger)

			// Act
			secret, err := usecase.Create(context.Background(), tt.userID, tt.secretType, tt.data, tt.metadata)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrInvalidSecretType) ||
					errors.Is(tt.expectedError, ErrEmptySecretData) {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
				assert.Nil(t, secret)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, tt.userID, secret.UserID)
				assert.Equal(t, tt.secretType, secret.Type)
				assert.NotEmpty(t, secret.ID)
			}

			mockRepo.AssertExpectations(t)
			mockEnc.AssertExpectations(t)
		})
	}
}

func TestSecretUsecase_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*MockSecretRepository)
		expectedCount int
		expectedError error
	}{
		{
			name:   "успешное получение нескольких секретов",
			userID: "user-123",
			setupMocks: func(repo *MockSecretRepository) {
				secrets := []*domain.Secret{
					{ID: "secret-1", UserID: "user-123", Type: "login"},
					{ID: "secret-2", UserID: "user-123", Type: "text"},
				}
				repo.On("GetByUserID", mock.Anything, "user-123").Return(secrets, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:   "пользователь без секретов",
			userID: "user-456",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("GetByUserID", mock.Anything, "user-456").Return([]*domain.Secret{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name:   "ошибка при получении",
			userID: "user-789",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("GetByUserID", mock.Anything, "user-789").Return(([]*domain.Secret)(nil), errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockSecretRepository)
			mockEnc := new(MockEncryptor)
			logger := newTestLogger()

			tt.setupMocks(mockRepo)

			usecase := NewSecretUsecase(mockRepo, mockEnc, logger)

			// Act
			secrets, err := usecase.GetAll(context.Background(), tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, secrets)
			} else {
				assert.NoError(t, err)
				assert.Len(t, secrets, tt.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSecretUsecase_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		secretID      string
		userID        string
		setupMocks    func(*MockSecretRepository)
		expectedError error
	}{
		{
			name:     "успешное получение секрета",
			secretID: "secret-123",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository) {
				secret := &domain.Secret{
					ID:     "secret-123",
					UserID: "user-123",
					Type:   "login",
				}
				repo.On("GetByID", mock.Anything, "secret-123", "user-123").Return(secret, nil)
			},
			expectedError: nil,
		},
		{
			name:     "секрет не найден",
			secretID: "secret-999",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("GetByID", mock.Anything, "secret-999", "user-123").Return((*domain.Secret)(nil), errors.New("not found"))
			},
			expectedError: ErrSecretNotFound,
		},
		{
			name:     "попытка доступа к чужому секрету",
			secretID: "secret-123",
			userID:   "user-456",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("GetByID", mock.Anything, "secret-123", "user-456").Return((*domain.Secret)(nil), errors.New("not found"))
			},
			expectedError: ErrSecretNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockSecretRepository)
			mockEnc := new(MockEncryptor)
			logger := newTestLogger()

			tt.setupMocks(mockRepo)

			usecase := NewSecretUsecase(mockRepo, mockEnc, logger)

			// Act
			secret, err := usecase.GetByID(context.Background(), tt.secretID, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, secret)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, tt.secretID, secret.ID)
				assert.Equal(t, tt.userID, secret.UserID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSecretUsecase_GetDecrypted(t *testing.T) {
	encryptedData := []byte("encrypted-data")
	decryptedData := []byte("decrypted-data")

	tests := []struct {
		name          string
		secretID      string
		userID        string
		setupMocks    func(*MockSecretRepository, *MockEncryptor)
		expectedData  []byte
		expectedError error
	}{
		{
			name:     "успешное получение и расшифровка",
			secretID: "secret-123",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				secret := &domain.Secret{
					ID:     "secret-123",
					UserID: "user-123",
					Type:   "login",
					Data:   encryptedData,
				}
				repo.On("GetByID", mock.Anything, "secret-123", "user-123").Return(secret, nil)
				enc.On("Decrypt", encryptedData).Return(decryptedData, nil)
			},
			expectedData:  decryptedData,
			expectedError: nil,
		},
		{
			name:     "секрет не найден",
			secretID: "secret-999",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				repo.On("GetByID", mock.Anything, "secret-999", "user-123").Return((*domain.Secret)(nil), errors.New("not found"))
			},
			expectedData:  nil,
			expectedError: ErrSecretNotFound,
		},
		{
			name:     "ошибка расшифровки",
			secretID: "secret-123",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				secret := &domain.Secret{
					ID:     "secret-123",
					UserID: "user-123",
					Data:   encryptedData,
				}
				repo.On("GetByID", mock.Anything, "secret-123", "user-123").Return(secret, nil)
				enc.On("Decrypt", encryptedData).Return([]byte(nil), errors.New("decryption failed"))
			},
			expectedData:  nil,
			expectedError: errors.New("decryption failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockSecretRepository)
			mockEnc := new(MockEncryptor)
			logger := newTestLogger()

			tt.setupMocks(mockRepo, mockEnc)

			usecase := NewSecretUsecase(mockRepo, mockEnc, logger)

			// Act
			secret, data, err := usecase.GetDecrypted(context.Background(), tt.secretID, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, secret)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, tt.expectedData, data)
			}

			mockRepo.AssertExpectations(t)
			mockEnc.AssertExpectations(t)
		})
	}
}

func TestSecretUsecase_Update(t *testing.T) {
	existingSecret := &domain.Secret{
		ID:        "secret-123",
		UserID:    "user-123",
		Type:      "login",
		Data:      []byte("old-encrypted"),
		Metadata:  `{"old":"metadata"}`,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	tests := []struct {
		name          string
		secretID      string
		userID        string
		secretType    string
		data          []byte
		metadata      string
		setupMocks    func(*MockSecretRepository, *MockEncryptor)
		expectedError error
	}{
		{
			name:       "успешное обновление",
			secretID:   "secret-123",
			userID:     "user-123",
			secretType: "text",
			data:       []byte("new data"),
			metadata:   `{"new":"metadata"}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				repo.On("GetByID", mock.Anything, "secret-123", "user-123").Return(existingSecret, nil)
				enc.On("Encrypt", []byte("new data")).Return([]byte("new-encrypted"), nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(s *domain.Secret) bool {
					return s.ID == "secret-123" && s.Type == "text"
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "секрет не найден",
			secretID:   "secret-999",
			userID:     "user-123",
			secretType: "login",
			data:       []byte("data"),
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				repo.On("GetByID", mock.Anything, "secret-999", "user-123").Return((*domain.Secret)(nil), errors.New("not found"))
			},
			expectedError: ErrSecretNotFound,
		},
		{
			name:       "невалидный тип",
			secretID:   "secret-123",
			userID:     "user-123",
			secretType: "invalid",
			data:       []byte("data"),
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				repo.On("GetByID", mock.Anything, "secret-123", "user-123").Return(existingSecret, nil)
			},
			expectedError: ErrInvalidSecretType,
		},
		{
			name:       "пустые данные",
			secretID:   "secret-123",
			userID:     "user-123",
			secretType: "login",
			data:       []byte{},
			metadata:   `{}`,
			setupMocks: func(repo *MockSecretRepository, enc *MockEncryptor) {
				repo.On("GetByID", mock.Anything, "secret-123", "user-123").Return(existingSecret, nil)
			},
			expectedError: ErrEmptySecretData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockSecretRepository)
			mockEnc := new(MockEncryptor)
			logger := newTestLogger()

			tt.setupMocks(mockRepo, mockEnc)

			usecase := NewSecretUsecase(mockRepo, mockEnc, logger)

			// Act
			err := usecase.Update(context.Background(), tt.secretID, tt.userID, tt.secretType, tt.data, tt.metadata)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrSecretNotFound) ||
					errors.Is(tt.expectedError, ErrInvalidSecretType) ||
					errors.Is(tt.expectedError, ErrEmptySecretData) {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockEnc.AssertExpectations(t)
		})
	}
}

func TestSecretUsecase_Delete(t *testing.T) {
	tests := []struct {
		name          string
		secretID      string
		userID        string
		setupMocks    func(*MockSecretRepository)
		expectedError error
	}{
		{
			name:     "успешное удаление",
			secretID: "secret-123",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("Delete", mock.Anything, "secret-123", "user-123").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "секрет не найден",
			secretID: "secret-999",
			userID:   "user-123",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("Delete", mock.Anything, "secret-999", "user-123").Return(errors.New("not found"))
			},
			expectedError: ErrSecretNotFound,
		},
		{
			name:     "попытка удалить чужой секрет",
			secretID: "secret-123",
			userID:   "user-456",
			setupMocks: func(repo *MockSecretRepository) {
				repo.On("Delete", mock.Anything, "secret-123", "user-456").Return(errors.New("not found"))
			},
			expectedError: ErrSecretNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockSecretRepository)
			mockEnc := new(MockEncryptor)
			logger := newTestLogger()

			tt.setupMocks(mockRepo)

			usecase := NewSecretUsecase(mockRepo, mockEnc, logger)

			// Act
			err := usecase.Delete(context.Background(), tt.secretID, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
