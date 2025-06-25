package main

import (
	"log"

	"github.com/s0n1cAK/yandex-metrics/internal/server"
	memStorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
)

// should not use ALL_CAPS in Go names; use CamelCase instead
// Не уверен, что это корректно, когда мы говорим за const
const (
	serverAddr = "localhost"
	serverPort = 8080
)

func main() {
	storage := memStorage.New()

	srv, err := server.New(serverAddr, serverPort, storage)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting server on %s:%v", serverAddr, serverPort)
	srv.MustStart()
}
