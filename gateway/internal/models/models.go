package models

import "time"

// NewsFullDetailed provides full information about the news.
type NewsFullDetailed struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	PubTime time.Time `json:"pub_time"`
	Link    string    `json:"link"`
}

// NewsShortDetailed provides a summary of the news.
type NewsShortDetailed struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

// Comment represents the structure of a user's comment on a news item.
type Comment struct {
	ID        int       `json:"id"`
	NewsID    int       `json:"news_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
