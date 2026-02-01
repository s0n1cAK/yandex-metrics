package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/s0n1cAK/yandex-metrics/internal/model"
)

type HTTPAuditObserver struct {
	client *retryablehttp.Client
	url    string
}

func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 5 * time.Second
	client.Backoff = retryablehttp.DefaultBackoff

	client.CheckRetry = retryablehttp.DefaultRetryPolicy

	return &HTTPAuditObserver{
		client: client,
		url:    url,
	}
}

func (h *HTTPAuditObserver) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, h.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
