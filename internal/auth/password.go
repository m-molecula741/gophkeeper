package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const PasswordCost = 12

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
}

func CheckPassword(password string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}