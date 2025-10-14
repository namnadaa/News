package main

import (
	"context"
	"gateway/internal/handler"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	hndlr := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(hndlr))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	h := handler.New(
		os.Getenv("NEWS_SERVICE_URL"),
		os.Getenv("COMMENTS_SERVICE_URL"),
		os.Getenv("CENSORSHIP_SERVICE_URL"),
	)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: h.Router(),
	}

	go func() {
		slog.Info("starting gateway service", "address", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down censorship service")

	sdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(sdCtx)
	if err != nil {
		slog.Error("http server shutdown error", "err", err)
	}
}
