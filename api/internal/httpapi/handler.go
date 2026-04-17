// Package httpapi wires calc.Calculate up to an HTTP endpoint. The handler is
// deliberately thin: decode JSON, delegate to calc, map the error to a status
// code, encode the JSON response.
package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/andrecalil/go-calc/api/internal/calc"
)

type calcRequest struct {
	Expression string `json:"expression"`
}

type successResponse struct {
	Result string `json:"result"`
}

type errorBody struct {
	Error string `json:"error"`
}

// NewHandler builds the root http.Handler with routing and middleware.
func NewHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/calculate", handleCalculate(logger))
	mux.HandleFunc("GET /healthz", handleHealthz)
	return withLogging(logger, withCORS(mux))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"status":"ok"}`)
}

func handleCalculate(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Content-Type check — lenient: accept "application/json" with or
		// without charset params. Missing or wrong type → 415.
		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(strings.ToLower(ct), "application/json") {
			writeError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
			return
		}

		var req calcRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		result, err := calc.Calculate(req.Expression)
		if err != nil {
			resp := mapError(err)
			// Log the underlying technical detail; send only the user-facing
			// message over the wire.
			logger.Info("calculation failed",
				"expression", req.Expression,
				"err", err.Error(),
				"status", resp.status,
			)
			writeError(w, resp.status, resp.message)
			return
		}

		_ = json.NewEncoder(w).Encode(successResponse{Result: result})
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: msg})
}
