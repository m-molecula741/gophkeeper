package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/m-molecula741/gophkeeper/internal/auth"
)

type userIDKey struct{}

func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing auth header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
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
