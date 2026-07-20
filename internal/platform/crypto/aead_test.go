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

var additionalDataCases = []struct {
	name string
	aad  []byte
}{
	{"WithoutAdditionalData", nil},
	{"WithAdditionalData", additionalData},
}

var badKeyCases = []struct {
	name string
	key  []byte
}{
	{"NilKey", nil},
	{"EmptyKey", []byte{}},
	{"KeyOf16", make([]byte, 16)},
	{"KeyOf24", make([]byte, 24)},
	{"KeyOf33", make([]byte, 33)},
}

func mustEncrypt(t *testing.T, plaintext, key, additionalData []byte) []byte {
	t.Helper()

	ciphertext, err := EncryptSecret(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("Failed to encrypt secret: %v", err)
	}

	return ciphertext
}

func TestDecryptSecretReturnsOriginalPlaintext(t *testing.T) {
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

	for _, ad := range additionalDataCases {
		t.Run(ad.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					ciphertext := mustEncrypt(t, tt.text, key, ad.aad)

					deciphered, err := DecryptSecret(ciphertext, key, ad.aad)
					if err != nil {
						t.Fatalf("Failed to decrypt secret: %v", err)
					}
					if !bytes.Equal(deciphered, tt.text) {
						t.Error("DecryptSecret() output does not match original plaintext")
					}
				})
			}
		})
	}
}

func TestEncryptSecretProducesUniqueCiphertexts(t *testing.T) {
	for _, ad := range additionalDataCases {
		t.Run(ad.name, func(t *testing.T) {
			ciphertext1 := mustEncrypt(t, plaintext, key, ad.aad)
			ciphertext2 := mustEncrypt(t, plaintext, key, ad.aad)

			if bytes.Equal(ciphertext1, ciphertext2) {
				t.Error("EncryptSecret() produced identical ciphertexts, want unique")
			}
		})
	}
}

func TestDecryptSecretRejectsWrongKey(t *testing.T) {
	wrongKey := make([]byte, len(key))
	copy(wrongKey, key)
	wrongKey[len(wrongKey)-1] ^= '0'

	for _, ad := range additionalDataCases {
		t.Run(ad.name, func(t *testing.T) {
			ciphertext := mustEncrypt(t, plaintext, key, ad.aad)

			_, err := DecryptSecret(ciphertext, wrongKey, ad.aad)
			if !errors.Is(err, ErrDecryptFailed) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
			}
		})
	}
}

func TestDecryptSecretRejectsWrongAdditionalData(t *testing.T) {
	alteredAdditionalData := make([]byte, len(additionalData))
	copy(alteredAdditionalData, additionalData)
	alteredAdditionalData[len(alteredAdditionalData)-1] = '1'

	ciphertext := mustEncrypt(t, plaintext, key, additionalData)

	tests := []struct {
		name string
		aad  []byte
	}{
		{"MissingAdditionalData", nil},
		{"AlteredAdditionalData", alteredAdditionalData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSecret(ciphertext, key, tt.aad)
			if !errors.Is(err, ErrDecryptFailed) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
			}
		})
	}
}

func TestDecryptSecretRejectsTamperedCiphertext(t *testing.T) {
	for _, ad := range additionalDataCases {
		t.Run(ad.name, func(t *testing.T) {
			ciphertext := mustEncrypt(t, plaintext, key, ad.aad)

			tests := []struct {
				name  string
				index int
			}{
				{"TamperedNonce", 0},
				{"TamperedBody", 13},
				{"TamperedTag", len(ciphertext) - 1},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tampered := make([]byte, len(ciphertext))
					copy(tampered, ciphertext)
					tampered[tt.index] ^= 1

					_, err := DecryptSecret(tampered, key, ad.aad)
					if !errors.Is(err, ErrDecryptFailed) {
						t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
					}
				})
			}
		})
	}
}

func TestDecryptSecretRejectsShortCiphertext(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"NilCiphertext", nil},
		{"EmptyCiphertext", []byte{}},
		{"BelowMinimumLength", make([]byte, 27)},
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

func TestDecryptSecretRejectsCiphertextAtMinimumLength(t *testing.T) {
	atMinimumLength := make([]byte, 28)

	_, err := DecryptSecret(atMinimumLength, key, nil)
	if !errors.Is(err, ErrDecryptFailed) {
		t.Errorf("DecryptSecret() error = %v, want %v", err, ErrDecryptFailed)
	}
}

func TestEncryptSecretRejectsEmptyPlaintext(t *testing.T) {
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

func TestEncryptSecretRejectsBadKeyLength(t *testing.T) {
	for _, tt := range badKeyCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptSecret(plaintext, tt.key, nil)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("EncryptSecret() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}

func TestDecryptSecretRejectsBadKeyLength(t *testing.T) {
	ciphertext := mustEncrypt(t, plaintext, key, nil)

	for _, tt := range badKeyCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSecret(ciphertext, tt.key, nil)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("DecryptSecret() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}
