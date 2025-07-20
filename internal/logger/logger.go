package logger

import (
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	var cfg zap.Config

	cfg = zap.NewDevelopmentConfig()

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}
