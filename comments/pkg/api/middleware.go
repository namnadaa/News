package api

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

type ctxKey string

const RequestIDKey ctxKey = "requestID"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// generateRandomID creates a random string of 6 characters.
func generateRandomID() string {
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, 6)
	for i := range b {
		b[i] = letters[src.Intn(len(letters))]
	}
	return string(b)
}

// requestIDMiddleware extracts or generates request_id and puts it into the context.
func (api *API) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("request_id")
		if id == "" {
			id = generateRandomID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		slog.Info("request received", "method", r.Method, "path", r.URL.Path, "request_id", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getRequestID returns the request_id from context.
func getRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

// jsonMiddleware sets the Content-Type header for all JSON responses.
func (api *API) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs request details after handler execution.
func (api *API) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		requestID := getRequestID(r.Context())

		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"ip", r.RemoteAddr,
			"request_id", requestID,
		)
	})
}
