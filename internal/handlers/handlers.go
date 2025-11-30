package handlers

import (
	"encoding/json"
	"io"
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

func NewImageHandler(storage *storage.ImageStorageService, imageProcessor imageprocessing.ImageProcessingService, cfg *config.Config) *ImageHandler {
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
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}

	processedFilename := "resize" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	resizeRequest := dto.ResizeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&resizeRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if resizeRequest.Width <= 0 || resizeRequest.Height <= 0 {
		http.Error(w, "unprocessable entity error", http.StatusUnprocessableEntity)
		return
	}

	img, err := h.ImageProcessor.Resize(inputPath, resizeRequest.Width, resizeRequest.Height)
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

func (h *ImageHandler) BlurImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}
	processedFilename := "blur" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	blurRequest := dto.BlurRequest{}
	if err := json.NewDecoder(r.Body).Decode(&blurRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if blurRequest.Sigma <= 0 {
		http.Error(w, "unprocessable entity error", http.StatusUnprocessableEntity)
		return
	}

	img, err := h.ImageProcessor.Blur(inputPath, blurRequest.Sigma)
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

func (h *ImageHandler) ContrastImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}
	processedFilename := "contrast" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	contrastRequest := dto.ContrastRequest{}
	if err := json.NewDecoder(r.Body).Decode(&contrastRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if contrastRequest.Percentage < -100 && contrastRequest.Percentage > 100 {
		http.Error(w, "unprocessable entity error", http.StatusUnprocessableEntity)
		return
	}

	img, err := h.ImageProcessor.Contrast(inputPath, contrastRequest.Percentage)
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

func (h *ImageHandler) BrightnessImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}
	processedFilename := "brightness" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	brightnessRequest := dto.BrightnessRequest{}
	if err := json.NewDecoder(r.Body).Decode(&brightnessRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if brightnessRequest.Percentage < -100 && brightnessRequest.Percentage > 100 {
		http.Error(w, "unprocessable entity error", http.StatusUnprocessableEntity)
		return
	}

	img, err := h.ImageProcessor.Brightness(inputPath, brightnessRequest.Percentage)
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

func (h *ImageHandler) SharpenImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}

	processedFilename := "sharpen" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	sharpenRequest := dto.SharpenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&sharpenRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if sharpenRequest.Sigma <= 0 {
		http.Error(w, "unprocessable entity error", http.StatusUnprocessableEntity)
		return
	}

	img, err := h.ImageProcessor.Sharpen(inputPath, sharpenRequest.Sigma)
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

func (h *ImageHandler) GrayscaleImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		http.Error(w, "preparing image error", http.StatusBadRequest)
		return
	}
	processedFilename := "grayscale" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	img, err := h.ImageProcessor.Grayscale(inputPath)
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

func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fileName := chi.URLParam(r, "filename")

	_, err := h.Storage.GetFilePath(id, fileName)
	if err != nil {
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	file, err := h.Storage.GetFile(id, fileName)
	if err != nil {
		http.Error(w, "get file error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileName)
	switch ext {
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpg")
	 case ".png":
		w.Header().Set("Content-Type", "image/png")
	}
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

// Подготавливает данные для работы с изображением, возвращает метаданные, исходный путь файла,
// в противном слчуае возвращает ошибку
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
