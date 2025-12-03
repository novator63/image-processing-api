package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"program/internal/config"
	"program/internal/dto"
	"program/internal/http/apierror"
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

func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) error {
	file, header, err := r.FormFile("file")
	if err != nil {
		return apierror.NewBadRequest("failed to read file", err)
	}
	defer file.Close()

	extension := filepath.Ext(header.Filename)

	if !slices.Contains(h.cfg.Images.AllowedFormats, extension) {
		return apierror.NewValidation("unsupported media type", nil)
	}

	id := h.Storage.GenerateID()
	path, err := h.Storage.SaveFile(id, file, extension)
	if err != nil {
		return apierror.NewInternal(err)
	}

	uploadResponse := dto.UploadResponse{
		ID:   id,
		Path: path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadResponse)
	return nil
}

func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	//TODO проверить работает ли это
	if err := h.Storage.Exists(id); err != nil {
		return apierror.NewNotFound("image not found", err)
	}

	if err := h.Storage.DeleteFile(id); err != nil {
		return apierror.NewInternal(err)
	}

	//изначально тут было - w.WriteHeader(http.StatusNoContent), решить что должно быть возвращеное
	return nil
}

func (h *ImageHandler) CropImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}

	processedFilename := "crop" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	cropRequest := dto.CropRequest{}
	if err := json.NewDecoder(r.Body).Decode(&cropRequest); err != nil {
		return apierror.NewBadRequest("invalid JSON body", err)
	}

	if cropRequest.Width <= 0 || cropRequest.Height <= 0 {
		return apierror.NewValidation("width and height must be > 0", nil)
	}

	img, err := h.ImageProcessor.Crop(inputPath, cropRequest.Width, cropRequest.Height)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) ResizeImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}

	processedFilename := "resize" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	resizeRequest := dto.ResizeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&resizeRequest); err != nil {
		return apierror.NewBadRequest("invalid JSON body", err)
	}

	if resizeRequest.Width <= 0 || resizeRequest.Height <= 0 {
		return apierror.NewValidation("width and height must be > 0", nil)
	}

	img, err := h.ImageProcessor.Resize(inputPath, resizeRequest.Width, resizeRequest.Height)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) BlurImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}
	processedFilename := "blur" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	blurRequest := dto.BlurRequest{}
	if err := json.NewDecoder(r.Body).Decode(&blurRequest); err != nil {
		return apierror.NewBadRequest("invalid JSON body", err)
	}

	if blurRequest.Sigma <= 0 {
		return apierror.NewValidation("sigma must be > 0", nil)
	}

	img, err := h.ImageProcessor.Blur(inputPath, blurRequest.Sigma)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) ContrastImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}
	processedFilename := "contrast" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	contrastRequest := dto.ContrastRequest{}
	if err := json.NewDecoder(r.Body).Decode(&contrastRequest); err != nil {
		return apierror.NewBadRequest("invalid JSON body", err)
	}

	if contrastRequest.Percentage < -100 || contrastRequest.Percentage > 100 {
		return apierror.NewValidation("percentage must be in range [-100,100]", nil)
	}

	img, err := h.ImageProcessor.Contrast(inputPath, contrastRequest.Percentage)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) BrightnessImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}
	processedFilename := "brightness" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	brightnessRequest := dto.BrightnessRequest{}
	if err := json.NewDecoder(r.Body).Decode(&brightnessRequest); err != nil {
		return apierror.NewBadRequest("invalid JSON body", err)
	}

	if brightnessRequest.Percentage < -100 || brightnessRequest.Percentage > 100 {
		return apierror.NewValidation("percentage must be between -100 and 100", nil)
	}

	img, err := h.ImageProcessor.Brightness(inputPath, brightnessRequest.Percentage)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) SharpenImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}

	processedFilename := "sharpen" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	sharpenRequest := dto.SharpenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&sharpenRequest); err != nil {
		return apierror.NewBadRequest("invalid JSON body", err)
	}

	if sharpenRequest.Sigma <= 0 {
		return apierror.NewValidation("sigma must be > 0", nil)
	}

	img, err := h.ImageProcessor.Sharpen(inputPath, sharpenRequest.Sigma)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) GrayscaleImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	metaData, inputPath, err := h.prepareImage(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apierror.NewNotFound("image not found", err)
		}
		return apierror.NewInternal(err)
	}
	processedFilename := "grayscale" + metaData.Extension
	outputURL := "/images/" + id + "/" + processedFilename

	img, err := h.ImageProcessor.Grayscale(inputPath)
	if err != nil {
		return apierror.NewInternal(err)
	}

	if err := h.Storage.SaveProcessedFile(id, processedFilename, img); err != nil {
		return apierror.NewInternal(err)
	}

	operationResponse := dto.OperationResponse{
		ID:   id,
		Path: outputURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(operationResponse)
	return nil
}

func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	fileName := chi.URLParam(r, "filename")

	_, err := h.Storage.GetFilePath(id, fileName)
	if err != nil {
		return apierror.NewNotFound("image not found", err)
	}

	file, err := h.Storage.GetFile(id, fileName)
	if err != nil {
		return apierror.NewInternal(err)
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
	return nil
}

func (h *ImageHandler) ImagesList(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	files, err := h.Storage.ListImages(id)
	if err != nil {
		return apierror.NewInternal(err)
	}

	listFilesResponse := dto.ListFilesResponse{
		ID:    id,
		Files: files,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listFilesResponse)
	return nil
}

func (h *ImageHandler) ImageIDsList(w http.ResponseWriter, r *http.Request) error {
	IDsList, err := h.Storage.ListImageIDs()
	if err != nil {
		return apierror.NewInternal(err)
	}

	listFilesResponse := dto.IDsListResponse{
		IDs: IDsList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listFilesResponse)
	return nil
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
