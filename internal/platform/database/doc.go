// Package database opens the PostgreSQL connection pool shared by the service.
//
// It owns the pool and its settings only — sizing, timeouts and health checks.
// Queries belong to the repository implementations built on top of it.
package database
