package agent

import (
	"fmt"
	"log"
	"net/http"
	"time"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

const fiveMinutes = time.Second * 300

type Storage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() map[string]models.Metrics
	Clear()
	Delete(key string)
}

type Agent struct {
	Storage        Storage
	LastReportTime time.Duration
	Client         *http.Client
	Scheme         string
	Server         string
}

func New(client *http.Client, server string, storage Storage) *Agent {
	return &Agent{
		Client:  client,
		Server:  server,
		Storage: storage,
	}
}

// https://gosamples.dev/range-over-ticker/

/*
 */
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
				log.Printf("CollectRuntime error: %s", err)
			}
			if err := agent.CollectRandomValue(); err != nil {
				log.Printf("CollectRandomValue error: %s", err)
			}
			if err := agent.CollectIncrementCounter("PollCount", 1); err != nil {
				log.Printf("CollectIncrementCounter error: %s", err)
			}

		case <-reportTicker.C:
			log.Println("Reporting metrics")
			if err := agent.Report(); err != nil {
				log.Printf("Report error: %s", err)
			}
		}
	}
}
