//go:build integration

package database

import (
	"context"
	"errors"
	"testing"

	"github.com/Harkaso/gauthly/internal/auth"
	"github.com/Harkaso/gauthly/internal/user"
	"github.com/google/uuid"
)

func TestUserRepositoryCreate(t *testing.T) {
	repo := NewUserRepository(testPool)
	tenantID := insertTenant(t, true)

	u := user.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        "user@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	var (
		gotTenant uuid.UUID
		gotEmail  string
		gotHash   string
	)
	const query = `SELECT tenant_id, email, password_hash FROM users WHERE id = $1`
	if err := testPool.QueryRow(context.Background(), query, u.ID).Scan(&gotTenant, &gotEmail, &gotHash); err != nil {
		t.Fatalf("Failed to read the stored user: %v", err)
	}
	if gotTenant != tenantID {
		t.Errorf("Create() stored tenant_id = %v, want %v", gotTenant, tenantID)
	}
	if gotEmail != u.Email {
		t.Errorf("Create() stored email = %q, want %q", gotEmail, u.Email)
	}
	if gotHash != u.PasswordHash {
		t.Errorf("Create() stored password_hash = %q, want %q", gotHash, u.PasswordHash)
	}
}

func TestUserRepositoryCreateRejectsDuplicateInTenant(t *testing.T) {
	repo := NewUserRepository(testPool)
	tenantID := insertTenant(t, true)
	const email = "duplicate@example.com"

	first := user.User{ID: uuid.New(), TenantID: tenantID, Email: email, PasswordHash: "hash"}
	if err := repo.Create(context.Background(), first); err != nil {
		t.Fatalf("Create() first user error = %v, want nil", err)
	}

	second := user.User{ID: uuid.New(), TenantID: tenantID, Email: email, PasswordHash: "hash"}
	if err := repo.Create(context.Background(), second); !errors.Is(err, auth.ErrEmailAlreadyExists) {
		t.Errorf("Create() duplicate error = %v, want %v", err, auth.ErrEmailAlreadyExists)
	}
}

func TestUserRepositoryCreateIsolatesTenants(t *testing.T) {
	repo := NewUserRepository(testPool)
	firstTenant := insertTenant(t, true)
	secondTenant := insertTenant(t, true)
	const email = "shared@example.com"

	first := user.User{ID: uuid.New(), TenantID: firstTenant, Email: email, PasswordHash: "hash"}
	if err := repo.Create(context.Background(), first); err != nil {
		t.Fatalf("Create() in first tenant error = %v, want nil", err)
	}

	second := user.User{ID: uuid.New(), TenantID: secondTenant, Email: email, PasswordHash: "hash"}
	if err := repo.Create(context.Background(), second); err != nil {
		t.Errorf("Create() same email in a second tenant error = %v, want nil (uniqueness is per tenant)", err)
	}
}

func TestUserRepositoryCreateHonoursCanceledContext(t *testing.T) {
	repo := NewUserRepository(testPool)
	tenantID := insertTenant(t, true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	u := user.User{ID: uuid.New(), TenantID: tenantID, Email: "canceled@example.com", PasswordHash: "hash"}
	if err := repo.Create(ctx, u); !errors.Is(err, context.Canceled) {
		t.Errorf("Create() error = %v, want %v", err, context.Canceled)
	}
}
