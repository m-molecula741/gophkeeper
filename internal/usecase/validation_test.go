package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		expectedError error
	}{
		{
			name:          "валидный email",
			email:         "test@example.com",
			expectedError: nil,
		},
		{
			name:          "валидный email с точками",
			email:         "user.name@example.co.uk",
			expectedError: nil,
		},
		{
			name:          "валидный email с цифрами",
			email:         "user123@test456.com",
			expectedError: nil,
		},
		{
			name:          "пустой email",
			email:         "",
			expectedError: ErrEmptyEmail,
		},
		{
			name:          "email с пробелами удаляются",
			email:         "  test@example.com  ",
			expectedError: nil,
		},
		{
			name:          "только пробелы",
			email:         "   ",
			expectedError: ErrEmptyEmail,
		},
		{
			name:          "email без @",
			email:         "testexample.com",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "email без домена",
			email:         "test@",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "email без имени пользователя",
			email:         "@example.com",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "email без доменной зоны",
			email:         "test@example",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "email с пробелами внутри",
			email:         "test @example.com",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "email с кириллицей",
			email:         "тест@example.com",
			expectedError: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		expectedError error
	}{
		{
			name:          "валидный пароль 6 символов",
			password:      "123456",
			expectedError: nil,
		},
		{
			name:          "валидный пароль 8 символов",
			password:      "password",
			expectedError: nil,
		},
		{
			name:          "валидный длинный пароль",
			password:      "verylongpasswordwithmanycharacters123!@#",
			expectedError: nil,
		},
		{
			name:          "пустой пароль",
			password:      "",
			expectedError: ErrEmptyPassword,
		},
		{
			name:          "пароль 1 символ",
			password:      "a",
			expectedError: ErrWeakPassword,
		},
		{
			name:          "пароль 5 символов",
			password:      "12345",
			expectedError: ErrWeakPassword,
		},
		{
			name:          "пароль ровно 6 символов",
			password:      "abcdef",
			expectedError: nil,
		},
		{
			name:          "пароль с пробелами тоже валиден",
			password:      "pass word",
			expectedError: nil,
		},
		{
			name:          "пароль с кириллицей",
			password:      "парольКириллица",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
