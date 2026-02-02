package agent

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"time"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/sync/errgroup"
)

const (
	MetricNameRandomValue = "RandomValue"
)

var runtimeMetrics = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

func uniqMetric(m string) string {
	return fmt.Sprintf("%s-%v", m, time.Now().UnixNano())
}

func (agent *Agent) CollectRuntime() error {
	op := "agent.CollectRuntime"

	var g errgroup.Group

	for _, metric := range runtimeMetrics {
		metric := metric

		g.Go(func() error {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			val := reflect.ValueOf(ms)

			field := val.FieldByName(metric)
			if !field.IsValid() {
				return fmt.Errorf("%s: metric %s not found", op, metric)
			}

			var metricToReport float64
			switch field.Kind() {
			case reflect.Uint64, reflect.Uint32:
				metricToReport = float64(field.Uint())
			case reflect.Float64:
				metricToReport = field.Float()
			default:
				return fmt.Errorf("%s: unsupported metric type for %s: %s", op, metric, field.Kind())
			}

			if err := agent.updateGaugeMetruc(metric, metricToReport); err != nil {
				return fmt.Errorf("%s: error: %w", op, err)
			}

			return nil
		})
	}

	return g.Wait()
}

func (agent *Agent) CollectRandomValue() error {
	op := "agent.CollectRandomValue"

	randFloat := rand.Float64()

	if err := agent.updateGaugeMetruc(MetricNameRandomValue, randFloat); err != nil {
		return fmt.Errorf("%s: error: %w", op, err)
	}
	return nil
}

func (agent *Agent) CollectIncrementCounter(id string, value int64) error {
	return agent.Storage.Set(id, models.Metrics{
		ID:    id,
		MType: models.Counter,
		Delta: &value,
	})
}

func (agent *Agent) CollectGopsutil() error {
	op := "agent.CollectGopsutil"

	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("%s: error: %w", op, err)
	}

	metrics := map[string]float64{
		"TotalMemory":     float64(v.Total),
		"FreeMemory":      float64(v.Free),
		"CPUutilization1": float64(v.UsedPercent),
	}

	for name, value := range metrics {
		if err := agent.updateGaugeMetruc(name, value); err != nil {
			return fmt.Errorf("%s: error updating %s: %w", op, name, err)
		}
	}

	return nil
}
