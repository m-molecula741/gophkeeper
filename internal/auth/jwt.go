// Package auth предоставляет функции аутентификации и авторизации пользователей,
// включая управление JWT токенами и хеширование паролей.
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager управляет созданием и проверкой JWT токенов.
type JWTManager struct {
	secretKey string
	tokenTTL  time.Duration
}

// NewJWTManager создает новый экземпляр JWTManager.
// Параметр secretKey используется для подписи токенов (HS256).
// Параметр tokenTTL определяет время жизни создаваемых токенов.
func NewJWTManager(secretKey string, tokenTTL time.Duration) *JWTManager {
	return &JWTManager{secretKey: secretKey, tokenTTL: tokenTTL}
}

// UserClaims представляет claims (утверждения) JWT токена пользователя.
// Содержит идентификатор пользователя и стандартные registered claims.
type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken создает новый JWT токен для указанного пользователя.
// Параметр userID - уникальный идентификатор пользователя.
// Возвращает подписанный JWT токен в виде строки.
func (m *JWTManager) GenerateToken(userID string) (string, error) {
	claims := &UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// VerifyToken проверяет валидность JWT токена и извлекает claims.
// Параметр tokenStr - JWT токен в виде строки.
// Возвращает UserClaims если токен валиден, иначе ошибку.
func (m *JWTManager) VerifyToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
