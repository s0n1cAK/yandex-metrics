package storage

// Фиг знает надо ли иметь два package storage дляserver и агента.
// И нужно ли чтобы этот был отдельным пакетом, а не входил в состав Agent
import (
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

type AgentStorage struct {
	storage []models.Metrics
}

func New() *AgentStorage {
	return &AgentStorage{
		storage: make([]models.Metrics, 0),
	}
}

func (s *AgentStorage) Set(m models.Metrics) {
	s.storage = append(s.storage, m)
}

func (s *AgentStorage) GetAll() []models.Metrics {
	return s.storage
}

func (s *AgentStorage) Clear() {
	s.storage = s.storage[:0]
}
