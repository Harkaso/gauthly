package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Harkaso/gauthly/internal/auth"
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

type whoamiResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

type routerArgs struct {
	tenantRepo  tenant.Repository
	authHandler *auth.Handler
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

func newRouter(cfg config.Config, args routerArgs) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, httpx.StatusResponse{Status: "ok"})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, httpx.ErrNotFound)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(tenant.Resolve(args.tenantRepo, cfg.TenantMode, cfg.DefaultTenantID))

		r.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
			id, _ := tenant.FromContext(r.Context())
			httpx.WriteJSON(w, http.StatusOK, whoamiResponse{TenantID: id})
		})

		r.Post("/auth/register", args.authHandler.Register)
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

	tenantRepo := database.NewTenantRepository(dbPool)
	userRepo := database.NewUserRepository(dbPool)
	authService := auth.NewAuthService(userRepo)
	authHandler := auth.NewHandler(authService)

	slog.Info("starting server", "addr", "http://localhost"+cfg.Port)

	return http.ListenAndServe(cfg.Port, newRouter(*cfg, routerArgs{
		tenantRepo:  tenantRepo,
		authHandler: authHandler,
	}))
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
