package metrics

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/s0n1cAK/yandex-metrics/internal/domain"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

type Repository interface {
	Set(id string, m models.Metrics) error
	Get(id string) (models.Metrics, bool)
	GetAll() (map[string]models.Metrics, error)
	SetAll(batch []models.Metrics) error
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type Service interface {
	Set(ctx context.Context, m models.Metrics) error
	Get(ctx context.Context, id, mtype string) (models.Metrics, error)
	ListIDs(ctx context.Context) ([]string, error)
	SetBatch(ctx context.Context, batch []models.Metrics) error
	Ping(ctx context.Context) error
}

type service struct {
	repo Repository
	ping Pinger
	log  *zap.Logger
}

func New(repo Repository, ping Pinger, log *zap.Logger) Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &service{repo: repo, ping: ping, log: log}
}

func (s *service) Set(ctx context.Context, m models.Metrics) error {
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
		if *m.Delta == 0 {
			return domain.ErrZeroCounter
		}
	default:
		return domain.ErrInvalidType
	}

	if err := s.repo.Set(m.ID, m); err != nil {
		return err
	}
	s.log.Info("metric set", zap.String("id", m.ID), zap.String("type", m.MType))
	return nil
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

func (s *service) SetBatch(ctx context.Context, batch []models.Metrics) error {
	if len(batch) == 0 {
		return domain.ErrInvalidPayload
	}
	for _, m := range batch {
		switch m.MType {
		case models.Gauge:
			if m.ID == "" || m.Value == nil {
				return domain.ErrInvalidPayload
			}
		case models.Counter:
			if m.ID == "" || m.Delta == nil || *m.Delta == 0 {
				return domain.ErrInvalidPayload
			}
		default:
			return domain.ErrInvalidType
		}
	}
	return s.repo.SetAll(batch)
}

func (s *service) Ping(ctx context.Context) error {
	if s.ping == nil {
		return errors.New("no pinger configured")
	}
	return s.ping.Ping(ctx)
}
