package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

func (agent *Agent) Report() error {
	OP := "Agent.Report"

	stotageMetrics := agent.Storage.GetAll()

	metrics := make([]models.Metrics, 0, len(stotageMetrics))

	for _, metric := range stotageMetrics {
		metrics = append(metrics, metric)
	}

	endpoint := fmt.Sprintf("%s/updates", agent.Server)

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err = gz.Write(payload)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	if err := gz.Close(); err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	fmt.Println(string(payload[:]))
	// Подумать о переходе на resty, но для начала узначать в чем выгода
	request, err := http.NewRequest(http.MethodPost, endpoint, &buf)
	if err != nil {
		return fmt.Errorf("%s: %s", OP, err)
	}

	request.Close = true
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	response, err := agent.Client.Do(request)
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
