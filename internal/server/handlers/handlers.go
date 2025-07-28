package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

type Storage interface {
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() map[string]models.Metrics
}

func SetMetricURL(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rMetric := models.Metrics{
			ID:    chi.URLParam(r, "metric"),
			MType: chi.URLParam(r, "type"),
		}

		if rMetric.ID == "" {
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		param := chi.URLParam(r, "value")
		switch rMetric.MType {
		case models.Gauge:

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

		err := s.Set(rMetric.ID, rMetric)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func SetMetricJSON(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}
		var rMetric models.Metrics
		if err = json.Unmarshal(bodyBytes, &rMetric); err != nil {
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		if rMetric.ID == "" {
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		switch rMetric.MType {
		case models.Gauge:
			if rMetric.Value == nil {
				http.Error(w, "Gauge can't be nil", http.StatusBadRequest)
				return
			}
		case models.Counter:
			if rMetric.Delta == nil {
				http.Error(w, "Counter can't be nil", http.StatusBadRequest)
				return
			}
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

		err = s.Set(rMetric.ID, rMetric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(rMetric); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func GetMetricJSON(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		var rMetric models.Metrics
		if err = json.Unmarshal(bodyBytes, &rMetric); err != nil {
			http.Error(w, "Metric can't be parsed", http.StatusBadRequest)
			return
		}

		if rMetric.ID == "" {
			http.Error(w, "Metric name not specified", http.StatusNotFound)
			return
		}

		switch rMetric.MType {
		case models.Gauge:
		case models.Counter:
		default:
			http.Error(w, "Insupported type of metrics", http.StatusBadRequest)
			return
		}

		payloadMetric, ok := s.Get(rMetric.ID)
		if !ok {
			http.Error(w, "Metric doesn't exist", http.StatusNotFound)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(payloadMetric); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func GetMetric(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rMetric := models.Metrics{
			ID:    chi.URLParam(r, "metric"),
			MType: chi.URLParam(r, "type"),
		}

		if rMetric.ID == "" || rMetric.MType == "" {
			http.Error(w, "Metric name or type not specified", http.StatusBadRequest)
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

func GetMetrics(s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []string

		for _, metirc := range s.GetAll() {
			metrics = append(metrics, metirc.ID)
		}

		// Говорим код ошибки, но без текста
		payload, err := json.Marshal(metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(string(payload)))
	}
}
