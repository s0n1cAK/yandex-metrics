package server

import "errors"

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
	return nil
}
