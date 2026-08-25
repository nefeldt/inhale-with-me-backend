package api

import (
	"errors"
	"net/http"

	"github.com/nfeldt/inhale-with-me/internal/store"
)

// writeStoreError maps a store sentinel error to an HTTP error response.
func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
	case errors.Is(err, store.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "resource already exists", nil)
	case errors.Is(err, store.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "you are not allowed to do that", nil)
	default:
		writeError(w, http.StatusInternalServerError, "internal", "internal server error", nil)
	}
}
