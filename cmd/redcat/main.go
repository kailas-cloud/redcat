package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"redcat/internal/api"
	"redcat/internal/config"
	"redcat/internal/service/places"
	"redcat/internal/storage/valkey"
)

func main() {
	cfg := config.FromEnv()

	cli, err := valkey.NewClient(cfg.ValkeyAddrs, cfg.ValkeyUser, cfg.ValkeyPass)
	if err != nil { log.Fatalf("valkey: %v", err) }
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := valkey.EnsurePlacesIndex(ctx, cli.R, cfg.IndexName, cfg.KeyPrefix); err != nil {
		log.Fatalf("ensure index: %v", err)
	}

	store := valkey.NewPlacesStorage(cli.R, cfg.IndexName, cfg.KeyPrefix)
	svc := places.New(store)

	s := api.New()
	api.Register(s.App(), api.Handlers{Places: svc})

	go func() {
		if err := s.App().Listen(cfg.HTTPAddr); err != nil {
			log.Printf("http closed: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	_ = s.App().Shutdown()
}