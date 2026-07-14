package crypto

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"testing"
)

var (
	jwt        = []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.0000000000000000000000000000000000000000000")
	seed       = bytes.Repeat([]byte("0"), ed25519.SeedSize)
	privateKey = ed25519.NewKeyFromSeed(seed)
	publicKey  = privateKey.Public().(ed25519.PublicKey)
)

func TestValidToken(t *testing.T) {
	sig, err := Sign(jwt, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	ok, err := VerifySignature(jwt, sig, publicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if !ok {
		t.Errorf("VerifySignature() = %v, want true", ok)
	}
}

func TestWrongPublicKey(t *testing.T) {
	wrongPublicKey := bytes.Repeat([]byte("1"), ed25519.PublicKeySize)
	sig, err := Sign(jwt, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	ok, err := VerifySignature(jwt, sig, wrongPublicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if ok {
		t.Errorf("VerifySignature() = %v, want false", ok)
	}
}

func TestTamperedData(t *testing.T) {
	sig, err := Sign(jwt, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	tamperedJWT := make([]byte, len(jwt))
	copy(tamperedJWT, jwt)
	tamperedJWT[len(tamperedJWT)-1] ^= 1

	ok, err := VerifySignature(tamperedJWT, sig, publicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if ok {
		t.Errorf("VerifySignature() = %v, want false", ok)
	}
}

func TestTamperedSignature(t *testing.T) {
	sig, err := Sign(jwt, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	tamperedSig := make([]byte, len(sig))
	copy(tamperedSig, sig)
	tamperedSig[len(tamperedSig)-1] ^= 1

	ok, err := VerifySignature(jwt, tamperedSig, publicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if ok {
		t.Errorf("VerifySignature() = %v, want false", ok)
	}
}

func TestEmptyData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"NilData", nil},
		{"EmptyData", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Sign(tt.data, privateKey)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("Sign() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}

func TestWrongSignatureLength(t *testing.T) {
	sig, err := Sign(jwt, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	smallSig := make([]byte, len(sig)-1)
	copy(smallSig, sig[:len(sig)-1])

	bigSig := append(append([]byte(nil), sig...), '0')

	tests := []struct {
		name string
		sig  []byte
	}{
		{"NilSignature", nil},
		{"EmptySignature", []byte{}},
		{"SmallSignature", smallSig},
		{"BigSignature", bigSig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifySignature(jwt, tt.sig, publicKey)
			if err != nil {
				t.Fatalf("Failed to verify signature: %v", err)
			}
			if ok {
				t.Errorf("VerifySignature() = %v, want false", ok)
			}
		})
	}
}

func TestBadPrivateKeyLength(t *testing.T) {
	tests := []struct {
		name       string
		privateKey []byte
	}{
		{"NilPrivateKey", nil},
		{"EmptyPrivateKey", []byte{}},
		{"SmallPrivateKey", bytes.Repeat([]byte("0"), ed25519.PrivateKeySize-1)},
		{"LargePrivateKey", bytes.Repeat([]byte("0"), ed25519.PrivateKeySize+1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Sign(jwt, tt.privateKey)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("Sign() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}

func TestBadPublicKeyLength(t *testing.T) {
	sig, err := Sign(jwt, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	tests := []struct {
		name      string
		publicKey []byte
	}{
		{"NilPublicKey", nil},
		{"EmptyPublicKey", []byte{}},
		{"SmallPublicKey", bytes.Repeat([]byte("0"), ed25519.PublicKeySize-1)},
		{"LargePublicKey", bytes.Repeat([]byte("0"), ed25519.PublicKeySize+1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifySignature(jwt, sig, tt.publicKey)
			if !errors.Is(err, ErrInvalidParamLength) {
				t.Errorf("VerifySignature() error = %v, want %v", err, ErrInvalidParamLength)
			}
		})
	}
}
