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
	"strings"

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

// UploadImage godoc
// @Summary      Загрузить изображение
// @Description  Принимает изображение (multipart/form-data), сохраняет и возвращает ID и путь
// @Tags         images
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Загружаемый файл"
// @Success      201   {object}  dto.UploadResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      422   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /images [post]
func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, int64(h.cfg.Images.MaxUploadSizeMb)<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrBodyReadAfterClose) ||
			strings.Contains(err.Error(), "http: request body too large") {
			return apierror.NewValidation("file is too large", nil)
		}

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

// DeleteImage godoc
// @Summary      Удалить набор изображений
// @Tags         images
// @Produce      json
// @Param        id   path   string  true  "ID набора"
// @Success      204  {string} string "deleted"
// @Failure      404  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /images/{id} [delete]
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

// CropImage godoc
// @Summary      Обрезать изображение
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id    path   string         true  "ID набора"
// @Param        body  body   dto.CropRequest  true  "Параметры обрезки"
// @Success      200   {object} dto.OperationResponse
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      422   {object} dto.ErrorResponse
// @Failure      500   {object} dto.ErrorResponse
// @Router       /images/{id}/crop [post]
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

// ResizeImage godoc
// @Summary      Изменить размер изображения
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id    path   string            true  "ID набора"
// @Param        body  body   dto.ResizeRequest  true  "Параметры resize"
// @Success      200   {object} dto.OperationResponse
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      422   {object} dto.ErrorResponse
// @Failure      500   {object} dto.ErrorResponse
// @Router       /images/{id}/resize [post]
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

// BlurImage godoc
// @Summary      Размытие изображения (Gaussian Blur)
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id    path   string         true  "ID набора"
// @Param        body  body   dto.BlurRequest  true  "Sigma (радиус размытия)"
// @Success      200   {object} dto.OperationResponse
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      422   {object} dto.ErrorResponse
// @Failure      500   {object} dto.ErrorResponse
// @Router       /images/{id}/blur [post]
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

// ContrastImage godoc
// @Summary      Изменить контрастность изображения
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id    path   string            true  "ID набора"
// @Param        body  body   dto.ContrastRequest  true  "[-100..100]"
// @Success      200   {object} dto.OperationResponse
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      422   {object} dto.ErrorResponse
// @Failure      500   {object} dto.ErrorResponse
// @Router       /images/{id}/contrast [post]
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

// BrightnessImage godoc
// @Summary      Изменить яркость изображения
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id    path   string              true  "ID набора"
// @Param        body  body   dto.BrightnessRequest  true  "[-100..100]"
// @Success      200   {object} dto.OperationResponse
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      422   {object} dto.ErrorResponse
// @Failure      500   {object} dto.ErrorResponse
// @Router       /images/{id}/brightness [post]
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

// SharpenImage godoc
// @Summary      Повысить резкость изображения
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id    path   string            true  "ID набора"
// @Param        body  body   dto.SharpenRequest  true  "Sigma (степень резкости)"
// @Success      200   {object} dto.OperationResponse
// @Failure      400   {object} dto.ErrorResponse
// @Failure      404   {object} dto.ErrorResponse
// @Failure      422   {object} dto.ErrorResponse
// @Failure      500   {object} dto.ErrorResponse
// @Router       /images/{id}/sharpen [post]
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

// GrayscaleImage godoc
// @Summary      Конвертировать изображение в grayscale
// @Tags         processing
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "ID набора"
// @Success      200  {object} dto.OperationResponse
// @Failure      404  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /images/{id}/grayscale [post]
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

// GetImage godoc
// @Summary      Получить изображение по имени файла
// @Tags         images
// @Produce      jpeg
// @Param        id       path   string  true  "ID набора"
// @Param        filename path   string  true  "Имя файла"
// @Success      200      {file}  binary
// @Failure      404      {object} dto.ErrorResponse
// @Failure      500      {object} dto.ErrorResponse
// @Router       /images/{id}/{filename} [get]
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

// ImagesList godoc
// @Summary      Получить список файлов набора
// @Tags         images
// @Produce      json
// @Param        id   path   string  true  "ID набора"
// @Success      200  {object} dto.ListFilesResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /images/{id} [get]
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

// ImageIDsList godoc
// @Summary      Получить список всех ID наборов изображений
// @Tags         images
// @Produce      json
// @Success      200  {object} dto.IDsListResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /images [get]
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
