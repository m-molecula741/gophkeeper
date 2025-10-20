package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	CreatedAt    time.Time
}
