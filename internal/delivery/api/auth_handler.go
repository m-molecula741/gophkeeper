package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/m-molecula741/gophkeeper/internal/usecase"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authUsecase AuthUsecaseInterface
	logger      *zap.Logger
}

func NewAuthHandler(authUsecase AuthUsecaseInterface, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		logger:      logger,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode request", zap.Error(err))
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.authUsecase.Register(r.Context(), req.Email, req.Password); err != nil {
		// Бизнес-ошибки (не системные)
		if errors.Is(err, usecase.ErrUserExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		// Ошибки валидации
		if errors.Is(err, usecase.ErrEmptyEmail) ||
			errors.Is(err, usecase.ErrInvalidEmail) ||
			errors.Is(err, usecase.ErrEmptyPassword) ||
			errors.Is(err, usecase.ErrWeakPassword) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Системные ошибки
		h.logger.Error("internal error during registration", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode request", zap.Error(err))
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, err := h.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		// Неправильные credentials (бизнес-логика)
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		// Ошибки валидации
		if errors.Is(err, usecase.ErrEmptyEmail) ||
			errors.Is(err, usecase.ErrInvalidEmail) ||
			errors.Is(err, usecase.ErrEmptyPassword) ||
			errors.Is(err, usecase.ErrWeakPassword) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Системные ошибки (например, JWT generation failed)
		h.logger.Error("internal error during login", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("login successful", zap.String("email", req.Email))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse{Token: token})
}
