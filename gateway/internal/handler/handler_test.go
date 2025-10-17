package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_newsListHandler(t *testing.T) {
	tests := []struct {
		name       string
		newsStatus int
		newsBody   string
		wantStatus int
		wantInBody string
		route      string
	}{
		{
			name:       "with page",
			newsStatus: http.StatusOK,
			newsBody:   `{"id":1,"title":"Test News","content":"content"}`,
			wantStatus: http.StatusOK,
			wantInBody: "Test News",
			route:      "/news?page=1&request_id=abc123",
		},
		{
			name:       "without page",
			newsStatus: http.StatusOK,
			newsBody:   `{"id":1,"title":"Test News","content":"content"}`,
			wantStatus: http.StatusOK,
			wantInBody: "Test News",
			route:      "/news?request_id=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.newsStatus != 0 {
					w.WriteHeader(tt.newsStatus)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				if tt.newsBody != "" {
					io.WriteString(w, tt.newsBody)
				}
			}))
			defer newsSrv.Close()

			h := New(newsSrv.URL, "", "")
			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("[%s] got %d, want %d", tt.name, rr.Code, tt.wantStatus)
			}

			if tt.wantInBody != "" && !strings.Contains(rr.Body.String(), tt.wantInBody) {
				t.Errorf("[%s] body = %s, want substring %q", tt.name, rr.Body.String(), tt.wantInBody)
			}
		})
	}
}

func TestHandler_newsFilterHandler(t *testing.T) {
	tests := []struct {
		name       string
		newsStatus int
		newsBody   string
		wantStatus int
		wantInBody string
		route      string
	}{
		{
			name:       "success",
			newsStatus: http.StatusOK,
			newsBody:   `{"id":1,"title":"Test News","content":"content"}`,
			wantStatus: http.StatusOK,
			wantInBody: "Test News",
			route:      "/news/filter?s=test&request_id=abc123",
		},
		{
			name:       "without s",
			newsStatus: http.StatusOK,
			newsBody:   `{"id":1,"title":"Test News","content":"content"}`,
			wantStatus: http.StatusBadRequest,
			route:      "/news/filter?test&request_id=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.newsStatus != 0 {
					w.WriteHeader(tt.newsStatus)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				if tt.newsBody != "" {
					io.WriteString(w, tt.newsBody)
				}
			}))
			defer newsSrv.Close()

			h := New(newsSrv.URL, "", "")
			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("[%s] got %d, want %d", tt.name, rr.Code, tt.wantStatus)
			}

			if tt.wantInBody != "" && !strings.Contains(rr.Body.String(), tt.wantInBody) {
				t.Errorf("[%s] body = %s, want substring %q", tt.name, rr.Body.String(), tt.wantInBody)
			}
		})
	}
}

func TestHandler_newsDetailedHandler(t *testing.T) {
	tests := []struct {
		name           string
		newsStatus     int
		newsBody       string
		commentsStatus int
		commentsBody   string
		wantStatus     int
		wantInBody     string
		route          string
	}{
		{
			name:           "success",
			newsStatus:     http.StatusOK,
			newsBody:       `{"id":1,"title":"Test News","content":"content"}`,
			commentsStatus: http.StatusOK,
			commentsBody:   `[{"id":"1","author":"Alex","content":"Nice"}]`,
			wantStatus:     http.StatusOK,
			wantInBody:     "Test News",
			route:          "/news/1?request_id=abc123",
		},
		{
			name:       "news service down",
			newsStatus: http.StatusInternalServerError,
			wantStatus: http.StatusInternalServerError,
			route:      "/news/1?request_id=abc123",
		},
		{
			name:           "news invalid JSON",
			newsStatus:     http.StatusOK,
			newsBody:       `[{invalid_json}]`,
			commentsStatus: http.StatusOK,
			commentsBody:   `[{"id":"1","author":"Alex","content":"Nice"}]`,
			wantStatus:     http.StatusInternalServerError,
			route:          "/news/1?request_id=abc123",
		},
		{
			name:           "comments service returns error",
			newsStatus:     http.StatusOK,
			newsBody:       `{"id":1,"title":"Test News","content":"content"}`,
			commentsStatus: http.StatusInternalServerError,
			wantStatus:     http.StatusInternalServerError,
			route:          "/news/1?request_id=abc123",
		},
		{
			name:           "comments invalid JSON",
			newsStatus:     http.StatusOK,
			newsBody:       `{"id":1,"title":"Test News","content":"content"}`,
			commentsStatus: http.StatusOK,
			commentsBody:   `invalid_json`,
			wantStatus:     http.StatusInternalServerError,
			route:          "/news/1?request_id=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.newsStatus != 0 {
					w.WriteHeader(tt.newsStatus)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				if tt.newsBody != "" {
					io.WriteString(w, tt.newsBody)
				}
			}))
			defer newsSrv.Close()

			commSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.commentsStatus != 0 {
					w.WriteHeader(tt.commentsStatus)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				if tt.commentsBody != "" {
					io.WriteString(w, tt.commentsBody)
				}
			}))
			defer commSrv.Close()

			h := New(newsSrv.URL, commSrv.URL, "")
			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("[%s] got %d, want %d", tt.name, rr.Code, tt.wantStatus)
			}

			if tt.wantInBody != "" && !strings.Contains(rr.Body.String(), tt.wantInBody) {
				t.Errorf("[%s] body = %s, want substring %q", tt.name, rr.Body.String(), tt.wantInBody)
			}
		})
	}
}

func TestHandler_addCommentHandler(t *testing.T) {
	tests := []struct {
		name             string
		censorshipStatus int
		censorshipBody   string
		commentsStatus   int
		commentsBody     string
		inputBody        string
		wantStatus       int
		wantInBody       string
		route            string
	}{
		{
			name:             "success",
			censorshipStatus: http.StatusOK,
			censorshipBody:   `{"status":"ok"}`,
			commentsStatus:   http.StatusOK,
			commentsBody:     `{"id":"1","news_id":"1","author":"Alex","content":"Nice post"}`,
			inputBody:        `{"author":"Alex","content":"Nice post"}`,
			wantStatus:       http.StatusOK,
			wantInBody:       "Nice post",
			route:            "/news/1/comment?request_id=abc123",
		},
		{
			name:             "rejected by censorship",
			censorshipStatus: http.StatusBadRequest,
			censorshipBody:   `comment contains banned words`,
			inputBody:        `{"author":"Alex","content":"qwerty"}`,
			wantStatus:       http.StatusBadRequest,
			wantInBody:       "comment rejected by censorship",
			route:            "/news/1/comment?request_id=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			censorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.censorshipStatus != 0 {
					w.WriteHeader(tt.censorshipStatus)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				if tt.censorshipBody != "" {
					io.WriteString(w, tt.censorshipBody)
				}
			}))
			defer censorSrv.Close()

			commSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.commentsStatus != 0 {
					w.WriteHeader(tt.commentsStatus)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				if tt.commentsBody != "" {
					io.WriteString(w, tt.commentsBody)
				}
			}))
			defer commSrv.Close()

			h := New("", commSrv.URL, censorSrv.URL)
			req := httptest.NewRequest(http.MethodPost, tt.route, strings.NewReader(tt.inputBody))
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("[%s] got %d, want %d", tt.name, rr.Code, tt.wantStatus)
			}

			if tt.wantInBody != "" && !strings.Contains(rr.Body.String(), tt.wantInBody) {
				t.Errorf("[%s] body = %s, want substring %q", tt.name, rr.Body.String(), tt.wantInBody)
			}
		})
	}
}
