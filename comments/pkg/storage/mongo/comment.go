package mongo

import (
	"comments/pkg/storage"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentsByNews returns all comments for a given news ID.
func (ms *MongoStorage) CommentsByNews(ctx context.Context, newsID string) ([]storage.Comment, error) {
	collection := ms.client.Database(ms.databaseName).Collection(ms.collectionName)

	filter := bson.M{"news_id": newsID, "allowed": true}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find comments: %w", err)
	}
	defer cur.Close(ctx)

	var data []storage.Comment
	for cur.Next(ctx) {
		var c storage.Comment
		err := cur.Decode(&c)
		if err != nil {
			return nil, fmt.Errorf("failed to decode comment: %w", err)
		}

		data = append(data, c)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return data, nil
}

// AddComment saves a new comment and returns it with generated ID.
func (ms *MongoStorage) AddComment(ctx context.Context, comment storage.Comment) (storage.Comment, error) {
	collection := ms.client.Database(ms.databaseName).Collection(ms.collectionName)

	comment.CreatedAt = time.Now()
	if !comment.Allowed {
		comment.Allowed = false
	}

	res, err := collection.InsertOne(ctx, comment)
	if err != nil {
		return storage.Comment{}, fmt.Errorf("failed to insert comment: %w", err)
	}

	if id, ok := res.InsertedID.(primitive.ObjectID); ok {
		comment.ID = id.Hex()
	}

	return comment, nil
}

// Clear clears the collection.
func (ms *MongoStorage) Clear(ctx context.Context) error {
	collection := ms.client.Database(ms.databaseName).Collection(ms.collectionName)

	err := collection.Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %v", err)
	}
	return nil
}
