package agent

import (
	"net/http"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/agent/storage"
)

type Config struct {
	Storage        *storage.AgentStorage
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

func New(client *http.Client, server string, storage *storage.AgentStorage, lastReportTime time.Time) *Config {
	return &Config{
		Client:         client,
		Server:         server,
		Storage:        storage,
		LastReportTime: lastReportTime,
	}
}
