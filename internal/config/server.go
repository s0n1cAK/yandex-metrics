package config

import (
	"errors"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/s0n1cAK/yandex-metrics/internal/lib"
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

	ServerFlagInit sync.Once
)

type ServerConfig struct {
	Endpoint      Endpoint   `env:"ADDRESS"`
	StoreInterval customTime `env:"STORE_INTERVAL"`
	File          string     `env:"FILE_STORAGE_PATH"`
	Restore       bool       `env:"RESTORE"`
	DSN           DSN        `env:"DATABASE_DSN"`
	Logger        *zap.Logger
	UseFile       bool
	UseDB         bool
	UseRAM        bool
}

func NewServerConfigWithFlags(fs *flag.FlagSet, args []string, log *zap.Logger) (*ServerConfig, error) {
	cfg := &ServerConfig{
		Endpoint:      defaultServerEndpoint,
		StoreInterval: customTime(defaultStoreInterval),
		File:          defaultFile,
		Restore:       defaultRestore,
		DSN:           DSN{},
		Logger:        log,
		UseFile:       false,
		UseDB:         false,
		UseRAM:        true,
	}

	fs.Var(&cfg.Endpoint, "a", "Address server")
	fs.Var(&cfg.StoreInterval, "i", "Store interval")
	fs.StringVar(&cfg.File, "f", cfg.File, "Storage file")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore from file")
	fs.Var(&cfg.DSN, "d", "Data Source Name")

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	checkFileENV := os.Getenv("FILE_STORAGE_PATH")
	checkFileFlag := lib.IsFlagPassed("f")
	if checkFileFlag || checkFileENV != "" {
		cfg.UseDB = false
		cfg.UseFile = true
		cfg.UseRAM = false
	}

	checkDBENV := os.Getenv("DATABASE_DSN")
	checkDBFlag := lib.IsFlagPassed("d")
	if checkDBFlag || checkDBENV != "" {
		cfg.UseDB = true
		cfg.UseFile = false
		cfg.UseRAM = false
	}

	return cfg, nil
}

func NewServerConfig(log *zap.Logger) (*ServerConfig, error) {
	return NewServerConfigWithFlags(flag.CommandLine, os.Args[1:], log)
}
