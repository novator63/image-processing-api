package storage

import (
	"crypto/rand"
	"encoding/hex"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type ImageStorage interface {
	GenerateID() string
	SaveFile(id string, file multipart.File, extension string) (string, error)
	GetFile(id string, fileName string) ([]byte, error)
	DeleteFile(id string) error
}

type ImageStorageService struct {
	StoragePath string
}

//Метод генерации случайного ID
func (f ImageStorageService) GenerateID() string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// Метод предназначен для сохранения файла на диске. Создаёт директорию в корневой папке хранилища по
// переданному id string. Копирует содержимое file multipart.File и сохраняет в cозданную ранее директорию.
// Поддерживается сохранение изображений разного типа расширений ext string. 
func (f ImageStorageService) SaveFile(id string, file multipart.File, ext string) (string, error) {
	if _, err := os.Stat(f.StoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(f.StoragePath, 0755); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(filepath.Join(f.StoragePath, id), 0755); err != nil {
		return "", err
	}

	dstPath := filepath.Join(f.StoragePath, id, "original" + ext)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, file); err != nil {
		return "", err
	}
	return dstPath, nil
}


//Возвращает файл из каталога. id stirng - айди каталога,  filename string -
//имя возвращаемого файла (cropped, original, resized, ...)
func (f ImageStorageService) GetFile(id string, fileName string) ([]byte, error) {
	readPath := filepath.Join(f.StoragePath, id, fileName)

	file, err := os.Open(readPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//Удаляет каталог по переданному ID
func (f ImageStorageService) DeleteFile(id string) error {
	err := os. RemoveAll(filepath.Join(f.StoragePath, id))
	if err != nil {
		return err
	}
	return nil
}

//TODO метод для сохранения обработанного изображения
func (f ImageStorageService) SaveProcessedFile(id string, img image.Image, operation string) error {
	
}