package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/agent"
	agentStorage "github.com/s0n1cAK/yandex-metrics/internal/agent/storage"
)

const (
	serverAddr = "localhost"
	serverPort = "8080"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	metricsStorage := agentStorage.New()
	endpoint := fmt.Sprintf("http://%s:%s", serverAddr, serverPort)

	agent := agent.New(&http.Client{}, endpoint, metricsStorage, time.Now())

	var count int64
	// Не очень нравится это структура. Для её исправления нужно чтобы report подразумевал сам сбор метрик. В принципе звучит норм, но лучше уточню
	for {
		count++

		// Собираем метрики
		err := agent.CollectRuntime()
		if err != nil {
			log.Printf("Error while reporting: %s", err)
		}
		agent.RandomValue()
		agent.Counter(count)

		// Отправляем метрики
		if time.Since(agent.LastReportTime) >= reportInterval {
			log.Printf("Reporting metrics")
			err = agent.Report()
			if err != nil {
				log.Printf("Error while reporting: %s", err)
			}
			agent.LastReportTime = time.Now()
		}
		time.Sleep(pollInterval)
	}
}
