package mongo

import (
	"comments/pkg/storage"
	"context"
	"testing"
)

func TestMongoStorage_CommentsByNews(t *testing.T) {
	connStr := "mongodb://localhost:27017"
	dbName := "commentsdb_test"
	collectionName := "comments"
	db, err := New(connStr, dbName, collectionName)
	if err != nil {
		t.Fatalf("failed to init mongo storage: %v", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("failed to close MongoDB: %v", err)
		}
	}()

	err = db.collection.Drop(context.Background())
	if err != nil {
		t.Fatalf("failed to drop collection: %v", err)
	}

	input := storage.Comment{
		NewsID:   "123",
		Author:   "Alex",
		Content:  "Test comment",
		ParentID: "",
	}

	_, err = db.AddComment(input)
	if err != nil {
		t.Fatalf("AddComment() unexpected error = %v", err)
	}

	comments, err := db.CommentsByNews("123")
	if err != nil {
		t.Fatalf("CommentByNews() unexpected error: %v", err)
	}

	if len(comments) != 1 {
		t.Errorf("got: %v, want: 1", len(comments))
	}

	if comments[0].NewsID != input.NewsID {
		t.Errorf("expected NewsID %s, got %s", input.NewsID, comments[0].NewsID)
	}

	if comments[0].Author != input.Author || comments[0].Content != input.Content {
		t.Errorf("got %+v, want %+v", comments[0], input)
	}
}

func TestMongoStorage_AddComment(t *testing.T) {
	connStr := "mongodb://localhost:27017"
	dbName := "commentsdb_test"
	collectionName := "comments"
	db, err := New(connStr, dbName, collectionName)
	if err != nil {
		t.Fatalf("failed to init mongo storage: %v", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("failed to close MongoDB: %v", err)
		}
	}()

	err = db.collection.Drop(context.Background())
	if err != nil {
		t.Fatalf("failed to drop collection: %v", err)
	}

	input := storage.Comment{
		NewsID:   "123",
		Author:   "Alex",
		Content:  "Test comment",
		ParentID: "",
	}

	got, err := db.AddComment(input)
	if err != nil {
		t.Fatalf("AddComment() unexpected error = %v", err)
	}

	if got.ID == "" {
		t.Errorf("AddComment() got empty ID")
	}

	if got.Author != input.Author || got.Content != input.Content {
		t.Errorf("AddComment() mismatch: got: %v, want: %v", got, input)
	}
}
