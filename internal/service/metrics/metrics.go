package metrics

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/s0n1cAK/yandex-metrics/internal/audit"
	"github.com/s0n1cAK/yandex-metrics/internal/domain"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

// Repository интерфейс определяет контракт для хранилища метрик.
type Repository interface {
	// Set устанавливает значение метрики с указанным идентификатором
	Set(id string, m models.Metrics) error
	// Get возвращает метрику с указанным идентификатором и флаг существования
	Get(id string) (models.Metrics, bool)
	// GetAll возвращает все метрики в хранилище
	GetAll() (map[string]models.Metrics, error)
	// SetAll устанавливает значения для пакета метрик
	SetAll(batch []models.Metrics) error
}

// Pinger интерфейс определяет контракт для проверки подключения к базе данных.
type Pinger interface {
	// Ping проверяет доступность базы данных
	Ping(ctx context.Context) error
}

// Service интерфейс определяет контракт для бизнес-логики метрик.
type Service interface {
	// Set устанавливает значение метрики
	Set(ctx context.Context, m models.Metrics, ip string) error
	// SetBatch устанавливает значения для пакета метрик
	SetBatch(ctx context.Context, batch []models.Metrics, ip string) error
	// Get возвращает значение метрики по идентификатору и типу
	Get(ctx context.Context, id, mtype string) (models.Metrics, error)
	// ListIDs возвращает список всех идентификаторов метрик
	ListIDs(ctx context.Context) ([]string, error)
	// Ping проверяет доступность базы данных
	Ping(ctx context.Context) error
}

type service struct {
	repo      Repository
	ping      Pinger
	log       *zap.Logger
	publisher audit.AuditPublisher
}

// New создает новый экземпляр сервиса метрик с заданными зависимостями.
func New(repo Repository, ping Pinger, log *zap.Logger, publisher audit.AuditPublisher) Service {
	return &service{repo: repo, ping: ping, log: log, publisher: publisher}
}

func (s *service) Set(ctx context.Context, m models.Metrics, ip string) error {
	if m.ID == "" {
		return domain.ErrInvalidPayload
	}
	switch m.MType {
	case models.Gauge:
		if m.Value == nil {
			return domain.ErrInvalidPayload
		}
	case models.Counter:
		if m.Delta == nil {
			return domain.ErrInvalidPayload
		}
	default:
		return domain.ErrInvalidType
	}

	if err := s.repo.Set(m.ID, m); err != nil {
		s.log.Error(err.Error())
		return err
	}

	s.notify([]models.Metrics{m}, ip)
	s.log.Info("metric set", zap.String("id", m.ID), zap.String("type", m.MType))
	return nil
}

func (s *service) SetBatch(ctx context.Context, batch []models.Metrics, ip string) error {
	if len(batch) == 0 {
		return domain.ErrInvalidPayload
	}
	for _, m := range batch {

		switch m.MType {
		case models.Gauge:
			s.log.Debug("metric set", zap.String("id", m.ID), zap.String("type", m.MType), zap.Float64("Value", *m.Value))
			if m.ID == "" || m.Value == nil {
				return domain.ErrInvalidPayload
			}
		case models.Counter:
			s.log.Debug("metric set", zap.String("id", m.ID), zap.String("type", m.MType), zap.Int64("Value", *m.Delta))
			if m.ID == "" || m.Delta == nil {
				return domain.ErrInvalidPayload
			}
		default:
			return domain.ErrInvalidType
		}
	}

	s.notify(batch, ip)
	return s.repo.SetAll(batch)
}

func (s *service) Get(ctx context.Context, id, mtype string) (models.Metrics, error) {
	if id == "" || mtype == "" {
		return models.Metrics{}, domain.ErrInvalidPayload
	}
	v, ok := s.repo.Get(id)
	if !ok {
		return models.Metrics{}, domain.ErrNotFound
	}
	if v.MType != mtype {
		return models.Metrics{}, domain.ErrNotFound
	}
	return v, nil
}

func (s *service) ListIDs(ctx context.Context) ([]string, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	for _, m := range items {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

func (s *service) Ping(ctx context.Context) error {
	if s.ping == nil {
		return errors.New("no pinger configured")
	}
	return s.ping.Ping(ctx)
}

func (s *service) notify(metrics []models.Metrics, ip string) {
	err := s.publisher.Publish(models.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ip,
	})
	if err != nil {
		s.log.Error(err.Error())
	}
}
