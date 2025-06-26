package agent

import (
	"fmt"
	"log"
	"net/http"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

func (agent *Config) Report() error {
	OP := "Agent.Report"
	var endpoint string

	err := agent.CollectRuntime()
	if err != nil {
		log.Printf("Error while reporting: %s", err)
	}
	agent.RandomValue()
	agent.IncrementCounter("PollCount", 1)

	for _, metric := range agent.Storage.GetAll() {
		switch metric.MType {
		case models.Gauge:
			endpoint = fmt.Sprintf("%s/update/%s/%s/%v", agent.Server, metric.MType, metric.ID, *metric.Value)
		case models.Counter:
			endpoint = fmt.Sprintf("%s/update/%s/%s/%v", agent.Server, metric.MType, metric.ID, *metric.Delta)
		default:
			return fmt.Errorf("Unknown type %s", metric.MType)
		}

		request, err := http.NewRequest(http.MethodPost, endpoint, nil)
		if err != nil {
			return fmt.Errorf("%s: %s", OP, err)
		}

		request.Close = true
		request.Header.Set("Content-Type", "text/plain")

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
