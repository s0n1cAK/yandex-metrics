package agent

import (
	"flag"
	"net/http"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

func LoadConfig(fs *flag.FlagSet, args []string, log *zap.Logger) (Config, error) {
	cfg := Config{
		Client:         &http.Client{},
		Endpoint:       DefaultEndpoint,
		ReportInterval: DefaultReportInterval,
		PollInterval:   DefaultPollInterval,
		Logger:         log,
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	fs.Var(&cfg.Endpoint, "a", "Server address, e.g. http://host:port")
	fs.Var(&cfg.ReportInterval, "r", "Report interval (e.g. 10s)")
	fs.Var(&cfg.PollInterval, "p", "Poll interval (e.g. 2s)")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func NewConfig(log *zap.Logger) (Config, error) {
	return LoadConfig(flag.CommandLine, os.Args[1:], log)
}
