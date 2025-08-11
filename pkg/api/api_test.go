package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"news/pkg/storage"
	"news/pkg/storage/postgres"
	"testing"
	"time"
)

func TestAPI_postHandler(t *testing.T) {
	db, err := postgres.New("postgres://news_user:strongpassword@localhost:5435/newsdb?sslmode=disable")
	if err != nil {
		t.Fatalf("Error connect to database: %v", err)
	}
	defer db.Close()
	api := New(db)

	post := storage.Post{
		ID:      1,
		Title:   "News-1",
		Content: "Content-1",
		PubTime: time.Now(),
		Link:    "http://news1/content1",
	}
	created, err := db.AddPost(post)
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	url := fmt.Sprintf("/news/%d", created.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)
	if !(rr.Code == http.StatusOK) {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	var got storage.Post
	err = json.Unmarshal(b, &got)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	if got.Title != post.Title {
		t.Errorf("got title %q, want %q", got.Title, post.Title)
	}
}
