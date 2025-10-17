package handler

import (
	"encoding/json"
	"fmt"
	"gateway/internal/models"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Handler is responsible for registering routes and processing HTTP requests.
type Handler struct {
	router               *mux.Router
	newsServiceURL       string
	commentsServiceURL   string
	censorshipServiceURL string
}

// NewHandler creates and initializes a new Handler instance.
func New(newsURL, commentsURL, censorshipURL string) *Handler {
	h := Handler{}
	h.router = mux.NewRouter()
	h.newsServiceURL = newsURL
	h.commentsServiceURL = commentsURL
	h.censorshipServiceURL = censorshipURL
	h.registerRoutes()
	return &h
}

// Router returns the request router.
func (h *Handler) Router() *mux.Router {
	return h.router
}

// RegisterRoutes registers all API Gateway routes.
func (h *Handler) registerRoutes() {
	h.router.Use(h.jsonMiddleware)
	h.router.Use(h.requestIDMiddleware)
	h.router.Use(h.loggingMiddleware)
	h.router.HandleFunc("/news", h.newsListHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/news/filter", h.newsFilterHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/news/{id}", h.newsDetailedHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/news/{id}/comment", h.addCommentHandler).Methods(http.MethodPost)
}

// newsListHandler proxies the request for the list of news.
func (h *Handler) newsListHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	page := r.URL.Query().Get("page")
	var url string

	if page != "" {
		url = fmt.Sprintf("%s/news?page=%s&request_id=%s", h.newsServiceURL, page, requestID)
	} else {
		limit := "40"
		url = fmt.Sprintf("%s/news/%s?request_id=%s", h.newsServiceURL, limit, requestID)
	}

	resp, err := http.Get(url)
	if err != nil {
		slog.Error("newsHandler: failed to get news list", "err", err, "request_id", requestID)
		http.Error(w, "failed to fetch news", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// newsFilterHandler proxies the request for the filter of news.
func (h *Handler) newsFilterHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	query := r.URL.Query().Get("s")
	if query == "" {
		http.Error(w, "missing search query", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("%s/news/filter?s=%s&request_id=%s", h.newsServiceURL, query, requestID)

	resp, err := http.Get(url)
	if err != nil {
		slog.Error("filterHandler: failed to get filtered news", "err", err, "request_id", requestID)
		http.Error(w, "failed to fetch filtered news", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// newsDetailedHandler concurrently fetches news and comments, merges their results, and returns a combined JSON response.
func (h *Handler) newsDetailedHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	var (
		newsData     models.NewsShortDetailed
		commentsData []models.Comment
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		newsURL := fmt.Sprintf("%s/news/new/%s?request_id=%s", h.newsServiceURL, id, requestID)
		newsResp, err := http.Get(newsURL)
		if err != nil {
			slog.Error("newsDetailedHandler: failed to fetch news", "err", err, "request_id", requestID)
			errCh <- fmt.Errorf("failed to fetch news: %w", err)
			return
		}
		defer newsResp.Body.Close()

		if newsResp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("news service returned %d", newsResp.StatusCode)
			return
		}

		err = json.NewDecoder(newsResp.Body).Decode(&newsData)
		if err != nil {
			slog.Error("newsDetailedHandler: failed to decode news", "err", err, "request_id", requestID)
			errCh <- fmt.Errorf("failed to decode news response: %w", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		commentsURL := fmt.Sprintf("%s/comments/%s?request_id=%s", h.commentsServiceURL, id, requestID)
		commentsResp, err := http.Get(commentsURL)
		if err != nil {
			slog.Error("newsDetailedHandler: failed to fetch comments", "err", err, "request_id", requestID)
			errCh <- fmt.Errorf("failed to fetch comments: %w", err)
			return
		}
		defer commentsResp.Body.Close()

		if commentsResp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("comments service returned %d", commentsResp.StatusCode)
			return
		}

		err = json.NewDecoder(commentsResp.Body).Decode(&commentsData)
		if err != nil {
			slog.Error("newsDetailedHandler: failed to decode comments", "err", err, "request_id", requestID)
			errCh <- fmt.Errorf("failed to decode news response: %w", err)
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			slog.Error("newsDetailedHandler: concurrent request failed", "err", err, "request_id", requestID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return

		}
	}

	detailed := models.NewsFullDetailed{
		News:     newsData,
		Comments: commentsData,
	}

	err := json.NewEncoder(w).Encode(detailed)
	if err != nil {
		slog.Error("newsDetailedHandler: failed to encode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// addCommentHandler proxies the request for creating a new comment with censorship validation.
func (h *Handler) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	id := mux.Vars(r)["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	censorURL := fmt.Sprintf("%s/check?request_id=%s", h.censorshipServiceURL, requestID)
	censorReq, err := http.NewRequest(http.MethodPost, censorURL, strings.NewReader(string(body)))
	if err != nil {
		slog.Error("addCommentHandler: failed to create censorship request", "err", err, "request_id", requestID)
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	censorReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	censorResp, err := client.Do(censorReq)
	if err != nil {
		slog.Error("addCommentHandler: failed to send censorship request", "err", err, "request_id", requestID)
		http.Error(w, "failed to send request to censorship service", http.StatusBadGateway)
		return
	}
	defer censorResp.Body.Close()

	if censorResp.StatusCode != http.StatusOK {
		slog.Warn("addCommentHandler: comment rejected by censorship", "status", censorResp.StatusCode, "request_id", requestID)
		http.Error(w, "comment rejected by censorship", http.StatusBadRequest)
		return
	}

	updated := fmt.Sprintf(`{"news_id":"%s",%s`, id, body[1:])
	url := fmt.Sprintf("%s/comments?request_id=%s", h.commentsServiceURL, requestID)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(updated))
	if err != nil {
		slog.Error("addCommentHandler: failed to create request", "err", err, "request_id", requestID)
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("addCommentHandler: failed to send request", "err", err, "request_id", requestID)
		http.Error(w, "failed to send request to comments service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
