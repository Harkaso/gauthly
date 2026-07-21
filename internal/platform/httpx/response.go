package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// StatusResponse is a minimal JSON body carrying a single machine-readable
// status word, for endpoints that report an outcome without returning a
// resource.
type StatusResponse struct {
	Status string `json:"status"`
}

func writeJSONHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")
}

// WriteJSON sends payload as a JSON response with the given status.
//
// The payload is encoded as-is: its own type defines the response shape, and the
// status is carried only by the HTTP status line, never duplicated in the body.
// A failure to encode is logged rather than returned, since the status has
// already been committed by then; the payload is never logged, as it may hold
// data that must not reach the logs.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSONHeaders(w)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
