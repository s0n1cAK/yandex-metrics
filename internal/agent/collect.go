package agent

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
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

// Вылядит не оч, как будто бы есть явно способ лучше и эффективнее, но я его пока не нашел
func (agent *Config) CollectRuntime() error {
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

		agent.Storage.Set(metirc, models.Metrics{
			ID:    metirc,
			MType: models.Gauge,
			Value: &metircToReport,
		})
	}
	return nil
}

func (agent *Config) RandomValue() {

	randFloat := rand.Float64()
	agent.Storage.Set(MetricNameRandomValue, models.Metrics{
		ID:    MetricNameRandomValue,
		MType: models.Gauge,
		Value: lib.FloatPtr(randFloat),
	})
}

func (agent *Config) IncrementCounter(ID string, value int64) {
	metric, ok := agent.Storage.Get(ID)

	var newDelta int64
	if ok && metric.MType == models.Counter && metric.Delta != nil {
		newDelta = *metric.Delta + value
	} else {
		newDelta = value
	}

	agent.Storage.Set(ID, models.Metrics{
		ID:    ID,
		MType: models.Counter,
		Delta: &newDelta,
	})
}
