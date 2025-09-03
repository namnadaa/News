package aggregator

import (
	"context"
	"net/http"
	"news/pkg/rss"
	"news/pkg/storage"
	"time"
)

// Aggregator polls RSS feeds and stores posts in the storage.
type Aggregator struct {
	feeds  []string
	period time.Duration
	client *http.Client
}

// New creates and initializes a new Aggregator.
func New(f []string, p time.Duration) *Aggregator {
	ag := Aggregator{
		feeds:  f,
		period: p,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	return &ag
}

// runOnce fetches all feeds once and sends posts to channels.
func (a *Aggregator) runOnce(ctx context.Context, posts chan<- []storage.Post, errs chan<- error) {
	for _, url := range a.feeds {
		select {
		case <-ctx.Done():
			return
		default:
		}

		news, err := rss.Parse(ctx, a.client, url)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case errs <- err:
			}
			continue
		}

		select {
		case <-ctx.Done():
			return
		case posts <- news:
		}
	}
}

// Run starts periodic fetching until the context is canceled.
func (a *Aggregator) Run(ctx context.Context, posts chan<- []storage.Post, errs chan<- error) {
	if a.period <= 0 {
		a.runOnce(ctx, posts, errs)
		return
	}

	ticker := time.NewTicker(a.period)

	a.runOnce(ctx, posts, errs)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.runOnce(ctx, posts, errs)
		}
	}
}
