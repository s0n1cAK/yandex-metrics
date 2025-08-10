package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/s0n1cAK/yandex-metrics/internal/domain"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

func BindMetricFromURL(r *http.Request) (models.Metrics, error) {
	id := chi.URLParam(r, "metric")
	mtype := chi.URLParam(r, "type")
	if id == "" || mtype == "" {
		return models.Metrics{}, domain.ErrInvalidPayload
	}

	m := models.Metrics{ID: id, MType: mtype}
	valStr := chi.URLParam(r, "value")

	switch mtype {
	case models.Gauge:
		if valStr == "" {
			return models.Metrics{}, domain.ErrInvalidPayload
		}
		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return models.Metrics{}, domain.ErrInvalidPayload
		}
		m.Value = &f

	case models.Counter:
		if valStr == "" {
			return models.Metrics{}, domain.ErrInvalidPayload
		}
		i, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return models.Metrics{}, domain.ErrInvalidPayload
		}
		m.Delta = &i

	default:
		return models.Metrics{}, domain.ErrInvalidType
	}

	return m, nil
}

func BindMetricFromJSON(r *http.Request) (models.Metrics, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return models.Metrics{}, err
	}
	defer r.Body.Close()

	if len(body) == 0 {
		return models.Metrics{}, domain.ErrInvalidPayload
	}

	var m models.Metrics
	if err := json.Unmarshal(body, &m); err != nil {
		return models.Metrics{}, domain.ErrInvalidPayload
	}
	if m.ID == "" || m.MType == "" {
		return models.Metrics{}, domain.ErrInvalidPayload
	}
	return m, nil
}

func BindBatchFromJSON(r *http.Request) ([]models.Metrics, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if len(body) == 0 {
		return nil, domain.ErrInvalidPayload
	}

	var batch []models.Metrics
	if err := json.Unmarshal(body, &batch); err != nil {
		return nil, domain.ErrInvalidPayload
	}
	if len(batch) == 0 {
		return nil, domain.ErrInvalidPayload
	}
	return batch, nil
}
