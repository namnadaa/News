package storage

import "time"

// Comment represents a comment for a news article.
type Comment struct {
	ID        string    `bson:"_id,omitempty"`
	NewsID    string    `bson:"news_id"`
	ParentID  string    `bson:"parent_id,omitempty"`
	Author    string    `bson:"author"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}

// Interface defines the behavior of a storage system for posts.
type Storage interface {
	CommentsByNews(newsID string) ([]Comment, error)
	AddComment(comment Comment) (Comment, error)
}
