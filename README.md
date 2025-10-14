# News Microservice

A modular microservice-based news aggregator built with **Go**, featuring:
- Independent services for **news**, **comments**, **censorship**, and **API gateway**
- RESTful APIs for each service
- Full-stack integration via API Gateway
- Docker-based deployment with PostgreSQL and MongoDB

---

## Overview

The system periodically collects articles from RSS feeds, stores them in a PostgreSQL database, and exposes them via an HTTP API.  
Users can view news, leave comments, and the censorship service automatically filters restricted words in the comments.

---

## Services

| Service | Description | Port | Tech |
|----------|-------------|------|------|
| **news** | Aggregates RSS feeds, stores and serves news | `8081` | Go + PostgreSQL |
| **comments** | Stores comments per news, validates via censorship service | `8082` | Go + MongoDB |
| **censorship** | Validates comment text, blocks banned words | `8083` | Go |
| **gateway** | Entry point for all clients, routes to services | `8080` | Go |

---

## Features
- Scheduled polling of multiple RSS feeds
- PostgreSQL for persistent news storage
- MongoDB for comments
- Censorship microservice for content moderation
- API Gateway unifying all endpoints
- Unit & Integration tests
- Docker & docker-compose support
- Structured JSON logging with slog
- End-to-end Request ID propagation across all services
- Built-in static serving of the frontend (Vue/Vuetify bundle in webapp/)

---


## How It Works
- On a timer, the news service iterates over all RSS URLs from config.json, parses title, description, pubDate, and link, and writes them to PostgreSQL.
- Each record has a unique index on link to prevent duplicates.
- The API Gateway proxies all requests to internal services (news, comments, censorship), attaching a request_id for full traceability in logs.
- The comments service stores user comments in MongoDB, while the censorship service validates each comment synchronously before it’s saved.
- The frontend (Vue/Vuetify) calls /news, /news/{id}, and /news/{id}/comment endpoints through the gateway and renders data dynamically.

---

## Data Storage

**PostgreSQL** (news service), table posts:

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

**MongoDB** (comments service), collection comments:

```go
type Comment struct {
    ID        string    `bson:"_id,omitempty"`
    NewsID    string    `bson:"news_id"`
    ParentID  string    `bson:"parent_id,omitempty"`
    Author    string    `bson:"author"`
    Content   string    `bson:"content"`
    CreatedAt time.Time `bson:"created_at"`
}
```
- news_id links a comment to its post.
- parent_id allows nested comment threads.
- created_at is automatically set when adding a new comment.

---

## Running

### build and start the system

```bash
docker compose up -d --build
```

### check running services

```bash
curl http://localhost:8080/news?page={n}
curl http://localhost:8080/news/filter?s={m}
curl http://localhost:8080/news/{id}
curl http://localhost:8080/news/{id}/comment
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

You can also test the full system manually using Postman via the API Gateway:
- GET /news — list paginated news
- GET /news/filter?s=keyword — search by title
- GET /news/{id} — get full details with comments
- POST /news/{id}/comment — add a new comment

---

## Project Structure

```
goNews/
├── censorship/       # Censorship validation service
├── comments/         # Comments service (MongoDB)
├── gateway/          # API Gateway
├── news/             # News service (PostgreSQL)
├── docker-compose.yaml
└── README.md
```
