// Package crypto provides the low-level cryptographic primitives used across
// the service: argon2id password hashing, AES-256-GCM authenticated encryption
// for secrets at rest, and Ed25519 signing for tokens.
//
// All exported functions are pure wrappers over the standard library and
// golang.org/x/crypto. They perform no key management, storage, or higher-level
// protocol handling.
package crypto
