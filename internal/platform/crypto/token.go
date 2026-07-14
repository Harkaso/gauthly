package crypto

import (
	"crypto/ed25519"
)

// Sign signs data with the Ed25519 private key and returns the signature.
// privateKey must be ed25519.PrivateKeySize bytes and data must not be empty.
func Sign(data []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidParamLength
	}

	if len(data) == 0 {
		return nil, ErrInvalidParamLength
	}

	return ed25519.Sign(privateKey, data), nil
}

// VerifySignature reports whether sig is a valid Ed25519 signature of data by publicKey,
// which must be ed25519.PublicKeySize bytes.
// An invalid signature returns (false, nil); only a malformed public key returns an error.
func VerifySignature(data, sig []byte, publicKey ed25519.PublicKey) (bool, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return false, ErrInvalidParamLength
	}

	return ed25519.Verify(publicKey, data, sig), nil
}
