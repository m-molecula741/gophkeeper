package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key length (must be 32 bytes for AES-256)")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// AESCipher предоставляет функции шифрования/дешифрования с использованием AES-256-GCM
type AESCipher struct {
	key []byte
}

// NewAESCipher создает новый AES cipher с указанным ключом
// key должен быть длиной 32 байта для AES-256
func NewAESCipher(key []byte) (*AESCipher, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return &AESCipher{key: key}, nil
}

// Encrypt шифрует данные используя AES-256-GCM
// Возвращает зашифрованные данные с nonce в начале
func (c *AESCipher) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Создаем nonce (число, используемое один раз)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Шифруем данные и добавляем nonce в начало
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt расшифровывает данные используя AES-256-GCM
// Ожидает, что nonce находится в начале ciphertext
func (c *AESCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Извлекаем nonce и зашифрованные данные
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Расшифровываем
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// GenerateKey генерирует случайный 32-байтный ключ для AES-256
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}
