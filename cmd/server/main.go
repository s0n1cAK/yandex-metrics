package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	config "github.com/s0n1cAK/yandex-metrics/internal/config/server"
	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/s0n1cAK/yandex-metrics/internal/server"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
	"github.com/spf13/pflag"

	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printMetaInfo()

	log, err := logger.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %s \n", err)
		os.Exit(1)
	}

	cfg, err := config.NewConfig(log)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatal("failed to create server config", zap.Error(err))
	}
	defer cfg.Logger.Sync()

	appCtx, appCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer appCancel()

	storage, err := storage.New(appCtx, cfg, log)
	if err != nil {
		log.Fatal("failed to create storage", zap.Error(err))
	}

	srv, err := server.New(&cfg, storage)
	if err != nil {
		log.Fatal("failed to create server", zap.Error(err))
	}

	err = srv.Start(appCtx)
	if err != nil {
		log.Fatal("error while starting server", zap.Error(err))
	}

}

func printMetaInfo() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}
