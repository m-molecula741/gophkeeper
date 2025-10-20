package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient клиент для работы с API сервера
type APIClient struct {
	serverURL  string
	token      string
	httpClient *http.Client
}

// NewAPIClient создает новый API клиент
func NewAPIClient(serverURL, token string) *APIClient {
	return &APIClient{
		serverURL: serverURL,
		token:     token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest запрос на логин
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse ответ на логин
type LoginResponse struct {
	Token string `json:"token"`
}

// SecretRequest запрос для создания/обновления секрета
type SecretRequest struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Metadata string `json:"metadata"`
}

// SecretResponse ответ с информацией о секрете
type SecretResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Register регистрирует нового пользователя
func (c *APIClient) Register(email, password string) error {
	req := RegisterRequest{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.serverURL+"/api/v1/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	return nil
}

// Login выполняет вход и возвращает токен
func (c *APIClient) Login(email, password string) (string, error) {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.serverURL+"/api/v1/login",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return loginResp.Token, nil
}

// CreateSecret создает новый секрет
func (c *APIClient) CreateSecret(secretType, data, metadata string) (*SecretResponse, error) {
	req := SecretRequest{
		Type:     secretType,
		Data:     data,
		Metadata: metadata,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.serverURL+"/api/v1/secrets", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create secret: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	var secret SecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &secret, nil
}

// ListSecrets возвращает список всех секретов
func (c *APIClient) ListSecrets() ([]SecretResponse, error) {
	httpReq, err := http.NewRequest(http.MethodGet, c.serverURL+"/api/v1/secrets", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list secrets: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	var secrets []SecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return secrets, nil
}

// GetSecret получает конкретный секрет с расшифрованными данными
func (c *APIClient) GetSecret(secretID string) (*SecretResponse, error) {
	httpReq, err := http.NewRequest(http.MethodGet, c.serverURL+"/api/v1/secrets/"+secretID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get secret: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	var secret SecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &secret, nil
}

// UpdateSecret обновляет существующий секрет
func (c *APIClient) UpdateSecret(secretID, secretType, data, metadata string) error {
	req := SecretRequest{
		Type:     secretType,
		Data:     data,
		Metadata: metadata,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPut, c.serverURL+"/api/v1/secrets/"+secretID, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update secret: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	return nil
}

// DeleteSecret удаляет секрет
func (c *APIClient) DeleteSecret(secretID string) error {
	httpReq, err := http.NewRequest(http.MethodDelete, c.serverURL+"/api/v1/secrets/"+secretID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete secret: %s (status: %d)", string(bodyBytes), resp.StatusCode)
	}

	return nil
}
