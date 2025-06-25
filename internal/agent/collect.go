package agent

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
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

		agent.Storage.Set(models.Metrics{
			ID:    metirc,
			MType: models.Gauge,
			Value: &metircToReport,
		})
	}
	return nil
}

func (agent *Config) RandomValue() {
	randFloat := rand.Float64()
	agent.Storage.Set(models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: &randFloat,
	})
}

func (agent *Config) Counter(value int64) {
	agent.Storage.Set(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &value,
	})
}
