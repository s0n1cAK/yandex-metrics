package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/s0n1cAK/yandex-metrics/internal/storage"
)

/*
Принимать метрики по протоколу HTTP методом POST. ✓
Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>, Content-Type: text/plain. ✓
При успешном приёме возвращать http.StatusOK. ✓
При попытке передать запрос без имени метрики возвращать http.StatusNotFound. ✓
При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest. ✓
*/

func SetMetric(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rMetric := models.Metrics{
			ID:    chi.URLParam(r, "metric"),
			MType: chi.URLParam(r, "type"),
		}

		if rMetric.ID == "" {
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		switch rMetric.MType {
		case models.Gauge:
			param := chi.URLParam(r, "value")
			if param == "" {
				http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
				return
			}

			val, err := strconv.ParseFloat(param, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rMetric.Value = &val
		case models.Counter:
			param := chi.URLParam(r, "value")
			if param == "" {
				http.Error(w, "Missing 'delta' parameter", http.StatusBadRequest)
				return
			}
			val, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if val == 0 {
				http.Error(w, "Counter can't be zero", http.StatusBadRequest)
				return
			}

			rMetric.Delta = &val
		default:
			http.Error(w, "Insupported type of metrics", http.StatusBadRequest)
			return
		}

		oldMetric, _ := s.Get(rMetric.ID)

		if rMetric.MType == models.Counter {
			var newDelta int64
			if oldMetric.Delta != nil {
				newDelta = *oldMetric.Delta + *rMetric.Delta
			} else {
				newDelta = *rMetric.Delta
			}
			rMetric.Delta = &newDelta
		}

		s.Set(rMetric.ID, rMetric)

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetric(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rMetric := models.Metrics{
			ID:    chi.URLParam(r, "metric"),
			MType: chi.URLParam(r, "type"),
		}

		if rMetric.ID == "" || rMetric.MType == "" {
			http.Error(w, "Metric name or type not specified", http.StatusNotFound)
			return
		}

		value, ok := s.Get(rMetric.ID)
		if !ok {
			http.Error(w, "Metric hasn't found", http.StatusNotFound)
			return
		}

		if value.MType != rMetric.MType {
			http.Error(w, "Metric with that type hasn't found", http.StatusNotFound)
			return
		}

		payload, err := json.Marshal(value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(string(payload)))

	}
}

func GetMetrics(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []string

		for _, metirc := range s.GetAll() {
			metrics = append(metrics, metirc.ID)
		}
		payload, err := json.Marshal(metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(string(payload)))
	}
}
