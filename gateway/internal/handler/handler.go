package handler

import (
	"encoding/json"
	"gateway/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Handler is responsible for registering routes and processing HTTP requests.
type Handler struct {
	router *mux.Router
}

// NewHandler creates and initializes a new Handler instance.
func NewHandler() *Handler {
	h := Handler{}
	h.router = mux.NewRouter()
	h.registerRoutes()
	return &h
}

// Router returns the request router.
func (h *Handler) Router() *mux.Router {
	return h.router
}

// RegisterRoutes registers all API Gateway routes.
func (h *Handler) registerRoutes() {
	h.router.HandleFunc("/news", h.newsHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/news/filter", h.filterHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/news/{id}", h.newHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/news/{n}/comment", h.commentHandler).Methods(http.MethodPost)
}

// newsHandler returns a list of news in a short format.
func (h *Handler) newsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	news := []models.NewsShortDetailed{
		{
			ID:    1,
			Title: "A new stub for the first news",
			Link:  "http://example.com/stub-1",
		},
		{
			ID:    2,
			Title: "A new stub for the second news",
			Link:  "http://example.com/stub-2",
		},
	}
	err := json.NewEncoder(w).Encode(news)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// filterHandler returns a filtered list of news items.
func (h *Handler) filterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	news := []models.NewsShortDetailed{
		{
			ID:    1,
			Title: "A new stub for the first news",
			Link:  "http://example.com/stub-1",
		},
		{
			ID:    2,
			Title: "A new stub for the second news",
			Link:  "http://example.com/stub-2",
		},
	}
	err := json.NewEncoder(w).Encode(news)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// newHandler returns full information about the news by ID.
func (h *Handler) newHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rawID := mux.Vars(r)["id"]
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	news := models.NewsFullDetailed{
		ID:      id,
		Title:   "A new stub for the first news",
		Content: "New content",
		PubTime: time.Now(),
		Link:    "http://example.com/stub-3",
	}
	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// commentHandler accepts the news comment and returns it with the added ID and creation time.
func (h *Handler) commentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var comment models.Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "failed to decode response", http.StatusBadRequest)
		return
	}

	comment.ID = 100
	comment.CreatedAt = time.Now()

	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}
