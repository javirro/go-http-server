package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/javier/go-http-server/internal/middleware"
)

// Envelope is the standard JSON response wrapper.
type Envelope[T any] struct {
	Data  T      `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// At this point headers are already sent; best effort logging.
		slog.ErrorContext(r.Context(), "failed to encode JSON response",
			slog.Any("error", err),
			slog.String("request_id", middleware.GetRequestID(r.Context())),
		)
	}
}

// ErrorResponse writes a standardised JSON error response.
func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	JSON(w, r, status, Envelope[any]{Error: message})
}

// DecodeJSON reads and decodes the request body into v.
// Returns false and writes a 400 response if decoding fails.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}
