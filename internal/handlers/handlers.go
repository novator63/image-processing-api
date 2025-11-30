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

	"github.com/go-chi/chi/v5"
)

type ImageHandler struct {
	Storage        storage.ImageStorage
	ImageProcessor imageprocessing.ImageProcessor
	cfg            *config.Config
}

func NewUploadHandler(storage *storage.ImageStorageService, imageProcessor imageprocessing.ImageProcessingService, cfg *config.Config) *ImageHandler {
	return &ImageHandler{
		Storage:        storage,
		ImageProcessor: imageProcessor,
		cfg:            cfg,
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadResponse)
}

func (h *ImageHandler) CropImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}

	processedFilename := "crop" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	cropRequest := dto.CropRequest{}
	if err := json.NewDecoder(r.Body).Decode(&cropRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if cropRequest.Width <= 0 || cropRequest.Height <= 0 {
		http.Error(w, "unprocessable entity error", http.StatusUnprocessableEntity)
		return
	}

	img, err := h.ImageProcessor.Crop(inputPath, cropRequest.Width, cropRequest.Height)
	if err != nil {
		http.Error(w, "crop image error", http.StatusInternalServerError)
		return
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		http.Error(w, "saving processed file error", http.StatusInternalServerError)
		return
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
}

func (h *ImageHandler) ResizeImage(w http.ResponseWriter, r *http.Request) {

}

func (h *ImageHandler) BlurImage(w http.ResponseWriter, r *http.Request) {

}

func (h *ImageHandler) ContrastImage(w http.ResponseWriter, r *http.Request) {

}
func (h *ImageHandler) BrightnessImage(w http.ResponseWriter, r *http.Request) {

}
func (h *ImageHandler) SharpenImage(w http.ResponseWriter, r *http.Request) {

}
func (h *ImageHandler) GrayscaleImage(w http.ResponseWriter, r *http.Request) {

}
func (h *ImageHandler) InvertImage(w http.ResponseWriter, r *http.Request) {

}

func (h *ImageHandler) prepareImage(id string) (metaData dto.Metadata, inputPath string, err error) {
	metaData, err = h.Storage.LoadMetadata(id)
	if err != nil {
		return dto.Metadata{}, "", err
	}

	originalFilename := "original" + metaData.Extension

	inputPath, err = h.Storage.GetFilePath(id, originalFilename)
	if err != nil {
		return dto.Metadata{}, "", err
	}

	return metaData, inputPath, nil
}
