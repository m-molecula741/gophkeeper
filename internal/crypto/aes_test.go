package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAESCipher(t *testing.T) {
	tests := []struct {
		name          string
		keyLength     int
		expectedError error
	}{
		{
			name:          "валидный ключ 32 байта (AES-256)",
			keyLength:     32,
			expectedError: nil,
		},
		{
			name:          "невалидный ключ 16 байт",
			keyLength:     16,
			expectedError: ErrInvalidKey,
		},
		{
			name:          "невалидный ключ 24 байта",
			keyLength:     24,
			expectedError: ErrInvalidKey,
		},
		{
			name:          "невалидный ключ 0 байт",
			keyLength:     0,
			expectedError: ErrInvalidKey,
		},
		{
			name:          "невалидный ключ 64 байта",
			keyLength:     64,
			expectedError: ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLength)
			cipher, err := NewAESCipher(key)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, cipher)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cipher)
			}
		})
	}
}

func TestAESCipher_Encrypt_Decrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "простой текст",
			plaintext: []byte("Hello, World!"),
		},
		{
			name:      "пустые данные",
			plaintext: []byte{},
		},
		{
			name:      "JSON данные",
			plaintext: []byte(`{"username":"user","password":"pass123"}`),
		},
		{
			name:      "бинарные данные",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name:      "длинный текст",
			plaintext: bytes.Repeat([]byte("A"), 1000),
		},
		{
			name:      "кириллица",
			plaintext: []byte("Привет, мир! 🔐"),
		},
	}

	// Создаем cipher с тестовым ключом
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	cipher, err := NewAESCipher(key)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := cipher.Encrypt(tt.plaintext)
			assert.NoError(t, err)
			assert.NotNil(t, ciphertext)

			// Зашифрованные данные должны отличаться от исходных (если не пустые)
			if len(tt.plaintext) > 0 {
				assert.NotEqual(t, tt.plaintext, ciphertext)
			}

			// Decrypt
			decrypted, err := cipher.Decrypt(ciphertext)
			assert.NoError(t, err)
			// Для пустых данных проверяем по длине
			if len(tt.plaintext) == 0 {
				assert.Empty(t, decrypted)
			} else {
				assert.Equal(t, tt.plaintext, decrypted)
			}
		})
	}
}

func TestAESCipher_Encrypt_DifferentNonce(t *testing.T) {
	// Тест проверяет, что одинаковые данные шифруются по-разному из-за разных nonce
	key := make([]byte, 32)
	cipher, err := NewAESCipher(key)
	require.NoError(t, err)

	plaintext := []byte("Same data")

	// Шифруем одни и те же данные дважды
	ciphertext1, err := cipher.Encrypt(plaintext)
	require.NoError(t, err)

	ciphertext2, err := cipher.Encrypt(plaintext)
	require.NoError(t, err)

	// Зашифрованные данные должны отличаться (разные nonce)
	assert.NotEqual(t, ciphertext1, ciphertext2)

	// Но оба должны расшифровываться в исходные данные
	decrypted1, err := cipher.Decrypt(ciphertext1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := cipher.Decrypt(ciphertext2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestAESCipher_Decrypt_InvalidData(t *testing.T) {
	key := make([]byte, 32)
	cipher, err := NewAESCipher(key)
	require.NoError(t, err)

	tests := []struct {
		name          string
		ciphertext    []byte
		expectedError error
	}{
		{
			name:          "слишком короткие данные",
			ciphertext:    []byte{0x01, 0x02},
			expectedError: ErrInvalidCiphertext,
		},
		{
			name:          "пустые данные",
			ciphertext:    []byte{},
			expectedError: ErrInvalidCiphertext,
		},
		{
			name:       "поврежденные данные",
			ciphertext: bytes.Repeat([]byte{0xFF}, 50),
			// Ошибка будет, но не ErrInvalidCiphertext
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := cipher.Decrypt(tt.ciphertext)
			assert.Error(t, err)
			assert.Nil(t, decrypted)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	// Генерируем несколько ключей
	key1, err := GenerateKey()
	assert.NoError(t, err)
	assert.Len(t, key1, 32)

	key2, err := GenerateKey()
	assert.NoError(t, err)
	assert.Len(t, key2, 32)

	// Ключи должны быть разными
	assert.NotEqual(t, key1, key2)

	// Проверяем, что ключ можно использовать для создания cipher
	cipher, err := NewAESCipher(key1)
	assert.NoError(t, err)
	assert.NotNil(t, cipher)

	// Проверяем, что можно шифровать/дешифровать
	plaintext := []byte("test data")
	ciphertext, err := cipher.Encrypt(plaintext)
	assert.NoError(t, err)

	decrypted, err := cipher.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESCipher_DifferentKeys(t *testing.T) {
	// Создаем два cipher с разными ключами
	key1 := bytes.Repeat([]byte{0x01}, 32)
	cipher1, err := NewAESCipher(key1)
	require.NoError(t, err)

	key2 := bytes.Repeat([]byte{0x02}, 32)
	cipher2, err := NewAESCipher(key2)
	require.NoError(t, err)

	plaintext := []byte("secret message")

	// Шифруем первым ключом
	ciphertext, err := cipher1.Encrypt(plaintext)
	require.NoError(t, err)

	// Пытаемся расшифровать вторым ключом (должно не получиться)
	decrypted, err := cipher2.Decrypt(ciphertext)
	assert.Error(t, err)
	assert.Nil(t, decrypted)

	// Расшифровка правильным ключом должна работать
	decrypted, err = cipher1.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
