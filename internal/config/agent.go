package config

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	"go.uber.org/zap"
)

const (
	defaultEndpoint   = "localhost"
	defaultReportTime = time.Second * 10
	defaultPollTime   = time.Second * 2
)

var (
	ErrInvalidDurationFormat = errors.New("invalid duration format")
	ErrInvalidNumericFormat  = errors.New("invalid numeric duration format")
)

type AgentConfig struct {
	Client     *http.Client
	Endpoint   endpoint   `env:"ADDRESS"`
	ReportTime customTime `env:"REPORT_INTERVAL"`
	PollTime   customTime `env:"POLL_INTERVAL"`
	Logger     *zap.Logger
}

func formatCustomTime(value string) (customTime, error) {
	var duration time.Duration
	var err error

	if lib.HasLetter(value) {
		duration, err = time.ParseDuration(value)
		if err != nil {
			return 0, ErrInvalidDurationFormat
		}
	} else {
		seconds, err := strconv.Atoi(value)
		if err != nil {
			return 0, ErrInvalidNumericFormat
		}
		duration = time.Duration(seconds) * time.Second
	}

	return customTime(duration), nil
}

type customTime time.Duration

func (ct *customTime) String() string {
	return time.Duration(*ct).String()
}

func (ct customTime) Duration() time.Duration {
	return time.Duration(ct)
}

func (ct *customTime) Set(value string) error {
	gValue, err := formatCustomTime(value)
	if err == nil {
		return err
	}
	*ct = gValue
	return nil
}

func (ct *customTime) UnmarshalText(text []byte) error {
	gValue, err := formatCustomTime(string(text[:]))
	if err == nil {
		return err
	}
	*ct = gValue
	return nil
}

func formatEndpoint(value string) (endpoint, error) {
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}

	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return "", errors.New("invalid endpoint format, must be scheme://host:port")
	}

	hostPort := strings.Split(u.Host, ":")
	if len(hostPort) != 2 {
		return "", errors.New("endpoint must include port")
	}

	port, err := strconv.Atoi(hostPort[1])
	if err != nil || port < 1 || port > 65535 {
		return "", errors.New("invalid port number")
	}

	return endpoint(value), err
}

type endpoint string

func (e *endpoint) String() string {
	return string(*e)
}

func (e *endpoint) Set(value string) error {
	gValue, err := formatEndpoint(string(value[:]))
	fmt.Println(gValue)
	if err != nil {
		return err
	}
	*e = gValue
	return nil
}

func (e *endpoint) UnmarshalText(text []byte) error {
	gValue, err := formatEndpoint(string(text[:]))
	if err != nil {
		return err
	}
	*e = gValue
	return nil
}

func NewAgentConfig(log *zap.Logger) (*AgentConfig, error) {
	cfg := &AgentConfig{
		Client:     &http.Client{},
		Endpoint:   "http://localhost:8080",
		ReportTime: customTime(defaultReportTime),
		PollTime:   customTime(defaultPollTime),
		Logger:     log,
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	flag.Var(&cfg.Endpoint, "a", "Server address in format scheme://host:port")
	flag.Var(&cfg.ReportTime, "r", "Frequency of sending metrics to the server")
	flag.Var(&cfg.PollTime, "p", "Frequency of polling metrics from the package")

	flag.Parse()

	return cfg, nil
}
