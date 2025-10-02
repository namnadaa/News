package mongo

import (
	"comments/pkg/storage"
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestModeration(t *testing.T) {
	connStr := "mongodb://localhost:27017"
	dbName := "commentsdb_test"
	collectionName := "comments"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := New(ctx, connStr, dbName, collectionName)
	if err != nil {
		t.Fatalf("failed to init mongo storage: %v", err)
	}

	defer func() {
		err := db.Close(ctx)
		if err != nil {
			t.Errorf("failed to close MongoDB: %v", err)
		}
	}()

	err = db.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to drop collection: %v", err)
	}

	clean := storage.Comment{
		NewsID:  "1",
		Author:  "Alice",
		Content: "Nice post!",
	}
	banned := storage.Comment{
		NewsID:  "1",
		Author:  "Bob",
		Content: "This is qwerty spam",
	}

	_, err = db.AddComment(ctx, clean)
	if err != nil {
		t.Fatalf("failed to insert clean comment: %v", err)
	}
	_, err = db.AddComment(ctx, banned)
	if err != nil {
		t.Fatalf("failed to insert clean comment: %v", err)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	go Moderation(ctx, ticker, db)
	time.Sleep(300 * time.Millisecond)

	collection := db.client.Database(db.databaseName).Collection(db.collectionName)

	var gotClean storage.Comment
	err = collection.FindOne(ctx, bson.M{"author": "Alice"}).Decode(&gotClean)
	if err != nil {
		t.Fatalf("failed to find clean comment: %v", err)
	}
	if !gotClean.Allowed {
		t.Errorf("expected clean comment to be allowed, got Allowed=%v", gotClean.Allowed)
	}

	var gotBanned storage.Comment
	err = collection.FindOne(ctx, bson.M{"author": "Bob"}).Decode(&gotBanned)
	if err != nil {
		t.Fatalf("failed to find banned comment: %v", err)
	}
	if gotBanned.Allowed {
		t.Errorf("expected banned comment to be blocked, got Allowed=%v", gotBanned.Allowed)
	}
}
