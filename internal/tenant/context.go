package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNotInContext is returned by RequireFromContext when ctx carries no
// tenant, which means the request never went through Resolve.
var ErrNotInContext = errors.New("tenant not in context")

type ctxKey struct{}

// WithTenant returns a copy of ctx carrying tenantID as the tenant of the
// request. It panics if tenantID is uuid.Nil: at this point an absent tenant
// is a programming error, not a condition callers are meant to handle.
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	if tenantID == uuid.Nil {
		panic("tenant: tenantID must not be nil")
	}

	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext returns the tenant identifier carried by ctx. The boolean
// reports whether one was present; when it is false the returned identifier is
// uuid.Nil and the caller must refuse to proceed rather than fall back to an
// unscoped query.
func FromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
	return id, ok
}

// RequireFromContext returns the tenant identifier carried by ctx, or
// ErrNotInContext if there is none. Repositories call it to refuse a query
// they could not scope to a tenant, instead of running it unscoped.
func RequireFromContext(ctx context.Context) (uuid.UUID, error) {
	id, ok := FromContext(ctx)
	if !ok {
		return uuid.Nil, ErrNotInContext
	}

	return id, nil
}
