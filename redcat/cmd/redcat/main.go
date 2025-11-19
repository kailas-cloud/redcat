package main

import (
	"github.com/redis/rueidis"
	"log"
	"net/http"
	"os"
	"redcat/internal/clients/embedder"
	"redcat/internal/http/api"
	"redcat/internal/service/categories"
	"redcat/internal/service/places"
	catStore "redcat/internal/storage/categories"
	plcStore "redcat/internal/storage/places"
)

func main() {
	redisAddr := getenv("REDIS_ADDR", "redis:6379")
	embedderURL := getenv("EMBEDDER_URL", "http://embedder:8000")
	httpAddr := getenv("HTTP_ADDR", ":8080")

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{redisAddr},
	})
	if err != nil {
		log.Fatalf("failed to create redis client: %v", err)
	}
	defer client.Close()

	mux := http.NewServeMux()
	server := api.New(
		categories.New(embedder.New(embedderURL), catStore.New(client)),
		places.New(plcStore.New(client)),
	)
	server.Routes(mux)

	log.Printf("RedCat listening on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
