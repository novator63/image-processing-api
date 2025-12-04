package main

import (
	"log"
	"net/http"
	"program/internal/config"
	"program/internal/handlers"
	"program/internal/http/handlerwrap"
	"program/internal/http/httpmiddleware"
	"program/internal/services/imageprocessing"
	"program/internal/services/storage"

	_ "program/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title 			Processing API
// @version 		1.0
// @description 	REST API для базовой обработки изображений
// @BasePath 		/
// @host 			localhost:8080

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
	router.Use(httpmiddleware.ErrorHandler)

	router.Get("/swagger/*", httpSwagger.WrapHandler)
	
	router.Post("/images", handlerwrap.WrapHandler(imageHandler.UploadImage))

	router.Get("/images", handlerwrap.WrapHandler(imageHandler.ImageIDsList))
	router.Get("/images/{id}", handlerwrap.WrapHandler(imageHandler.ImagesList))
	router.Get("/images/{id}/{filename}", handlerwrap.WrapHandler(imageHandler.GetImage))
	router.Delete("/images/{id}", handlerwrap.WrapHandler(imageHandler.DeleteImage))

	router.Post("/images/{id}/crop", handlerwrap.WrapHandler(imageHandler.CropImage))
	router.Post("/images/{id}/resize", handlerwrap.WrapHandler(imageHandler.ResizeImage))
	router.Post("/images/{id}/blur", handlerwrap.WrapHandler(imageHandler.BlurImage))
	router.Post("/images/{id}/contrast", handlerwrap.WrapHandler(imageHandler.ContrastImage))
	router.Post("/images/{id}/brightness", handlerwrap.WrapHandler(imageHandler.BrightnessImage))
	router.Post("/images/{id}/sharpen", handlerwrap.WrapHandler(imageHandler.SharpenImage))
	router.Post("/images/{id}/grayscale", handlerwrap.WrapHandler(imageHandler.GrayscaleImage))

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("starting server error: %s", err)
	}
}
