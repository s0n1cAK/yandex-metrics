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

	cfgPath, err := resolveConfigPath(args)
	if err != nil {
		return Config{}, err
	}
	if cfgPath != "" {
		fc, err := loadAgentFileConfig(cfgPath)
		if err != nil {
			return Config{}, err
		}
		if err := applyAgentFileConfig(&cfg, fc); err != nil {
			return Config{}, err
		}
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

	fs.StringVar(new(string), "c", "", "Path to config file (JSON)")
	fs.StringVar(new(string), "config", "", "Path to config file (JSON)")

	fs.StringVar(&cfg.GRPCAddress, "grpc-address", cfg.GRPCAddress, "gRPC server address")

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
