package api

import (
	"log/slog"
	"net/http"
)

// headersMiddleWare sets the JSON content type.
// func (api *API) headersMiddleWare(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")
// 		next.ServeHTTP(w, r)
// 	})
// }

// loggingMiddleWare logs the HTTP method and path.
func (api *API) loggingMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request received", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
