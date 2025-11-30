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
	imageHandler := handlers.NewImageHandler(&storage, imageProcessor, cfg)

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

	router.Post("/images", imageHandler.UploadImage) 

	router.Get("/images", imageHandler.ImageIDsList)
	router.Get("/images/{id}", imageHandler.ImagesList)
	router.Get("/images/{id}/{filename}", imageHandler.GetImage)
	router.Delete("/images/{id}", imageHandler.DeleteImage)

	router.Post("/images/{id}/crop", imageHandler.CropImage)
	router.Post("/images/{id}/resize", imageHandler.ResizeImage)
	router.Post("/images/{id}/blur", imageHandler.BlurImage)
	router.Post("/images/{id}/contrast", imageHandler.ContrastImage)
	router.Post("/images/{id}/brightness", imageHandler.BrightnessImage)
	router.Post("/images/{id}/sharpen", imageHandler.SharpenImage)
	router.Post("/images/{id}/grayscale", imageHandler.GrayscaleImage)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("starting server error: %s", err)
	}
}
