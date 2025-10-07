package main

import (
	"gateway/internal/handler"
	"log"
	"net/http"
)

func main() {
	h := handler.NewHandler()

	log.Println("API Gateway listening on :8080")
	err := http.ListenAndServe(":8080", h.Router())
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
