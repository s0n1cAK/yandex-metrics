package audit

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/s0n1cAK/yandex-metrics/internal/model"
)

type HTTPAuditObserver struct {
	url string
}

func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{url: url}
}

func (h *HTTPAuditObserver) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
