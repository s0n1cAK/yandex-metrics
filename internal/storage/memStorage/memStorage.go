package memstorage

import (
	"fmt"
	"sync"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

func deepCopy(metrics models.Metrics) models.Metrics {
	clone := metrics
	if metrics.Delta != nil {
		d := *metrics.Delta
		clone.Delta = &d
	}
	if metrics.Value != nil {
		v := *metrics.Value
		clone.Value = &v
	}
	return clone

}

type MemStorage struct {
	values map[string]models.Metrics
	mu     sync.RWMutex
}

func New() *MemStorage {
	return &MemStorage{
		values: make(map[string]models.Metrics),
	}
}

func (s *MemStorage) Set(key string, value models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if key == "" {
		return fmt.Errorf("empty key")
	}

	if value.ID == "" {
		return fmt.Errorf("metric name is nil")
	}

	if value.MType != models.Gauge && value.MType != models.Counter {
		return fmt.Errorf("%s unsupported type of metric", value.MType)
	}

	sValue, ok := s.values[key]
	if ok && sValue.MType != value.MType {
		return fmt.Errorf("%s already in storage with type %s", sValue.ID, sValue.MType)
	}

	s.values[key] = value
	return nil
}

func (s *MemStorage) Get(key string) (models.Metrics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.values[key]
	return deepCopy(val), ok
}

func (s *MemStorage) GetAll() map[string]models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make(map[string]models.Metrics, len(s.values))

	for name, data := range s.values {
		metrics[name] = deepCopy(data)
	}
	return metrics
}

func (s *MemStorage) SetAll(metrics []models.Metrics) {
	for _, value := range metrics {
		err := s.Set(value.ID, value)
		if err != nil {
			fmt.Printf("ошибка при Set: %v\n", err)
		}
	}
}

func (s *MemStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.values {
		delete(s.values, k)
	}
}

func (s *MemStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
}
