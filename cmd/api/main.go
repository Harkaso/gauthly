package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Harkaso/gauthly/internal/config"
	"github.com/Harkaso/gauthly/internal/platform/database"
	"github.com/Harkaso/gauthly/internal/platform/httpx"
	"github.com/Harkaso/gauthly/internal/tenant"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ctxTimeout = 20 * time.Second
)

type statusResponse struct {
	Status string `json:"status"`
}

type whoamiResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

func newLogHandler(format string) slog.Handler {
	if format == "json" {
		return slog.NewJSONHandler(os.Stdout, nil)
	}
	return slog.NewTextHandler(os.Stdout, nil)
}

func connectDB(connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	return database.Connect(ctx, connString)
}

func newRouter(cfg config.Config, repo tenant.Repository) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, httpx.ErrNotFound)
	})

	r.Group(func(r chi.Router) {
		r.Use(tenant.Resolve(repo, cfg.TenantMode, cfg.DefaultTenantID))

		r.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
			id, _ := tenant.FromContext(r.Context())
			httpx.WriteJSON(w, http.StatusOK, whoamiResponse{TenantID: id})
		})
	})

	return r
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(newLogHandler(cfg.LogFormat)))
	dbPool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	repo := database.NewTenantRepository(dbPool)

	slog.Info("starting server", "addr", "http://localhost"+cfg.Port)

	return http.ListenAndServe(cfg.Port, newRouter(*cfg, repo))
}

func main() {
	err := run()
	if err != nil {
		slog.Error("failed to run", "error", err)
		// TODO : graceful shutdown
		os.Exit(1)
	}
	// TODO : graceful shutdown
}
