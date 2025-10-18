-- Создаём расширение для UUID (если ещё не создано)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL CHECK (email <> ''),
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска по email
CREATE INDEX idx_users_email ON users (email);

-- Таблица секретов
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('login', 'text', 'binary', 'card')),
    metadata TEXT NOT NULL DEFAULT '{}',
    data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для выборки всех секретов пользователя
CREATE INDEX idx_secrets_user_id ON secrets (user_id);