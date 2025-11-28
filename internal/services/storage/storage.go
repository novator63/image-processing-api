package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type ImageStorage interface {
	GenerateID() (string, error)
	SaveFile(id string, file multipart.File, extension string) (string, error)
	GetFile() error
	DeleteFile() error
}

type FileSytstemStorage struct {
	storagePath string
}

func (f FileSytstemStorage) GenerateID() (string, error) {

}

// Функция предназначена для сохранения файла на диске. Создаёт директорию в корневой папке хранилища по
// переданному id string. Копирует содержимое file multipart.File и сохраняет в cозданную ранее директорию.
// Поддерживается сохранение изображений разного типа расширений ext string. 
func (f FileSytstemStorage) SaveFile(id string, file multipart.File, ext string) (string, error) {
	if _, err := os.Stat(f.storagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(f.storagePath, 0755); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(f.storagePath + "/" + id, 0755); err != nil {
		return "", err
	}
	dstPath := filepath.Join(f.storagePath, id, "original" + ext)

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

func (f FileSytstemStorage) GetFile() error {

}

func (f FileSytstemStorage) DeleteFile() error {

}
