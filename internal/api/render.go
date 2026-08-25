package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// writeError writes the standard error envelope.
func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	writeJSON(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message, Fields: fields}})
}

// writeValidation writes a 422 with per-field messages.
func writeValidation(w http.ResponseWriter, fields map[string]string) {
	writeError(w, http.StatusUnprocessableEntity, "validation_error", "one or more fields are invalid", fields)
}

// decodeJSON strictly decodes the request body into dst (1 MiB cap, no unknown fields).
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
