package handlers

import (
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

func Set(s storage.Storage) http.HandlerFunc {
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
			param := chi.URLParam(r, "delta")
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
