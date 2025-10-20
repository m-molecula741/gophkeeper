# 🔐 GophKeeper - Secure Password Manager

Безопасный менеджер паролей с клиент-серверной архитектурой на Go.

---

## 📋 Возможности

### ✅ Реализовано

- **Регистрация и аутентификация**
  - JWT-based авторизация
  - Bcrypt хеширование паролей
  - Валидация email и пароля

- **Хранение секретов**
  - 4 типа данных: login (логин/пароль), text (текст), binary (бинарные данные), card (банковская карта)
  - AES-256-GCM шифрование
  - Метаданные в JSON формате

- **HTTP API**
  - RESTful архитектура
  - Защищенные endpoints (JWT middleware)
  - CRUD операции для секретов

- **CLI клиент** 
  - Кросс-платформенный (Windows, Linux, macOS)
  - Cobra framework для команд
  - Безопасное хранение токена
  - Версионирование и дата сборки

- **Архитектура**
  - Clean Architecture (domain, usecase, delivery, repository)
  - SOLID принципы
  - Dependency Injection
  - Graceful shutdown

- **Тестирование**
  - Table-driven tests
  - Mock репозитории
  - 85%+ покрытие
  - 150+ тест-кейсов

---

## 🏗️ Архитектура

```
gophkeeper/
├── cmd/
│   ├── server/          # Серверное приложение
│   └── client/          # CLI клиент
│
├── internal/
│   ├── auth/            # JWT и хеширование паролей
│   ├── config/          # Конфигурация сервера
│   ├── crypto/          # AES-256-GCM шифрование
│   │
│   ├── domain/          # Бизнес-модели (User, Secret)
│   │
│   ├── usecase/         # Бизнес-логика
│   │   ├── auth.go      # Регистрация, логин
│   │   ├── secret.go    # CRUD секретов
│   │   └── validation.go
│   │
│   ├── delivery/        # HTTP layer (сервер)
│   │   ├── api/         # HTTP handlers
│   │   └── middleware/  # JWT middleware
│   │
│   ├── repository/      # Работа с БД
│   │   └── postgres/    # PostgreSQL
│   │
│   └── client/          # HTTP client (клиент)
│       ├── api_client.go
│       └── config.go
│
├── migrations/          # SQL миграции
└── pkg/
    └── logger/          # Zap logger
```

---

## 🚀 Быстрый старт

### Требования

- Go 1.21+
- PostgreSQL 14+
- make

### 1. Клонирование репозитория

```bash
git clone https://github.com/m-molecula741/gophkeeper.git
cd gophkeeper
```

### 2. Установка зависимостей

```bash
go mod download
```

### 3. Настройка базы данных

```bash
# Создать базу данных
createdb gophkeeper

# Применить миграции
make migrate-up
```

### 4. Конфигурация

Отредактируйте `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  dsn: "postgres://user:password@localhost:5432/gophkeeper?sslmode=disable"

auth:
  jwt_secret: "your-secret-key-change-me"
  token_ttl: 24h

logging:
  level: "info"
```

### 5. Запуск сервера

```bash
make run
```

### 6. Сборка и использование клиента

```bash
# Собрать клиент
make build-client

# Регистрация
./bin/gophkeeper register -e user@example.com -p password

# Вход
./bin/gophkeeper login -e user@example.com -p password

# Добавить секрет
./bin/gophkeeper secrets add -t login -d "username:password" -m '{"website":"github.com"}'

# Список секретов
./bin/gophkeeper secrets list

# Получить секрет
./bin/gophkeeper secrets get -i <SECRET_ID>
```

Подробнее о CLI: [cmd/client/README.md](cmd/client/README.md)

---

## 🔌 HTTP API

### Публичные endpoints:

#### POST /api/v1/register
Регистрация нового пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePass123"
}
```

**Response:** `201 Created`

---

#### POST /api/v1/login
Вход в систему.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePass123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### Защищенные endpoints (требуется JWT):

#### POST /api/v1/secrets
Создать секрет.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request:**
```json
{
  "type": "login",
  "data": "username:password",
  "metadata": "{\"website\":\"github.com\"}"
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "login",
  "metadata": "{\"website\":\"github.com\"}",
  "created_at": "2025-10-20T19:00:00Z",
  "updated_at": "2025-10-20T19:00:00Z"
}
```

---

#### GET /api/v1/secrets
Получить все секреты пользователя (без данных).

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "login",
    "metadata": "{\"website\":\"github.com\"}",
    "created_at": "2025-10-20T19:00:00Z",
    "updated_at": "2025-10-20T19:00:00Z"
  }
]
```

