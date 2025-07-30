package config

import (
	"errors"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

const (
	defaultServerEndpoint = "http://localhost:8080"
	defaultStoreInterval  = time.Second * 300
	defaultFile           = "Metrics.data"
	defaultRestore        = true
	defaultDSN            = "host=localhost user=test_user password=test_password dbname=metrics sslmode=disable"
)

var (
	ErrInvalidAddressFormat = errors.New("need address in a form host:port")

	ServerFlagInit sync.Once
)

type ServerConfig struct {
	Endpoint      Endpoint   `env:"ADDRESS"`
	StoreInterval customTime `env:"STORE_INTERVAL"`
	File          string     `env:"FILE_STORAGE_PATH"`
	Restore       bool       `env:"RESTORE"`
	DSN           string     `end:"DATABASE_DSN"`
	Logger        *zap.Logger
}

func NewServerConfigWithFlags(fs *flag.FlagSet, args []string, log *zap.Logger) (*ServerConfig, error) {
	cfg := &ServerConfig{
		Endpoint:      defaultServerEndpoint,
		StoreInterval: customTime(defaultStoreInterval),
		File:          defaultFile,
		Restore:       defaultRestore,
		DSN:           defaultDSN,
		Logger:        log,
	}

	fs.Var(&cfg.Endpoint, "a", "Address server")
	fs.Var(&cfg.StoreInterval, "i", "Store interval")
	fs.StringVar(&cfg.File, "f", cfg.File, "Storage file")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore from file")
	fs.StringVar(&cfg.DSN, "d", cfg.DSN, "Data Source Name")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewServerConfig(log *zap.Logger) (*ServerConfig, error) {
	return NewServerConfigWithFlags(flag.CommandLine, os.Args[1:], log)
}
