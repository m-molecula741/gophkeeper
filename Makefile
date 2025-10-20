.PHONY: help test test-verbose test-cover test-cover-html lint run build clean migrate-up migrate-down deps

# Цвета для вывода
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

help: ## Показать эту справку
	@echo "GophKeeper - Менеджер паролей"
	@echo ""
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(RESET)"
	go mod download
	go mod tidy

test: ## Запустить все тесты
	@echo "$(GREEN)Запуск тестов...$(RESET)"
	go test ./...

test-verbose: ## Запустить тесты с подробным выводом
	@echo "$(GREEN)Запуск тестов (verbose)...$(RESET)"
	go test ./... -v

test-cover: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(RESET)"
	go test ./... -cover

test-cover-html: ## Сгенерировать HTML отчет о покрытии
	@echo "$(GREEN)Генерация HTML отчета о покрытии...$(RESET)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет сохранен в coverage.html$(RESET)"

test-usecase: ## Запустить только usecase тесты
	@echo "$(GREEN)Запуск usecase тестов...$(RESET)"
	go test ./internal/usecase/... -v

test-handler: ## Запустить только handler тесты
	@echo "$(GREEN)Запуск handler тестов...$(RESET)"
	go test ./internal/delivery/api/... -v

lint: ## Запустить линтер
	@echo "$(GREEN)Запуск линтера...$(RESET)"
	golangci-lint run ./...

run: ## Запустить сервер
	@echo "$(GREEN)Запуск сервера...$(RESET)"
	go run cmd/server/main.go

build: ## Собрать бинарник сервера
	@echo "$(GREEN)Сборка сервера...$(RESET)"
	go build -o bin/server cmd/server/main.go
	@echo "$(GREEN)Бинарник сохранен в bin/server$(RESET)"

build-client: ## Собрать бинарник клиента
	@echo "$(GREEN)Сборка клиента...$(RESET)"
	go build -ldflags="-X 'main.version=$(shell git describe --tags --always --dirty)' -X 'main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'" -o bin/gophkeeper cmd/client/main.go
	@echo "$(GREEN)Бинарник сохранен в bin/gophkeeper$(RESET)"

build-all: ## Собрать сервер и клиент
	@echo "$(GREEN)Сборка всех компонентов...$(RESET)"
	@$(MAKE) build
	@$(MAKE) build-client
	@echo "$(GREEN)Сборка завершена!$(RESET)"

clean: ## Очистить временные файлы
	@echo "$(GREEN)Очистка...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

migrate-up: ## Применить миграции
	@echo "$(GREEN)Применение миграций...$(RESET)"
	migrate -path migrations -database "$(DB_DSN)" up

migrate-down: ## Откатить миграции
	@echo "$(YELLOW)Откат миграций...$(RESET)"
	migrate -path migrations -database "$(DB_DSN)" down

migrate-create: ## Создать новую миграцию (make migrate-create name=add_users_table)
	@echo "$(GREEN)Создание миграции $(name)...$(RESET)"
	migrate create -ext sql -dir migrations -seq $(name)

db-setup: ## Создать БД и применить миграции
	@echo "$(GREEN)Создание БД gophkeeper...$(RESET)"
	createdb gophkeeper || true
	@echo "$(GREEN)Применение миграций...$(RESET)"
	migrate -path migrations -database "postgresql://postgres:741852963@localhost:5432/gophkeeper?sslmode=disable" up

docker-up: ## Запустить PostgreSQL в Docker
	@echo "$(GREEN)Запуск PostgreSQL...$(RESET)"
	docker run --name gophkeeper-postgres \
		-e POSTGRES_PASSWORD=741852963 \
		-e POSTGRES_DB=gophkeeper \
		-p 5432:5432 \
		-d postgres:14

docker-down: ## Остановить PostgreSQL Docker контейнер
	@echo "$(YELLOW)Остановка PostgreSQL...$(RESET)"
	docker stop gophkeeper-postgres
	docker rm gophkeeper-postgres

# Примеры использования
examples: ## Показать примеры API запросов
	@echo "$(GREEN)Примеры API запросов:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Регистрация:$(RESET)"
	@echo 'curl -X POST http://localhost:8080/api/v1/register \\'
	@echo '  -H "Content-Type: application/json" \\'
	@echo '  -d '"'"'{"email": "user@example.com", "password": "password123"}'"'"''
	@echo ""
	@echo "$(YELLOW)Логин:$(RESET)"
	@echo 'curl -X POST http://localhost:8080/api/v1/login \\'
	@echo '  -H "Content-Type: application/json" \\'
	@echo '  -d '"'"'{"email": "user@example.com", "password": "password123"}'"'"''
	@echo ""
	@echo "$(YELLOW)Использование токена:$(RESET)"
	@echo 'curl http://localhost:8080/api/v1/secrets \\'
	@echo '  -H "Authorization: Bearer YOUR_TOKEN_HERE"'

# По умолчанию показываем справку
.DEFAULT_GOAL := help

