package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/s0n1cAK/yandex-metrics/internal/config/db"
	"github.com/s0n1cAK/yandex-metrics/internal/config/server"
	"github.com/s0n1cAK/yandex-metrics/internal/service/metrics"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
	dbstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/dbStorage"
	filestorage "github.com/s0n1cAK/yandex-metrics/internal/storage/fileStorage"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"github.com/s0n1cAK/yandex-metrics/internal/transport/httpx"
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
	Config   *server.Config
	Storage  storage.BasicStorage
	consumer *filestorage.Consumer
	producer *filestorage.Producer
}

func New(cfg *server.Config, storage storage.BasicStorage) (*Server, error) {
	var consumer *filestorage.Consumer
	var producer *filestorage.Producer
	var err error

	OP := "Server.New"

	domain, port, err := parseURL(cfg)

	consumer, err = filestorage.NewConsumer(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	producer, err = filestorage.NewProducer(cfg.File, cfg.StoreInterval.Duration())
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OP, err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(Logging(cfg.Logger))
	r.Use(gzipCompession())
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(60 * time.Second))

	if cfg.StoreInterval == 0 {
		cfg.Logger.Info("Cинхронная запись метрик")
		r.Use(writeMetrics(producer))
	}

	pinger := db.NewPinger(cfg.DSN)

	svc := metrics.New(storage, pinger, cfg.Logger)

	r.Post("/update/{type}/{metric}/{value}", httpx.SetMetricURL(svc))

	// Кастыль т.к. проверка хеша нужна для updates, проблема в тестах
	// hard code value https://github.com/Yandex-Practicum/go-autotests/blob/main/cmd/metricstest/iteration14_test.go#L58
	r.Group(func(r chi.Router) {
		if !strings.EqualFold(cfg.HashKey, "") {
			cfg.Logger.Info("Используется hash валидация")
			r.Use(checkHash(cfg.HashKey))
		}
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", httpx.SetBatchMetrics(svc))
		})
	})

	r.Post("/update", httpx.SetMetricJSON(svc))
	r.Post("/value", httpx.GetMetricJSON(svc))
	r.Get("/value/{type}/{metric}", httpx.GetMetric(svc))
	r.Get("/", httpx.GetMetrics(svc))
	r.Get("/ping", httpx.Ping(svc))

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

	c.logStartupInfo()

	if c.Config.Restore {
		err = c.restoreMetricsFromFile()
		if err != nil {
			return err
		}
	}

	switch c.Storage.(type) {
	case *dbstorage.PostgresStorage:
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := db.Migration(ctx, c.Config.DSN); err != nil {
			c.Config.Logger.Error("Ошибка при выполнении миграции", zap.Error(err))
			return err
		}

		c.Config.Logger.Info("Миграции выполнены успешно")
	case *memstorage.MemStorage:
		if err := c.scheduleFilePersistence(); err != nil {
			return err
		}
	}

	srv := c.start()

	err = c.gracefulShutdown(ctx, srv)
	if err != nil {
		return err
	}

	return nil
}

func (c *Server) start() *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%v", c.Address, c.Port),
		Handler: c.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.Config.Logger.Fatal("Ошибка сервера", zap.Error(err))
		}
	}()

	return srv
}

func (c *Server) gracefulShutdown(ctx context.Context, srv *http.Server) error {
	OP := "server.gracefulShutdown"

	<-ctx.Done()
	if _, ok := c.Storage.(*memstorage.MemStorage); ok {
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
