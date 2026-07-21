package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

var (
	testTenantID        = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testDefaultTenantID = uuid.MustParse("00000000-0000-4000-8000-000000000000")
)

type fakeRepo struct {
	tenant Tenant
	err    error
}

func (f fakeRepo) GetByID(context.Context, uuid.UUID) (Tenant, error) {
	return f.tenant, f.err
}

type spyHandler struct {
	called   bool
	tenantID uuid.UUID
	found    bool
}

func (s *spyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.called = true
	s.tenantID, s.found = FromContext(r.Context())
	w.WriteHeader(http.StatusOK)
}

func serve(t *testing.T, repo Repository, mode Mode, defaultID uuid.UUID, header string) (*httptest.ResponseRecorder, *spyHandler) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set(headerTenantID, header)
	}

	rec := httptest.NewRecorder()
	next := &spyHandler{}
	Resolve(repo, mode, defaultID)(next).ServeHTTP(rec, req)

	return rec, next
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

func TestResolveRejectsRequest(t *testing.T) {
	activeRepo := fakeRepo{tenant: Tenant{ID: testTenantID, IsActive: true}}

	tests := []struct {
		name       string
		mode       Mode
		repo       Repository
		header     string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "MissingHeaderInB2B",
			mode:       B2B,
			repo:       activeRepo,
			header:     "",
			wantStatus: http.StatusBadRequest,
			wantCode:   "tenant_required",
		},
		{
			name:       "MalformedHeaderInB2B",
			mode:       B2B,
			repo:       activeRepo,
			header:     "not-a-uuid",
			wantStatus: http.StatusBadRequest,
			wantCode:   "tenant_invalid",
		},
		{
			name:       "HeaderPresentInB2C",
			mode:       B2C,
			repo:       activeRepo,
			header:     testTenantID.String(),
			wantStatus: http.StatusBadRequest,
			wantCode:   "tenant_forbidden",
		},
		{
			name:       "UnknownTenant",
			mode:       B2B,
			repo:       fakeRepo{err: ErrTenantNotFound},
			header:     testTenantID.String(),
			wantStatus: http.StatusNotFound,
			wantCode:   "tenant_not_found",
		},
		{
			name:       "DisabledTenant",
			mode:       B2B,
			repo:       fakeRepo{tenant: Tenant{ID: testTenantID, IsActive: false}},
			header:     testTenantID.String(),
			wantStatus: http.StatusNotFound,
			wantCode:   "tenant_not_found",
		},
		{
			name:       "RepositoryFailure",
			mode:       B2B,
			repo:       fakeRepo{err: errors.New("connection refused")},
			header:     testTenantID.String(),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, next := serve(t, tt.repo, tt.mode, testDefaultTenantID, tt.header)

			if rec.Code != tt.wantStatus {
				t.Errorf("Resolve() status = %v, want %v", rec.Code, tt.wantStatus)
			}
			if got := errorCode(t, rec); got != tt.wantCode {
				t.Errorf("Resolve() error code = %v, want %v", got, tt.wantCode)
			}
			if next.called {
				t.Error("Resolve() called the next handler, want the request rejected")
			}
		})
	}
}

func TestResolveServesRequest(t *testing.T) {
	tests := []struct {
		name   string
		mode   Mode
		header string
		wantID uuid.UUID
	}{
		{
			name:   "HeaderInB2B",
			mode:   B2B,
			header: testTenantID.String(),
			wantID: testTenantID,
		},
		{
			name:   "NoHeaderInB2C",
			mode:   B2C,
			header: "",
			wantID: testDefaultTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := fakeRepo{tenant: Tenant{ID: tt.wantID, IsActive: true}}
			rec, next := serve(t, repo, tt.mode, testDefaultTenantID, tt.header)

			if rec.Code != http.StatusOK {
				t.Errorf("Resolve() status = %v, want %v", rec.Code, http.StatusOK)
			}
			if !next.called {
				t.Fatal("Resolve() did not call the next handler, want the request served")
			}
			if !next.found {
				t.Error("Resolve() left no tenant in the request context, want one")
			}
			if next.tenantID != tt.wantID {
				t.Errorf("Resolve() context tenant = %v, want %v", next.tenantID, tt.wantID)
			}
		})
	}
}

func TestResolveUsesStoredTenantOverHeader(t *testing.T) {
	storedID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	repo := fakeRepo{tenant: Tenant{ID: storedID, IsActive: true}}

	_, next := serve(t, repo, B2B, testDefaultTenantID, testTenantID.String())

	if !next.found {
		t.Fatal("Resolve() left no tenant in the request context, want one")
	}
	if next.tenantID != storedID {
		t.Errorf("Resolve() context tenant = %v, want %v", next.tenantID, storedID)
	}
}

func TestResolveHidesWhetherTenantExists(t *testing.T) {
	header := testTenantID.String()

	unknown, _ := serve(t, fakeRepo{err: ErrTenantNotFound}, B2B, testDefaultTenantID, header)
	disabled, _ := serve(t, fakeRepo{tenant: Tenant{ID: testTenantID, IsActive: false}}, B2B, testDefaultTenantID, header)

	if unknown.Code != disabled.Code {
		t.Errorf("Resolve() status = %v for an unknown tenant and %v for a disabled one, want identical", unknown.Code, disabled.Code)
	}
	if unknown.Body.String() != disabled.Body.String() {
		t.Errorf("Resolve() body = %q for an unknown tenant and %q for a disabled one, want identical", unknown.Body.String(), disabled.Body.String())
	}
}

func TestResolvePanicsOnInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		mode      Mode
		defaultID uuid.UUID
	}{
		{name: "UndeclaredMode", mode: Mode(42), defaultID: testDefaultTenantID},
		{name: "B2CWithoutDefaultTenant", mode: B2C, defaultID: uuid.Nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Resolve() did not panic on an invalid configuration, want panic")
				}
			}()

			Resolve(fakeRepo{}, tt.mode, tt.defaultID)
		})
	}
}

func TestModeZeroValueIsB2B(t *testing.T) {
	if Mode(0) != B2B {
		t.Errorf("Mode(0) = %v, want B2B (%v)", Mode(0), B2B)
	}
}
