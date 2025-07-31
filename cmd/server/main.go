package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/s0n1cAK/yandex-metrics/internal/server"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
	dbstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/dbStorage"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
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

	appCtx, appCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer appCancel()

	var storage storage.BasicStorage

	if cfg.UseDB {
		storage, err = dbstorage.NewPostgresStorage(appCtx, cfg.DSN)
		if err != nil {
			log.Fatal("failed to create storage", zap.Error(err))
		}
	}

	if cfg.UseFile || cfg.UseRAM {
		storage = memstorage.New()

	}

	srv, err := server.New(cfg, storage)
	if err != nil {
		log.Fatal("failed to create server", zap.Error(err))
	}

	err = srv.Start(appCtx)
	if err != nil {
		log.Fatal("error while starting server", zap.Error(err))
	}

}
