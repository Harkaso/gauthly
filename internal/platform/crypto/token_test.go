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

func mustSign(t *testing.T, data []byte, privateKey ed25519.PrivateKey) []byte {
	t.Helper()

	sig, err := Sign(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	return sig
}

func TestVerifySignatureAcceptsValidSignature(t *testing.T) {
	sig := mustSign(t, jwt, privateKey)

	ok, err := VerifySignature(jwt, sig, publicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if !ok {
		t.Errorf("VerifySignature() = %v, want true", ok)
	}
}

func TestVerifySignatureRejectsWrongPublicKey(t *testing.T) {
	wrongPublicKey := bytes.Repeat([]byte("1"), ed25519.PublicKeySize)
	sig := mustSign(t, jwt, privateKey)

	ok, err := VerifySignature(jwt, sig, wrongPublicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if ok {
		t.Errorf("VerifySignature() = %v, want false", ok)
	}
}

func TestVerifySignatureRejectsTamperedData(t *testing.T) {
	sig := mustSign(t, jwt, privateKey)

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

func TestVerifySignatureRejectsTamperedSignature(t *testing.T) {
	sig := mustSign(t, jwt, privateKey)

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

func TestVerifySignatureRejectsBadSignatureLength(t *testing.T) {
	sig := mustSign(t, jwt, privateKey)

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

func TestVerifySignatureRejectsBadPublicKeyLength(t *testing.T) {
	sig := mustSign(t, jwt, privateKey)

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

func TestSignRejectsEmptyData(t *testing.T) {
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

func TestSignRejectsBadPrivateKeyLength(t *testing.T) {
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
