package agent

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

func LoadConfig(fs *flag.FlagSet, args []string, log *zap.Logger) (Config, error) {
	cfg := Config{
		Client:         &retryablehttp.Client{},
		Endpoint:       DefaultEndpoint,
		ReportInterval: DefaultReportInterval,
		PollInterval:   DefaultPollInterval,
		Logger:         log,
		Hash:           DefaultHashKey,
		RateLimit:      DefaultRateLimit,
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	fs.Var(&cfg.Endpoint, "a", "Server address, e.g. http://host:port")
	fs.Var(&cfg.ReportInterval, "r", "Report interval (e.g. 10s)")
	fs.Var(&cfg.PollInterval, "p", "Poll interval (e.g. 2s)")
	fs.StringVar(&cfg.Hash, "k", cfg.Hash, "Key to make hash")
	fs.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "Request rate limit to server")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to public key (PEM)")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	configureRetries(cfg)

	if err := ValidateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func NewConfig(log *zap.Logger) (Config, error) {
	return LoadConfig(flag.CommandLine, os.Args[1:], log)
}
