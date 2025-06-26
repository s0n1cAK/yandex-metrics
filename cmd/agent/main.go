package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/agent"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
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
	metricsStorage := memstorage.New()
	endpoint := fmt.Sprintf("http://%s:%s", serverAddr, serverPort)

	agent := agent.New(&http.Client{}, endpoint, metricsStorage, time.Now())

	for {
		// Отправляем метрики
		if time.Since(agent.LastReportTime) >= reportInterval {
			log.Printf("Reporting metrics")
			err := agent.Report()
			if err != nil {
				log.Printf("Error while reporting: %s", err)
			}
			agent.LastReportTime = time.Now()
		}
		time.Sleep(pollInterval)
	}
}
