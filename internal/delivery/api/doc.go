// Package api предоставляет HTTP handlers для RESTful API GophKeeper.
//
// Основные handlers:
//   - AuthHandler: регистрация (POST /api/v1/register) и логин (POST /api/v1/login)
//   - SecretHandler: CRUD операции для секретов
//
// Все защищенные endpoints требуют JWT токен в Authorization header.
//
// Пример запроса:
//
//	POST /api/v1/login
//	Content-Type: application/json
//
//	{
//	  "email": "user@example.com",
//	  "password": "password123"
//	}
//
// Пример ответа:
//
//	{
//	  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	}
package api
