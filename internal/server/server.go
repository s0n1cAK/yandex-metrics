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
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/s0n1cAK/yandex-metrics/internal/server/handlers"
	filestorage "github.com/s0n1cAK/yandex-metrics/internal/storage/fileStorage"
	"go.uber.org/zap"
)

const (
	minPort = 0
	maxPort = 65535
)

type Storage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	SetAll([]models.Metrics)
	GetAll() map[string]models.Metrics
}

type Server struct {
	Address string
	Port    int
	Router  *chi.Mux
	Config  *config.ServerConfig
	Storage Storage
}

func New(cfg *config.ServerConfig, storage Storage) (*Server, error) {
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

	r := chi.NewRouter()
	r.Use(Logging(cfg.Logger))
	r.Use(gzipCompession())
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", handlers.GetMetrics(storage))
	r.Get("/value/{type}/{metric}", handlers.GetMetric(storage))

	r.Post("/value", handlers.GetMetricJSON(storage))
	r.Post("/value/", handlers.GetMetricJSON(storage))
	r.Post("/update", handlers.SetMetricJSON(storage))
	r.Post("/update/", handlers.SetMetricJSON(storage))
	r.Post("/update/{type}/{metric}/{value}", handlers.SetMetricURL(storage))

	return &Server{
		Address: domain,
		Port:    post,
		Router:  r,
		Config:  cfg,
		Storage: storage,
	}, nil
}

func (c *Server) Start() error {
	OP := "Server.Start"

	c.Config.Logger.Info("Starting server",
		zap.String("Address", c.Address),
		zap.Int("Port", c.Port),
		zap.String("File", c.Config.File),
		zap.Bool("Restore", c.Config.Restore),
	)

	// Read metrics from file
	if c.Config.Restore {
		file, err := filestorage.NewConsumer(c.Config.File)
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}
		defer file.Close()

		data, err := file.ReadFile()
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		c.Storage.SetAll(data)
	}

	if c.Config.StoreInterval > 0 {
		file, err := filestorage.NewProducer(c.Config.File, c.Config.StoreInterval.Duration())
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}
		ticker := time.NewTicker(c.Config.StoreInterval.Duration())
		go func() {
			for range ticker.C {
				err := file.WriteMetrics(c.Storage.GetAll())
				if err != nil {
					c.Config.Logger.Error("Ошибка при сохранении метрик", zap.Error(err))
				} else {
					c.Config.Logger.Info("Метрики сохранены в файл")
				}
			}
		}()
	}

	err := http.ListenAndServe(
		fmt.Sprintf("%s:%v", c.Address, c.Port),
		c.Router,
	)

	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}
	return nil
}
