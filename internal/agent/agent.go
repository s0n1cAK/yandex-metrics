package agent

import (
	"net/http"
	"time"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

type Storage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() map[string]models.Metrics
	Clear()
}

type Config struct {
	Storage        Storage
	LastReportTime time.Time
	Client         *http.Client
	Server         string
}

type Agent interface {
	CollectRuntime() error
	RandomValue()
	Counter(value int64) error
	Report() error
}

func New(client *http.Client, server string, storage Storage, lastReportTime time.Time) *Config {
	return &Config{
		Client:         client,
		Server:         server,
		Storage:        storage,
		LastReportTime: lastReportTime,
	}
}
