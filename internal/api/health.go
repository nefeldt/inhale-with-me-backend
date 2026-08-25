package api

import (
	"net/http"
	"time"
)

// handleHealth is a public liveness probe.
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}
