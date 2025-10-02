package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage wraps MongoDB client and collection.
type MongoStorage struct {
	client         *mongo.Client
	databaseName   string
	collectionName string
}

// New connects to MongoDB and returns a MongoStorage instance.
func New(content, dbName, colName string) (*MongoStorage, error) {
	mongoOpts := options.Client().ApplyURI(content)
	client, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot ping MongoDB: %v", err)
	}

	s := MongoStorage{
		client:         client,
		databaseName:   dbName,
		collectionName: colName,
	}
	return &s, nil
}

// Close disconnects from MongoDB.
func (ms *MongoStorage) Close() error {
	err := ms.client.Disconnect(context.Background())
	if err != nil {
		return fmt.Errorf("failed to disconnect to database: %v", err)
	}
	return nil
}
