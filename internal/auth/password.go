package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordCost определяет сложность алгоритма bcrypt для хеширования паролей.
// Значение 12 обеспечивает баланс между безопасностью и производительностью.
const PasswordCost = 12

// HashPassword создает bcrypt хеш пароля.
// Параметр password - пароль в открытом виде.
// Возвращает bcrypt хеш, который можно безопасно хранить в базе данных.
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
}

// CheckPassword проверяет соответствие пароля и его хеша.
// Параметр password - пароль в открытом виде.
// Параметр hash - bcrypt хеш пароля из базы данных.
// Возвращает nil если пароль корректен, иначе ошибку.
func CheckPassword(password string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}
