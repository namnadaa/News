package aggregator

import (
	"context"
	"log/slog"
	"net/http"
	"news/pkg/rss"
	"news/pkg/storage"
	"time"
)

type Aggregator struct {
	store  storage.Storage
	feeds  []string
	period time.Duration
	client *http.Client
}

func New(s storage.Storage, f []string, p time.Duration) *Aggregator {
	ag := Aggregator{
		store:  s,
		feeds:  f,
		period: p,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	return &ag
}

func (a *Aggregator) RunOnce(ctx context.Context) {
	for _, url := range a.feeds {
		select {
		case <-ctx.Done():
			return
		default:
		}

		posts, err := rss.Parse(ctx, a.client, url)
		if err != nil {
			slog.Error("runOnce: couldn't load RSS feed", "url", url, "err", err)
			continue
		}

		for _, post := range posts {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := a.store.AddPost(post)
			if err != nil {
				slog.Error("runOnce: couldn't add post to database", "err", err)
				continue
			}
		}
	}
}
