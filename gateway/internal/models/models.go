package models

import "time"

// NewsFullDetailed provides full information about the news.
type NewsFullDetailed struct {
	News     NewsShortDetailed `json:"news"`
	Comments []Comment         `json:"comments,omitempty"`
}

// NewsShortDetailed provides a summary of the news.
type NewsShortDetailed struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	PubTime time.Time `json:"pub_time"`
	Link    string    `json:"link"`
}

// Comment represents the structure of a user's comment on a news item.
type Comment struct {
	ID        string    `json:"id"`
	NewsID    string    `json:"news_id"`
	ParentID  string    `json:"parent_id,omitempty"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Allowed   bool      `json:"allowed,omitempty"`
}
