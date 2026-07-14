package crypto

import (
	"bytes"
	"errors"
	"testing"
)

var (
	key            = []byte("abcdefghijklmnopqrstuvwxyz012345")
	plaintext      = []byte("this-is-a-secret")
	additionalData = []byte("00000000-0000-4000-8000-000000000000")
)

func TestValidAEADWithoutAdditionalData(t *testing.T) {
	tests := []struct {
		name string
		text []byte
	}{
		{"BasicPlaintext", plaintext},
		{"PlaintextOf1", []byte("0")},
		{"PlaintextOf50", bytes.Repeat([]byte("0"), 50)},
		{"PlaintextOf100", bytes.Repeat([]byte("0"), 100)},
		{"PlaintextOf1000", bytes.Repeat([]byte("0"), 1000)},
		{"PlaintextOf10000", bytes.Repeat([]byte("0"), 10000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := EncryptSecret(tt.text, key, nil)
			if err != nil {
				t.Fatalf("Failed to encrypt secret: %v", err)
			}
			deciphered, err := DecryptSecret(ciphertext, key, nil)
			if err != nil {
				t.Fatalf("Failed to decrypt secret: %v", err)
			}
			if !bytes.Equal(deciphered, tt.text) {
				t.Error("DecryptSecret() output does not match original plaintext")
			}
		})
	}
}

func TestValidAEADWithAdditionalData(t *testing.T) {
	tests := []struct {
		name string
		text []byte
	}{
		{"BasicPlaintext", plaintext},
		{"PlaintextOf1", []byte("0")},
		{"PlaintextOf50", bytes.Repeat([]byte("0"), 50)},
		{"PlaintextOf100", bytes.Repeat([]byte("0"), 100)},
		{"PlaintextOf1000", bytes.Repeat([]byte("0"), 1000)},
		{"PlaintextOf10000", bytes.Repeat([]byte("0"), 10000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := EncryptSecret(tt.text, key, additionalData)
			if err != nil {
				t.Fatalf("Failed to encrypt secret: %v", err)
			}
			deciphered, err := DecryptSecret(ciphertext, key, additionalData)
			if err != nil {
				t.Fatalf("Failed to decrypt secret: %v", err)
			}
			if !bytes.Equal(deciphered, tt.text) {
				t.Error("DecryptSecret() output does not match original plaintext")
			}
		})
	}
}

func TestUniqueCiphertextWithoutAdditionalData(t *testing.T) {
	ciphertext1, err := EncryptSecret(plaintext, key, nil)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}
	ciphertext2, err := EncryptSecret(plaintext, key, nil)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("EncryptSecret() produced identical ciphertexts, want unique")
	}
}

func TestUniqueCiphertextWithAdditionalData(t *testing.T) {
	ciphertext1, err := EncryptSecret(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}
	ciphertext2, err := EncryptSecret(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("EncryptSecret() produced identical ciphertexts, want unique")
	}
}

func TestWrongKeyWithoutAdditionalData(t *testing.T) {
	key2 := make([]byte, len(key))
	copy(key2, key)
	key2[len(key2)-1] ^= '0'

	ciphertext, err := EncryptSecret(plaintext, key, nil)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}
	_, err = DecryptSecret(ciphertext, key2, nil)
	if !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
	}
}

func TestWrongKeyWithAdditionalData(t *testing.T) {
	key2 := make([]byte, len(key))
	copy(key2, key)
	key2[len(key2)-1] ^= '0'

	ciphertext, err := EncryptSecret(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}
	_, err = DecryptSecret(ciphertext, key2, additionalData)
	if !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
	}
}

func TestWrongAdditionalData(t *testing.T) {
	additionalData2 := make([]byte, len(additionalData))
	copy(additionalData2, additionalData)
	additionalData2[len(additionalData2)-1] = '1'

	ciphertext, err := EncryptSecret(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	_, err = DecryptSecret(ciphertext, key, nil)
	if !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
	}

	_, err = DecryptSecret(ciphertext, key, additionalData2)
	if !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
	}
}

func TestTamperedCiphertextWithoutAdditionalData(t *testing.T) {
	ciphertext, err := EncryptSecret(plaintext, key, nil)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	badCipherNonce := make([]byte, len(ciphertext))
	copy(badCipherNonce, ciphertext)
	badCipherNonce[0] ^= 1

	badCipherBody := make([]byte, len(ciphertext))
	copy(badCipherBody, ciphertext)
	badCipherBody[13] ^= 1

	badCipherTag := make([]byte, len(ciphertext))
	copy(badCipherTag, ciphertext)
	badCipherTag[len(ciphertext)-1] ^= 1

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"BadNoncePart", badCipherNonce},
		{"BadBodyPart", badCipherBody},
		{"BadTagPart", badCipherTag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSecret(tt.ciphertext, key, nil)
			if !errors.Is(err, ErrDecryptFailed) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
			}
		})
	}
}

func TestTamperedCiphertextWithAdditionalData(t *testing.T) {
	ciphertext, err := EncryptSecret(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	badCipherNonce := make([]byte, len(ciphertext))
	copy(badCipherNonce, ciphertext)
	badCipherNonce[0] ^= 1

	badCipherBody := make([]byte, len(ciphertext))
	copy(badCipherBody, ciphertext)
	badCipherBody[13] ^= 1

	badCipherTag := make([]byte, len(ciphertext))
	copy(badCipherTag, ciphertext)
	badCipherTag[len(ciphertext)-1] ^= 1

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"BadNoncePart", badCipherNonce},
		{"BadBodyPart", badCipherBody},
		{"BadTagPart", badCipherTag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSecret(tt.ciphertext, key, additionalData)
			if !errors.Is(err, ErrDecryptFailed) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
			}
		})
	}
}

func TestBadCipherLength(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"NilCipher", nil},
		{"EmptyCipher", []byte{}},
		{"BelowThreshold", make([]byte, 27)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSecret(tt.ciphertext, key, nil)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}

func TestCipherAtMinimumLength(t *testing.T) {
	atThreshold := make([]byte, 28)
	_, err := DecryptSecret(atThreshold, key, nil)
	if !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
	}
}

func TestEmptyPlaintext(t *testing.T) {
	tests := []struct {
		name string
		text []byte
	}{
		{"NilPlaintext", nil},
		{"EmptyPlaintext", []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptSecret(tt.text, key, nil)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("EncryptSecret() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}

func TestBadKeyLength(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{"NilKey", nil},
		{"EmptyKey", []byte{}},
		{"KeyOf16", make([]byte, 16)},
		{"KeyOf24", make([]byte, 24)},
		{"KeyOf33", make([]byte, 33)},
	}

	for _, tt := range tests {
		t.Run("Encrypt_"+tt.name, func(t *testing.T) {
			_, err := EncryptSecret(plaintext, tt.key, nil)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("EncryptSecret() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}

	ciphertext, err := EncryptSecret(plaintext, key, nil)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	for _, tt := range tests {
		t.Run("Decrypt_"+tt.name, func(t *testing.T) {
			_, err := DecryptSecret(ciphertext, tt.key, nil)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}
