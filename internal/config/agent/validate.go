package agent

import "errors"

var (
	ErrEmptyEndpoint = errors.New("endpoint is empty")
	ErrBadReport     = errors.New("report interval must be > 0")
	ErrBadPoll       = errors.New("poll interval must be > 0")
)

func ValidateConfig(cfg Config) error {
	if cfg.Endpoint == "" {
		return ErrEmptyEndpoint
	}
	if cfg.ReportInterval.Duration() <= 0 {
		return ErrBadReport
	}
	if cfg.PollInterval.Duration() <= 0 {
		return ErrBadPoll
	}
	return nil
}
