package apierror

import (
	"errors"
	"net/http"
)

type APIError struct {
	Err        error  `json:"-"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Err.Error()
}

func NewBadRequest(message string, err error) *APIError {
	if err == nil {
		err = errors.New(message)
	}
	return &APIError{
		Err:        err,
		Message:    message,
		Code:       "BAD_REQUEST",
		StatusCode: http.StatusBadRequest,
	}
}
func NewNotFound(message string, err error) *APIError {
	if err == nil {
		err = errors.New(message)
	}
	return &APIError{
		Err:        err,
		Message:    message,
		Code:       "NOT_FOUND",
		StatusCode: http.StatusNotFound,
	}
}
func NewValidation(message string, err error) *APIError {
	if err == nil {
		err = errors.New(message)
	}
	return &APIError{
		Err:        err,
		Message:    message,
		Code:       "VALIDATION_ERROR",
		StatusCode: http.StatusUnprocessableEntity,
	}
}
func NewInternal(err error) *APIError {
	if err == nil {
		err = errors.New("internal error")
	}
	return &APIError{
		Err:        err,
		Message:    "Internal server error",
		Code:       "INTERNAL",
		StatusCode: http.StatusInternalServerError,
	}
}
