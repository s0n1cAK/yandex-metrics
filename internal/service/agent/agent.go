package agent

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/s0n1cAK/yandex-metrics/internal/config/agent"
	"github.com/s0n1cAK/yandex-metrics/internal/crypt"
	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"go.uber.org/zap"
)

const fiveMinutes = time.Second * 300

type Storage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() (map[string]models.Metrics, error)
	Clear()
	Delete(key string)
}

type Agent struct {
	Storage        Storage
	Client         *retryablehttp.Client
	Server         string
	Logger         *zap.Logger
	hash           string
	publicKey      *rsa.PublicKey
	PollInterval   time.Duration
	ReportInterval time.Duration
	realIP         string
	httpLimiter    chan struct{}
}

func New(log *zap.Logger, storage Storage) *Agent {
	cfg, err := agent.NewConfig(log)
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		log.Fatal("Error while parsing config", zap.Error(err))
	}

	ip, err := localIPForRemote(cfg.Endpoint.HostPort())
	if err != nil {
		log.Fatal("failed to resolve agent ip", zap.Error(err))
	}

	a := &Agent{
		Client:         cfg.Client,
		Server:         cfg.Endpoint.String(),
		Storage:        storage,
		Logger:         cfg.Logger,
		hash:           cfg.Hash,
		PollInterval:   cfg.PollInterval.Duration(),
		ReportInterval: cfg.ReportInterval.Duration(),
		realIP:         ip,
		httpLimiter:    make(chan struct{}, cfg.RateLimit),
	}

	if cfg.CryptoKey != "" {
		pub, err := crypt.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			a.Logger.Fatal("Error while loading public key", zap.Error(err))
		}
		a.publicKey = pub
	}

	return a
}

// https://gosamples.dev/range-over-ticker/

func (agent *Agent) Run(ctx context.Context) error {
	if agent.PollInterval < time.Second {
		return fmt.Errorf("poll can't be lower that 2 seconds")
	}

	if agent.PollInterval > agent.ReportInterval {
		return fmt.Errorf("poll can't be higher that reportInterval")
	}

	if agent.ReportInterval > fiveMinutes {
		return fmt.Errorf("report can't be higher that 5 minutes")
	}

	pollTicker := time.NewTicker(agent.PollInterval)
	reportTicker := time.NewTicker(agent.ReportInterval)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			if err := agent.CollectRuntime(); err != nil {
				agent.Logger.Error("CollectRuntime error:", zap.Error(err))
			}
			if err := agent.CollectRandomValue(); err != nil {
				agent.Logger.Error("CollectRandomValue error:", zap.Error(err))
			}
			if err := agent.CollectIncrementCounter("PollCount", 1); err != nil {
				agent.Logger.Error("CollectIncrementCounter error:", zap.Error(err))
			}
			if err := agent.CollectGopsutil(); err != nil {
				agent.Logger.Error("CollectGopsutil error:", zap.Error(err))
			}

		case <-reportTicker.C:
			agent.Logger.Info("Reporting metrics")
			if err := agent.Report(); err != nil {
				agent.Logger.Error("Error while reporting:", zap.Error(err))
			}

		case <-ctx.Done():
			agent.Logger.Info("Shutdown signal received, flushing metrics")

			_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := agent.Report(); err != nil {
				agent.Logger.Error("Final report failed", zap.Error(err))
			}
			return nil
		}
	}
}

func (agent *Agent) updateGaugeMetruc(name string, value float64) error {
	err := agent.Storage.Set(uniqMetric(name), models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: lib.FloatPtr(value),
	})
	return err
}

func (agent *Agent) updateCounterMetruc(name string, value int64) error {
	err := agent.Storage.Set(uniqMetric(name), models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: lib.IntPtr(value),
	})
	return err
}
