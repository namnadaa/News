# News (RSS Aggregator)

A lightweight news aggregator in Go. It periodically polls RSS feeds, stores articles in PostgreSQL, and serves them via an HTTP API. The frontend (Vue + Vuetify) is built into static files and served by the same server.

---

## Features

- Scheduled polling of multiple RSS feeds
-	HTML stripping and parsing of various RSS date formats
-	Persistence in PostgreSQL (uniqueness by link)
-	HTTP API for listing and detailed view
-	Built-in static serving of the frontend (Vue/Vuetify bundle in webapp/)
-	Dockerfile + docker-compose (with Postgres and a test DB)
-	Concise JSON logs (slog)
-	Integration tests for the API

---

## How It Works
- On a timer, the aggregator iterates over all URLs from config.json → parses title/description/pubDate/link → writes to the DB.
- There’s a unique index on the link field to avoid duplicates (on repeats you’ll see error 23505 in logs — that’s OK).
- The API returns JSON; the frontend does fetch('/news/40') and renders cards.

---

## API Overview

### GET /news/{n}

Returns the latest n publications:

```json
[
  {
    "ID": 123,
    "Title": "Title",
    "Content": "Short text",
    "PubTime": 1757930940,
    "Link": "https://example.com/post"
  },
  {
    "ID": 124,
    "Title": "Title 2",
    "Content": "Short text 2",
    "PubTime": 1757964816,
    "Link": "https://example.com/post2"
  }
]
```

### GET /news/new/{id}

Returns a single publication by ID:

```json
[
  {
    "ID": 123,
    "Title": "Title",
    "Content": "Short text",
    "PubTime": 1757930940,
    "Link": "https://example.com/post"
  }
]
```

### (Optional) POST /news

Manual insert for debugging:

```json
{
  "Title": "Title",
  "Content": "Text",
  "PubTime": "2025-09-15T12:00:00Z",
  "Link": "https://example.com/post"
}
```

---

## Data Storage

PostgreSQL, table posts:

```SQL
CREATE TABLE posts (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL CHECK (char_length(title) <= 255),
  content TEXT NOT NULL,
  pub_time TIMESTAMP NOT NULL,
  link TEXT NOT NULL UNIQUE
);
```

- ID is assigned by the DB (BIGSERIAL).
- UNIQUE (link) protects against duplicates on each RSS poll.

---

## Running

### build and start (Docker Compose)

```bash
docker compose up -d --build
```

### running (locally)

```bash
go run ./cmd/server/main.go
```

---

## Configuration

config.json:

```json
{
  "rss": [
    "https://tproger.ru/feed/",
    "https://vc.ru/rss",
    "https://habr.com/ru/rss/all/all/?fl=ru"
  ],
  "request_period": 5
}
```

- rss — list of feeds (duplicates are removed)
- request_period — polling period in minutes (<= 0 — one pass and exit)

---

## Tests

API integration tests use a separate test DB (port 5436 in compose).

---

## Project Structure

```
News/
├── cmd/
│   └── server/           # Entry point (main.go)
├── initdb/               # SQL schema and initial setup
├── internal/
│   └── config/           # Read/normalize config.json
├── pkg/
│   ├── aggregator/       # Timed RSS polling
│   ├── api/              # HTTP API + static files (gorilla/mux)
│   ├── rss/              # RSS parser (XML → models)
│   └── storage/
│       ├── postgres/     # Pgx pool, SQL, Add/Post/Posts methods
│       └── storage.go    # Storage interface and Post model
├── webapp/               # Built Vue/Vuetify static files
├── .gitignore
├── Dockerfile
├── README.md
├── config.json           # RSS list and polling period
├── docker-compose.yaml
├── go.mod                # Module definition
└── go.sum
```
