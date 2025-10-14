package main

import (
	"context"
	"log/slog"
	"net/http"
	"news/internal/config"
	"news/pkg/aggregator"
	"news/pkg/api"
	"news/pkg/storage"
	"news/pkg/storage/postgres"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	var cnf config.Config
	err := cnf.Load("./config.json")
	if err != nil {
		slog.Error("could not read configuration file", "err", err)
		os.Exit(1)
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://news_user:strongpassword@localhost:5435/newsdb?sslmode=disable"
	}
	db, err := postgres.New(connStr)
	if err != nil {
		slog.Error("could not create DB storage", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	apiSrv := api.New(db)

	postsCh := make(chan []storage.Post)
	errsCh := make(chan error)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ag := aggregator.New(cnf.RSS, time.Duration(cnf.RequestPeriod)*time.Minute)
	go ag.Run(ctx, postsCh, errsCh)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case posts, ok := <-postsCh:
				if !ok {
					return
				}
				for _, post := range posts {
					saved, err := db.AddPost(post)
					if err != nil {
						slog.Error("could not add post", "link", post.Link, "err", err)
						continue
					}
					slog.Info("post added", "post", saved.ID, "link", saved.Link)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errsCh:
				if !ok {
					return
				}
				slog.Error("aggregator error", "err", err)
			}
		}
	}()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: apiSrv.Router(),
	}

	go func() {
		slog.Info("starting news service", "address", srv.Addr)
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down news service")

	sdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(sdCtx)
	if err != nil {
		slog.Error("http server shutdown error", "err", err)
	}
}
