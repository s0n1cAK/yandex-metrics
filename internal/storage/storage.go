package storage

import models "github.com/s0n1cAK/yandex-metrics/internal/model"

type Storage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() map[string]models.Metrics
}
