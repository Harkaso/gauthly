package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	minSaltLength = 8
	saltLength    = 16
	minHashLength = 16
	hashLength    = 32
	tCost         = 3
	mCost         = 64 * 1024
	pCost         = 1
	argon2Version = argon2.Version
	algorithmName = "argon2id"
)

// ErrInvalidPHCString indicates that a PHC-formatted hash string is malformed
// or does not describe the expected argon2id parameters.
var ErrInvalidPHCString = errors.New("invalid PHC string")

type phcString struct {
	mCost uint32
	tCost uint32
	pCost uint8
	salt  []byte
	hash  []byte
}

func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)

	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	return salt, nil
}

func toPHCString(salt, hash []byte) string {
	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		algorithmName, argon2Version, mCost,
		tCost, pCost,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func parsePHCString(phcStr string) (phcString, error) {
	parts := strings.Split(phcStr, "$")
	if len(parts) != 6 {
		return phcString{}, ErrInvalidPHCString
	}

	if parts[1] != algorithmName {
		return phcString{}, ErrInvalidPHCString
	}

	if version, err := strconv.Atoi(strings.TrimPrefix(parts[2], "v=")); err != nil || version != argon2Version {
		return phcString{}, ErrInvalidPHCString
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return phcString{}, ErrInvalidPHCString
	}

	m, err := strconv.Atoi(strings.TrimPrefix(params[0], "m="))
	if err != nil {
		return phcString{}, ErrInvalidPHCString
	}

	t, err := strconv.Atoi(strings.TrimPrefix(params[1], "t="))
	if err != nil {
		return phcString{}, ErrInvalidPHCString
	}

	p, err := strconv.Atoi(strings.TrimPrefix(params[2], "p="))
	if err != nil {
		return phcString{}, ErrInvalidPHCString
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return phcString{}, ErrInvalidPHCString
	}
	if len(salt) < minSaltLength {
		return phcString{}, ErrInvalidPHCString
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return phcString{}, ErrInvalidPHCString
	}
	if len(hash) < minHashLength {
		return phcString{}, ErrInvalidPHCString
	}

	return phcString{
		mCost: uint32(m),
		tCost: uint32(t),
		pCost: uint8(p),
		salt:  salt,
		hash:  hash,
	}, nil
}

// HashPassword hashes password with argon2id using a fresh random salt and
// returns the result encoded as a PHC string.
func HashPassword(password []byte) (string, error) {
	salt, err := generateSalt(saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(password, salt, tCost, mCost, pCost, hashLength)
	return toPHCString(salt, hash), nil
}

// VerifyPassword reports whether password matches the argon2id hash encoded in phcStr,
// using a constant-time comparison. It returns ErrInvalidPHCString if phcStr is malformed.
func VerifyPassword(password []byte, phcStr string) (bool, error) {
	parsedPHC, err := parsePHCString(phcStr)
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey(password, parsedPHC.salt, parsedPHC.tCost, parsedPHC.mCost, parsedPHC.pCost, uint32(len(parsedPHC.hash)))
	return subtle.ConstantTimeCompare(hash, parsedPHC.hash) == 1, nil
}
