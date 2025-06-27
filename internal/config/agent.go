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
	ReportTime customTime
	PollTime   customTime
}
type customTime time.Duration

func (ct *customTime) String() string {
	return time.Duration(*ct).String()
}

func (ct customTime) Duration() time.Duration {
	return time.Duration(ct)
}

func (ct *customTime) Set(value string) error {
	var duration time.Duration
	var err error

	if strings.HasSuffix(value, "s") {
		duration, err = time.ParseDuration(value)
		if err != nil {
			return errors.New("invalid duration format '10s'")
		}
	} else {
		seconds, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("invalid numeric duration format")
		}
		duration = time.Duration(seconds) * time.Second
	}

	*ct = customTime(duration)
	return nil
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
		Endpoint:   "http://localhost:8080",
		ReportTime: customTime(defaultReportTime),
		PollTime:   customTime(defaultPollTime),
	}

	flag.Var(&cfg.Endpoint, "a", "Server address in format scheme://host:port")
	flag.Var(&cfg.ReportTime, "r", "Frequency of sending metrics to the server")
	flag.Var(&cfg.PollTime, "p", "Frequency of polling metrics from the package")

	flag.Parse()

	return cfg
}
