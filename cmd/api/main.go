package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Harkaso/gauthly/internal/config"
	"github.com/Harkaso/gauthly/internal/platform/database"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ctxTimeout = 20 * time.Second
)

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func connectDB(connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	return database.Connect(ctx, connString)
}

func startServer(port string) error {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(StatusResponse{Status: "ok"}); err != nil {
			log.Printf("failed to write response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "not found"}); err != nil {
			log.Printf("failed to write response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})

	return http.ListenAndServe(port, r)
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbPool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	log.Printf("server started on http://localhost%s", cfg.Port)
	err = startServer(cfg.Port)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}
