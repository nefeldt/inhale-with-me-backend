package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

func (a *API) handleListCustomTypes(w http.ResponseWriter, r *http.Request) {
	types, err := a.store.ListCustomTypes(currentUserID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"types": types})
}

type customTypeRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (a *API) handleCreateCustomType(w http.ResponseWriter, r *http.Request) {
	var req customTypeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}
	name := strings.ToLower(strings.TrimSpace(req.Name))
	if !validTypeName(name) {
		writeValidation(w, map[string]string{"name": "1-30 characters: letters, numbers, space, . _ -"})
		return
	}
	if model.SessionType(name).Valid() {
		writeValidation(w, map[string]string{"name": "that is already a built-in type"})
		return
	}
	ct, err := a.store.UpsertCustomType(currentUserID(r), name, cleanColor(req.Color))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ct)
}

func (a *API) handleDeleteCustomType(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(chi.URLParam(r, "name"))
	if err := a.store.DeleteCustomType(currentUserID(r), name); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
