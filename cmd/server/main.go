package main

import (
	"fmt"
	"os"

	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/s0n1cAK/yandex-metrics/internal/server"
	memStorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %s \n", err)
		os.Exit(1)
	}

	cfg, err := config.NewServerConfig(log)
	if err != nil {
		log.Fatal("failed to create server config", zap.Error(err))
	}
	defer cfg.Logger.Sync()

	storage := memStorage.New()

	srv, err := server.New(cfg, storage)
	if err != nil {
		log.Fatal("failed to create server", zap.Error(err))
	}

	err = srv.Start()
	if err != nil {
		log.Fatal("error while starting server", zap.Error(err))
	}

}
