package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sconfig "github.com/s0n1cAK/yandex-metrics/internal/config/server"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/s0n1cAK/yandex-metrics/internal/server"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"go.uber.org/zap"
)

func Example() {
	// Создаем сервер с in-memory хранилищем для демонстрации
	storage := memstorage.New()

	// Создаем логгер
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Создаем конфигурацию сервера
	cfg := &sconfig.Config{
		Endpoint:      customtype.Endpoint("http://localhost:8080"),
		StoreInterval: customtype.Time(0),
		File:          "/tmp/test_metrics.data",
		Restore:       false,
		DSN:           customtype.DSN{},
		HashKey:       "",
		AuditFile:     "",
		AuditURL:      "",
		Logger:        logger,
	}

	// Создаем экземпляр сервера
	srv, err := server.New(cfg, storage)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// Создаем контекст с таймаутом для сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Запускаем сервер в горутине
	go func() {
		if err := srv.Start(ctx); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Даем серверу время для запуска
	time.Sleep(100 * time.Millisecond)

	// Обновление счетчика с помощью JSON
	counterMetric := model.Metrics{
		ID:    "request_count",
		MType: model.Counter,
		Delta: func() *int64 { v := int64(1); return &v }(),
	}

	jsonData, _ := json.Marshal(counterMetric)

	resp, err := http.Post("http://localhost:8080/update", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	} else {
		fmt.Printf("Counter update response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Обновление gauge с помощью URL-параметров
	resp, err = http.Post("http://localhost:8080/update/gauge/cpu_usage/0.75", "", nil)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	} else {
		fmt.Printf("Gauge update response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Групповое обновление нескольких метрик
	batch := []model.Metrics{
		{
			ID:    "batch_counter",
			MType: model.Counter,
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
		{
			ID:    "batch_gauge",
			MType: model.Gauge,
			Value: func() *float64 { v := 3.14; return &v }(),
		},
	}

	batchData, _ := json.Marshal(batch)
	resp, err = http.Post("http://localhost:8080/updates", "application/json", bytes.NewBuffer(batchData))
	if err != nil {
		fmt.Printf("Error sending batch request: %v\n", err)
	} else {
		fmt.Printf("Batch update response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Получение значения метрики с помощью JSON
	getMetric := model.Metrics{
		ID:    "request_count",
		MType: model.Counter,
	}

	jsonData, _ = json.Marshal(getMetric)
	resp, err = http.Post("http://localhost:8080/value", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error getting metric: %v\n", err)
	} else {
		fmt.Printf("Get metric response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Получение значения метрики с помощью URL
	resp, err = http.Get("http://localhost:8080/value/counter/request_count")
	if err != nil {
		fmt.Printf("Error getting metric by URL: %v\n", err)
	} else {
		fmt.Printf("Get metric by URL response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Получение всех метрик
	resp, err = http.Get("http://localhost:8080/")
	if err != nil {
		fmt.Printf("Error getting all metrics: %v\n", err)
	} else {
		fmt.Printf("Get all metrics response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Проверка сервера
	resp, err = http.Get("http://localhost:8080/ping")
	if err != nil {
		fmt.Printf("Error pinging server: %v\n", err)
	} else {
		fmt.Printf("Ping response status: %s\n", resp.Status)
		resp.Body.Close()
	}

	cancel()
}
