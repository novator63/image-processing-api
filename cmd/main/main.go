package main

import (
	"encoding/json"
	"log"
	"net/http"
	"program/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	router := chi.NewRouter()

	server := &http.Server{
		Addr:        cfg.HTTPServer.Addres,
		Handler:     router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	router.Use(middleware.Logger)
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(cfg.Timeout))

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{"message":"pong"}
		json.NewEncoder(w).Encode(resp)
	})

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("starting server error: %s", err)
	}
}