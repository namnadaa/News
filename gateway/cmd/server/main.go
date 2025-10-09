package main

import (
	"gateway/internal/handler"
	"log"
	"net/http"
	"os"
)

func main() {
	h := handler.New(
		os.Getenv("NEWS_SERVICE_URL"),
		os.Getenv("COMMENTS_SERVICE_URL"),
	)

	log.Println("API Gateway listening on :8080")
	err := http.ListenAndServe(":8080", h.Router())
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
