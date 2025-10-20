package usecase

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmptyEmail    = errors.New("email cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrInvalidEmail  = errors.New("invalid email format")
	ErrWeakPassword  = errors.New("password must be at least 6 characters")
)

// emailRegex проверяет базовый формат email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

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
