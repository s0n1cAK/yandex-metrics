package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
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

		// Подумать о переходе на resty, но для начала узначать в чем выгода
		request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		request.Close = true
		request.Header.Set("Content-Type", "application/json")

		agent.Logger.Info("Sending metric", zap.String("id", metric.ID))
		response, err := agent.Client.Do(request)
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: bad status: %s", OP, response.Status)
		}
		defer response.Body.Close()
	}

	agent.Storage.Clear()
	return nil
}
