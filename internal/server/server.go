package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/s0n1cAK/yandex-metrics/internal/server/handlers"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
)

const (
	minPort = 0
	maxPort = 65535
)

type (
	Server struct {
		sAddr  string
		sPort  int
		Router *chi.Mux
	}
)

func New(cfg *config.ServerConfig, storage storage.Storage) (*Server, error) {
	OP := "Server.New"

	if cfg.Port <= minPort || cfg.Port >= maxPort {
		return nil, fmt.Errorf("%s: %v is not an valid port", OP, cfg.Port)
	}

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware(cfg.Logger))
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", handlers.GetMetrics(storage))
	r.Get("/value/{type}/{metric}", handlers.GetMetric(storage))

	r.Post("/value/", handlers.GetMetricJSON(storage))
	r.Post("/update", handlers.SetMetricJSON(storage))
	r.Post("/update/", handlers.SetMetricJSON(storage))
	r.Post("/update/{type}/{metric}/{value}", handlers.SetMetricURL(storage))

	return &Server{
		sAddr:  cfg.Address,
		sPort:  cfg.Port,
		Router: r,
	}, nil
}

func (c *Server) Start() error {
	OP := "Server.Start"

	err := http.ListenAndServe(
		fmt.Sprintf("%s:%v", c.sAddr, c.sPort),
		c.Router,
	)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}
	return nil
}
