package config

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

const (
	defaultAgentEndpoint = "http://localhost:8080"
	defaultReportTime    = time.Second * 10
	defaultPollTime      = time.Second * 2
)

var (
	ErrInvalidDurationFormat = errors.New("invalid duration format")
	ErrInvalidNumericFormat  = errors.New("invalid numeric duration format")

	AgentFlagInit sync.Once
)

type AgentConfig struct {
	Client     *http.Client
	Endpoint   Endpoint   `env:"ADDRESS"`
	ReportTime customTime `env:"REPORT_INTERVAL"`
	PollTime   customTime `env:"POLL_INTERVAL"`
	Logger     *zap.Logger
}

func NewAgentConfigWithFlags(fs *flag.FlagSet, args []string, log *zap.Logger) (*AgentConfig, error) {
	cfg := &AgentConfig{
		Client:     &http.Client{},
		Endpoint:   defaultAgentEndpoint,
		ReportTime: customTime(defaultReportTime),
		PollTime:   customTime(defaultPollTime),
		Logger:     log,
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	fs.Var(&cfg.Endpoint, "a", "Server address in format scheme://host:port")
	fs.Var(&cfg.ReportTime, "r", "Frequency of sending metrics to the server")
	fs.Var(&cfg.PollTime, "p", "Frequency of polling metrics from the package")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewAgentConfig(log *zap.Logger) (*AgentConfig, error) {
	return NewAgentConfigWithFlags(flag.CommandLine, os.Args[1:], log)
}
