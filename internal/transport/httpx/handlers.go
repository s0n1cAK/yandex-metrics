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

func SetMetricURL(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := BindMetricFromURL(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if err := svc.Set(r.Context(), m); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func SetMetricJSON(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := BindMetricFromJSON(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if err := svc.Set(r.Context(), m); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(m)
	}
}

func SetBatchMetrics(svc metrics.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		batch, err := BindBatchFromJSON(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if err := svc.SetBatch(r.Context(), batch); err != nil {
			WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

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
