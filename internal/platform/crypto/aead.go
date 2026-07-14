package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const aes256KeySize = 32

// ErrDecryptFailed indicates that the ciphertext could not be authenticated and decrypted,
// because it was tampered with or the key or additional data
// does not match the one used for encryption.
var ErrDecryptFailed = errors.New("decrypt failed")

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != aes256KeySize {
		return nil, ErrInvalidParamLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return aead, nil
}

// EncryptSecret encrypts plaintext with AES-256-GCM and returns the 12-byte nonce
// prepended to the ciphertext. key must be 32 bytes and plaintext must not be empty.
// additionalData is authenticated but not encrypted, and the same value
// must be supplied to DecryptSecret.
func EncryptSecret(plaintext, key, additionalData []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, ErrInvalidParamLength
	}

	aead, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, additionalData)

	return ciphertext, nil
}

// DecryptSecret authenticates and decrypts a nonce-prefixed ciphertext produced
// by EncryptSecret with the same key and additionalData. key must be 32 bytes.
// It returns ErrDecryptFailed if authentication fails.
func DecryptSecret(ciphertext, key, additionalData []byte) ([]byte, error) {
	aead, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aead.NonceSize()+aead.Overhead() {
		return nil, ErrInvalidParamLength
	}
	noncePart := ciphertext[:aead.NonceSize()]
	ciphertextPart := ciphertext[aead.NonceSize():]

	plaintext, err := aead.Open(nil, noncePart, ciphertextPart, additionalData)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	return plaintext, nil
}
