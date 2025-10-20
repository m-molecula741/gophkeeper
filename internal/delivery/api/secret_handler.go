package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/m-molecula741/gophkeeper/internal/delivery/middleware"
	"github.com/m-molecula741/gophkeeper/internal/domain"
	"github.com/m-molecula741/gophkeeper/internal/usecase"
	"go.uber.org/zap"
)

// SecretUsecaseInterface определяет методы для работы с секретами
type SecretUsecaseInterface interface {
	Create(ctx context.Context, userID, secretType string, data []byte, metadata string) (*domain.Secret, error)
	GetAll(ctx context.Context, userID string) ([]*domain.Secret, error)
	GetByID(ctx context.Context, secretID, userID string) (*domain.Secret, error)
	GetDecrypted(ctx context.Context, secretID, userID string) (*domain.Secret, []byte, error)
	Update(ctx context.Context, secretID, userID, secretType string, data []byte, metadata string) error
	Delete(ctx context.Context, secretID, userID string) error
}

type SecretHandler struct {
	secretUsecase SecretUsecaseInterface
	logger        *zap.Logger
}

func NewSecretHandler(secretUsecase SecretUsecaseInterface, logger *zap.Logger) *SecretHandler {
	return &SecretHandler{
		secretUsecase: secretUsecase,
		logger:        logger,
	}
}

// Структуры запросов и ответов

type createSecretRequest struct {
	Type     string `json:"type"`
	Data     string `json:"data"`     // Base64 encoded
	Metadata string `json:"metadata"` // JSON string
}

type updateSecretRequest struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Metadata string `json:"metadata"`
}

type secretResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type secretWithDataResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Data      string `json:"data"` // Base64 encoded
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateSecret создает новый секрет
// POST /api/v1/secrets
func (h *SecretHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (установлен middleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode request", zap.Error(err))
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Декодируем данные из base64
	data := []byte(req.Data) // В реальном приложении нужно декодировать из base64

	secret, err := h.secretUsecase.Create(r.Context(), userID, req.Type, data, req.Metadata)
	if err != nil {
		// Валидационные ошибки
		if errors.Is(err, usecase.ErrInvalidSecretType) ||
			errors.Is(err, usecase.ErrEmptySecretData) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Системные ошибки
		h.logger.Error("failed to create secret", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := secretResponse{
		ID:        secret.ID,
		Type:      secret.Type,
		Metadata:  secret.Metadata,
		CreatedAt: secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetAllSecrets возвращает все секреты пользователя (без данных)
// GET /api/v1/secrets
func (h *SecretHandler) GetAllSecrets(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	secrets, err := h.secretUsecase.GetAll(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get secrets", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := make([]secretResponse, 0, len(secrets))
	for _, s := range secrets {
		response = append(response, secretResponse{
			ID:        s.ID,
			Type:      s.Type,
			Metadata:  s.Metadata,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSecret возвращает конкретный секрет с расшифрованными данными
// GET /api/v1/secrets/{id}
func (h *SecretHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "secret id is required", http.StatusBadRequest)
		return
	}

	secret, data, err := h.secretUsecase.GetDecrypted(r.Context(), secretID, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrSecretNotFound) {
			http.Error(w, "secret not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get secret", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := secretWithDataResponse{
		ID:        secret.ID,
		Type:      secret.Type,
		Data:      string(data), // В реальном приложении нужно закодировать в base64
		Metadata:  secret.Metadata,
		CreatedAt: secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateSecret обновляет секрет
// PUT /api/v1/secrets/{id}
func (h *SecretHandler) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "secret id is required", http.StatusBadRequest)
		return
	}

	var req updateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode request", zap.Error(err))
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	data := []byte(req.Data)

	err = h.secretUsecase.Update(r.Context(), secretID, userID, req.Type, data, req.Metadata)
	if err != nil {
		if errors.Is(err, usecase.ErrSecretNotFound) {
			http.Error(w, "secret not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, usecase.ErrInvalidSecretType) ||
			errors.Is(err, usecase.ErrEmptySecretData) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error("failed to update secret", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteSecret удаляет секрет
// DELETE /api/v1/secrets/{id}
func (h *SecretHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "secret id is required", http.StatusBadRequest)
		return
	}

	err = h.secretUsecase.Delete(r.Context(), secretID, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrSecretNotFound) {
			http.Error(w, "secret not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete secret", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
