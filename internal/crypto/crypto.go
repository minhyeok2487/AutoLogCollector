package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const (
	keyDir  = "config"
	keyFile = "encryption.key"
	keySize = 32 // AES-256
)

// LoadOrGenerateKey loads the encryption key from disk, or generates a new one if it doesn't exist.
func LoadOrGenerateKey() ([]byte, error) {
	path := filepath.Join(keyDir, keyFile)

	key, err := os.ReadFile(path)
	if err == nil && len(key) == keySize {
		return key, nil
	}

	// Generate new key
	key = make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	// Ensure directory exists
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, err
	}

	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a base64-encoded string.
func Encrypt(plaintext string, key []byte) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// nonce is prepended to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext.
// Returns the plaintext string on success, or the original input if decryption fails
// (for backward compatibility with previously stored plaintext passwords).
func Decrypt(encoded string, key []byte) string {
	if encoded == "" {
		return ""
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded // Not base64 = plaintext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return encoded
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return encoded
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return encoded // Too short = plaintext
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return encoded // Decryption failed = plaintext
	}

	return string(plaintext)
}

// EncryptFields encrypts password and enablePassword fields, returning an error if encryption fails.
func EncryptFields(password, enablePassword string, key []byte) (encPassword, encEnablePassword string, err error) {
	encPassword, err = Encrypt(password, key)
	if err != nil {
		return "", "", errors.New("failed to encrypt password: " + err.Error())
	}

	encEnablePassword, err = Encrypt(enablePassword, key)
	if err != nil {
		return "", "", errors.New("failed to encrypt enable password: " + err.Error())
	}

	return encPassword, encEnablePassword, nil
}

// DecryptFields decrypts password and enablePassword fields.
// Falls back to plaintext if decryption fails (backward compatibility).
func DecryptFields(password, enablePassword string, key []byte) (string, string) {
	return Decrypt(password, key), Decrypt(enablePassword, key)
}
