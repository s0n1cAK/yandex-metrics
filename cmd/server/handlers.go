package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

/*
Принимать метрики по протоколу HTTP методом POST. ✓
Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>, Content-Type: text/plain. ✓
При успешном приёме возвращать http.StatusOK. ✓
При попытке передать запрос без имени метрики возвращать http.StatusNotFound. ✓
При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest. ✓
*/

func (s *MemStorage) SetHandler(w http.ResponseWriter, r *http.Request) {
	rMetric := Metric{ID: chi.URLParam(r, "metric"), MType: chi.URLParam(r, "type")}

	if rMetric.ID == "" {
		http.Error(w, "Metric name not specified", http.StatusNotFound)
		return
	}

	switch rMetric.MType {

	case "gauge":
		val, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rMetric.Value = val

	case "counter":
		val, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if val == 0 {
			http.Error(w, "Counter can't be zero", http.StatusBadRequest)
			return
		}

		rMetric.Delta = val

	default:
		http.Error(w, "Insupported type of metrics", http.StatusBadRequest)
		return
	}

	oldMetric := s.values[rMetric.ID]

	var newDelta int64
	if rMetric.Delta != 0 {
		newDelta = oldMetric.Delta + rMetric.Delta
	}

	updatedMetric := Metric{
		ID:    rMetric.ID,
		MType: rMetric.MType,
		Value: rMetric.Value,
		Delta: newDelta,
	}

	s.values[rMetric.ID] = updatedMetric

	fmt.Println(rMetric)
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
