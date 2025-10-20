package domain

import "time"

// Secret представляет приватные данные пользователя.
// Поддерживает хранение логинов/паролей, текста, бинарных данных и банковских карт.
type Secret struct {
	// ID - уникальный идентификатор секрета (UUID)
	ID string
	// UserID - идентификатор владельца секрета
	UserID string
	// Type - тип секрета: "login", "text", "binary", "card"
	Type string
	// Metadata - произвольная метаинформация в формате JSON
	Metadata string
	// Data - зашифрованные данные секрета (AES-256-GCM)
	Data []byte
	// CreatedAt - дата и время создания секрета
	CreatedAt time.Time
	// UpdatedAt - дата и время последнего изменения
	UpdatedAt time.Time
}
