package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/m-molecula741/gophkeeper/internal/auth"
)

type userIDKey struct{}

var ErrUserIDNotFound = errors.New("user ID not found in context")

// AuthMiddleware проверяет JWT токен и добавляет userID в контекст
func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing auth header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				// Bearer prefix не найден
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.VerifyToken(tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Передаём userID в контекст
			ctx := context.WithValue(r.Context(), userIDKey{}, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID извлекает userID из контекста запроса
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok || userID == "" {
		return "", ErrUserIDNotFound
	}
	return userID, nil
}
