package handlerwrap

import (
	"context"
	"net/http"
)

type ctxKey string

// ключик по которому достаем
const errorKey ctxKey = "handler_error"

func WrapHandler(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := h(w, r); err != nil {
            // тута кладем
            ctx := context.WithValue(r.Context(), errorKey, err)
            // все равно не понимаю как это работает :(
            *r = *r.WithContext(ctx)
        }
    }
}

func GetError(r *http.Request) error {
    // тута достаем
    if err, ok := r.Context().Value(errorKey).(error); ok {
        return err
    }
    return nil
}
