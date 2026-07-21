package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Harkaso/gauthly/internal/platform/httpx"
)

const maxBodySize = 4 << 10

// Handler adapts the authentication use cases to HTTP.
type Handler struct {
	service *AuthService
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	errPasswordTooShort = httpx.APIError{
		Status:  http.StatusBadRequest,
		Code:    "password_too_short",
		Message: "password must be at least 12 characters",
	}

	errPasswordTooLong = httpx.APIError{
		Status:  http.StatusBadRequest,
		Code:    "password_too_long",
		Message: "password must be at most 128 characters",
	}

	errInvalidEmail = httpx.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_email",
		Message: "email must be a valid email address",
	}
)

// NewHandler returns a Handler backed by service.
func NewHandler(service *AuthService) *Handler {
	return &Handler{service: service}
}

// Register handles POST /api/v1/auth/register. It decodes the credentials,
// delegates to AuthService.Register within the request's tenant, and answers
// 202 Accepted.
//
// The response is identical whether the account was created or the email was
// already registered, so that it never reveals which addresses exist. Only an
// invalid request, a malformed body or a password outside the length policy, is
// reported distinctly, since those concern the input rather than account
// existence.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	reqBody := http.MaxBytesReader(w, r.Body, maxBodySize)

	var req registerRequest
	err := json.NewDecoder(reqBody).Decode(&req)
	if err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest)
		return
	}

	_, err = h.service.Register(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, ErrPasswordTooShort):
		httpx.WriteError(w, errPasswordTooShort)
		return
	case errors.Is(err, ErrPasswordTooLong):
		httpx.WriteError(w, errPasswordTooLong)
		return
	case errors.Is(err, ErrInvalidEmail):
		httpx.WriteError(w, errInvalidEmail)
		return
	case err != nil && !errors.Is(err, ErrEmailAlreadyExists):
		httpx.WriteError(w, httpx.ErrInternalServer)
		return
	}

	httpx.WriteJSON(w, http.StatusAccepted, httpx.StatusResponse{Status: "accepted"})
}
