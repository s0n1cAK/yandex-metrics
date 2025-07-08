package main

import (
	"log"

	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/server"
	memStorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
)

func main() {
	cfg := config.NewServerConfig()
	storage := memStorage.New()

	srv, err := server.New(cfg.Address, cfg.Port, storage)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting server on %s:%v", cfg.Address, cfg.Port)
	err = srv.Start()
	if err != nil {
		log.Fatal(err)
	}

}
