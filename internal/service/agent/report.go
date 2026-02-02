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
	"github.com/s0n1cAK/yandex-metrics/internal/crypt"
	"github.com/s0n1cAK/yandex-metrics/internal/hash"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

func (agent *Agent) Report() error {
	op := "Agent.Report"

	stotageMetrics, err := agent.Storage.GetAll()
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	metrics := make([]models.Metrics, 0, len(stotageMetrics))

	for _, metric := range stotageMetrics {
		metrics = append(metrics, metric)
	}

	endpoint := fmt.Sprintf("%s/updates", agent.Server)

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	hash := hash.GetHashHex(payload, agent.hash)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(payload); err != nil {
		return fmt.Errorf("%s: gzip write: %w", op, err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("%s: gzip close: %w", op, err)
	}

	out := buf.Bytes()
	encrypted := false
	if agent.publicKey != nil {
		enc, err := crypt.EncryptHybrid(agent.publicKey, out)
		if err != nil {
			return fmt.Errorf("%s: encrypt payload: %w", op, err)
		}
		out = enc
		encrypted = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(out))
	if err != nil {
		return fmt.Errorf("%s: new request: %w", op, err)
	}

	request.Close = true
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("HashSHA256", hash)
	if encrypted {
		request.Header.Set("X-Encrypted", "1")
	}

	response, err := agent.requestWithLimit(ctx, request)
	if err != nil {
		return fmt.Errorf("%s: do request: %w", op, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("%s: bad status: %s; body: %s", op, response.Status, string(body))
	}

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
