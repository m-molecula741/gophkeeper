package domain

import "time"

type Secret struct {
	ID        string
	UserID    string
	Type      string // "login", "text", "binary", "card"
	Metadata  string // JSON
	Data      []byte // зашифрованные данные
	CreatedAt time.Time
	UpdatedAt time.Time
}