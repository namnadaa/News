package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"news/pkg/storage"
	"news/pkg/storage/postgres"
	"strings"
	"testing"
	"time"
)

func TestAPI_filterHandler(t *testing.T) {
	connstr := "postgres://news_user_test:strongpasswordtest@localhost:5436/newsdb_test?sslmode=disable"

	db, err := postgres.New(connstr)
	if err != nil {
		t.Fatalf("Could not create DB storage: %v", err)
	}
	defer db.Close()
	api := New(db)

	err = db.ClearPosts()
	if err != nil {
		t.Fatalf("Failed to clear posts: %v", err)
	}

	posts := []storage.Post{
		{
			Title:   "Go concurrency",
			Content: "Channels and goroutines",
			PubTime: time.Now(),
			Link:    "http://example.com/go-concurrency",
		},
		{
			Title:   "Python basics",
			Content: "Intro to Python",
			PubTime: time.Now(),
			Link:    "http://example.com/python",
		},
	}

	for _, p := range posts {
		_, err := db.AddPost(p)
		if err != nil {
			t.Fatalf("Failed to insert post: %v", err)
		}
	}

	t.Run("valid search returns result", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/filter?s=go", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Code error: got %d, want %d", rr.Code, http.StatusOK)
		}

		var got []postDTO
		err := json.Unmarshal(rr.Body.Bytes(), &got)
		if err != nil {
			t.Fatalf("The server response could not be decoded: %v", err)
		}

		if len(got) != 1 || !strings.Contains(strings.ToLower(got[0].Title), "go") {
			t.Errorf("unexpected result: %+v", got)
		}
	})

	t.Run("no search param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/filter", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}

		if !strings.Contains(rr.Body.String(), "missing search parameter") {
			t.Errorf("unexpected body: %s", rr.Body.String())
		}
	})

	t.Run("no posts found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/filter?s=nonexistent", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("got %d, want %d", rr.Code, http.StatusNotFound)
		}

		if !strings.Contains(rr.Body.String(), "no posts found") {
			t.Errorf("unexpected body: %s", rr.Body.String())
		}
	})
}

func TestAPI_postHandler(t *testing.T) {
	connstr := "postgres://news_user_test:strongpasswordtest@localhost:5436/newsdb_test?sslmode=disable"

	db, err := postgres.New(connstr)
	if err != nil {
		t.Fatalf("Could not create DB storage: %v", err)
	}
	defer db.Close()
	api := New(db)

	err = db.ClearPosts()
	if err != nil {
		t.Fatalf("Failed to clear posts: %v", err)
	}

	post := storage.Post{
		Title:   "News-1",
		Content: "Content-1",
		PubTime: time.Now(),
		Link:    "http://example.com/news1",
	}

	created, err := db.AddPost(post)
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	t.Run("valid id post", func(t *testing.T) {
		url := fmt.Sprintf("/news/new/%d", created.ID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Code error: got %d, want %d", rr.Code, http.StatusOK)
		}

		var got postDTO
		err := json.Unmarshal(rr.Body.Bytes(), &got)
		if err != nil {
			t.Fatalf("The server response could not be decoded: %v", err)
		}

		if got.Title != post.Title {
			t.Errorf("Got title %q, want %q", got.Title, post.Title)
		}
	})

	t.Run("invalid id post", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/new/999999", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusNotFound)
		}
	})
}

func TestAPI_postsHandler(t *testing.T) {
	connstr := "postgres://news_user_test:strongpasswordtest@localhost:5436/newsdb_test?sslmode=disable"

	db, err := postgres.New(connstr)
	if err != nil {
		t.Fatalf("Error connect to database: %v", err)
	}
	defer db.Close()
	api := New(db)

	err = db.ClearPosts()
	if err != nil {
		t.Fatalf("Failed to clear posts: %v", err)
	}

	posts := []storage.Post{
		{
			Title:   "News-1",
			Content: "Content-1",
			PubTime: time.Now(),
			Link:    "http://news1/content1",
		},
		{
			Title:   "News-2",
			Content: "Content-2",
			PubTime: time.Now(),
			Link:    "http://news2/content2",
		},
	}

	for _, p := range posts {
		if _, err := db.AddPost(p); err != nil {
			t.Fatalf("Failed to insert post: %v", err)
		}
	}

	t.Run("get n posts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/2", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
		}

		var data []postDTO
		err := json.Unmarshal(rr.Body.Bytes(), &data)
		if err != nil {
			t.Fatalf("The server response could not be decoded: %v", err)
		}

		if len(data) != 2 {
			t.Fatalf("Got %d post(s), want 2", len(data))
		}
	})

	t.Run("invalid n format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/abc", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("pagination first page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news?page=1", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
		}

		var resp struct {
			News       []postDTO
			Pagination storage.Pagination
		}

		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if len(resp.News) == 0 {
			t.Errorf("expected some posts, got 0")
		}
		if resp.Pagination.CurrentPage != 1 {
			t.Errorf("expected page 1, got %d", resp.Pagination.CurrentPage)
		}
	})

	t.Run("invalid page format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news?page=abc", nil)
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})
}

func TestAPI_addPostHandler(t *testing.T) {
	connstr := "postgres://news_user_test:strongpasswordtest@localhost:5436/newsdb_test?sslmode=disable"

	db, err := postgres.New(connstr)
	if err != nil {
		t.Fatalf("Error connect to database: %v", err)
	}
	defer db.Close()
	api := New(db)

	if err := db.ClearPosts(); err != nil {
		t.Fatalf("Failed to clear posts: %v", err)
	}

	post := storage.Post{
		Title:   "News-1",
		Content: "Content-1",
		PubTime: time.Now(),
		Link:    "http://news1/content1",
	}

	t.Run("valid post", func(t *testing.T) {
		body, _ := json.Marshal(post)
		req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Code error: got %d, want %d", rr.Code, http.StatusOK)
		}

		var got storage.Post
		err := json.Unmarshal(rr.Body.Bytes(), &got)
		if err != nil {
			t.Fatalf("The server response could not be decoded: %v", err)
		}

		if got.Title != post.Title {
			t.Errorf("Got title %q, want %q", got.Title, post.Title)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidJSON := `{"title": "Broken Post"`
		req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewBufferString(invalidJSON))
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Code error: got %d, want %d", rr.Code, http.StatusOK)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "failed to decode response") {
			t.Errorf("unexpected body: %s", body)
		}
	})

	t.Run("duplicate request", func(t *testing.T) {
		body, _ := json.Marshal(post)
		req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		api.r.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("duplicate link status=%d, want 500", rr.Code)
		}
	})
}
