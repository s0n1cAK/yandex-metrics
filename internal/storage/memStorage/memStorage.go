package memstorage

import (
	"fmt"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

// Добавить Mutex
type MemStorage struct {
	values map[string]models.Metrics
}

func New() *MemStorage {
	return &MemStorage{
		values: make(map[string]models.Metrics),
	}
}

func (s *MemStorage) Set(key string, value models.Metrics) error {
	if key == "" {
		return fmt.Errorf("empty key")
	}

	if value.ID == "" {
		return fmt.Errorf("metric name is nil")
	}

	if value.MType != models.Gauge && value.MType != models.Counter {
		return fmt.Errorf("%s unsupported type of metric", value.MType)
	}

	s.values[key] = value
	return nil
}

func (s *MemStorage) Get(key string) (models.Metrics, bool) {
	val, ok := s.values[key]
	return val, ok
}

func (s *MemStorage) GetAll() map[string]models.Metrics {
	return s.values
}

func (s *MemStorage) Clear() {
	for k := range s.values {
		delete(s.values, k)
	}
}

func (s *MemStorage) Delete(key string) {
	delete(s.values, key)
}
