package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testPayload struct {
	Value string `json:"value"`
}

var testJSON = testPayload{Value: "ok"}

func TestWriteJSONStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusCreated, testJSON)

	if rec.Code != http.StatusCreated {
		t.Errorf("WriteJSON() status = %v, want %v", rec.Code, http.StatusCreated)
	}
}

func TestWriteJSONHeaders(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusOK, testJSON)

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"ContentType", "Content-Type", "application/json"},
		{"ContentTypeOptions", "X-Content-Type-Options", "nosniff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rec.Header().Get(tt.header); got != tt.want {
				t.Errorf("WriteJSON() %s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestWriteJSONBody(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusOK, testJSON)

	var body testPayload
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if body != testJSON {
		t.Errorf("WriteJSON() body = %+v, want %+v", body, testJSON)
	}
}

// TestWriteJSONDropsStaleContentLength covers a length left by an earlier
// writer, which would no longer match the body being sent.
func TestWriteJSONDropsStaleContentLength(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Length", "9999")

	WriteJSON(rec, http.StatusOK, testJSON)

	if got := rec.Header().Get("Content-Length"); got != "" {
		t.Errorf("WriteJSON() Content-Length = %q, want empty", got)
	}
}

// TestWriteJSONSurvivesUnwritableResponse covers the encoding failure path: the
// status is already committed by then, so the failure can only be logged. The
// test asserts that WriteJSON returns rather than panicking.
func TestWriteJSONSurvivesUnwritableResponse(t *testing.T) {
	WriteJSON(&unwritableWriter{}, http.StatusOK, testJSON)
}
