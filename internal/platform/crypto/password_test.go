package crypto

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
)

var password = []byte("TestP@ssw0rd")

func mustHash(t *testing.T, password []byte) string {
	t.Helper()

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	return hashed
}

func formatPHC(algorithm, version, params, salt, hash string) string {
	return fmt.Sprintf("$%s$%s$%s$%s$%s", algorithm, version, params, salt, hash)
}

func TestVerifyPasswordAcceptsMatchingPassword(t *testing.T) {
	hashed := mustHash(t, password)

	ok, err := VerifyPassword(password, hashed)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !ok {
		t.Errorf("VerifyPassword() = %v, want true", ok)
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hashed := mustHash(t, []byte("TestP@ssw0rd1"))

	ok, err := VerifyPassword([]byte("TestP@ssw0rd2"), hashed)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if ok {
		t.Errorf("VerifyPassword() = %v, want false", ok)
	}
}

func TestHashPasswordProducesUniqueHashes(t *testing.T) {
	hash1 := mustHash(t, password)
	hash2 := mustHash(t, password)

	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes, want unique per call")
	}
}

func TestVerifyPasswordUsesStoredCosts(t *testing.T) {
	const (
		storedMCost = 32 * 1024
		storedTCost = 2
		storedPCost = 1
	)

	salt := bytes.Repeat([]byte("s"), saltLength)
	hash := argon2.IDKey(password, salt, storedTCost, storedMCost, storedPCost, hashLength)

	phc := formatPHC(
		algorithmName,
		fmt.Sprintf("v=%d", argon2Version),
		fmt.Sprintf("m=%d,t=%d,p=%d", storedMCost, storedTCost, storedPCost),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	ok, err := VerifyPassword(password, phc)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !ok {
		t.Errorf("VerifyPassword() = %v, want true", ok)
	}
}

func TestVerifyPasswordRejectsMalformedPHCString(t *testing.T) {
	salt := []byte("<|testing-salt|>")
	hash := argon2.IDKey(password, salt, tCost, mCost, pCost, hashLength)

	version := fmt.Sprintf("v=%d", argon2Version)
	params := fmt.Sprintf("m=%d,t=%d,p=%d", mCost, tCost, pCost)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	tests := []struct {
		name string
		phc  string
	}{
		{"EmptyString", ""},
		{"NotPHCString", "this is not a PHC string"},
		{"BadAlgorithmName", formatPHC("argon2", version, params, encodedSalt, encodedHash)},
		{"NoAlgorithmName", formatPHC("", version, params, encodedSalt, encodedHash)},
		{"BadAlgorithmVersion", formatPHC(algorithmName, fmt.Sprintf("v=%d", argon2Version-2), params, encodedSalt, encodedHash)},
		{"NoAlgorithmVersion", formatPHC(algorithmName, "v=", params, encodedSalt, encodedHash)},
		{"BadAlgorithmVersionFormat", formatPHC(algorithmName, "v=v", params, encodedSalt, encodedHash)},
		{"NoMemoryCost", formatPHC(algorithmName, version, fmt.Sprintf("m=,t=%d,p=%d", tCost, pCost), encodedSalt, encodedHash)},
		{"BadMemoryCostFormat", formatPHC(algorithmName, version, fmt.Sprintf("m=m,t=%d,p=%d", tCost, pCost), encodedSalt, encodedHash)},
		{"NoTimeCost", formatPHC(algorithmName, version, fmt.Sprintf("m=%d,t=,p=%d", mCost, pCost), encodedSalt, encodedHash)},
		{"BadTimeCostFormat", formatPHC(algorithmName, version, fmt.Sprintf("m=%d,t=t,p=%d", mCost, pCost), encodedSalt, encodedHash)},
		{"NoThreadCost", formatPHC(algorithmName, version, fmt.Sprintf("m=%d,t=%d,p=", mCost, tCost), encodedSalt, encodedHash)},
		{"BadThreadCostFormat", formatPHC(algorithmName, version, fmt.Sprintf("m=%d,t=%d,p=p", mCost, tCost), encodedSalt, encodedHash)},
		{"NoBase64SaltEncoding", formatPHC(algorithmName, version, params, string(salt), encodedHash)},
		{"NoSalt", formatPHC(algorithmName, version, params, "", encodedHash)},
		{"TooShortSalt", formatPHC(algorithmName, version, params, base64.RawStdEncoding.EncodeToString([]byte("salt")), encodedHash)},
		{"NoBase64HashEncoding", formatPHC(algorithmName, version, params, encodedSalt, "<--------|testing-hash|-------->")},
		{"NoHash", formatPHC(algorithmName, version, params, encodedSalt, "")},
		{"TooShortHash", formatPHC(algorithmName, version, params, encodedSalt, base64.RawStdEncoding.EncodeToString([]byte("hash")))},
		{"TooManyFields", fmt.Sprintf("$%s$%s$m=%d$t=%d$p=%d$%s$%s", algorithmName, version, mCost, tCost, pCost, encodedSalt, encodedHash)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifyPassword(password, tt.phc)
			if !errors.Is(err, ErrInvalidPHCString) {
				t.Errorf("VerifyPassword() error = %v, want %v", err, ErrInvalidPHCString)
			}
			if ok {
				t.Errorf("VerifyPassword() = %v, want false", ok)
			}
		})
	}
}
