package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	_ = http.ListenAndServe(":8080", nil)
}
