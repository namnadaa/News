package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// API handles HTTP requests and routes.
type API struct {
	r *mux.Router
}

// New creates and initializes a new API instance.
func New() *API {
	api := API{}
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

// Router returns the request router.
func (api *API) Router() *mux.Router {
	return api.r
}

// Registration of API methods in the request router.
func (api *API) endpoints() {
	api.r.Use(api.jsonMiddleware)
	api.r.Use(api.requestIDMiddleware)
	api.r.Use(api.loggingMiddleware)
	api.r.HandleFunc("/check", api.checkHandler).Methods(http.MethodPost)
}

func (api *API) checkHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	var req struct {
		Content string
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("checkHandler: failed to decode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to decode response", http.StatusBadRequest)
		return
	}

	if containsBannedWords(req.Content) {
		slog.Warn("checkHandler: comment blocked by censorship", "request_id", requestID)
		http.Error(w, "comment contains banned words", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		slog.Error("checkHandler: failed to encode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// containsBannedWords returns true if the text contains forbidden words.
func containsBannedWords(text string) bool {
	bannedWords := []string{"qwerty", "йцукен", "zxcvbnm"}
	text = strings.ToLower(text)
	for _, w := range bannedWords {
		if strings.Contains(text, w) {
			return true
		}
	}

	return false
}
