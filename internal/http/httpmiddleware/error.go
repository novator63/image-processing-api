package httpmiddleware

import (
	"encoding/json"
	"net/http"
	"program/internal/http/handlerwrap"
	"program/internal/http/apierror"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		err := handlerwrap.GetError(r)
		if err == nil {
			return
		}

		// TODO  отдельное логирование отдельные client error response
		if apiErr, ok := err.(*apierror.APIError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiErr.StatusCode)
			json.NewEncoder(w).Encode(apiErr)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apierror.NewInternal(err))
	})
}
