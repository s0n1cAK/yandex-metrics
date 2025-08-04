package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/s0n1cAK/yandex-metrics/internal/config/db"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
	"go.uber.org/zap"
)

func SetMetricURL(s storage.BasicStorage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rMetric := models.Metrics{
			ID:    chi.URLParam(r, "metric"),
			MType: chi.URLParam(r, "type"),
		}

		if rMetric.ID == "" {
			logger.Warn("отсутствует ID метрики", zap.String("uri", r.RequestURI))
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		param := chi.URLParam(r, "value")
		switch rMetric.MType {
		case models.Gauge:
			if param == "" {
				logger.Warn("не передано значение для gauge", zap.String("id", rMetric.ID))
				http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
				return
			}
			val, err := strconv.ParseFloat(param, 64)
			if err != nil {
				logger.Error("ошибка преобразования gauge", zap.String("param", param), zap.Error(err))
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rMetric.Value = &val
		case models.Counter:
			if param == "" {
				logger.Warn("не передан delta для counter", zap.String("id", rMetric.ID))
				http.Error(w, "Missing 'delta' parameter", http.StatusBadRequest)
				return
			}
			val, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				logger.Error("ошибка преобразования counter", zap.String("param", param), zap.Error(err))
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if val == 0 {
				logger.Warn("counter == 0", zap.String("id", rMetric.ID))
				http.Error(w, "Counter can't be zero", http.StatusBadRequest)
				return
			}
			rMetric.Delta = &val
		default:
			logger.Warn("неподдерживаемый тип метрики", zap.String("type", rMetric.MType))
			http.Error(w, "Unsupported type of metrics", http.StatusBadRequest)
			return
		}

		err := s.Set(rMetric.ID, rMetric)
		if err != nil {
			logger.Error("ошибка сохранения метрики", zap.String("id", rMetric.ID), zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger.Info("метрика установлена", zap.String("id", rMetric.ID), zap.String("type", rMetric.MType))
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func SetMetricJSON(s storage.BasicStorage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("ошибка чтения тела запроса", zap.Error(err))
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}
		var rMetric models.Metrics
		if err = json.Unmarshal(bodyBytes, &rMetric); err != nil {
			logger.Error("ошибка парсинга JSON", zap.ByteString("body", bodyBytes), zap.Error(err))
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		if rMetric.ID == "" {
			logger.Warn("пустой ID метрики", zap.Any("metric", rMetric))
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		switch rMetric.MType {
		case models.Gauge:
			if rMetric.Value == nil {
				logger.Warn("gauge: значение nil", zap.String("id", rMetric.ID))
				http.Error(w, "Gauge can't be nil", http.StatusBadRequest)
				return
			}
		case models.Counter:
			if rMetric.Delta == nil {
				logger.Warn("counter: значение nil", zap.String("id", rMetric.ID))
				http.Error(w, "Counter can't be nil", http.StatusBadRequest)
				return
			}
		default:
			logger.Warn("неподдерживаемый тип", zap.String("type", rMetric.MType))
			http.Error(w, "Unsupported type of metrics", http.StatusBadRequest)
			return
		}

		err = s.Set(rMetric.ID, rMetric)
		if err != nil {
			logger.Error("ошибка записи метрики", zap.String("id", rMetric.ID), zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(rMetric); err != nil {
			logger.Error("ошибка сериализации ответа", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		logger.Info("метрика установлена (JSON)", zap.String("id", rMetric.ID), zap.String("type", rMetric.MType))
	}
}

func GetMetricJSON(s storage.BasicStorage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("ошибка чтения тела запроса", zap.Error(err))
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		var rMetric models.Metrics
		if err = json.Unmarshal(bodyBytes, &rMetric); err != nil {
			logger.Error("ошибка парсинга JSON", zap.ByteString("body", bodyBytes), zap.Error(err))
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		if rMetric.ID == "" {
			logger.Warn("пустой ID", zap.Any("metric", rMetric))
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		switch rMetric.MType {
		case models.Gauge, models.Counter:
		default:
			logger.Warn("неподдерживаемый тип", zap.String("type", rMetric.MType))
			http.Error(w, "Unsupported type of metrics", http.StatusBadRequest)
			return
		}

		payloadMetric, ok := s.Get(rMetric.ID)
		if !ok {
			logger.Warn("метрика не найдена", zap.String("id", rMetric.ID))
			http.Error(w, "Metric doesn't exist", http.StatusNotFound)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(payloadMetric); err != nil {
			logger.Error("ошибка сериализации ответа", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func GetMetric(s storage.BasicStorage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rMetric := models.Metrics{
			ID:    chi.URLParam(r, "metric"),
			MType: chi.URLParam(r, "type"),
		}

		if rMetric.ID == "" || rMetric.MType == "" {
			logger.Warn("пустой ID или тип метрики", zap.String("id", rMetric.ID), zap.String("type", rMetric.MType))
			http.Error(w, "Metric name or type not specified", http.StatusBadRequest)
			return
		}

		value, ok := s.Get(rMetric.ID)
		if !ok {
			logger.Warn("метрика не найдена", zap.String("id", rMetric.ID))
			http.Error(w, "Metric hasn't found", http.StatusNotFound)
			return
		}

		if value.MType != rMetric.MType {
			logger.Warn("тип не совпадает", zap.String("expected", value.MType), zap.String("got", rMetric.MType))
			http.Error(w, "Metric with that type hasn't found", http.StatusNotFound)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		switch value.MType {
		case models.Counter:
			w.Write([]byte(strconv.FormatInt(*value.Delta, 10)))
		case models.Gauge:
			w.Write([]byte(strconv.FormatFloat(*value.Value, 'f', -1, 64)))
		}
	}
}

func GetMetrics(s storage.BasicStorage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sMetrics, err := s.GetAll()
		if err != nil {
			logger.Error("ошибка получения всех метрик", zap.Error(err))
			http.Error(w, "Ошибка чтения метрик", http.StatusInternalServerError)
			return
		}

		var metrics []string
		for _, metirc := range sMetrics {
			metrics = append(metrics, metirc.ID)
		}

		payload, err := json.Marshal(metrics)
		if err != nil {
			logger.Error("ошибка сериализации списка метрик", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}
}

func PingDB(DSN string, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		err := db.PingDB(ctx, DSN)
		w.Header().Add("Content-Type", "text/plain")

		if err != nil {
			logger.Error("база данных недоступна", zap.Error(err))
			http.Error(w, "Fail to check DB", http.StatusInternalServerError)
			return
		}

		logger.Info("база данных доступна")
		w.WriteHeader(http.StatusOK)
	}
}

func SetBatchMetrics(s storage.BasicStorage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("ошибка чтения тела запроса", zap.Error(err))
			http.Error(w, "Metrics can't be parsed", http.StatusBadRequest)
			return
		}

		if len(bodyBytes) == 0 {
			logger.Warn("тело запроса пустое")
			http.Error(w, "No metrics", http.StatusBadRequest)
			return
		}

		var rMetrics []models.Metrics
		if err = json.Unmarshal(bodyBytes, &rMetrics); err != nil {
			logger.Error("ошибка парсинга JSON", zap.ByteString("body", bodyBytes), zap.Error(err))
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		if err = s.SetAll(rMetrics); err != nil {
			logger.Error("ошибка сохранения метрик", zap.Error(err))
			http.Error(w, "Error while sending metrics", http.StatusInternalServerError)
			return
		}

		logger.Info("метрики сохранены пакетом", zap.Int("count", len(rMetrics)))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
