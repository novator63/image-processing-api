package main

import (
	"log"
	"net/http"
	"program/internal/config"
	"program/internal/handlers"
	"program/internal/services/imageprocessing"
	"program/internal/services/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()
	storage := storage.ImageStorageService{StoragePath: cfg.StoragePath}
	imageProcessor := imageprocessing.ImageProcessingService{}
	imageHandler := handlers.NewUploadHandler(&storage, imageProcessor, cfg)

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

	router.Post("/upload", imageHandler.UploadImage)
	router.Route("/images/{id}", func(r chi.Router) {
		r.Post("/crop", imageHandler.CropImage)
		r.Post("/resize", imageHandler.ResizeImage)
		r.Post("/blur", imageHandler.BlurImage)
	})

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("starting server error: %s", err)
	}
}