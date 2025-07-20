package config

import (
	"errors"
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

const (
	defaultServerEndpoint = "http://localhost:8080"
	defaultStoreInterval  = time.Second * 300
	defaultFile           = "Metrics.data"
	defaultRestore        = true
)

var (
	ErrInvalidAddressFormat = errors.New("need address in a form host:port")
)

type ServerConfig struct {
	Endpoint      Endpoint   `env:"ADDRESS"`
	StoreInterval customTime `env:"STORE_INTERVAL"`
	File          string     `env:"FILE_STORAGE_PATH"`
	Restore       bool       `env:"RESTORE"`
	Logger        *zap.Logger
}

func NewServerConfig(log *zap.Logger) (*ServerConfig, error) {
	cfg := &ServerConfig{
		Endpoint:      defaultServerEndpoint,
		StoreInterval: customTime(defaultStoreInterval),
		File:          defaultFile,
		Restore:       defaultRestore,
		Logger:        log,
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	flag.Var(&cfg.Endpoint, "a", "Address server")
	flag.Var(&cfg.StoreInterval, "i", "Frequency of saving metrics to file")
	flag.StringVar(&cfg.File, "f", cfg.File, "File for metrics")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore metrics from file")
	flag.Parse()

	return cfg, nil
}
