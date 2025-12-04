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

type ListFilesResponse struct {
	ID    string   `json:"id"`
	Files []string `json:"files"`
}

type IDsListResponse struct {
	IDs []string `json:"ids"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Metadata struct {
	Extension string `json:"extension"`
	ID        string `json:"ID"`
}
