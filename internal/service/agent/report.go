package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/s0n1cAK/yandex-metrics/internal/hash"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

func (agent *Agent) Report() error {
	OP := "Agent.Report"

	stotageMetrics, err := agent.Storage.GetAll()
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	metrics := make([]models.Metrics, 0, len(stotageMetrics))

	for _, metric := range stotageMetrics {
		metrics = append(metrics, metric)
	}

	endpoint := fmt.Sprintf("%s/updates", agent.Server)

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	hash := hash.GetHashHex(payload, agent.Hash)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err = gz.Write(payload)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	if err := gz.Close(); err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	request.Close = true

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("HashSHA256", hash)

	response, err := agent.requestWithLimit(ctx, request)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		return fmt.Errorf("%s: bad status: %s; body: %s", OP, response.Status, string(body))
	}
	response.Body.Close()

	for _, metric := range stotageMetrics {
		if metric.MType == models.Gauge {
			agent.Storage.Delete(metric.ID)
		}
	}

	return nil
}

func (agent *Agent) requestWithLimit(ctx context.Context, req *retryablehttp.Request) (*http.Response, error) {
	select {
	case agent.httpLimiter <- struct{}{}:
		defer func() {
			<-agent.httpLimiter
		}()
		return agent.Client.Do(req)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
