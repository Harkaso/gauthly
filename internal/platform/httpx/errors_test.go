package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testError = APIError{
	Status:  http.StatusNotFound,
	Code:    "resource_not_found",
	Message: "resource not found",
}

func TestWriteErrorStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, testError)

	if rec.Code != testError.Status {
		t.Errorf("WriteError() status = %v, want %v", rec.Code, testError.Status)
	}
}

func TestWriteErrorHeaders(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, testError)

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
				t.Errorf("WriteError() %s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestWriteErrorBody(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, testError)

	var body errorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode error body: %v", err)
	}

	if body.Error.Code != testError.Code {
		t.Errorf("WriteError() code = %v, want %v", body.Error.Code, testError.Code)
	}
	if body.Error.Message != testError.Message {
		t.Errorf("WriteError() message = %q, want %q", body.Error.Message, testError.Message)
	}
}

func TestWriteErrorBodyOmitsStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, testError)

	if strings.Contains(rec.Body.String(), "status") {
		t.Errorf("WriteError() body = %q, want no status field", rec.Body.String())
	}
}

type unwritableWriter struct {
	header http.Header
}

func (u *unwritableWriter) Header() http.Header {
	if u.header == nil {
		u.header = http.Header{}
	}

	return u.header
}

func (u *unwritableWriter) Write([]byte) (int, error) {
	return 0, errors.New("connection reset by peer")
}

func (u *unwritableWriter) WriteHeader(int) {}

func TestWriteErrorSurvivesUnwritableResponse(t *testing.T) {
	WriteError(&unwritableWriter{}, testError)
}

func TestWriteErrorDropsStaleContentLength(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Length", "9999")

	WriteError(rec, testError)

	if got := rec.Header().Get("Content-Length"); got != "" {
		t.Errorf("WriteError() Content-Length = %q, want empty", got)
	}
}
