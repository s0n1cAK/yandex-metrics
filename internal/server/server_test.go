package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/config/server"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

func BenchmarkServerSetMetricURL(b *testing.B) {
	logger, _ := zap.NewProduction()
	cfg := &server.Config{
		Endpoint: customtype.Endpoint("http://localhost:8080"),
		Logger:   logger,
		File:     "/tmp/test.data",
	}
	storage := memstorage.New()
	srv, err := New(cfg, storage)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/100.0", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.Router.ServeHTTP(w, req)
	}
}

func BenchmarkServerSetMetricJSON(b *testing.B) {
	logger, _ := zap.NewProduction()
	cfg := &server.Config{
		Endpoint: customtype.Endpoint("http://localhost:8080"),
		Logger:   logger,
		File:     "/tmp/test.data",
	}
	storage := memstorage.New()
	srv, err := New(cfg, storage)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	jsonData := `{"id":"test_metric","type":"gauge","value":100.0}`
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		srv.Router.ServeHTTP(w, req)
	}
}

func BenchmarkServerGetMetric(b *testing.B) {
	logger, _ := zap.NewProduction()
	cfg := &server.Config{
		Endpoint: customtype.Endpoint("http://localhost:8080"),
		Logger:   logger,
		File:     "/tmp/test.data",
	}
	storage := memstorage.New()
	srv, err := New(cfg, storage)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/100.0", nil)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodGet, "/value/gauge/test_metric", nil)
	w = httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.Router.ServeHTTP(w, req)
	}
}

func BenchmarkServerGetAllMetrics(b *testing.B) {
	logger, _ := zap.NewProduction()
	cfg := &server.Config{
		Endpoint: customtype.Endpoint("http://localhost:8080"),
		Logger:   logger,
		File:     "/tmp/test.data",
	}
	storage := memstorage.New()
	srv, err := New(cfg, storage)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/gauge/test_metric_%d/%d.0", i, i), nil)
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.Router.ServeHTTP(w, req)
	}
}

func BenchmarkServerSetBatchMetrics(b *testing.B) {
	logger, _ := zap.NewProduction()
	cfg := &server.Config{
		Endpoint: customtype.Endpoint("http://localhost:8080"),
		Logger:   logger,
		File:     "/tmp/test.data",
	}
	storage := memstorage.New()
	srv, err := New(cfg, storage)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	jsonData := `[{"id":"test_metric_1","type":"gauge","value":100.0},{"id":"test_metric_2","type":"counter","delta":1}]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, req)
	}
}
