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
	Post(newsID int) (Post, error)
	Posts(limit int) ([]Post, error)
	AddPost(p Post) (Post, error)
	SearchPosts(search string) ([]Post, error)
}
