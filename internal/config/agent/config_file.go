package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type agentFileConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
}

func resolveConfigPath(args []string) (string, error) {
	var path string
	fs := flag.NewFlagSet("cfg", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&path, "c", "", "config file path")
	fs.StringVar(&path, "config", "", "config file path")
	if err := fs.Parse(args); err != nil {
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

func loadAgentFileConfig(path string) (agentFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return agentFileConfig{}, fmt.Errorf("read config file: %w", err)
	}

	var fc agentFileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return agentFileConfig{}, fmt.Errorf("unmarshal config json: %w", err)
	}
	return fc, nil
}

func normalizeAddress(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "http://" + s
}

func applyAgentFileConfig(cfg *Config, fc agentFileConfig) error {
	if fc.Address != nil {
		if err := cfg.Endpoint.Set(normalizeAddress(*fc.Address)); err != nil {
			return fmt.Errorf("bad address in config: %w", err)
		}
	}
	if fc.ReportInterval != nil {
		if err := cfg.ReportInterval.Set(*fc.ReportInterval); err != nil {
			return fmt.Errorf("bad report_interval in config: %w", err)
		}
	}
	if fc.PollInterval != nil {
		if err := cfg.PollInterval.Set(*fc.PollInterval); err != nil {
			return fmt.Errorf("bad poll_interval in config: %w", err)
		}
	}
	if fc.CryptoKey != nil {
		cfg.CryptoKey = *fc.CryptoKey
	}
	return nil
}
