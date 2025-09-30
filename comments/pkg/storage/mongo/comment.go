package mongo

import (
	"comments/pkg/storage"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentsByNews returns all comments for a given news ID.
func (ms *MongoStorage) CommentsByNews(newsID string) ([]storage.Comment, error) {
	filter := bson.M{"news_id": newsID}

	cur, err := ms.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find comments: %w", err)
	}
	defer cur.Close(context.Background())

	var data []storage.Comment
	for cur.Next(context.Background()) {
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
func (ms *MongoStorage) AddComment(comment storage.Comment) (storage.Comment, error) {
	res, err := ms.collection.InsertOne(context.Background(), comment)
	if err != nil {
		return storage.Comment{}, fmt.Errorf("failed to insert comment: %w", err)
	}

	if id, ok := res.InsertedID.(primitive.ObjectID); ok {
		comment.ID = id.Hex()
	}

	return comment, nil
}
