package api

import (
	"context"
)

type AuthUsecaseInterface interface {
    Register(ctx context.Context, email, password string) error
    Login(ctx context.Context, email, password string) (string, error)
}