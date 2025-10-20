// Package postgres реализует репозитории для PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// PostgresClient — обёртка над пулом подключений к PostgreSQL.
type PostgresClient struct {
	db *sql.DB
}

// New создаёт новый клиент PostgreSQL из DSN.
// Использует pgx как драйвер через stdlib адаптер.
func New(ctx context.Context, dsn string) (*PostgresClient, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Преобразуем pgx.Config в *sql.DB через stdlib
	db := stdlib.OpenDB(*config)

	// Проверяем соединение
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresClient{db: db}, nil
}

// DB возвращает *sql.DB для использования в репозиториях.
// Это позволяет репозиториям выполнять запросы, но не управлять жизненным циклом.
func (c *PostgresClient) DB() *sql.DB {
	return c.db
}

// Close закрывает пул подключений.
func (c *PostgresClient) Close() error {
	return c.db.Close()
}