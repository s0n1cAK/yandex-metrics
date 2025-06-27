package main

import (
	"net/http"

	"github.com/s0n1cAK/yandex-metrics/internal/agent"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
)

func main() {
	cfg := config.NewAgentConfig()
	metricsStorage := memstorage.New()

	agent := agent.New(&http.Client{}, string(cfg.Endpoint), metricsStorage)

	agent.Run(cfg.PollTime.Duration(), cfg.ReportTime.Duration())
}
