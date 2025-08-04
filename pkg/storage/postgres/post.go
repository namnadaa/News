package postgres

import (
	"context"
	"fmt"
	"news/pkg/storage"
)

// Post retrieves a post by its ID from the database.
func (ps *PostgresStorage) Post(postID int) (storage.Post, error) {
	var p storage.Post
	err := ps.db.QueryRow(context.Background(), `
	SELECT
		id,
		title,
		content,
		pub_time,
		link
	FROM 
		posts
	WHERE
		id = $1;
	`,
		postID,
	).Scan(
		&p.ID,
		&p.Title,
		&p.Content,
		&p.PubTime,
		&p.Link,
	)
	if err != nil {
		return p, fmt.Errorf("failed to execute query for Post: %v", err)
	}

	return p, nil
}

// Posts returns a list of posts ordered by pub_time (newest first).
// Supports optional limit and offset for pagination.
func (ps *PostgresStorage) Posts(limit, offset int) ([]storage.Post, error) {
	rows, err := ps.db.Query(context.Background(), `
	SELECT 
		id, 
		title, 
		content,
		pub_time,
		link
	FROM 
		posts
	ORDER BY 
		pub_time DESC
	LIMIT $1 OFFSET $2;
	`,
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for Posts: %v", err)
	}
	defer rows.Close()

	var posts []storage.Post
	for rows.Next() {
		var p storage.Post
		err = rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.PubTime,
			&p.Link,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %v", err)
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return posts, nil
}

// CreatePost adds a new post to the database.
func (ps *PostgresStorage) CreatePost(p storage.Post) (storage.Post, error) {
	var post storage.Post
	err := ps.db.QueryRow(context.Background(), `
	INSERT INTO posts (title, content, pub_time, link)
	VALUES ($1, $2, $3, $4)
	RETURNING 
		id, title, content, pub_time, link;
	`,
		p.Title, p.Content, p.PubTime, p.Link,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.PubTime,
		&post.Link,
	)
	if err != nil {
		return post, fmt.Errorf("failed to create post: %v", err)
	}

	return post, nil
}
