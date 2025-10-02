package mongo

import (
	"comments/pkg/storage"
	"context"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var bannedWords = []string{"qwerty", "йцукен", "zxcvbnm"}

// containsBannedWords returns true if the text contains forbidden words.
func containsBannedWords(text string) bool {
	text = strings.ToLower(text)
	for _, w := range bannedWords {
		if strings.Contains(text, w) {
			return true
		}
	}

	return false
}

// Moderation periodically checks new comments and marks them as allowed or blocked.
func Moderation(ctx context.Context, ticker *time.Ticker, db *MongoStorage) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			collection := db.client.Database(db.databaseName).Collection(db.collectionName)
			cur, err := collection.Find(ctx, bson.M{"allowed": false})
			if err != nil {
				slog.Error("failed to find unmoderated comments", "err", err)
				continue
			}

			for cur.Next(ctx) {
				var c storage.Comment
				if err := cur.Decode(&c); err != nil {
					slog.Error("failed to decode comment", "err", err)
					continue
				}

				allowed := !containsBannedWords(c.Content)
				id, err := primitive.ObjectIDFromHex(c.ID)
				if err != nil {
					slog.Error("invalid ObjectID", "id", c.ID, "err", err)
					continue
				}

				_, err = collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"allowed": allowed}})
				if err != nil {
					slog.Error("failed to update comment moderation status", "err", err)
					continue
				}

				if !allowed {
					slog.Warn("comment blocked by moderation", "id", c.ID, "content", c.Content)
				}
			}

			err = cur.Close(ctx)
			if err != nil {
				slog.Error("failed to close cursor", "err", err)
			}
		}
	}
}
