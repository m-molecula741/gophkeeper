package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIClient_Register(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		statusCode     int
		responseBody   string
		expectedError  bool
		errorContains  string
	}{
		{
			name:          "успешная регистрация",
			email:         "test@example.com",
			password:      "password123",
			statusCode:    http.StatusCreated,
			responseBody:  "",
			expectedError: false,
		},
		{
			name:          "пользователь уже существует",
			email:         "existing@example.com",
			password:      "password",
			statusCode:    http.StatusConflict,
			responseBody:  "user already exists",
			expectedError: true,
			errorContains: "registration failed",
		},
		{
			name:          "невалидный email",
			email:         "invalid-email",
			password:      "password",
			statusCode:    http.StatusBadRequest,
			responseBody:  "invalid email format",
			expectedError: true,
			errorContains: "registration failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/register", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)
				
				var req RegisterRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.email, req.Email)
				assert.Equal(t, tt.password, req.Password)
				
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()
			
			client := NewAPIClient(server.URL, "")
			
			// Act
			err := client.Register(tt.email, tt.password)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIClient_Login(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		statusCode    int
		token         string
		expectedError bool
		errorContains string
	}{
		{
			name:          "успешный логин",
			email:         "user@example.com",
			password:      "password123",
			statusCode:    http.StatusOK,
			token:         "test-jwt-token",
			expectedError: false,
		},
		{
			name:          "неправильный пароль",
			email:         "user@example.com",
			password:      "wrong",
			statusCode:    http.StatusUnauthorized,
			token:         "",
			expectedError: true,
			errorContains: "login failed",
		},
		{
			name:          "пользователь не найден",
			email:         "notfound@example.com",
			password:      "password",
			statusCode:    http.StatusUnauthorized,
			token:         "",
			expectedError: true,
			errorContains: "login failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/login", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)
				
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					resp := LoginResponse{Token: tt.token}
					json.NewEncoder(w).Encode(resp)
				} else {
					w.Write([]byte("invalid credentials"))
				}
			}))
			defer server.Close()
			
			client := NewAPIClient(server.URL, "")
			
			// Act
			token, err := client.Login(tt.email, tt.password)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, token)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.token, token)
			}
		})
	}
}

func TestAPIClient_CreateSecret(t *testing.T) {
	tests := []struct {
		name          string
		secretType    string
		data          string
		metadata      string
		token         string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "успешное создание",
			secretType:    "login",
			data:          "user:pass",
			metadata:      `{"website":"example.com"}`,
			token:         "valid-token",
			statusCode:    http.StatusCreated,
			expectedError: false,
		},
		{
			name:          "без авторизации",
			secretType:    "text",
			data:          "data",
			metadata:      "{}",
			token:         "",
			statusCode:    http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name:          "невалидный тип",
			secretType:    "invalid",
			data:          "data",
			metadata:      "{}",
			token:         "valid-token",
			statusCode:    http.StatusBadRequest,
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/secrets", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)
				
				if tt.token != "" {
					assert.Equal(t, "Bearer "+tt.token, r.Header.Get("Authorization"))
				}
				
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusCreated {
					resp := SecretResponse{
						ID:   "secret-123",
						Type: tt.secretType,
					}
					json.NewEncoder(w).Encode(resp)
				}
			}))
			defer server.Close()
			
			client := NewAPIClient(server.URL, tt.token)
			
			// Act
			secret, err := client.CreateSecret(tt.secretType, tt.data, tt.metadata)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, secret)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, "secret-123", secret.ID)
			}
		})
	}
}

func TestAPIClient_ListSecrets(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		
		w.WriteHeader(http.StatusOK)
		secrets := []SecretResponse{
			{ID: "secret-1", Type: "login"},
			{ID: "secret-2", Type: "text"},
		}
		json.NewEncoder(w).Encode(secrets)
	}))
	defer server.Close()
	
	client := NewAPIClient(server.URL, "test-token")
	
	// Act
	secrets, err := client.ListSecrets()
	
	// Assert
	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Equal(t, "secret-1", secrets[0].ID)
	assert.Equal(t, "secret-2", secrets[1].ID)
}

func TestAPIClient_GetSecret(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets/secret-123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		
		w.WriteHeader(http.StatusOK)
		secret := SecretResponse{
			ID:   "secret-123",
			Type: "login",
			Data: "decrypted-data",
		}
		json.NewEncoder(w).Encode(secret)
	}))
	defer server.Close()
	
	client := NewAPIClient(server.URL, "test-token")
	
	// Act
	secret, err := client.GetSecret("secret-123")
	
	// Assert
	require.NoError(t, err)
	assert.Equal(t, "secret-123", secret.ID)
	assert.Equal(t, "login", secret.Type)
	assert.Equal(t, "decrypted-data", secret.Data)
}

func TestAPIClient_DeleteSecret(t *testing.T) {
	tests := []struct {
		name          string
		secretID      string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "успешное удаление",
			secretID:      "secret-123",
			statusCode:    http.StatusNoContent,
			expectedError: false,
		},
		{
			name:          "секрет не найден",
			secretID:      "secret-999",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/secrets/"+tt.secretID, r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			
			client := NewAPIClient(server.URL, "test-token")
			
			// Act
			err := client.DeleteSecret(tt.secretID)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIClient_UpdateSecret(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets/secret-123", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		
		var req SecretRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "text", req.Type)
		assert.Equal(t, "new-data", req.Data)
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	client := NewAPIClient(server.URL, "test-token")
	
	// Act
	err := client.UpdateSecret("secret-123", "text", "new-data", "{}")
	
	// Assert
	assert.NoError(t, err)
}

