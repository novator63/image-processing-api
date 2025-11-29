package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"program/internal/config"
	"program/internal/dto"
	"program/internal/services/imageprocessing"
	"program/internal/services/storage"
	"slices"
)

type ImageHandler struct {
	Storage storage.ImageStorage
	ImageProcessor imageprocessing.ImageProcessor
	cfg     *config.Config
}

func NewUploadHandler(storage storage.ImageStorageService, imageProcessor imageprocessing.ImageProcessingService, cfg *config.Config) *ImageHandler {
	return &ImageHandler{
		Storage: storage,
		ImageProcessor: imageProcessor,
		cfg:     cfg,
	}
}

func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "reading file error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	extension := filepath.Ext(header.Filename)

	if !slices.Contains(h.cfg.Images.AllowedFormats, extension) {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	id := h.Storage.GenerateID()
	path, err := h.Storage.SaveFile(id, file, extension)
	if err != nil {
		http.Error(w, "saving file error", http.StatusInternalServerError)
		return
	}

	uploadResponse := dto.UploadResponse{
		ID:   id,
		Path: path,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uploadResponse)
}

func (h *ImageHandler) CropImage(w http.ResponseWriter, r *http.Request) {

}