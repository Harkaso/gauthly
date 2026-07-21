package tenant

import (
	"net/http"

	"github.com/Harkaso/gauthly/internal/platform/httpx"
)

var (
	errTenantRequired = httpx.APIError{
		Status:  http.StatusBadRequest,
		Code:    "tenant_required",
		Message: "tenant must be provided",
	}

	errTenantForbidden = httpx.APIError{
		Status:  http.StatusBadRequest,
		Code:    "tenant_forbidden",
		Message: "tenant must not be provided",
	}

	errTenantInvalid = httpx.APIError{
		Status:  http.StatusBadRequest,
		Code:    "tenant_invalid",
		Message: "tenant must be a valid UUID",
	}

	errTenantNotFound = httpx.APIError{
		Status:  http.StatusNotFound,
		Code:    "tenant_not_found",
		Message: "tenant not found",
	}
)
