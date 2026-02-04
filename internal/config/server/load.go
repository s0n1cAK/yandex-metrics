package server

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func LoadConfig(fs *pflag.FlagSet, args []string, logger *zap.Logger) (Config, error) {
	cfg := Config{
		Endpoint:      DefaultEndpoint,
		StoreInterval: DefaultStoreInterval,
		File:          DefaultFile,
		Restore:       DefaultRestore,
		DSN:           customtype.DSN{},
		Logger:        logger,
	}

	cfgPath, err := resolveConfigPath(args)
	if err != nil {
		return Config{}, err
	}
	if cfgPath != "" {
		fc, err := loadServerFileConfig(cfgPath)
		if err != nil {
			return Config{}, err
		}
		if err := applyServerFileConfig(&cfg, fc); err != nil {
			return Config{}, err
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	fs.StringVarP(new(string), "config", "c", "", "Path to config file (JSON)")

	fs.VarP(&cfg.Endpoint, "endpoint", "a", "Server listen address, e.g. http://host:port")
	fs.VarP(&cfg.StoreInterval, "store-interval", "i", "Store interval (e.g. 5m)")
	fs.StringVarP(&cfg.File, "file", "f", cfg.File, "Storage file path")
	fs.BoolVarP(&cfg.Restore, "restore", "r", cfg.Restore, "Restore metrics from file on start")
	fs.VarP(&cfg.DSN, "dsn", "d", "Database DSN")

	fs.StringVarP(&cfg.HashKey, "hash-key", "k", cfg.HashKey, "Hash key to validate request from agent")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to private key (PEM)")

	fs.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Path to audit file")
	fs.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "URL of audit endpoint")

	fs.StringVarP(&cfg.TrustedSubnet, "trusted-subnet", "t", cfg.TrustedSubnet, "Trusted subnet in CIDR")

	fs.StringVar(&cfg.GRPCAddress, "grpc-address", cfg.GRPCAddress, "gRPC listen address")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if err := ValidateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func NewConfig(log *zap.Logger) (Config, error) {
	return LoadConfig(pflag.CommandLine, os.Args[1:], log)
}
