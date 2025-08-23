package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"time"

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

// В будущем переписать на структуры с нужными полями, которая будет заполняться, из-за того что reflect
// тяжелый и медленный пакет
func (agent *Agent) CollectRuntime() error {
	OP := "agent.CollectRuntime"

	var g errgroup.Group

	for _, metric := range runtimeMetrics {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		val := reflect.ValueOf(ms)

		g.Go(func() error {
			field := val.FieldByName(metric)

			var metricToReport float64
			switch field.Kind() {
			case reflect.Uint64:
				metricToReport = float64(field.Uint())
			case reflect.Uint32:
				metricToReport = float64(field.Uint())
			case reflect.Float64:
				metricToReport = field.Float()
			default:
				return fmt.Errorf("%s: unsupported metric type: %s", OP, field.Kind())
			}

			err := agent.updateGaugeMetruc(metric, metricToReport)
			if err != nil {
				return fmt.Errorf("%s: Error: %s", OP, err)
			}

			return nil
		})

	}
	return g.Wait()
}

func (agent *Agent) CollectRandomValue() error {
	OP := "agent.CollectRandomValue"

	randFloat := rand.Float64()

	err := agent.updateGaugeMetruc(MetricNameRandomValue, randFloat)
	if err != nil {
		return fmt.Errorf("%s: Error: %s", OP, err)
	}
	return nil
}

func (agent *Agent) CollectIncrementCounter(ID string, value int64) error {
	OP := "agent.CollectIncrementCounter"

	err := agent.updateCounterMetruc(ID, value)
	if err != nil {
		return fmt.Errorf("%s: Error: %s", OP, err)
	}

	return nil
}

func (agent *Agent) CollectGopsutil(ctx context.Context, errs chan<- error) {
	OP := "agent.CollectGopsutil"

	v, err := mem.VirtualMemory()
	if err != nil {
		errs <- fmt.Errorf("%s: Error: %s", OP, err)
	}

	totalMemoryValue := float64(v.Total)
	freeMemoryValue := float64(v.Free)
	usePersentValue := float64(v.UsedPercent)

	err = agent.updateGaugeMetruc("TotalMemory", totalMemoryValue)
	if err != nil {
		errs <- fmt.Errorf("%s: Error: %s", OP, err)
		return
	}
	err = agent.updateGaugeMetruc("FreeMemory", freeMemoryValue)
	if err != nil {
		errs <- fmt.Errorf("%s: Error: %s", OP, err)
		return
	}
	err = agent.updateGaugeMetruc("CPUutilization1", usePersentValue)
	if err != nil {
		errs <- fmt.Errorf("%s: Error: %s", OP, err)
		return
	}
}
