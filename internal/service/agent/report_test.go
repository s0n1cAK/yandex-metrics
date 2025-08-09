package agent

// До 1:1
// httptest для client request
// В ближайщем будущем

// func (agent *Config) Report() error {
// 	var endpoint string

// 	for _, metric := range agent.Storage.Get() {
// 		switch metric.MType {
// 		case models.Gauge:
// 			endpoint = fmt.Sprintf("%s/update/%s/%s/%v", agent.Server, metric.MType, metric.ID, *metric.Value)
// 		case models.Counter:
// 			endpoint = fmt.Sprintf("%s/update/%s/%s/%v", agent.Server, metric.MType, metric.ID, *metric.Delta)
// 		default:
// 			return fmt.Errorf("Unknown type %s", metric.MType)
// 		}

// 		request, err := http.NewRequest(http.MethodPost, endpoint, nil)
// 		if err != nil {
// 			return err
// 		}

// 		request.Close = true
// 		request.Header.Set("Content-Type", "text/plain")

// 		response, err := agent.Client.Do(request)
// 		if err != nil {
// 			return err
// 		}
// 		if response.StatusCode != http.StatusOK {
// 			return fmt.Errorf("bad status: %s", response.Status)
// 		}
// 		response.Body.Close()
// 	}

// 	agent.Storage.Clear()
// 	return nil
// }
