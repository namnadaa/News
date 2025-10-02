package api

import (
	"bytes"
	"comments/pkg/storage"
	"comments/pkg/storage/mongo"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAPI_commentsByNewsHandler(t *testing.T) {
	connStr := "mongodb://localhost:27017"
	dbName := "commentsdb_test"
	collectionName := "comments"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := mongo.New(ctx, connStr, dbName, collectionName)
	if err != nil {
		t.Fatalf("Could not create DB storage: %v", err)
	}

	defer func() {
		err := db.Close(ctx)
		if err != nil {
			t.Errorf("failed to close MongoDB: %v", err)
		}
	}()

	api := New(db)

	err = db.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to drop collection: %v", err)
	}

	comment1 := storage.Comment{
		NewsID:   "123",
		Author:   "Alex",
		Content:  "Test comment",
		ParentID: "",
	}
	comment2 := storage.Comment{
		NewsID:   "123",
		Author:   "Bob",
		Content:  "Test comment 2",
		ParentID: "",
	}

	_, err = db.AddComment(ctx, comment1)
	if err != nil {
		t.Fatalf("Failed to insert comment: %v", err)
	}
	_, err = db.AddComment(ctx, comment2)
	if err != nil {
		t.Fatalf("Failed to insert comment: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/comments/123", nil)
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	var got []storage.Comment
	err = json.Unmarshal(b, &got)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	if len(got) == 0 {
		t.Fatalf("Got 0, want %v", len(got))
	}

	if got[0].Content != comment1.Content {
		t.Errorf("Got content %q, want %q", got[0].Content, comment1.Content)
	}

	errReq := httptest.NewRequest(http.MethodGet, "/comments/999", nil)
	errRr := httptest.NewRecorder()
	api.r.ServeHTTP(errRr, errReq)
	if errRr.Code != http.StatusNotFound {
		t.Errorf("Code error: got %d, want %d", errRr.Code, http.StatusNotFound)
	}
}

func TestAPI_addCommentHandler(t *testing.T) {
	connStr := "mongodb://localhost:27017"
	dbName := "commentsdb_test"
	collectionName := "comments"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := mongo.New(ctx, connStr, dbName, collectionName)
	if err != nil {
		t.Fatalf("Could not create DB storage: %v", err)
	}

	defer func() {
		err := db.Close(ctx)
		if err != nil {
			t.Errorf("failed to close MongoDB: %v", err)
		}
	}()

	api := New(db)

	err = db.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to drop collection: %v", err)
	}

	comment := storage.Comment{
		NewsID:   "123",
		Author:   "Alex",
		Content:  "Test comment",
		ParentID: "",
	}

	body, _ := json.Marshal(comment)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	api.r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Code error: got %d, want %d", rr.Code, http.StatusOK)
	}

	var got storage.Comment
	err = json.Unmarshal(rr.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("The server response could not be decoded: %v", err)
	}

	if got.Content != comment.Content {
		t.Errorf("Got content %q, want %q", got.Content, comment.Content)
	}

	invalidJSON := `{"title": "Broken Post"`
	reqInv := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBufferString(invalidJSON))
	rrInv := httptest.NewRecorder()
	api.addCommentHandler(rrInv, reqInv)
	if rrInv.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rrInv.Code, http.StatusBadRequest)
	}

	bodyInv := rrInv.Body.String()
	if !strings.Contains(bodyInv, "failed to decode response") {
		t.Errorf("unexpected body: %s", bodyInv)
	}
}
