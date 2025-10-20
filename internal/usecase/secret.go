package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/m-molecula741/gophkeeper/internal/domain"
	"go.uber.org/zap"
)

// Encryptor определяет интерфейс для шифрования/дешифрования данных
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// SecretUsecase реализует бизнес-логику работы с секретами
type SecretUsecase struct {
	secretRepo SecretRepository
	encryptor  Encryptor
	logger     *zap.Logger
}

// NewSecretUsecase создает новый usecase для работы с секретами
func NewSecretUsecase(secretRepo SecretRepository, encryptor Encryptor, logger *zap.Logger) *SecretUsecase {
	return &SecretUsecase{
		secretRepo: secretRepo,
		encryptor:  encryptor,
		logger:     logger,
	}
}

// Create создает новый секрет
func (u *SecretUsecase) Create(ctx context.Context, userID, secretType string, data []byte, metadata string) (*domain.Secret, error) {
	// Валидация
	if err := ValidateSecretType(secretType); err != nil {
		return nil, err
	}
	if err := ValidateSecretData(data); err != nil {
		return nil, err
	}

	// Шифруем данные
	encryptedData, err := u.encryptor.Encrypt(data)
	if err != nil {
		u.logger.Error("failed to encrypt secret data", zap.Error(err))
		return nil, err
	}

	now := time.Now()
	secret := &domain.Secret{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      secretType,
		Metadata:  metadata,
		Data:      encryptedData,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.secretRepo.Create(ctx, secret); err != nil {
		u.logger.Error("failed to create secret", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	u.logger.Info("secret created successfully",
		zap.String("secret_id", secret.ID),
		zap.String("user_id", userID),
		zap.String("type", secretType))

	return secret, nil
}

// GetAll возвращает все секреты пользователя
func (u *SecretUsecase) GetAll(ctx context.Context, userID string) ([]*domain.Secret, error) {
	secrets, err := u.secretRepo.GetByUserID(ctx, userID)
	if err != nil {
		u.logger.Error("failed to get secrets", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	u.logger.Debug("retrieved secrets", zap.String("user_id", userID), zap.Int("count", len(secrets)))
	return secrets, nil
}

// GetByID возвращает секрет по ID (только если принадлежит пользователю)
func (u *SecretUsecase) GetByID(ctx context.Context, secretID, userID string) (*domain.Secret, error) {
	secret, err := u.secretRepo.GetByID(ctx, secretID, userID)
	if err != nil {
		u.logger.Debug("secret not found",
			zap.String("secret_id", secretID),
			zap.String("user_id", userID))
		return nil, ErrSecretNotFound
	}

	u.logger.Debug("retrieved secret",
		zap.String("secret_id", secretID),
		zap.String("user_id", userID))
	return secret, nil
}

// GetDecrypted возвращает секрет с расшифрованными данными
func (u *SecretUsecase) GetDecrypted(ctx context.Context, secretID, userID string) (*domain.Secret, []byte, error) {
	secret, err := u.GetByID(ctx, secretID, userID)
	if err != nil {
		return nil, nil, err
	}

	// Дешифруем данные
	decryptedData, err := u.encryptor.Decrypt(secret.Data)
	if err != nil {
		u.logger.Error("failed to decrypt secret data",
			zap.Error(err),
			zap.String("secret_id", secretID))
		return nil, nil, err
	}

	return secret, decryptedData, nil
}

// Update обновляет секрет
func (u *SecretUsecase) Update(ctx context.Context, secretID, userID, secretType string, data []byte, metadata string) error {
	// Проверяем существование секрета
	existingSecret, err := u.secretRepo.GetByID(ctx, secretID, userID)
	if err != nil {
		u.logger.Debug("secret not found for update",
			zap.String("secret_id", secretID),
			zap.String("user_id", userID))
		return ErrSecretNotFound
	}

	// Валидация
	if err := ValidateSecretType(secretType); err != nil {
		return err
	}
	if err := ValidateSecretData(data); err != nil {
		return err
	}

	// Шифруем новые данные
	encryptedData, err := u.encryptor.Encrypt(data)
	if err != nil {
		u.logger.Error("failed to encrypt secret data", zap.Error(err))
		return err
	}

	// Обновляем секрет
	existingSecret.Type = secretType
	existingSecret.Metadata = metadata
	existingSecret.Data = encryptedData
	existingSecret.UpdatedAt = time.Now()

	if err := u.secretRepo.Update(ctx, existingSecret); err != nil {
		u.logger.Error("failed to update secret",
			zap.Error(err),
			zap.String("secret_id", secretID))
		return err
	}

	u.logger.Info("secret updated successfully",
		zap.String("secret_id", secretID),
		zap.String("user_id", userID))

	return nil
}

// Delete удаляет секрет
func (u *SecretUsecase) Delete(ctx context.Context, secretID, userID string) error {
	if err := u.secretRepo.Delete(ctx, secretID, userID); err != nil {
		u.logger.Debug("failed to delete secret",
			zap.String("secret_id", secretID),
			zap.String("user_id", userID))
		return ErrSecretNotFound
	}

	u.logger.Info("secret deleted successfully",
		zap.String("secret_id", secretID),
		zap.String("user_id", userID))

	return nil
}
