package crypto

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
)

var password = []byte("TestP@ssw0rd")

func TestSamePassword(t *testing.T) {
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	ok, err := VerifyPassword(password, hashed)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !ok {
		t.Errorf("VerifyPassword() = %v, want true", ok)
	}
}

func TestSamePasswordDifferentSalt(t *testing.T) {
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes, want unique per call")
	}
}

func TestDifferentPassword(t *testing.T) {
	password1 := []byte("TestP@ssw0rd1")

	hashed1, err := HashPassword(password1)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	password2 := []byte("TestP@ssw0rd2")

	ok, err := VerifyPassword(password2, hashed1)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if ok {
		t.Errorf("VerifyPassword() = %v, want false", ok)
	}
}

func TestBadPHCStringFormat(t *testing.T) {
	salt := []byte("<|testing-salt|>")

	hash := argon2.IDKey(password, salt, tCost, mCost, pCost, hashLength)

	tests := []struct {
		name string
		phc  string
	}{
		{"EmptyString", ""},
		{"NotPHCString", "this is not a PHC string"},
		{"BadAlgorithmName", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
			"argon2", argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoAlgorithmName", fmt.Sprintf("$$v=%d$m=%d,t=%d,p=%d$%s$%s",
			argon2Version, mCost, tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"BadAlgorithmVersion", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version-2, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoAlgorithmVersion", fmt.Sprintf("$%s$v=$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, mCost, tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"BadAlgorithmVersionFormat", fmt.Sprintf("$%s$v=v$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, mCost, tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoMemoryCost", fmt.Sprintf("$%s$v=%d$m=,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version, tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"BadMemoryCostFormat", fmt.Sprintf("$%s$v=%d$m=m,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version, tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoTimeCost", fmt.Sprintf("$%s$v=%d$m=%d,t=,p=%d$%s$%s",
			algorithmName, argon2Version, mCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"BadTimeCostFormat", fmt.Sprintf("$%s$v=%d$m=%d,t=t,p=%d$%s$%s",
			algorithmName, argon2Version, mCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoThreadCost", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=$%s$%s",
			algorithmName, argon2Version, mCost, tCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"BadThreadCostFormat", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=p$%s$%s",
			algorithmName, argon2Version, mCost, tCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoBase64SaltEncoding", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version, mCost,
			tCost, pCost, salt,
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoSalt", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$$%s",
			algorithmName, argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"TooShortSalt", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString([]byte("salt")),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
		{"NoBase64HashEncoding", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			"<--------|testing-hash|-------->",
		)},
		{"NoHash", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$",
			algorithmName, argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
		)},
		{"TooShortHash", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
			algorithmName, argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString([]byte("hash")),
		)},
		{"BadPHCStringFormat", fmt.Sprintf("$%s$v=%d$m=%d$t=%d$p=%d$%s$%s",
			algorithmName, argon2Version, mCost,
			tCost, pCost,
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hash),
		)},
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
