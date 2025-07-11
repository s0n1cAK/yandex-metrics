package main

import (
	"log"
	"net/http"

	"github.com/s0n1cAK/yandex-metrics/internal/agent"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
)

func main() {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatal("Error while parsing env:", err)
	}

	metricsStorage := memstorage.New()

	agent := agent.New(&http.Client{}, string(cfg.Endpoint), metricsStorage)

	agent.Run(cfg.PollTime.Duration(), cfg.ReportTime.Duration())
}
