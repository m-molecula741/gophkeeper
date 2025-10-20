// Package postgres реализует слой доступа к данным с использованием PostgreSQL.
//
// Основные компоненты:
//   - UserRepository: CRUD операции для пользователей
//   - SecretRepository: CRUD операции для секретов
//   - Client: подключение к PostgreSQL с поддержкой пула соединений
//
// Все репозитории реализуют интерфейсы из usecase пакета,
// обеспечивая инверсию зависимостей согласно Clean Architecture.
package postgres
