package usecase

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// Auth validation errors
	ErrEmptyEmail    = errors.New("email cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrInvalidEmail  = errors.New("invalid email format")
	ErrWeakPassword  = errors.New("password must be at least 6 characters")

	// Secret validation errors
	ErrInvalidSecretType = errors.New("invalid secret type (must be: login, text, binary, card)")
	ErrEmptySecretData   = errors.New("secret data cannot be empty")
	ErrInvalidMetadata   = errors.New("invalid metadata format (must be valid JSON)")
)

// emailRegex проверяет базовый формат email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Допустимые типы секретов
var validSecretTypes = map[string]bool{
	"login":  true,
	"text":   true,
	"binary": true,
	"card":   true,
}

// ValidateEmail проверяет корректность email
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmptyEmail
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword проверяет корректность пароля
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 6 {
		return ErrWeakPassword
	}
	return nil
}

// ValidateSecretType проверяет, что тип секрета допустимый
func ValidateSecretType(secretType string) error {
	if !validSecretTypes[secretType] {
		return ErrInvalidSecretType
	}
	return nil
}

// ValidateSecretData проверяет, что данные секрета не пустые
func ValidateSecretData(data []byte) error {
	if len(data) == 0 {
		return ErrEmptySecretData
	}
	return nil
}
