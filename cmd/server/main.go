package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// should not use ALL_CAPS in Go names; use CamelCase instead
// Не уверен, что это корректно, когда мы говорим за const
const (
	serverAddr = "localhost"
	serverPort = "8080"
)

type Metric struct {
	ID    string
	MType string
	Value float64
	Delta int64
}

type MemStorage struct {
	values map[string]Metric
}

type Storage interface {
	SetHandler(w http.ResponseWriter, r *http.Request)
}

func main() {
	storage := &MemStorage{
		values: make(map[string]Metric),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/update/{type}/{metric}/{value}", storage.SetHandler)

	sAddr := fmt.Sprint(serverAddr, ":", serverPort)

	log.Printf("Starting server on %s:%s", serverAddr, serverPort)
	err := http.ListenAndServe(sAddr, r)
	if err != nil {
		panic(err)
	}
}
