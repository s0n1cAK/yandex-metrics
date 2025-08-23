package main

import (
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/s0n1cAK/yandex-metrics/internal/service/agent"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.NewLogger()
	if err != nil {
		log.Fatal("failed to init logger", zap.Error(err))
	}
	defer log.Sync()

	metricsStorage := memstorage.New()

	agent := agent.New(log, metricsStorage)

	log.Info("Agent started",
		zap.String("endpoint", agent.Server),
		zap.Duration("poll_interval", agent.PollInterval),
		zap.Duration("report_interval", agent.ReportInterval),
	)
	err = agent.Run()
	if err != nil {
		log.Fatal("Error: %w \n", zap.Error(err))
	}
}
