package metrics

import (
	"context"
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/audit"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

type mockPinger struct{}

func (m *mockPinger) Ping(ctx context.Context) error {
	return nil
}

func BenchmarkServiceSet(b *testing.B) {
	logger, _ := zap.NewProduction()
	repo := memstorage.New()
	publisher := &audit.AuditPublisher{}
	service := New(repo, &mockPinger{}, logger, *publisher)

	ctx := context.Background()
	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: func() *float64 { v := 100.0; return &v }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.Set(ctx, metric, "127.0.0.1")
	}
}

func BenchmarkServiceSetBatch(b *testing.B) {
	logger, _ := zap.NewProduction()
	repo := memstorage.New()
	publisher := &audit.AuditPublisher{}
	service := New(repo, &mockPinger{}, logger, *publisher)

	ctx := context.Background()
	batch := make([]models.Metrics, 100)
	for i := range batch {
		value := float64(i)
		batch[i] = models.Metrics{
			ID:    "test_metric_" + string(rune(i+'0')),
			MType: models.Gauge,
			Value: &value,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.SetBatch(ctx, batch, "127.0.0.1")
	}
}

func BenchmarkServiceGet(b *testing.B) {
	logger, _ := zap.NewProduction()
	repo := memstorage.New()
	publisher := &audit.AuditPublisher{}
	service := New(repo, &mockPinger{}, logger, *publisher)

	ctx := context.Background()
	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: func() *float64 { v := 100.0; return &v }(),
	}
	_ = service.Set(ctx, metric, "127.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Get(ctx, "test_metric", models.Gauge)
	}
}

func BenchmarkServiceListIDs(b *testing.B) {
	logger, _ := zap.NewProduction()
	repo := memstorage.New()
	publisher := &audit.AuditPublisher{}
	service := New(repo, &mockPinger{}, logger, *publisher)

	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		value := float64(i)
		metric := models.Metrics{
			ID:    "test_metric_" + string(rune(i+'0')),
			MType: models.Gauge,
			Value: &value,
		}
		_ = service.Set(ctx, metric, "127.0.0.1")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ListIDs(ctx)
	}
}

func BenchmarkMemStorageSet(b *testing.B) {
	storage := memstorage.New()
	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: func() *float64 { v := 100.0; return &v }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.Set("test_metric", metric)
	}
}

func BenchmarkMemStorageGet(b *testing.B) {
	storage := memstorage.New()
	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: func() *float64 { v := 100.0; return &v }(),
	}
	_ = storage.Set("test_metric", metric)

	for b.Loop() {
		_, _ = storage.Get("test_metric")
	}
}

func BenchmarkMemStorageGetAll(b *testing.B) {
	storage := memstorage.New()
	for i := 0; i < 1000; i++ {
		value := float64(i)
		metric := models.Metrics{
			ID:    "test_metric_" + string(rune(i+'0')),
			MType: models.Gauge,
			Value: &value,
		}
		_ = storage.Set("test_metric_"+string(rune(i+'0')), metric)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetAll()
	}
}
