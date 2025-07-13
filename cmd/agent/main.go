package main

import (
	"fmt"
	"os"

	"github.com/s0n1cAK/yandex-metrics/internal/agent"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %s \n", err)
		os.Exit(1)
	}
	defer log.Sync()

	cfg, err := config.NewAgentConfig(log)
	if err != nil {
		log.Fatal("Error while parsing env", zap.Error(err))
	}

	metricsStorage := memstorage.New()

	agent := agent.New(*cfg, metricsStorage)

	log.Info("Agent started",
		zap.String("endpoint", cfg.Endpoint.String()),
		zap.Duration("poll_interval", cfg.PollTime.Duration()),
		zap.Duration("_interval", cfg.PollTime.Duration()),
	)
	agent.Run(cfg.PollTime.Duration(), cfg.ReportTime.Duration())
}
