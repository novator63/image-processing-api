package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"program/internal/dto"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageStorage interface {
	DeleteFile(id string) error
	Exists(id string) error
	GenerateID() string
	GetFile(id string, fileName string) (*os.File, error)
	GetFilePath(id, fileName string) (string, error)
	LoadMetadata(id string) (dto.Metadata, error)
	ListImages(id string) ([]string, error)
	ListImageIDs() ([]string, error)
	SaveFile(id string, file multipart.File, extension string) (string, error)
	SaveProcessedFile(id, fileName string, img image.Image) error
}

type ImageStorageService struct {
	StoragePath string
}

const (
	metadataFileName = "meta.json"
	originalFileName = "original"
)

// Метод генерации случайного ID
func (f *ImageStorageService) GenerateID() string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// Метод предназначен для сохранения файла на диске. Создаёт директорию в корневой папке хранилища по
// переданному id string. Копирует содержимое file multipart.File и сохраняет в cозданную ранее директорию.
// Поддерживается сохранение изображений разного типа расширений ext string.
func (f *ImageStorageService) SaveFile(id string, file multipart.File, ext string) (string, error) {
	if _, err := os.Stat(f.StoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(f.StoragePath, 0755); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(filepath.Join(f.StoragePath, id), 0755); err != nil {
		return "", err
	}

	ext = strings.ToLower(ext)
	dstPath := filepath.Join(f.StoragePath, id, originalFileName+ext)
	dstURL := filepath.Join(id, originalFileName+ext)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	// Формируем метаданные
	metaData := dto.Metadata{
		Extension: ext,
		ID:        id,
	}

	if _, err = io.Copy(dstFile, file); err != nil {
		return "", err
	}

	if err := f.saveMetadata(id, metaData); err != nil {
		return "", err
	}

	return dstURL, nil
}

// Возвращает файл из каталога. id stirng - айди каталога,  filename string -
// имя возвращаемого файла (cropped, original, resized, ...)
func (f *ImageStorageService) GetFile(id string, fileName string) (*os.File, error) {
	readPath := filepath.Join(f.StoragePath, id, filepath.Base(fileName))

	file, err := os.Open(readPath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Удаляет каталог по переданному ID
func (f *ImageStorageService) DeleteFile(id string) error {
	err := os.RemoveAll(filepath.Join(f.StoragePath, id))
	if err != nil {
		return err
	}
	return nil
}

// Метод для сохранения обработанного изображения
func (f *ImageStorageService) SaveProcessedFile(id, fileName string, img image.Image) error {
	outputPath := filepath.Join(f.StoragePath, id, fileName)
	if err := imaging.Save(img, outputPath); err != nil {
		return err
	}
	return nil
}

// Метод считывает метаданные загруженного файла
func (f *ImageStorageService) LoadMetadata(id string) (dto.Metadata, error) {

	var metaData dto.Metadata
	metadataPath := filepath.Join(f.StoragePath, id, metadataFileName)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return dto.Metadata{}, err
	}

	if err := json.Unmarshal(data, &metaData); err != nil {
		return dto.Metadata{}, err
	}

	return metaData, nil
}

// Приватный метод, сохраняет метаданные
func (f *ImageStorageService) saveMetadata(id string, metaData dto.Metadata) error {
	data, err := json.Marshal(metaData)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(f.StoragePath, id, metadataFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return err
	}

	return nil
}

func (f *ImageStorageService) ListImages(id string) ([]string, error) {
	dirPath := filepath.Join(f.StoragePath, id)
	if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var fileNames []string
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

func (f *ImageStorageService) ListImageIDs() ([]string, error) {
	if _, err := os.Stat(f.StoragePath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var IDList []string
	files, err := os.ReadDir(f.StoragePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		IDList = append(IDList, file.Name())
	}

	return IDList, nil
}

// Возвращает путь до файла, также провереят его на существование
func (f *ImageStorageService) GetFilePath(id, fileName string) (string, error) {
	fileName = filepath.Base(fileName)
	filePath := filepath.Join(f.StoragePath, id, fileName)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return filePath, nil
}

func (s *ImageStorageService) Exists(id string) error {
	path := filepath.Join(s.StoragePath, id)
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return nil
}
