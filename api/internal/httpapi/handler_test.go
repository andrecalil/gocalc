package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler() http.Handler {
	return NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func doRequest(t *testing.T, h http.Handler, method, path, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCalculateHandler_Success(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/calculate", "application/json", `{"expression":"(1+1)*4"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var body successResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Result != "8" {
		t.Errorf("result = %q, want %q", body.Result, "8")
	}
}

func TestCalculateHandler_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{"empty expression", `{"expression":""}`, http.StatusBadRequest, "expression is empty"},
		{"invalid expression", `{"expression":"1+"}`, http.StatusBadRequest, "invalid expression"},
		{
			"unclosed parenthesis",
			`{"expression":"(1+1"}`,
			http.StatusBadRequest,
			"Please review your equation, it seems there's a missing parenthesis",
		},
		{"division by zero", `{"expression":"1/0"}`, http.StatusUnprocessableEntity, "division by zero"},
	}
	h := newTestHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(t, h, "POST", "/api/calculate", "application/json", tc.body)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Error != tc.wantError {
				t.Errorf("error = %q, want %q", body.Error, tc.wantError)
			}
		})
	}
}

func TestCalculateHandler_MalformedJSON(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/calculate", "application/json", `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error == "" {
		t.Errorf("expected error message, got empty")
	}
}

func TestCalculateHandler_WrongContentType(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/calculate", "text/plain", `{"expression":"1+1"}`)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", rec.Code)
	}
}

func TestCalculateHandler_EmptyBody(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/calculate", "application/json", ``)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCalculateHandler_MissingField(t *testing.T) {
	// {} decodes fine but yields empty expression → 400 empty
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/calculate", "application/json", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCalculateHandler_WrongMethod(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/api/calculate", "application/json", "")
	// Go 1.22+ pattern router returns 405 for wrong method on registered path.
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/healthz", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "OPTIONS", "/api/calculate", "", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Errorf("missing CORS origin header")
	}
}
