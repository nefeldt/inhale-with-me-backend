package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type registerDeviceRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

// handleRegisterDevice stores the caller's APNs device token for push delivery.
func (a *API) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req registerDeviceRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		writeValidation(w, map[string]string{"token": "token is required"})
		return
	}
	d, err := a.store.RegisterDevice(currentUserID(r), token, req.Platform)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

// handleUnregisterDevice removes a device token (e.g. on logout).
func (a *API) handleUnregisterDevice(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteDevice(currentUserID(r), chi.URLParam(r, "token")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
