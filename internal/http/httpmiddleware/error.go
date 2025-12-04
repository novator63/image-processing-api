package httpmiddleware

import (
	"encoding/json"
	"log"
	"net/http"
	"program/internal/dto"
	"program/internal/http/apierror"
	"program/internal/http/handlerwrap"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		err := handlerwrap.GetError(r)
		if err == nil {
			return
		}

		if apiErr, ok := err.(*apierror.APIError); ok {
			log.Printf(
				"ERROR: %v | STATUS=%d | PATH=%s | METHOD=%s",
				apiErr.Err,
				apiErr.StatusCode,
				r.URL.Path,
				r.Method,
			)

			errorRespone := dto.ErrorResponse{
				Code:    apiErr.StatusCode,
				Message: apiErr.Message,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiErr.StatusCode)
			json.NewEncoder(w).Encode(errorRespone)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apierror.NewInternal(err))
	})
}
