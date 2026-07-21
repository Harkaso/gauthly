package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/Harkaso/gauthly/internal/auth"
	"github.com/Harkaso/gauthly/internal/user"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ auth.UserRepository = (*UserRepository)(nil)

const uniqueViolation = "23505"

// UserRepository stores user accounts in PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository returns a repository writing through pool. The pool is
// borrowed, not owned: closing it remains the caller's responsibility.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts u. It returns auth.ErrEmailAlreadyExists when a user with the
// same email already exists in the tenant, which the UNIQUE(email, tenant_id)
// constraint enforces. Any other error reports a database failure and wraps its
// cause.
func (r *UserRepository) Create(ctx context.Context, u user.User) error {
	const query = `INSERT INTO users (id, tenant_id, email, password_hash) VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, query, u.ID, u.TenantID, u.Email, u.PasswordHash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return auth.ErrEmailAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
