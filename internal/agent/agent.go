package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/config/agent"
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
	Client         *http.Client
	Server         string
	Logger         *zap.Logger
}

func New(cfg agent.Config, storage Storage) *Agent {
	return &Agent{
		Client:  cfg.Client,
		Server:  cfg.Endpoint.String(),
		Storage: storage,
		Logger:  cfg.Logger,
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

		case <-reportTicker.C:
			agent.Logger.Info("Reporting metrics")
			err := agent.Report()
			if err != nil {
				agent.Logger.Error("Error while reporting:", zap.Error(err))
			}
		}
	}
}
