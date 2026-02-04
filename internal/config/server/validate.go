package server

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrEmptyEndpoint = errors.New("endpoint is empty")
	ErrBadStore      = errors.New("store interval must be > 0")
	ErrEmptyFile     = errors.New("file path is empty while restore enabled")
)

func ValidateConfig(cfg Config) error {
	if cfg.Endpoint == "" {
		return ErrEmptyEndpoint
	}
	if cfg.StoreInterval.Duration() <= 0 {
		return ErrBadStore
	}
	if cfg.Restore && cfg.File == "" {
		return ErrEmptyFile
	}
	if cfg.TrustedSubnet != "" {
		if _, _, err := net.ParseCIDR(cfg.TrustedSubnet); err != nil {
			return fmt.Errorf("invalid trusted_subnet CIDR: %w", err)
		}
	}
	return nil
}
