// Package domain содержит бизнес-модели системы GophKeeper.
package domain

import "time"

// User представляет пользователя системы.
type User struct {
	// ID - уникальный идентификатор пользователя (UUID)
	ID string
	// Email - электронная почта пользователя (уникальная)
	Email string
	// PasswordHash - bcrypt хеш пароля пользователя
	PasswordHash []byte
	// CreatedAt - дата и время регистрации пользователя
	CreatedAt time.Time
}
