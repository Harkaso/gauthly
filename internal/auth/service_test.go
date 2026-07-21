package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Harkaso/gauthly/internal/platform/crypto"
	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/Harkaso/gauthly/internal/user"
	"github.com/google/uuid"
)

const validPassword = "correct horse battery"

var testTenantID = uuid.MustParse("11111111-1111-4111-8111-111111111111")

type fakeUserRepo struct {
	called  bool
	created user.User
	err     error
}

func (f *fakeUserRepo) Create(_ context.Context, u user.User) error {
	f.called = true
	f.created = u
	return f.err
}

func ctxWithTenant(id uuid.UUID) context.Context {
	return tenant.WithTenant(context.Background(), id)
}

func TestRegisterCreatesUser(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewAuthService(repo)

	got, err := svc.Register(ctxWithTenant(testTenantID), "  User@Example.COM  ", validPassword)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if !repo.called {
		t.Fatal("Register() did not call the repository, want a user created")
	}
	if got != repo.created {
		t.Errorf("Register() = %v, want the stored user %v", got, repo.created)
	}
	if got.TenantID != testTenantID {
		t.Errorf("Register() tenant = %v, want %v", got.TenantID, testTenantID)
	}
	if got.Email != "user@example.com" {
		t.Errorf("Register() email = %q, want %q (trimmed and lowercased)", got.Email, "user@example.com")
	}
	if got.ID == uuid.Nil {
		t.Error("Register() id = uuid.Nil, want a generated identifier")
	}

	ok, err := crypto.VerifyPassword([]byte(validPassword), got.PasswordHash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v, want nil", err)
	}
	if !ok {
		t.Error("Register() stored a hash that does not verify against the password")
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{"PasswordTooShort", "user@example.com", strings.Repeat("a", 11), ErrPasswordTooShort},
		{"PasswordTooLong", "user@example.com", strings.Repeat("a", 129), ErrPasswordTooLong},
		{"InvalidEmail", "not-an-email", validPassword, ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{}
			svc := NewAuthService(repo)

			_, err := svc.Register(ctxWithTenant(testTenantID), tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Register() error = %v, want %v", err, tt.wantErr)
			}
			if repo.called {
				t.Error("Register() reached the repository, want the request rejected before persistence")
			}
		})
	}
}

func TestRegisterRequiresTenant(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewAuthService(repo)

	_, err := svc.Register(context.Background(), "user@example.com", validPassword)
	if !errors.Is(err, tenant.ErrNotInContext) {
		t.Errorf("Register() error = %v, want %v", err, tenant.ErrNotInContext)
	}
	if repo.called {
		t.Error("Register() reached the repository without a tenant, want the request rejected")
	}
}

func TestRegisterPropagatesRepositoryError(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
	}{
		{"DuplicateEmail", ErrEmailAlreadyExists},
		{"OtherFailure", errors.New("connection refused")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{err: tt.repoErr}
			svc := NewAuthService(repo)

			_, err := svc.Register(ctxWithTenant(testTenantID), "user@example.com", validPassword)
			if !errors.Is(err, tt.repoErr) {
				t.Errorf("Register() error = %v, want %v", err, tt.repoErr)
			}
			if !repo.called {
				t.Error("Register() did not reach the repository, want it called")
			}
		})
	}
}

func TestCheckPasswordLength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"Minimum", strings.Repeat("a", 12), nil},
		{"Maximum", strings.Repeat("a", 128), nil},
		{"TooShort", strings.Repeat("a", 11), ErrPasswordTooShort},
		{"TooLong", strings.Repeat("a", 129), ErrPasswordTooLong},
		{"CountsRunesNotBytes", strings.Repeat("é", 12), nil},
		{"MultibyteTooShort", strings.Repeat("é", 11), ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPasswordLength(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("checkPasswordLength(%d runes) error = %v, want %v", utf8.RuneCountInString(tt.password), err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"TrimsAndLowercases", "  User@Example.COM  ", "user@example.com", nil},
		{"AlreadyClean", "user@example.com", "user@example.com", nil},
		{"DisplayNameForm", "User <User@Example.com>", "user@example.com", nil},
		{"Invalid", "not-an-email", "", ErrInvalidEmail},
		{"Empty", "", "", ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEmail(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("normalizeEmail(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("normalizeEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