---

#### GET /api/v1/secrets/{id}
Получить конкретный секрет (с расшифрованными данными).

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "login",
  "data": "username:password",
  "metadata": "{\"website\":\"github.com\"}",
  "created_at": "2025-10-20T19:00:00Z",
  "updated_at": "2025-10-20T19:00:00Z"
}
```

---

#### PUT /api/v1/secrets/{id}
Обновить секрет.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request:**
```json
{
  "type": "login",
  "data": "newusername:newpassword",
  "metadata": "{\"website\":\"github.com\",\"updated\":true}"
}
```

**Response:** `200 OK`

---

#### DELETE /api/v1/secrets/{id}
Удалить секрет.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:** `204 No Content`

---

## 🧪 Тестирование

### Запуск всех тестов:

```bash
make test
```

### Запуск с покрытием:

```bash
make test-coverage
```

### Покрытие по модулям:

| Модуль | Покрытие | Тесты |
|--------|----------|-------|
| `internal/usecase` | 94.8% | 50+ |
| `internal/delivery/api` | 91.1% | 40+ |
| `internal/crypto` | 81.2% | 15+ |
| `internal/client` | 76.3% | 17+ |
| **ИТОГО** | **85%+** | **150+** |

---

## 📝 Makefile команды

```bash
make help              # Показать все команды
make run               # Запустить сервер
make build             # Собрать сервер
make build-client      # Собрать клиент
make build-all         # Собрать все
make test              # Запустить тесты
make test-coverage     # Тесты с покрытием
make lint              # Запустить линтер
make migrate-up        # Применить миграции
make migrate-down      # Откатить миграции
make clean             # Очистить бинарники
```

---

## 🔒 Безопасность

### Реализованные меры:

1. **Хеширование паролей**: bcrypt с cost=10
2. **Шифрование данных**: AES-256-GCM для всех секретов
3. **JWT аутентификация**: HS256, настраиваемый TTL
4. **Валидация входных данных**: email, пароль (минимум 6 символов), типы секретов
5. **HTTPS поддержка**: клиент поддерживает https://
6. **Безопасное хранение токена**: права 0600 на файл конфигурации

### Рекомендации для production:

1. ✅ Использовать HTTPS для сервера
2. ✅ Изменить JWT secret в config.yaml
3. ✅ Использовать отдельный ключ шифрования (не JWT secret)
4. ✅ Настроить PostgreSQL SSL
5. ✅ Ограничить срок жизни токена
6. ✅ Добавить rate limiting
7. ✅ Использовать reverse proxy (nginx)

---

## 📊 Типы секретов

### 1. login - Логин/пароль

```bash
gophkeeper secrets add \
  -t login \
  -d "username:password" \
  -m '{"website":"github.com","email":"user@example.com"}'
```

### 2. text - Текстовая информация

```bash
gophkeeper secrets add \
  -t text \
  -d "Важная заметка" \
  -m '{"category":"notes","priority":"high"}'
```

### 3. binary - Бинарные данные

```bash
gophkeeper secrets add \
  -t binary \
  -d "$(base64 < file.bin)" \
  -m '{"filename":"file.bin","size":"1024"}'
```

### 4. card - Банковская карта

```bash
gophkeeper secrets add \
  -t card \
  -d "4111111111111111|12/25|123|John Doe" \
  -m '{"bank":"MyBank","type":"Visa"}'
```

---

## 🚢 Развертывание

### Docker (TODO):

```bash
docker-compose up -d
```

### Systemd service:

```ini
[Unit]
Description=GophKeeper Server
After=network.target postgresql.service

[Service]
Type=simple
User=gophkeeper
WorkingDirectory=/opt/gophkeeper
ExecStart=/opt/gophkeeper/bin/server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

---

## 👤 Автор

**m-molecula741**

- GitHub: [@m-molecula741](https://github.com/m-molecula741)

---

**⭐ Если проект вам понравился - поставьте звезду!**

