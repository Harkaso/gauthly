package tenant

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID uuid.UUID
		wantOK bool
	}{
		{
			name:   "ValidTenant",
			ctx:    WithTenant(context.Background(), testTenantID),
			wantID: testTenantID,
			wantOK: true,
		},
		{
			name:   "NoTenant",
			ctx:    context.Background(),
			wantID: uuid.Nil,
			wantOK: false,
		},
		{
			name:   "InvalidTypeValue",
			ctx:    context.WithValue(context.Background(), ctxKey{}, testTenantID.String()),
			wantID: uuid.Nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := FromContext(tt.ctx)
			if gotOK != tt.wantOK {
				t.Errorf("FromContext() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotID != tt.wantID {
				t.Errorf("FromContext() id = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestRequireFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantID  uuid.UUID
		wantErr error
	}{
		{
			name:   "ValidTenant",
			ctx:    WithTenant(context.Background(), testTenantID),
			wantID: testTenantID,
		},
		{
			name:    "NoTenant",
			ctx:     context.Background(),
			wantID:  uuid.Nil,
			wantErr: ErrNotInContext,
		},
		{
			name:    "InvalidTypeValue",
			ctx:     context.WithValue(context.Background(), ctxKey{}, testTenantID.String()),
			wantID:  uuid.Nil,
			wantErr: ErrNotInContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := RequireFromContext(tt.ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("RequireFromContext() error = %v, want %v", err, tt.wantErr)
			}
			if gotID != tt.wantID {
				t.Errorf("RequireFromContext() id = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestWithTenantPanicsOnNilTenant(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("WithTenant() did not panic on uuid.Nil, want panic")
		}
	}()

	WithTenant(context.Background(), uuid.Nil)
}
