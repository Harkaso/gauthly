package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Harkaso/gauthly/internal/config"
	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/google/uuid"
)

var (
	testTenantID        = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testDefaultTenantID = uuid.MustParse("00000000-0000-4000-8000-000000000000")
)

// fakeRepo is a tenant.Repository whose answer is fixed by the test, so the
// router can be exercised without a database.
type fakeRepo struct {
	tenant tenant.Tenant
	err    error
}

func (f fakeRepo) GetByID(context.Context, uuid.UUID) (tenant.Tenant, error) {
	return f.tenant, f.err
}

// request runs one request through the assembled router and returns the
// recorded response. An empty header is not sent at all.
func request(t *testing.T, repo tenant.Repository, mode tenant.Mode, method, target, header string) *httptest.ResponseRecorder {
	t.Helper()

	cfg := config.Config{TenantMode: mode, DefaultTenantID: testDefaultTenantID}
	req := httptest.NewRequest(method, target, nil)
	if header != "" {
		req.Header.Set("X-Tenant-ID", header)
	}

	rec := httptest.NewRecorder()
	newRouter(cfg, repo).ServeHTTP(rec, req)

	return rec
}

func TestRouterHealthzNeedsNoTenant(t *testing.T) {
	rec := request(t, fakeRepo{}, tenant.B2B, http.MethodGet, "/healthz", "")

	if rec.Code != http.StatusOK {
		t.Errorf("GET /healthz status = %v, want %v", rec.Code, http.StatusOK)
	}

	var body statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("GET /healthz status = %q, want %q", body.Status, "ok")
	}
}

func TestRouterNotFound(t *testing.T) {
	rec := request(t, fakeRepo{}, tenant.B2B, http.MethodGet, "/does-not-exist", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /does-not-exist status = %v, want %v", rec.Code, http.StatusNotFound)
	}
}

func TestRouterResolvesTenant(t *testing.T) {
	activeRepo := fakeRepo{tenant: tenant.Tenant{ID: testTenantID, IsActive: true}}

	tests := []struct {
		name       string
		repo       tenant.Repository
		mode       tenant.Mode
		header     string
		wantStatus int
	}{
		{
			name:       "ValidTenant",
			repo:       activeRepo,
			mode:       tenant.B2B,
			header:     testTenantID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "MissingHeaderInB2B",
			repo:       activeRepo,
			mode:       tenant.B2B,
			header:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "UnknownTenant",
			repo:       fakeRepo{err: tenant.ErrTenantNotFound},
			mode:       tenant.B2B,
			header:     testTenantID.String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "NoHeaderInB2C",
			repo:       fakeRepo{tenant: tenant.Tenant{ID: testDefaultTenantID, IsActive: true}},
			mode:       tenant.B2C,
			header:     "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := request(t, tt.repo, tt.mode, http.MethodGet, "/whoami", tt.header)

			if rec.Code != tt.wantStatus {
				t.Errorf("GET /whoami status = %v, want %v", rec.Code, tt.wantStatus)
			}
		})
	}
}

// TestRouterWhoamiReturnsStoredTenant proves the full chain: the router runs
// Resolve, which validates the tenant through the repository and places it in
// the context, where the handler reads it back. The identifier returned is the
// stored one, never the header, so a client cannot assert a tenant it does not
// own.
func TestRouterWhoamiReturnsStoredTenant(t *testing.T) {
	storedID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	repo := fakeRepo{tenant: tenant.Tenant{ID: storedID, IsActive: true}}

	rec := request(t, repo, tenant.B2B, http.MethodGet, "/whoami", testTenantID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /whoami status = %v, want %v", rec.Code, http.StatusOK)
	}

	var body whoamiResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if body.TenantID != storedID {
		t.Errorf("GET /whoami tenant_id = %v, want %v", body.TenantID, storedID)
	}
}

func TestNewLogHandler(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"JSONFormat", "json"},
		{"TextFormat", "text"},
		{"UnknownFormatFallsBackToText", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if newLogHandler(tt.format) == nil {
				t.Error("newLogHandler() = nil, want a handler")
			}
		})
	}
}
