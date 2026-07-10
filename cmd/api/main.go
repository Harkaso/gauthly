package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	port string = ":8080"
	url  string = "http://localhost" + port
)

func startServer() error {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			log.Printf("Failed to write response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return http.ListenAndServe(port, r)
}

func main() {
	log.Printf("Server started on : %s\n", url)
	err := startServer()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
