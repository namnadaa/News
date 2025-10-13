package handler

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
func (h *Handler) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("request_id")
		if id == "" {
			id = generateRandomID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		slog.Info("request received", "path", r.URL.Path, "request_id", id)

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
func (h *Handler) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
