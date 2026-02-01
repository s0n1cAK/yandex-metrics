package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/s0n1cAK/yandex-metrics/internal/domain"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/s0n1cAK/yandex-metrics/internal/service/metrics"
)

// SetMetricURL возвращает HTTP-обработчик для обновления метрики через URL-параметры.
// Принимает тип метрики, имя и значение в URL и устанавливает новое значение метрики.
// Пример: POST /update/counter/requests/10
func SetMetricURL(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		m, err := BindMetricFromURL(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if err := svc.Set(r.Context(), m, ip); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

// SetMetricJSON возвращает HTTP-обработчик для обновления метрики через JSON-тело запроса.
// Принимает метрику в формате JSON и устанавливает новое значение.
// Пример: POST /update {"id":"requests","type":"counter","delta":1}
func SetMetricJSON(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		m, err := BindMetricFromJSON(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if err := svc.Set(r.Context(), m, ip); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(m)
	}
}

// SetBatchMetrics возвращает HTTP-обработчик для обновления нескольких метрик за один запрос.
// Принимает массив метрик в формате JSON и устанавливает их значения.
// Пример: POST /updates [{"id":"requests","type":"counter","delta":1},{"id":"cpu","type":"gauge","value":0.7}]
func SetBatchMetrics(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		batch, err := BindBatchFromJSON(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if err := svc.SetBatch(r.Context(), batch, ip); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

// GetMetricJSON возвращает HTTP-обработчик для получения метрики в формате JSON.
// Принимает запрашиваемую метрику в формате JSON и возвращает её текущее значение.
// Пример: POST /value {"id":"requests","type":"counter"}
func GetMetricJSON(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := BindMetricFromJSON(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		res, err := svc.Get(r.Context(), m.ID, m.MType)
		if err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

// GetMetric возвращает HTTP-обработчик для получения метрики через URL-параметры.
// Принимает тип и имя метрики в URL и возвращает её текущее значение.
// Пример: GET /value/counter/requests
func GetMetric(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "metric")
		mtype := chi.URLParam(r, "type")
		if id == "" || mtype == "" {
			WriteError(w, domain.ErrInvalidPayload)
			return
		}
		res, err := svc.Get(r.Context(), id, mtype)
		if err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		switch res.MType {
		case models.Counter:
			w.Write([]byte(strconv.FormatInt(*res.Delta, 10)))
		case models.Gauge:
			w.Write([]byte(strconv.FormatFloat(*res.Value, 'f', -1, 64)))
		}
	}
}

// GetMetrics возвращает HTTP-обработчик для получения всех метрик.
// Возвращает список всех зарегистрированных метрик в формате JSON.
func GetMetrics(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids, err := svc.ListIDs(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ids)
	}
}

// Ping возвращает HTTP-обработчик для проверки доступности сервера.
// Проверяет подключение к базе данных (если используется) и возвращает статус.
func Ping(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := svc.Ping(ctx); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}
