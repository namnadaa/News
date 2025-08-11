package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"news/pkg/storage"
	"news/pkg/storage/postgres"
	"strings"
	"testing"
	"time"
)

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
		t.Fatalf("Failed to clear post: %v", err)
	}

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
		t.Errorf("Got title %q, want %q", got.Title, post.Title)
	}
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
		t.Fatalf("Failed to clear post: %v", err)
	}

	post1 := storage.Post{
		ID:      1,
		Title:   "News-1",
		Content: "Content-1",
		PubTime: time.Now(),
		Link:    "http://news1/content1",
	}
	post2 := storage.Post{
		ID:      2,
		Title:   "News-2",
		Content: "Content-2",
		PubTime: time.Now(),
		Link:    "http://news2/content2",
	}

	_, err = db.AddPost(post1)
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}
	_, err = db.AddPost(post2)
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/news", nil)
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)
	if !(rr.Code == http.StatusOK) {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	var data []storage.Post
	err = json.Unmarshal(b, &data)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	const wantLen = 2
	if len(data) != wantLen {
		t.Fatalf("Got %d order(s), want %d", len(data), wantLen)
	}
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

	body, _ := json.Marshal(post)
	req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)

	if !(rr.Code == http.StatusOK) {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	var got storage.Post
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	if got.Title != post.Title {
		t.Errorf("Got title %q, want %q", got.Title, post.Title)
	}

	invalidJSON := `{"title": "Broken Post"`

	req1 := httptest.NewRequest(http.MethodPost, "/news", bytes.NewBufferString(invalidJSON))
	rr1 := httptest.NewRecorder()

	api.addPostHandler(rr1, req1)

	if rr1.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr1.Code, http.StatusInternalServerError)
	}

	body1 := rr1.Body.String()
	if !strings.Contains(body1, "failed to decode response") {
		t.Errorf("unexpected body: %s", body1)
	}
}
