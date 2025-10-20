// Package middleware предоставляет HTTP middleware для GophKeeper.
//
// Основные компоненты:
//   - AuthMiddleware: проверка JWT токена и извлечение userID в контекст
//   - GetUserID: извлечение userID из контекста запроса
//
// Пример использования:
//
//	authMW := middleware.AuthMiddleware(jwtManager)
//	mux.Handle("/api/v1/secrets", authMW(http.HandlerFunc(handler.GetAllSecrets)))
//
//	// В handler'е:
//	userID, err := middleware.GetUserID(r.Context())
package middleware
