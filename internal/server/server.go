package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/s0n1cAK/yandex-metrics/internal/server/handlers"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
	filestorage "github.com/s0n1cAK/yandex-metrics/internal/storage/fileStorage"
	"go.uber.org/zap"
)

const (
	minPort = 0
	maxPort = 65535
)

type Server struct {
	Address  string
	Port     int
	Router   *chi.Mux
	Config   *config.ServerConfig
	Storage  storage.BasicStorage
	Consumer *filestorage.Consumer
	Producer *filestorage.Producer
}

func New(cfg *config.ServerConfig, storage storage.BasicStorage) (*Server, error) {
	var err error

	OP := "Server.New"

	if cfg.Logger == nil {
		cfg.Logger, err = logger.NewLogger()
		if err != nil {
			return nil, fmt.Errorf("%s: %s", OP, err)
		}
	}

	address := strings.Split(cfg.Endpoint.HostPort(), ":")

	domain := address[0]
	post, err := strconv.Atoi(address[1])
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	if post <= minPort || post >= maxPort {
		return nil, fmt.Errorf("%s: %v is not an valid port", OP, post)
	}

	if cfg.File == "" {
		cfg.File = "Metrics.data"
	}
	consumer, err := filestorage.NewConsumer(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	producer, err := filestorage.NewProducer(cfg.File, cfg.StoreInterval.Duration())
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	r := chi.NewRouter()
	r.Use(Logging(cfg.Logger))
	r.Use(gzipCompession())
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	if cfg.StoreInterval == 0 {
		r.Use(writeMetrics(producer))
	}

	r.Get("/", handlers.GetMetrics(storage))
	r.Get("/value/{type}/{metric}", handlers.GetMetric(storage))
	r.Get("/ping", handlers.PingDB(cfg.DSN))

	r.Post("/value", handlers.GetMetricJSON(storage))
	r.Post("/value/", handlers.GetMetricJSON(storage))
	r.Post("/update", handlers.SetMetricJSON(storage))
	r.Post("/update/", handlers.SetMetricJSON(storage))
	r.Post("/update/{type}/{metric}/{value}", handlers.SetMetricURL(storage))

	return &Server{
		Address:  domain,
		Port:     post,
		Router:   r,
		Config:   cfg,
		Storage:  storage,
		Consumer: consumer,
		Producer: producer,
	}, nil
}

func (c *Server) logStartupInfo() {
	c.Config.Logger.Info("Starting server",
		zap.String("Address", c.Address),
		zap.Int("Port", c.Port),
		zap.String("File", c.Config.File),
		zap.Bool("Restore", c.Config.Restore),
	)
}

func (c *Server) restoreMetricsFromFile() error {
	OP := "Server.Start.restoreMetricsFromFile"

	if c.Config.Restore {
		defer c.Consumer.Close()

		data, err := c.Consumer.ReadFile()
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		c.Storage.SetAll(data)
	}
	return nil
}

func (c *Server) scheduleFilePersistence() error {
	if c.Config.StoreInterval > 0 {
		if c.Config.StoreInterval > 0 {
			ticker := time.NewTicker(c.Config.StoreInterval.Duration())

			go func() {
				for range ticker.C {
					err := c.Producer.WriteMetrics(c.Storage.GetAll())
					if err != nil {
						c.Config.Logger.Error("Ошибка при сохранении метрик", zap.Error(err))
					} else {
						c.Config.Logger.Info("Метрики сохранены в файл (по таймеру)")
					}
				}
			}()
		}
	}
	return nil
}

func (c *Server) Start() error {
	OP := "Server.Start"

	c.logStartupInfo()

	// Read metrics from file
	err := c.restoreMetricsFromFile()
	if err != nil {
		return err
	}

	err = c.scheduleFilePersistence()
	if err != nil {
		return err
	}

	if err = http.ListenAndServe(
		fmt.Sprintf("%s:%v", c.Address, c.Port),
		c.Router,
	); err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	return nil
}
