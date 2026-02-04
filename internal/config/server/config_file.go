package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

type serverFileConfig struct {
	Address       *string `json:"address"`
	Restore       *bool   `json:"restore"`
	StoreInterval *string `json:"store_interval"`
	StoreFile     *string `json:"store_file"`
	DatabaseDSN   *string `json:"database_dsn"`
	CryptoKey     *string `json:"crypto_key"`
	TrustedSubnet *string `json:"trusted_subnet"`
}

func resolveConfigPath(args []string) (string, error) {
	var path string
	fs := pflag.NewFlagSet("cfg", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVarP(&path, "config", "c", "", "Path to config file (JSON)")

	if err := parseOnlyConfigFlag(fs, args); err != nil {
		return "", err
	}

	if path != "" {
		return path, nil
	}

	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath, nil
	}

	return "", nil
}
func parseOnlyConfigFlag(fs *pflag.FlagSet, args []string) error {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-c" || a == "--config":
			if i+1 < len(args) {
				_ = fs.Set("config", args[i+1])
				i++
			}
		case len(a) > 3 && a[:3] == "-c=":
			_ = fs.Set("config", a[3:])
		case len(a) > 9 && a[:9] == "--config=":
			_ = fs.Set("config", a[9:])
		}
	}
	return nil
}

func loadServerFileConfig(path string) (serverFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return serverFileConfig{}, fmt.Errorf("read config file: %w", err)
	}

	var fc serverFileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return serverFileConfig{}, fmt.Errorf("unmarshal config json: %w", err)
	}
	return fc, nil
}

func applyServerFileConfig(cfg *Config, fc serverFileConfig) error {
	if fc.Address != nil {
		if err := cfg.Endpoint.Set(*fc.Address); err != nil {
			return fmt.Errorf("bad address in config: %w", err)
		}
	}
	if fc.Restore != nil {
		cfg.Restore = *fc.Restore
	}
	if fc.StoreInterval != nil {
		if err := cfg.StoreInterval.Set(*fc.StoreInterval); err != nil {
			return fmt.Errorf("bad store_interval in config: %w", err)
		}
	}
	if fc.StoreFile != nil {
		cfg.File = *fc.StoreFile
	}
	if fc.DatabaseDSN != nil {
		dsn := *fc.DatabaseDSN
		if dsn == "" {
			return nil
		}
		if err := cfg.DSN.Set(dsn); err != nil {
			return fmt.Errorf("bad database_dsn in config: %w", err)
		}
	}
	if fc.CryptoKey != nil {
		cfg.CryptoKey = *fc.CryptoKey
	}
	if fc.TrustedSubnet != nil {
		cfg.TrustedSubnet = strings.TrimSpace(*fc.TrustedSubnet)
	}
	return nil
}
