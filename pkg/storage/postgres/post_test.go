package postgres

import (
	"context"
	"news/pkg/storage"
	"testing"
	"time"
)

func TestPostgresStorage_Post(t *testing.T) {
	connstr := "postgres://news_user_test:strongpasswordtest@localhost:5436/newsdb_test?sslmode=disable"

	ps, err := New(connstr)
	if err != nil {
		t.Fatalf("could not create DB storage: %v", err)
	}
	defer ps.db.Close()

	err = ps.ClearPosts()
	if err != nil {
		t.Fatalf("failed to clear post: %v", err)
	}

	rows, err := ps.db.Query(context.Background(), `
		INSERT INTO posts (title, content, pub_time, link)
		VALUES 
			('Новость 1', 'Содержание 1', $1, 'https://example.com/news/1'),
			('Новость 2', 'Содержание 2', $2, 'https://example.com/news/2'),
			('Новость 3', 'Содержание 3', $3, 'https://example.com/news/3')
		RETURNING id;
	`, time.Now(), time.Now().Add(-time.Hour), time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("failed to scan returned id: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iteration error: %v", err)
	}

	if len(ids) < 3 {
		t.Fatalf("got - %d, wand - 3 inserted posts", len(ids))
	}

	for _, id := range ids {
		post, err := ps.Post(id)
		if err != nil {
			t.Errorf("failed to get post with ID %d: %v", id, err)
		} else {
			t.Logf("Post ID %d: %+v", id, post)
		}
	}

	_, err = ps.Post(99999)
	if err == nil {
		t.Error("expected error for non-existent post ID, got nil")
	}
}

func TestPostgresStorage_Posts(t *testing.T) {
	connstr := "postgres://news_user_test:strongpasswordtest@localhost:5436/newsdb_test?sslmode=disable"

	ps, err := New(connstr)
	if err != nil {
		t.Fatalf("could not create DB storage: %v", err)
	}
	defer ps.db.Close()

	err = ps.ClearPosts()
	if err != nil {
		t.Fatalf("failed to clear post: %v", err)
	}

	rows, err := ps.db.Query(context.Background(), `
		INSERT INTO posts (title, content, pub_time, link)
		VALUES 
			('Новость 1', 'Содержание 1', $1, 'https://example.com/news/1'),
			('Новость 2', 'Содержание 2', $2, 'https://example.com/news/2'),
			('Новость 3', 'Содержание 3', $3, 'https://example.com/news/3')
		RETURNING id;
	`, time.Now(), time.Now().Add(-time.Hour), time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}
	defer rows.Close()

	posts, err := ps.Posts(10, 0)
	if err != nil {
		t.Fatalf("failed to get posts: %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("got - %d, expected - 3 posts", len(posts))
	}

	for i := 1; i < len(posts); i++ {
		if posts[i].PubTime.After(posts[i-1].PubTime) {
			t.Errorf("posts not sorted by pub_time DESC: post %d (%v) after post %d (%v)",
				i, posts[i].PubTime, i-1, posts[i-1].PubTime)
		}
	}

	pagedPosts, err := ps.Posts(1, 1)
	if err != nil {
		t.Fatalf("failed to get paginated posts: %v", err)
	}

	if len(pagedPosts) != 1 {
		t.Errorf("got - %d, expected - 1 post from pagination", len(pagedPosts))
	}
	t.Logf("Paginated post: %+v", pagedPosts[0])
}

func TestPostgresStorage_AddPost(t *testing.T) {
	connstr := "postgres://news_user:strongpassword@localhost:5435/newsdb?sslmode=disable"

	ps, err := New(connstr)
	if err != nil {
		t.Fatalf("could not create DB storage: %v", err)
	}
	defer ps.db.Close()

	err = ps.ClearPosts()
	if err != nil {
		t.Fatalf("failed to clear post: %v", err)
	}

	post := storage.Post{
		Title:   "Новость 1",
		Content: "Содержание 1",
		PubTime: time.Now(),
		Link:    "https://example.com/news/1",
	}

	data, err := ps.AddPost(post)
	if err != nil {
		t.Fatalf("could not create post: %v", err)
	}

	if data.ID == 0 {
		t.Errorf("expected non-zero ID")
	}

	if data.Title != post.Title || data.Link != post.Link {
		t.Errorf("created post doesn't match input post\nExpected: %+v\nGot: %+v", post, data)
	}
	t.Logf("Created post: %+v", data)
}
