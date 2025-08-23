package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/s0n1cAK/yandex-metrics/internal/config/agent"
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
	LastReportTime time.Duration
	Client         *retryablehttp.Client
	Server         string
	Logger         *zap.Logger
	Hash           string
	httpLimiter    chan struct{}
}

func New(cfg agent.Config, storage Storage) *Agent {
	return &Agent{
		Client:      cfg.Client,
		Server:      cfg.Endpoint.String(),
		Storage:     storage,
		Logger:      cfg.Logger,
		Hash:        cfg.Hash,
		httpLimiter: make(chan struct{}, cfg.RateLimit),
	}
}

// https://gosamples.dev/range-over-ticker/

func (agent *Agent) Run(pollInterval, reportInterval time.Duration) error {
	if pollInterval < time.Second {
		return fmt.Errorf("PollInterval can't be lower that 2 seconds")
	}

	if pollInterval > reportInterval {
		return fmt.Errorf("PollInterval can't be higher that reportInterval")
	}

	if reportInterval > fiveMinutes {
		return fmt.Errorf("reportInterval can't be higher that 5 minutes")
	}

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errs := make(chan error)

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

			agent.CollectGopsutil(ctx, errs)

		case <-reportTicker.C:
			agent.Logger.Info("Reporting metrics")
			err := agent.Report()
			if err != nil {
				agent.Logger.Error("Error while reporting:", zap.Error(err))
			}
		case err := <-errs:
			agent.Logger.Error("Error while reporting from channel:", zap.Error(err))
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
