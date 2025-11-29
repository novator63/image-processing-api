package imageprocessing

import (
	"image"
	"github.com/disintegration/imaging"
)

type ImageProcessor interface {
	Crop(inputPath string, width int, height int) (image.Image, error)
	Resize(inputPath string, width int, height int) (image.Image, error)
	Blur(inputPath string, sigma float64) (image.Image, error)
	Contrast(inputPath string, percentage float64) (image.Image, error)
	Brightness(inputPath string, percentage float64) (image.Image, error)
	Sharpen(inputPath string, sigma float64) (image.Image, error)
	Grayscale(inputPath string) (image.Image, error)
	Invert(inputPath string) (image.Image, error)
}

type ImageProcessingService struct{}

func (p ImageProcessingService) Crop(inputPath string, width int, height int) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.CropCenter(src, width, height)
	return src, nil
}

func (p ImageProcessingService) Resize(inputPath string, width int, height int) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.Resize(src, width, height, imaging.Lanczos)
	return src, nil
}

func (p ImageProcessingService) Blur(inputPath string, sigma float64) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.Blur(src, sigma)
	return src, nil
}

func (p ImageProcessingService) Contrast(inputPath string, percentage float64) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.AdjustContrast(src, percentage)
	return src, nil
}

func (p ImageProcessingService) Brightness(inputPath string, percentage float64) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.AdjustBrightness(src, percentage)
	return src, nil
}

func (p ImageProcessingService) Sharpen(inputPath string, sigma float64) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.Sharpen(src, sigma)
	return src, nil
}

func (p ImageProcessingService) Grayscale(inputPath string) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.Grayscale(src)
	return src, nil
}

func (p ImageProcessingService) Invert(inputPath string) (image.Image, error) {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, err
	}
	src = imaging.Invert(src)
	return src, nil
}