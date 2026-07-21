package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrTenantNotFound is returned by Repository.GetByID when no tenant matches
// the requested identifier.
var ErrTenantNotFound = errors.New("tenant not found")

// Tenant is the subset of a stored tenant needed to admit or reject a request.
// It deliberately carries no descriptive field.
type Tenant struct {
	ID       uuid.UUID
	IsActive bool
}

// Repository gives access to the stored tenants.
type Repository interface {
	// GetByID returns the tenant identified by id. It returns ErrTenantNotFound
	// if no such tenant exists; any other error reports a failure of the
	// underlying store, not an absent tenant.
	GetByID(ctx context.Context, id uuid.UUID) (Tenant, error)
}
