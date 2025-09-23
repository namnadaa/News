DROP TABLE IF EXISTS posts;

CREATE TABLE posts (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL CHECK (char_length(title) <= 255),
	content TEXT NOT NULL,
	pub_time TIMESTAMP NOT NULL,
	link TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_posts_pub_time ON posts(pub_time DESC);