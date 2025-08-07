package storage

import "time"

// Post represents a single news article.
type Post struct {
	ID      int
	Title   string
	Content string
	PubTime time.Time
	Link    string
}

// Interface defines the behavior of a storage system for posts.
type Storage interface {
	Post(postID int) (Post, error)
	Posts(limit, offset int) ([]Post, error)
	AddPost(p Post) (Post, error)
}
