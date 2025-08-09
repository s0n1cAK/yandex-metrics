package server

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"go.uber.org/zap"
)

func LoadConfig(fs *flag.FlagSet, args []string, logger *zap.Logger) (Config, error) {
	cfg := Config{
		Endpoint:      DefaultEndpoint,
		StoreInterval: DefaultStoreInterval,
		File:          DefaultFile,
		Restore:       DefaultRestore,
		DSN:           customtype.DSN{},
		Logger:        logger,
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	fs.Var(&cfg.Endpoint, "a", "Server listen address, e.g. http://host:port")
	fs.Var(&cfg.StoreInterval, "i", "Store interval (e.g. 5m)")
	fs.StringVar(&cfg.File, "f", cfg.File, "Storage file path")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore metrics from file on start")
	fs.Var(&cfg.DSN, "d", "Database DSN")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func NewConfig(log *zap.Logger) (Config, error) {
	return LoadConfig(flag.CommandLine, os.Args[1:], log)
}
