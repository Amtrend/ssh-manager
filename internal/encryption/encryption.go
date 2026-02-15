package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

var encryptionKey []byte // Global variable for encryption key

// SetEncryptionKey sets the encryption key.
func SetEncryptionKey(key []byte) {
	if len(key) != 32 {
		panic("Invalid encryption key size: key must be 32 bytes")
	}
	encryptionKey = key
}

// Encrypt encrypts text using AES.
func Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// GCM — standard for encrypting data in databases
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and add nonce to the beginning
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(text), nil)

	return fmt.Sprintf("%x", ciphertext), nil // We use hex for simplicity in the database
}

// Decrypt decrypts text using AES.
func Decrypt(encryptedHex string) (string, error) {
	var ciphertext []byte
	_, err := fmt.Sscanf(encryptedHex, "%x", &ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
