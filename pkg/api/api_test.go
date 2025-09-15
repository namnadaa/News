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

	url := fmt.Sprintf("/news/new/%d", created.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	var got postDTO
	err = json.Unmarshal(b, &got)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	if got.Title != post.Title {
		t.Errorf("Got title %q, want %q", got.Title, post.Title)
	}

	errReq := httptest.NewRequest(http.MethodGet, "/news/new/999999", nil)
	errRr := httptest.NewRecorder()
	api.r.ServeHTTP(errRr, errReq)
	if errRr.Code != http.StatusNotFound {
		t.Errorf("Code error: got %d, want %d", errRr.Code, http.StatusNotFound)
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
		t.Fatalf("Failed to clear posts: %v", err)
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

	req := httptest.NewRequest(http.MethodGet, "/news/2", nil)
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	var data []postDTO
	err = json.Unmarshal(b, &data)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	const wantLen = 2
	if len(data) != wantLen {
		t.Fatalf("Got %d post(s), want %d", len(data), wantLen)
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
	if rr.Code != http.StatusOK {
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
	reqInv := httptest.NewRequest(http.MethodPost, "/news", bytes.NewBufferString(invalidJSON))
	rrInv := httptest.NewRecorder()
	api.addPostHandler(rrInv, reqInv)
	if rrInv.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rrInv.Code, http.StatusBadRequest)
	}

	bodyInv := rrInv.Body.String()
	if !strings.Contains(bodyInv, "failed to decode response") {
		t.Errorf("unexpected body: %s", bodyInv)
	}

	reqDup := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
	rrDup := httptest.NewRecorder()
	api.r.ServeHTTP(rrDup, reqDup)
	if rrDup.Code != http.StatusInternalServerError {
		t.Errorf("duplicate link status=%d, want 500", rrDup.Code)
	}
}
