package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"go.uber.org/zap"
)

func (agent *Agent) Report() error {
	OP := "Agent.Report"
	for _, metric := range agent.Storage.GetAll() {

		endpoint := fmt.Sprintf("%s/update", agent.Server)
		if metric.MType != models.Gauge && metric.MType != models.Counter {
			return fmt.Errorf("unknown type %s", metric.MType)
		}

		payload, err := json.MarshalIndent(metric, "", "\t")
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

		// Подумать о переходе на resty, но для начала узначать в чем выгода
		request, err := http.NewRequest(http.MethodPost, endpoint, &buf)
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		request.Close = true
		request.Header.Set("Content-Encoding", "gzip")
		request.Header.Set("Content-Type", "application/json")

		if metric.Delta != nil {
			agent.Logger.Info("Sending metric", zap.String("id", metric.ID), zap.Int64("Delta", *metric.Delta))
		}
		if metric.Value != nil {
			agent.Logger.Info("Sending metric", zap.String("id", metric.ID), zap.Float64("Value", *metric.Value))
		}

		response, err := agent.Client.Do(request)
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}
		if err := response.Body.Close(); err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(response.Body)
			return fmt.Errorf("%s: bad status: %s; body: %s", OP, response.Status, string(body))
		}

		if metric.MType == models.Gauge {
			agent.Storage.Delete(metric.ID)
		}
	}
	return nil
}
