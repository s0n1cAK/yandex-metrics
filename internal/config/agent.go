package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEndpoint   = "localhost"
	defaultReportTime = time.Second * 10
	defaultPollTime   = time.Second * 2
)

type AgentConfig struct {
	Endpoint   endpoint
	ReportTime time.Duration
	PollTime   time.Duration
}

type endpoint string

func (e *endpoint) String() string {
	return string(*e)
}

func (e *endpoint) Set(value string) error {
	i := strings.LastIndex(value, ":")
	if i == -1 {
		return errors.New("must be in format host:port")
	}

	portStr := value[i+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("invalid port number")
	}
	*e = endpoint(value)

	return nil
}

func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{
		Endpoint: "http://localhost:8080",
	}

	flag.Var(&cfg.Endpoint, "a", "Server address in format scheme://host:port")
	flag.DurationVar(&cfg.ReportTime, "r", defaultReportTime, "Frequency of sending metrics to the server")
	flag.DurationVar(&cfg.PollTime, "p", defaultPollTime, "Frequency of polling metrics from the package")

	flag.Parse()

	return cfg
}
