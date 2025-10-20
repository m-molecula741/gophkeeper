// Package client предоставляет HTTP клиент для взаимодействия с GophKeeper API
// и функции управления конфигурацией CLI приложения.
//
// Основные компоненты:
//   - APIClient: HTTP клиент для всех API endpoints (регистрация, логин, CRUD секретов)
//   - Config: управление конфигурацией и JWT токеном
//
// Пример использования:
//
//	cfg, err := client.LoadConfig("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	apiClient := client.NewAPIClient(cfg.ServerURL)
//	token, err := apiClient.Login("user@example.com", "password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg.SetToken(token)
//	cfg.Save()
package client
