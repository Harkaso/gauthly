package tenant

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Harkaso/gauthly/internal/platform/httpx"
	"github.com/google/uuid"
)

const headerTenantID = "X-Tenant-ID"

// Resolve returns a middleware that establishes the tenant of each request and
// stores its identifier in the request context for the handlers that follow.
//
// The request is rejected and next is not called whenever the tenant cannot be
// established: header missing or unexpected for mode, identifier malformed,
// tenant unknown, or tenant disabled. Unknown and disabled tenants produce the
// same response, so that responses never disclose which identifiers exist.
//
// Resolve panics if mode is not one of the declared modes, or if mode is B2C
// and defaultTenantID is uuid.Nil. Both are configuration errors and are
// reported at startup rather than on the first request.
func Resolve(repo Repository, mode Mode, defaultTenantID uuid.UUID) func(http.Handler) http.Handler {
	if mode != B2C && mode != B2B {
		panic(fmt.Sprintf("tenant: invalid mode %d", mode))
	}
	if mode == B2C && defaultTenantID == uuid.Nil {
		panic("tenant: defaultTenantID must be set in B2C mode")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get(headerTenantID)
			var tenantID uuid.UUID

			if mode == B2C {
				if header != "" {
					httpx.WriteError(w, errTenantForbidden)
					return
				}

				tenantID = defaultTenantID
			} else {
				if header == "" {
					httpx.WriteError(w, errTenantRequired)
					return
				}

				id, err := uuid.Parse(header)
				if err != nil {
					httpx.WriteError(w, errTenantInvalid)
					return
				}
				tenantID = id
			}

			t, err := repo.GetByID(r.Context(), tenantID)
			if errors.Is(err, ErrTenantNotFound) {
				httpx.WriteError(w, errTenantNotFound)
				return
			}
			if err != nil {
				slog.ErrorContext(r.Context(), "failed to get tenant", "tenant_id", tenantID, "error", err)
				httpx.WriteError(w, httpx.ErrInternalServer)
				return
			}
			if !t.IsActive {
				httpx.WriteError(w, errTenantNotFound)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithTenant(r.Context(), t.ID)))
		})
	}
}
