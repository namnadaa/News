package main

import (
	"comments/pkg/api"
	"comments/pkg/storage/mongo"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connStr := os.Getenv("MONGO_URI")
	if connStr == "" {
		connStr = "mongodb://comments-mongo:27017"
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "commentsdb"
	}

	collectionName := os.Getenv("MONGO_COLLECTION")
	if collectionName == "" {
		collectionName = "comments"
	}

	db, err := mongo.New(ctx, connStr, dbName, collectionName)
	if err != nil {
		slog.Error("could not create DB storage", "err", err)
		os.Exit(1)
	}

	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.Close(closeCtx)
		if err != nil {
			slog.Error("failed to close MongoDB", "err", err)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	go mongo.Moderation(ctx, ticker, db)

	api := api.New(db)
	srv := &http.Server{
		Addr:    ":8083",
		Handler: api.Router(),
	}

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("http server failed", "err", err)
		stop()
	}

	<-ctx.Done()
	slog.Info("shutting down")

	sdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(sdCtx)
	if err != nil {
		slog.Error("http server shutdown error", "err", err)
	}
}
