package storage

import (
	"context"
	"fmt"

	"github.com/s0n1cAK/yandex-metrics/internal/config/server"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	dbstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/dbStorage"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

type BasicStorage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() (map[string]models.Metrics, error)
	SetAll([]models.Metrics) error
}

func New(ctx context.Context, cfg server.Config, log *zap.Logger) (BasicStorage, error) {
	if cfg.DSN.String() == "" {
		return nil, fmt.Errorf("DB string is empty")
	}
	if cfg.DSN.Host != "" && cfg.DSN.Port != "" {
		s, err := dbstorage.NewPostgresStorage(ctx, cfg.DSN)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	s := memstorage.New()
	return s, nil
}
