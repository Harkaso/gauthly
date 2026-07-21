package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorCode identifies an error condition in a form clients can match on.
// It is part of the public API contract: unlike a message,
// it is never rephrased.
type ErrorCode string

// APIError is an error response: the status to send, the code clients match on,
// and the message a developer reads. Callers declare one value per condition and
// hand it to WriteError, so that a given condition always answers identically.
type APIError struct {
	Status  int
	Code    ErrorCode
	Message string
}

// ErrInternalServer is the response to a failure the caller can do nothing about. It
// is defined here rather than per package so that every domain reports such a
// failure identically, and it carries no detail: the cause belongs in the logs.
var ErrInternalServer = APIError{
	Status:  http.StatusInternalServerError,
	Code:    "internal_error",
	Message: "internal error",
}

// ErrNotFound is the response to a request for a route or resource that does
// not exist. Like ErrInternalServer, it is shared so that every unmatched
// request answers identically.
var ErrNotFound = APIError{
	Status:  http.StatusNotFound,
	Code:    "not_found",
	Message: "not found",
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// WriteError sends e as a JSON error response.
//
// It must be called before anything else is written to w: sending the status
// freezes the headers, and any later change to them is silently dropped.
// A failure to encode the body is logged rather than returned,
// since the status has already been committed by then.
func WriteError(w http.ResponseWriter, e APIError) {
	writeJSONHeaders(w)
	w.WriteHeader(e.Status)
	body := errorBody{Error: errorDetail{Code: e.Code, Message: e.Message}}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to write error response", "code", e.Code, "error", err)
	}
}
