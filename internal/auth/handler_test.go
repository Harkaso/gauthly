package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Harkaso/gauthly/internal/platform/httpx"
	"github.com/Harkaso/gauthly/internal/tenant"
)

const validRequestBody = `{"email":"user@example.com","password":"correct horse battery"}`

func serveRegister(t *testing.T, repo *fakeUserRepo, withTenant bool, body string) *httptest.ResponseRecorder {
	t.Helper()

	h := NewHandler(NewAuthService(repo))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	if withTenant {
		req = req.WithContext(tenant.WithTenant(req.Context(), testTenantID))
	}

	rec := httptest.NewRecorder()
	h.Register(rec, req)

	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode error body: %v", err)
	}

	return body.Error.Code
}

func TestRegisterHandlerAcceptsValidRequest(t *testing.T) {
	repo := &fakeUserRepo{}

	rec := serveRegister(t, repo, true, validRequestBody)
	if rec.Code != http.StatusAccepted {
		t.Errorf("Register() status = %v, want %v", rec.Code, http.StatusAccepted)
	}
	if !repo.called {
		t.Error("Register() did not reach the service, want a user created")
	}

	var body httpx.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if body.Status != "accepted" {
		t.Errorf("Register() body status = %q, want %q", body.Status, "accepted")
	}
}

func TestRegisterHandlerRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{"PasswordTooShort", `{"email":"user@example.com","password":"short"}`, "password_too_short"},
		{"PasswordTooLong", `{"email":"user@example.com","password":"` + strings.Repeat("a", 129) + `"}`, "password_too_long"},
		{"InvalidEmail", `{"email":"not-an-email","password":"correct horse battery"}`, "invalid_email"},
		{"MalformedBody", `{"email":`, "bad_request"},
		{"BodyTooLarge", `{"email":"user@example.com","password":"` + strings.Repeat("a", 5000) + `"}`, "bad_request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveRegister(t, &fakeUserRepo{}, true, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Register() status = %v, want %v", rec.Code, http.StatusBadRequest)
			}
			if got := errorCode(t, rec); got != tt.wantCode {
				t.Errorf("Register() error code = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

func TestRegisterHandlerRequiresTenant(t *testing.T) {
	rec := serveRegister(t, &fakeUserRepo{}, false, validRequestBody)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Register() without a tenant status = %v, want %v", rec.Code, http.StatusInternalServerError)
	}
	if got := errorCode(t, rec); got != "internal_error" {
		t.Errorf("Register() error code = %q, want %q", got, "internal_error")
	}
}

func TestRegisterHandlerHidesWhetherEmailExists(t *testing.T) {
	created := serveRegister(t, &fakeUserRepo{}, true, validRequestBody)
	duplicate := serveRegister(t, &fakeUserRepo{err: ErrEmailAlreadyExists}, true, validRequestBody)

	if created.Code != duplicate.Code {
		t.Errorf("Register() status = %v for a new email and %v for an existing one, want identical", created.Code, duplicate.Code)
	}
	if created.Body.String() != duplicate.Body.String() {
		t.Errorf("Register() body = %q for a new email and %q for an existing one, want identical", created.Body.String(), duplicate.Body.String())
	}
}
