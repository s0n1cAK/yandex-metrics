package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/config/db"
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
	consumer *filestorage.Consumer
	producer *filestorage.Producer
}

func New(cfg *config.ServerConfig, storage storage.BasicStorage) (*Server, error) {
	var consumer *filestorage.Consumer
	var producer *filestorage.Producer
	var err error

	OP := "Server.New"

	if cfg.Logger == nil {
		cfg.Logger, err = logger.NewLogger()
		if err != nil {
			return nil, fmt.Errorf("%s: %s", OP, err)
		}
	}

	parts := strings.Split(cfg.Endpoint.HostPort(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%s: invalid endpoint format", OP)
	}

	domain := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	if port <= minPort || port >= maxPort {
		return nil, fmt.Errorf("%s: %v is not an valid port", OP, port)
	}

	if cfg.File == "" {
		cfg.File = "Metrics.data"
	}
	consumer, err = filestorage.NewConsumer(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	producer, err = filestorage.NewProducer(cfg.File, cfg.StoreInterval.Duration())
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	r := chi.NewRouter()
	r.Use(Logging(cfg.Logger))
	r.Use(gzipCompession())
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	if cfg.UseFile {
		if cfg.StoreInterval == 0 {
			cfg.Logger.Info("Cинхронная запись метрик")
			r.Use(writeMetrics(producer))
		}
	}

	r.Get("/", handlers.GetMetrics(storage, cfg.Logger))
	r.Get("/value/{type}/{metric}", handlers.GetMetric(storage, cfg.Logger))
	if cfg.UseDB {
		r.Get("/ping", handlers.PingDB(cfg.DSN.String(), cfg.Logger))
	}

	r.Post("/value", handlers.GetMetricJSON(storage, cfg.Logger))
	r.Post("/value/", handlers.GetMetricJSON(storage, cfg.Logger))
	r.Post("/update", handlers.SetMetricJSON(storage, cfg.Logger))
	r.Post("/update/", handlers.SetMetricJSON(storage, cfg.Logger))
	r.Post("/updates", handlers.SetBatchMetrics(storage, cfg.Logger))
	r.Post("/updates/", handlers.SetBatchMetrics(storage, cfg.Logger))
	r.Post("/update/{type}/{metric}/{value}", handlers.SetMetricURL(storage, cfg.Logger))

	return &Server{
		Address:  domain,
		Port:     port,
		Router:   r,
		Config:   cfg,
		Storage:  storage,
		consumer: consumer,
		producer: producer,
	}, nil
}

func (c *Server) logStartupInfo() {
	c.Config.Logger.Info("Включаю сервер",
		zap.String("Address", c.Address),
		zap.Int("Port", c.Port),
		zap.String("File", c.Config.File),
		zap.Bool("Restore", c.Config.Restore),
		zap.Bool("Database", c.Config.UseDB),
		zap.Bool("File", c.Config.UseFile),
		zap.Bool("Memory", c.Config.UseRAM),
	)
}

func (c *Server) restoreMetricsFromFile() error {
	OP := "Server.Start.restoreMetricsFromFile"

	if c.Config.Restore {
		defer c.consumer.Close()

		data, err := c.consumer.ReadFile()
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		c.Storage.SetAll(data)
	}
	return nil
}

func (c *Server) scheduleFilePersistence() error {
	if c.Config.StoreInterval > 0 {
		ticker := time.NewTicker(c.Config.StoreInterval.Duration())

		go func() {
			for range ticker.C {
				metrics, err := c.Storage.GetAll()
				if err != nil {
					c.Config.Logger.Error("Ошибка при сохранении метрик", zap.Error(err))
				}
				err = c.producer.WriteMetrics(metrics)
				if err != nil {
					c.Config.Logger.Error("Ошибка при сохранении метрик", zap.Error(err))
				} else {
					c.Config.Logger.Info("Метрики сохранены в файл (по таймеру)")
				}
			}
		}()
	}
	return nil
}

func (c *Server) Start(ctx context.Context) error {
	var err error

	OP := "Server.Start"

	c.logStartupInfo()

	if c.Config.Restore {
		// Read metrics from file
		err = c.restoreMetricsFromFile()
		if err != nil {
			return err
		}
	}

	if c.Config.UseFile {
		err = c.scheduleFilePersistence()
		if err != nil {
			return err
		}
	}
	if c.Config.UseDB {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err = db.InitMigration(ctx, c.Config.DSN)
		if err != nil {
			c.Config.Logger.Error("Ошибка при выполнении миграции", zap.Error(err))
			return err
		}
		c.Config.Logger.Info("Миграции выполнены успешно")
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%v", c.Address, c.Port),
		Handler: c.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.Config.Logger.Fatal("Ошибка сервера", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if c.Config.UseFile {
		metrics, err := c.Storage.GetAll()
		if err != nil {
			c.Config.Logger.Error("Ошибка при сохранении метрик", zap.Error(err))
		}
		err = c.producer.WriteMetrics(metrics)
		if err != nil {
			c.Config.Logger.Error("Ошибка при сохранении метрик", zap.Error(err))
		} else {
			c.Config.Logger.Info("Метрики сохранены в файл по завершению программы")
		}
	}
	c.Config.Logger.Info("Выключаю сервер...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: Попытка остановки сервера завершилась с ошибкой: %w", OP, err)
	}

	return nil
}
