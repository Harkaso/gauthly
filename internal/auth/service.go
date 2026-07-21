package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/Harkaso/gauthly/internal/platform/crypto"
	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/Harkaso/gauthly/internal/user"
	"github.com/google/uuid"
)

const (
	minPasswordLength = 12
	maxPasswordLength = 128
)

var (
	// ErrPasswordTooShort is returned by Register when the password has fewer
	// than the minimum number of characters.
	ErrPasswordTooShort = errors.New("password too short")

	// ErrPasswordTooLong is returned by Register when the password exceeds the
	// maximum number of characters, a bound that also caps the work handed to
	// the password hash.
	ErrPasswordTooLong = errors.New("password too long")

	// ErrInvalidEmail is returned by Register when the email is not a single,
	// syntactically valid address.
	ErrInvalidEmail = errors.New("invalid email")

	// ErrEmailAlreadyExists is returned by a UserRepository when a user with the
	// same email already exists in the tenant.
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserRepository stores user accounts for the authentication service. It is the
// port through which AuthService persists a user; the concrete implementation
// lives in the database layer.
type UserRepository interface {
	// Create inserts u. It returns ErrEmailAlreadyExists when a user with the
	// same email already exists in the tenant, which the store enforces through
	// the UNIQUE(email, tenant_id) constraint.
	Create(ctx context.Context, u user.User) error
}

// AuthService carries out the authentication use cases, starting with user
// registration. It depends only on ports, never on a concrete store or
// transport.
type AuthService struct {
	users UserRepository
}

// NewAuthService returns an AuthService backed by users.
func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{users: users}
}

func checkPasswordLength(password string) error {
	n := utf8.RuneCountInString(password)
	if n < minPasswordLength {
		return ErrPasswordTooShort
	}
	if n > maxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", ErrInvalidEmail
	}

	return strings.ToLower(addr.Address), nil
}

// Register creates a user account within the tenant carried by ctx.
//
// The tenant is read from ctx, never from the caller's input, so a context that
// never passed through the tenant middleware is rejected rather than served
// without one. The password is validated against the length policy, then hashed
// with argon2id; the plaintext is never stored, logged, or returned. The email
// is trimmed, parsed, and lowercased so that uniqueness is case-insensitive
// within the tenant.
//
// It returns the stored user, whose PasswordHash must never be serialized to a
// client. Register does not detect an already-registered email: that is left to
// the repository, which reports it as it translates the store's uniqueness
// constraint.
func (s *AuthService) Register(ctx context.Context, email, password string) (user.User, error) {
	tenantID, err := tenant.RequireFromContext(ctx)
	if err != nil {
		return user.User{}, err
	}

	err = checkPasswordLength(password)
	if err != nil {
		return user.User{}, err
	}

	email, err = normalizeEmail(email)
	if err != nil {
		return user.User{}, err
	}

	passwordHash, err := crypto.HashPassword([]byte(password))
	if err != nil {
		return user.User{}, err
	}

	u := user.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = s.users.Create(ctx, u)
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}
