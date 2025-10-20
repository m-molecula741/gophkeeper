// Package crypto предоставляет функции шифрования и дешифрования данных
// с использованием AES-256-GCM алгоритма.
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
	// ErrInvalidKey возвращается при использовании ключа неправильной длины.
	// Для AES-256 требуется ключ длиной 32 байта.
	ErrInvalidKey = errors.New("invalid encryption key length (must be 32 bytes for AES-256)")

	// ErrInvalidCiphertext возвращается при попытке расшифровать поврежденные данные.
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// AESCipher предоставляет функции шифрования и дешифрования данных
// с использованием алгоритма AES-256-GCM.
// AES-256-GCM обеспечивает аутентифицированное шифрование,
// гарантируя как конфиденциальность, так и целостность данных.
type AESCipher struct {
	key []byte
}

// NewAESCipher создает новый экземпляр AESCipher с указанным ключом.
// Параметр key должен иметь длину 32 байта для AES-256.
// Возвращает ошибку ErrInvalidKey, если длина ключа некорректна.
func NewAESCipher(key []byte) (*AESCipher, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return &AESCipher{key: key}, nil
}

// Encrypt шифрует данные используя алгоритм AES-256-GCM.
// Для каждого вызова генерируется уникальный nonce,
// который добавляется в начало возвращаемого ciphertext.
// Параметр plaintext - незашифрованные данные.
// Возвращает зашифрованные данные с nonce в начале.
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

// Decrypt расшифровывает данные используя алгоритм AES-256-GCM.
// Ожидает, что ciphertext содержит nonce в начале (как возвращает Encrypt).
// Параметр ciphertext - зашифрованные данные с nonce.
// Возвращает расшифрованные данные или ошибку ErrInvalidCiphertext.
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

// GenerateKey генерирует криптографически стойкий случайный ключ
// длиной 32 байта для использования с AES-256.
// Использует crypto/rand для генерации случайных чисел.
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}
