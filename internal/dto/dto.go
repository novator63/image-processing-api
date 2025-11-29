package dto

type CropRequest struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ResizeRequest struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type BlurRequest struct {
	Sigma float64 `json:"sigma"`
}

type ContrastRequest struct {
	Percentage float64 `json:"percentage"`
}

type BrightnessRequest struct {
	Percentage float64 `json:"percentage"`
}

type SharpenRequest struct {
	Sigma float64 `json:"sigma"`
}

type GrayscaleRequest struct{}

type InvertRequest struct{}

type UploadResponse struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type OperationResponse struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Status	int	   `json: "status"'`
	Message string `json:"message"`
}