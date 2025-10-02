package api

import (
	"log/slog"
	"net/http"
)

// loggingMiddleWare logs the HTTP method and path.
func (api *API) loggingMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request received", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
