// Package user defines the user account, the identity the authentication and
// authorization domains operate on.
//
// It holds the entity only; persistence and use cases live in the packages that
// consume it. A user always belongs to exactly one tenant, and its password is
// stored hashed, never in plaintext.
package user
