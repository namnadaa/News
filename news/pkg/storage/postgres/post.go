package postgres

import (
	"context"
	"fmt"
	"math"
	"news/pkg/storage"
)

// Post retrieves a post by its ID from the database.
func (ps *PostgresStorage) Post(newsID int) (storage.Post, error) {
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
		newsID,
	).Scan(
		&p.ID,
		&p.Title,
		&p.Content,
		&p.PubTime,
		&p.Link,
	)
	if err != nil {
		return p, fmt.Errorf("failed to execute query for Post: %w", err)
	}

	return p, nil
}

// Posts returns a list of posts ordered by pub_time.
func (ps *PostgresStorage) Posts(limit int) ([]storage.Post, error) {
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
	LIMIT $1;
	`,
		limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for Posts: %w", err)
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
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return posts, nil
}

// GetPostsPaginated returns a paginated list of posts and pagination info.
func (ps *PostgresStorage) GetPostsPaginated(page, perPage int) ([]storage.Post, storage.Pagination, error) {
	var totalCount int
	err := ps.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM posts`).Scan(&totalCount)
	if err != nil {
		return nil, storage.Pagination{}, fmt.Errorf("failed to count posts: %w", err)
	}

	offset := (page - 1) * perPage
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
		perPage, offset)
	if err != nil {
		return nil, storage.Pagination{}, fmt.Errorf("failed to fetch posts: %w", err)
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
			return nil, storage.Pagination{}, fmt.Errorf("failed to scan post row: %w", err)
		}

		posts = append(posts, p)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

	pagination := storage.Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		PerPage:     perPage,
	}

	return posts, pagination, nil
}

// AddPost adds a new post to the database.
func (ps *PostgresStorage) AddPost(p storage.Post) (storage.Post, error) {
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
		return post, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

// SearchPosts returns posts whose titles contain the given substring.
func (ps *PostgresStorage) SearchPosts(search string) ([]storage.Post, error) {
	rows, err := ps.db.Query(context.Background(), `
	SELECT 
		id, 
		title, 
		content,
		pub_time,
		link
	FROM 
		posts
	WHERE 
		title
	ILIKE '%' || $1 || '%';
	`,
		search)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for Posts: %w", err)
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
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return posts, nil
}

// ClearPosts clears the table.
func (ps *PostgresStorage) ClearPosts() error {
	_, err := ps.db.Exec(context.Background(), `DELETE FROM posts;`)
	if err != nil {
		return fmt.Errorf("failed to clean up posts table: %w", err)
	}
	return nil
}
