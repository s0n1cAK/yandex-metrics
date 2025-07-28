package agent

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

const (
	MetricNameRandomValue = "RandomValue"
)

// Лучше используй типизированные константы
// Подумать как это лучше всего сделать;

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
	for _, metirc := range runtimeMetrics {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		val := reflect.ValueOf(ms)
		field := val.FieldByName(metirc)

		var metircToReport float64
		switch field.Kind() {
		case reflect.Uint64:
			metircToReport = float64(field.Uint())
		case reflect.Uint32:
			metircToReport = float64(field.Uint())
		case reflect.Float64:
			metircToReport = field.Float()
		default:
			return fmt.Errorf("%s: unsupported metric type: %s", OP, field.Kind())
		}

		err := agent.Storage.Set(uniqMetric(metirc), models.Metrics{
			ID:    metirc,
			MType: models.Gauge,
			Value: &metircToReport,
		})
		if err != nil {
			return fmt.Errorf("%s: Error: %s", OP, err)
		}
	}
	return nil
}

func (agent *Agent) CollectRandomValue() error {
	OP := "agent.CollectRandomValue"

	randFloat := rand.Float64()
	err := agent.Storage.Set(uniqMetric(MetricNameRandomValue), models.Metrics{
		ID:    MetricNameRandomValue,
		MType: models.Gauge,
		Value: lib.FloatPtr(randFloat),
	})
	if err != nil {
		return fmt.Errorf("%s: Error: %s", OP, err)
	}
	return nil
}

func (agent *Agent) CollectIncrementCounter(ID string, value int64) error {
	OP := "agent.CollectIncrementCounter"

	metric, ok := agent.Storage.Get(ID)

	var newDelta int64
	if ok && metric.MType == models.Counter && metric.Delta != nil {
		newDelta = *metric.Delta + value
	} else {
		newDelta = value
	}
	fmt.Println(metric, newDelta)
	err := agent.Storage.Set(ID, models.Metrics{
		ID:    ID,
		MType: models.Counter,
		Delta: lib.IntPtr(newDelta),
	})
	if err != nil {
		return fmt.Errorf("%s: Error: %s", OP, err)
	}

	return nil
}
