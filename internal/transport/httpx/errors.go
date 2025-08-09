package httpx

import (
	"errors"
	"net/http"

	"github.com/s0n1cAK/yandex-metrics/internal/domain"
)

func WriteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidPayload), errors.Is(err, domain.ErrInvalidType):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrZeroCounter):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
