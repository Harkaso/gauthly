package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ tenant.Repository = (*TenantRepository)(nil)

// TenantRepository reads tenants from PostgreSQL.
type TenantRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRepository returns a repository reading through pool. The pool
// is borrowed, not owned: closing it remains the caller's responsibility.
func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// GetByID returns the tenant identified by id. It returns
// tenant.ErrTenantNotFound if no row matches. Any other
// error reports a database failure and wraps its cause.
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (tenant.Tenant, error) {
	const query = `SELECT id, is_active FROM tenants WHERE id = $1`

	var t tenant.Tenant

	err := r.pool.QueryRow(ctx, query, id).Scan(&t.ID, &t.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return tenant.Tenant{}, tenant.ErrTenantNotFound
	}
	if err != nil {
		return tenant.Tenant{}, fmt.Errorf("get tenant: %w", err)
	}

	return t, nil
}
