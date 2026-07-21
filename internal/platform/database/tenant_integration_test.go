//go:build integration

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage  = "postgres:18"
	migrationsDir  = "../../../migrations"
	seededTenantID = "00000000-0000-4000-8000-000000000000"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase("gauthly_test"),
		postgres.WithUsername("gauthly"),
		postgres.WithPassword("gauthly"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}

	code, err := runTests(ctx, container, m)

	if terminateErr := testcontainers.TerminateContainer(container); terminateErr != nil {
		log.Printf("Failed to terminate postgres container: %v", terminateErr)
	}
	if err != nil {
		log.Fatalf("Failed to prepare the database: %v", err)
	}

	os.Exit(code)
}

func runTests(ctx context.Context, container *postgres.PostgresContainer, m *testing.M) (int, error) {
	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return 0, fmt.Errorf("read connection string: %w", err)
	}

	if err := migrateUp(connString); err != nil {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}

	testPool, err = Connect(ctx, connString)
	if err != nil {
		return 0, fmt.Errorf("connect to database: %w", err)
	}
	defer testPool.Close()

	return m.Run(), nil
}

func migrateUp(connString string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetLogger(goose.NopLogger())

	return goose.Up(db, migrationsDir)
}

func insertTenant(t *testing.T, isActive bool) uuid.UUID {
	t.Helper()

	const query = `INSERT INTO tenants (id, name, slug, is_active) VALUES ($1, $2, $3, $4)`

	id := uuid.New()
	if _, err := testPool.Exec(context.Background(), query, id, id.String(), id.String(), isActive); err != nil {
		t.Fatalf("Failed to insert tenant: %v", err)
	}

	return id
}

func TestTenantRepositoryGetByID(t *testing.T) {
	repo := NewTenantRepository(testPool)

	tests := []struct {
		name       string
		tenantID   func(t *testing.T) uuid.UUID
		wantActive bool
		wantErr    error
	}{
		{
			name:       "ActiveTenant",
			tenantID:   func(t *testing.T) uuid.UUID { return insertTenant(t, true) },
			wantActive: true,
		},
		{
			name:       "DisabledTenant",
			tenantID:   func(t *testing.T) uuid.UUID { return insertTenant(t, false) },
			wantActive: false,
		},
		{
			name:       "SeededDefaultTenant",
			tenantID:   func(*testing.T) uuid.UUID { return uuid.MustParse(seededTenantID) },
			wantActive: true,
		},
		{
			name:     "UnknownTenant",
			tenantID: func(*testing.T) uuid.UUID { return uuid.New() },
			wantErr:  tenant.ErrTenantNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.tenantID(t)

			got, err := repo.GetByID(context.Background(), id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetByID() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			if got.ID != id {
				t.Errorf("GetByID() id = %v, want %v", got.ID, id)
			}
			if got.IsActive != tt.wantActive {
				t.Errorf("GetByID() isActive = %v, want %v", got.IsActive, tt.wantActive)
			}
		})
	}
}

func TestTenantRepositoryGetByIDHonoursCanceledContext(t *testing.T) {
	repo := NewTenantRepository(testPool)
	id := insertTenant(t, true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetByID(ctx, id)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("GetByID() error = %v, want %v", err, context.Canceled)
	}
}
