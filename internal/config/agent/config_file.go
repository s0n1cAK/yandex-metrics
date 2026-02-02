package agent

import (
	"encoding/json"
	"fmt"
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
	for i := 0; i < len(args); i++ {
		a := args[i]

		switch a {
		case "-c", "-config", "--config":
			if i+1 < len(args) {
				return args[i+1], nil
			}
			return "", fmt.Errorf("%s requires a value", a)
		}

		if strings.HasPrefix(a, "-c=") {
			return strings.TrimPrefix(a, "-c="), nil
		}
		if strings.HasPrefix(a, "-config=") {
			return strings.TrimPrefix(a, "-config="), nil
		}
		if strings.HasPrefix(a, "--config=") {
			return strings.TrimPrefix(a, "--config="), nil
		}
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
